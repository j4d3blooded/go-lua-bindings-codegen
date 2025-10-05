package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"strings"
)

func CreateLuaBindings(w io.Writer, bindInfo *LuaLibraryBinding) error {

	templates, err := template.New("").Parse(TEMPLATE_STR)
	if err != nil {
		return fmt.Errorf("error parsing templates: %w", err)
	}

	fmt.Fprintf(w, "package %v\n", bindInfo.PackageName)
	io.WriteString(w, "import (\n")
	io.WriteString(w, `"github.com/Shopify/go-lua"`)
	io.WriteString(w, "\n")
	io.WriteString(w, `"fmt"`)
	io.WriteString(w, "\n")

	for _, pkg := range bindInfo.Imports {
		if strings.TrimSpace(pkg) != "" {
			fmt.Fprintf(w, "%v\n", pkg)
		}
	}
	io.WriteString(w, ")\n")

	templates.ExecuteTemplate(w, "AddInstaller", bindInfo)

	for _, funcToBind := range bindInfo.BindFuncs {
		fmt.Fprintf(w, "func _LuaBinding%v(l *lua.State) int {\n", funcToBind.Name)

		if err := templates.ExecuteTemplate(w, "ArgCountCheck", len(funcToBind.Arguments)); err != nil {
			return fmt.Errorf("error inserting arg count check: %w", err)
		}

		for i, arg := range funcToBind.Arguments {
			data := map[string]any{
				"index":      i,
				"targetType": arg.Type,
			}

			targetTemplate := templates.Lookup(fmt.Sprintf("%vArgTypeCheck", arg.Type))

			if targetTemplate == nil {
				targetTemplate = templates.Lookup("DefaultArgTypeCheck")
			}

			err := targetTemplate.Execute(w, data)
			if err != nil {
				return fmt.Errorf("error executing template for arg %v: %w", i, err)
			}

		}

		if len(funcToBind.Results) > 0 {

			for i := range funcToBind.Results {
				fmt.Fprintf(w, "r%v", i)
				if i+1 != len(funcToBind.Results) {
					io.WriteString(w, ", ")
				}
			}

			io.WriteString(w, " := ")
		}

		fmt.Fprintf(w, "%v(\n", funcToBind.Name)

		for i := range funcToBind.Arguments {
			fmt.Fprintf(w, "arg%v,\n", i)
		}

		io.WriteString(w, ")\n")

		for i, arg := range funcToBind.Results {
			targetTemplate := templates.Lookup(fmt.Sprintf("%vRetTypePush", arg.Type))

			if targetTemplate == nil {
				targetTemplate = templates.Lookup("DefaultRetTypePush")
			}

			err := targetTemplate.Execute(w, i)
			if err != nil {
				return fmt.Errorf("error executing template for ret %v: %w", i, err)
			}
		}

		fmt.Fprintf(w, "return %v\n", len(funcToBind.Results))
		fmt.Fprintf(w, "}\n\n")
	}

	for _, callerFunc := range bindInfo.CallerFuncs {
		signature := bytes.NewBufferString("func(")
		fmt.Fprintf(w, "func _Get%vCaller(ls _LuaService, script string) (", callerFunc.Name)

		for i, arg := range callerFunc.Arguments {
			fmt.Fprintf(signature, "r%v %v", i, arg.Type)
			if i+1 != len(callerFunc.Arguments) {
				io.WriteString(signature, ", ")
			}
		}

		io.WriteString(signature, ") (")
		for i, res := range callerFunc.Results {
			io.WriteString(signature, res.Type)
			if i+1 != len(callerFunc.Results) {
				io.WriteString(signature, ", ")
			}
		}

		io.WriteString(signature, ")")

		sigStr := signature.String()
		io.WriteString(w, sigStr)
		io.WriteString(w, ", error){\n")

		templates.ExecuteTemplate(w, "ParseCheck", nil)

		io.WriteString(w, "return ")
		io.WriteString(w, sigStr)
		io.WriteString(w, "{\n")

		templates.ExecuteTemplate(w, "PrepState", nil)

		for i, arg := range callerFunc.Arguments {
			targetTemplate := templates.Lookup(fmt.Sprintf("%vRetTypePush", arg.Type))

			if targetTemplate == nil {
				targetTemplate = templates.Lookup("DefaultRetTypePush")
			}

			err := targetTemplate.Execute(w, i)
			if err != nil {
				return fmt.Errorf("error executing template for argument %v: %w", i, err)
			}

		}

		io.WriteString(w, "l.ProtectedCall(0, lua.MultipleReturns, 0)\n")

		for i, res := range callerFunc.Results {
			data := map[string]any{
				"index":      i,
				"targetType": res.Type,
			}

			targetTemplate := templates.Lookup(fmt.Sprintf("%vArgTypeCheck", res.Type))

			if targetTemplate == nil {
				targetTemplate = templates.Lookup("DefaultArgTypeCheck")
			}

			fmt.Fprintf(w, "l.Global(\"%v\")\n", res.Name)
			err := targetTemplate.Execute(w, data)
			if err != nil {
				return fmt.Errorf("error executing template for arg %v: %w", i, err)
			}
		}

		if len(callerFunc.Results) > 0 {
			io.WriteString(w, "return ")
			for i := range callerFunc.Results {
				fmt.Fprintf(w, "arg%v", i)
				if i+1 != len(callerFunc.Results) {
					io.WriteString(w, ", ")
				}
			}
		}

		fmt.Fprintf(w, "}, nil")
		fmt.Fprintf(w, "}")
	}

	return nil
}
