package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"
	"strings"
)

func findImportPath(file *ast.File, targetPackageName string) (path string) {
	for _, imp := range file.Imports {
		// imp.Path.Value is quoted ("github.com/someone/iter")
		if imp.Name != nil && imp.Name.Name == targetPackageName {
			return imp.Path.Value
		}
		// If no alias, infer from the last segment of the import path
		unquoted := imp.Path.Value[1 : len(imp.Path.Value)-1]
		last := unquoted[strings.LastIndex(unquoted, "/")+1:]
		if last == targetPackageName {
			return imp.Path.Value
		}
	}
	return ""
}

func typeInfo(expr ast.Expr) (pkgName, typeName string) {
	switch t := expr.(type) {
	case *ast.Ident:
		// Simple type like "int" or "string"
		return "", t.Name

	case *ast.SelectorExpr:
		// pkg.Type
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name, t.Sel.Name
		}

	case *ast.StarExpr:
		// pointer type (*T) — unwrap and recurse
		a, b := typeInfo(t.X)
		return a, "*" + b

	case *ast.IndexExpr:
		// generic single type param — recurse into t.X
		return typeInfo(t.X)

	case *ast.IndexListExpr:
		// generic multiple type params — recurse into t.X
		return typeInfo(t.X)

	case *ast.ArrayType:
		// array or slice — recurse into element type
		return typeInfo(t.Elt)

	case *ast.MapType:
		// map type — recurse into value type
		return typeInfo(t.Value)
	}

	// Unhandled or complex type (e.g. func(...) ...)
	return "", ""
}

// exprString returns Go source code for an ast.Expr.
func exprString(fset *token.FileSet, expr ast.Expr) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, expr); err != nil {
		return fmt.Sprintf("<error: %v>", err)
	}
	// Optionally format it nicely
	src, err := format.Source(buf.Bytes())
	if err != nil {
		return buf.String() // fallback if formatting fails
	}
	return string(src)
}

func checkIfFunctionShouldBind(decl ast.Decl) (*ast.FuncDecl, bool) {
	funcDecl, ok := decl.(*ast.FuncDecl)
	if !ok {
		return nil, false
	}

	// Check if there's a Doc comment group
	if funcDecl.Doc == nil {
		return nil, false
	}

	isLuaBinding := false

	// Search for the specific comment
	for _, comment := range funcDecl.Doc.List {
		if comment.Text == "//lua:bind" {
			isLuaBinding = true
		}
	}

	return funcDecl, isLuaBinding
}

func checkIfFunctionIsCaller(decl ast.Decl) (*ast.FuncDecl, bool) {
	funcDecl, ok := decl.(*ast.FuncDecl)
	if !ok {
		return nil, false
	}

	// Check if there's a Doc comment group
	if funcDecl.Doc == nil {
		return nil, false
	}

	isLuaBinding := false

	// Search for the specific comment
	for _, comment := range funcDecl.Doc.List {
		if comment.Text == "//lua:caller" {
			isLuaBinding = true
		}
	}

	return funcDecl, isLuaBinding
}
