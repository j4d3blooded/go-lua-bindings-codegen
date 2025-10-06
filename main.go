package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"log/slog"
	"strconv"
	"strings"
	"text/template"

	"go/parser"

	"go/token"
	"os"
	"path/filepath"

	_ "embed"
)

//go:embed "templates.txt"
var TEMPLATE_STR string

func addBindingParamInfo(
	file *ast.File, arg *ast.Field, fset *token.FileSet,
	bindInfo *LuaFunctionBindingInfo, isParam bool,
) (importPath []string) {
	pkgName, _ := typeInfo(arg.Type)
	fullTypeName := exprString(fset, arg.Type)

	if pkgName != "" {

		ip := findImportPath(file, pkgName)
		importPath = append(importPath, ip)
	}

	if arg.Names == nil {
		param := BindFuncParamInfo{
			Name: "",
			Type: fullTypeName,
		}

		if isParam {
			bindInfo.Arguments = append(bindInfo.Arguments, param)
		} else {
			bindInfo.Results = append(bindInfo.Results, param)
		}
	}

	for _, argName := range arg.Names {

		param := BindFuncParamInfo{
			Name: argName.Name,
			Type: fullTypeName,
		}

		if isParam {
			bindInfo.Arguments = append(bindInfo.Arguments, param)
		} else {
			bindInfo.Results = append(bindInfo.Results, param)
		}
	}

	return
}

// scanDir scans a directory for Go files and prints out functions with the //lua:bind comment
func GetLuaBindingFuncs(dir string) (*LuaLibraryBinding, error) {
	fset := token.NewFileSet()

	// Parse all Go files in the directory into a package
	pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse dir: %w", err)
	}

	lib := &LuaLibraryBinding{}

	for pkgToBindName, pkg := range pkgs {
		lib.PackageName = pkgToBindName
		for _, file := range pkg.Files {
			// Walk through declarations in the file
			for _, decl := range file.Decls {
				funcDecl, shouldBind := checkIfFunctionShouldBind(decl)
				_, isCaller := checkIfFunctionIsCaller(decl)

				if !(shouldBind || isCaller) {
					continue
				}

				bindInfo := LuaFunctionBindingInfo{
					Name: funcDecl.Name.Name,
				}

				if funcDecl.Type.Params != nil {
					for _, arg := range funcDecl.Type.Params.List {
						lib.Imports = append(lib.Imports, addBindingParamInfo(file, arg, fset, &bindInfo, true)...)
					}
				}

				if funcDecl.Type.Results != nil {
					for _, res := range funcDecl.Type.Results.List {
						addBindingParamInfo(file, res, fset, &bindInfo, false)
					}
				}

				sb := strings.Builder{}
				for _, comment := range funcDecl.Doc.List {
					cleaned := strings.TrimLeft(comment.Text, "/")
					cleaned = strings.TrimSpace(cleaned)
					if !strings.HasPrefix(cleaned, "lua:") {
						sb.WriteString(cleaned)
					}
				}

				bindInfo.Description = sb.String()

				for _, comment := range funcDecl.Doc.List {
					cleaned := strings.TrimLeft(comment.Text, "/")
					cleaned = strings.TrimSpace(cleaned)

					isParam := false
					canContinue := false

					if after, ok := strings.CutPrefix(cleaned, "lua:param:"); ok {
						cleaned = after
						isParam = true
						canContinue = true
					}

					if after, ok := strings.CutPrefix(cleaned, "lua:retrn:"); ok {
						cleaned = after
						canContinue = true
					}

					if !canContinue {
						continue
					}

					segments := strings.Split(cleaned, "-")
					paramName := strings.TrimSpace(segments[0])
					paramDesc := strings.TrimSpace(segments[1])

					var toSearch *[]BindFuncParamInfo

					if isParam {
						toSearch = &bindInfo.Arguments
					} else {
						toSearch = &bindInfo.Results
					}

					targetIndex, _ := strconv.Atoi(paramName)
					found := false

				l:
					for i, cParam := range *toSearch {
						if (isParam && cParam.Name == paramName) || i == targetIndex {
							found = true
							(*toSearch)[i].Description = paramDesc
							break l
						}
					}

					if !found {
						slog.Warn(
							"Bound parameter/argument has no description set",
							"Name", paramName, "Function", bindInfo.Name, "IsParam", isParam,
						)
					}
				}

				if shouldBind {
					lib.BindFuncs = append(lib.BindFuncs, bindInfo)
				} else if isCaller {
					lib.CallerFuncs = append(lib.CallerFuncs, bindInfo)
				}
			}
		}
	}

	return lib, nil
}

