package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"

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

				for _, arg := range funcDecl.Type.Params.List {
					lib.Imports = append(lib.Imports, addBindingParamInfo(file, arg, fset, &bindInfo, true)...)
				}

				for _, res := range funcDecl.Type.Results.List {
					addBindingParamInfo(file, res, fset, &bindInfo, false)
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
	Name string
	Type string
}

type LuaFunctionBindingInfo struct {
	Name      string
	Arguments []BindFuncParamInfo
	Results   []BindFuncParamInfo
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
)

func main() {

	flag.StringVar(&TARGET_DIR, "dir", ".", "directory to build binding from")
	flag.StringVar(&LIB_NAME, "name", "boundlib", "name for lua library")
	flag.StringVar(&OUTPUT_FILE, "out", "lua_GEN.go", "output file name")

	flag.Parse()

	TARGET_DIR, _ = filepath.Abs(TARGET_DIR)
	bindInfo, err := GetLuaBindingFuncs(TARGET_DIR)
	if err != nil {
		panic(fmt.Errorf("error generating bindings: %w", err))
	}

	genFile := filepath.Join(TARGET_DIR, OUTPUT_FILE)

	b := &bytes.Buffer{}

	bindInfo.LibName = LIB_NAME
	if err := CreateLuaBindings(b, bindInfo); err != nil {
		panic(fmt.Errorf("error creating bindings: %w", err))
	}

	// formatted, err := format.Source(b.Bytes())
	formatted := b.Bytes()

	if err != nil {
		panic(fmt.Errorf("error formatting generated code: %w", err))
	}

	if err := os.WriteFile(genFile, formatted, os.ModePerm); err != nil {
		panic(fmt.Errorf("error writing code: %w", err))
	}

}