type BindFuncParamInfo struct {
	Name        string
	Type        string
	Description string
}

type LuaFunctionBindingInfo struct {
	Name        string
	Arguments   []BindFuncParamInfo
	Results     []BindFuncParamInfo
	Description string
}

type LuaLibraryBinding struct {
	PackageName string
	Imports     []string
	LibName     string
	Name        string
	BindFuncs   []LuaFunctionBindingInfo
	CallerFuncs []LuaFunctionBindingInfo
}

var (
	TARGET_DIR  string
	LIB_NAME    string
	OUTPUT_FILE string
	STUB_FILE   string
)

func main() {

	flag.StringVar(&TARGET_DIR, "dir", ".", "directory to build binding from")
	flag.StringVar(&LIB_NAME, "name", "boundlib", "name for lua library")
	flag.StringVar(&OUTPUT_FILE, "out", "lua_GEN.go", "output file name")
	flag.StringVar(&STUB_FILE, "stub", "stub.lua", "output stub file name")

	flag.Parse()

	TARGET_DIR, _ = filepath.Abs(TARGET_DIR)
	bindInfo, err := GetLuaBindingFuncs(TARGET_DIR)
	if err != nil {
		panic(fmt.Errorf("error generating bindings: %w", err))
	}

	genFile := filepath.Join(TARGET_DIR, OUTPUT_FILE)
	stubFile := filepath.Join(TARGET_DIR, STUB_FILE)

	b := &bytes.Buffer{}

	bindInfo.LibName = LIB_NAME

	t := template.New("")

	t = t.Funcs(template.FuncMap{
		"execOr": func(data any, fallback, target string) (string, error) {
			toExec := target
			if temp := t.Lookup(target); temp == nil {
				toExec = fallback
			}

			tempWrite := bytes.Buffer{}
			if err := t.ExecuteTemplate(&tempWrite, toExec, data); err != nil {
				return "", fmt.Errorf("error executing subtemplate: %w", err)
			}
			return tempWrite.String(), nil
		},
		"map": func(pairs ...any) (map[string]any, error) {
			if len(pairs)%2 != 0 {
				return nil, errors.New("misaligned map")
			}

			m := make(map[string]any, len(pairs)/2)

			for i := 0; i < len(pairs); i += 2 {
				key, ok := pairs[i].(string)

				if !ok {
					return nil, fmt.Errorf("cannot use type %T as map key", pairs[i])
				}
				m[key] = pairs[i+1]
			}
			return m, nil
		},
		"tableIndex": func(len, curIndex int) int {
			return (len - 1) - curIndex
		},
		"stackIndex": func(curIndex int) int {
			return -(curIndex + 1)
		},
		"joinParamNamesLua": func(params []BindFuncParamInfo) string {

			strs := []string{}
			for _, param := range params {
				strs = append(strs, param.Name)
			}

			return strings.Join(strs, ", ")
		},
	})

	t, err = t.Parse(TEMPLATE_STR)
	if err != nil {
		panic(fmt.Errorf("error parsing template: %w", err))
	}

	if err := t.ExecuteTemplate(b, "main", bindInfo); err != nil {
		panic(fmt.Errorf("error executing template: %w", err))
	}

	formatted, err := format.Source(b.Bytes())
	// formatted := b.Bytes()

	if err != nil {
		panic(fmt.Errorf("error formatting generated code: %w", err))
	}

	if err := os.WriteFile(genFile, formatted, os.ModePerm); err != nil {
		panic(fmt.Errorf("error writing code: %w", err))
	}

	f, err := os.Create(stubFile)
	if err != nil {
		panic(fmt.Errorf("error creating stub file: %w", err))
	}

	defer f.Close()

	if err := t.ExecuteTemplate(f, "luaStub", bindInfo); err != nil {
		panic(fmt.Errorf("error creating lua stub: %w", err))
	}
}
