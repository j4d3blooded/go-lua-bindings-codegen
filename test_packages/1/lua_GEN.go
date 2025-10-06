package main

import (
	"fmt"
	"github.com/Shopify/go-lua"
	"io"
	"strings"
)

type _LuaService interface {
	AddLibraryFunction(libName, funcName string, f lua.Function)
	GetState() *lua.State
}

const _LIB_NAME = "boundlib"

func InstallLuaExtension(ls _LuaService) {

	ls.AddLibraryFunction(
		_LIB_NAME,
		"IsEven",
		_LuaBindingIsEven,
	)

}

func _LuaBindingIsEven(l *lua.State) int {
	if l.Top() != 1 {
		lua.Errorf(l, "incorret number of arguments.")
		return 0
	}

	arg0, isTyped := l.ToInteger(0)

	if !isTyped {
		lua.Errorf(l, "argument 0 is incorrect type, should be int")
		return 0
	}

	res0 := IsEven(
		arg0,
	)

	l.PushLightUserData(res0)

	return 1
}

func _GetLuaCaller_ShouldKill(ls _LuaService, src io.Reader) (func(name string, value int) (string, int, error), error) {
	pBytes, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("error reading lua program: %w", err)
	}
	program := string(pBytes)

	err = ls.GetState().Load(strings.NewReader(program), program, "")
	if err != nil {
		return nil, fmt.Errorf("error parsing lua program: %w", err)
	}

	return func(arg0 string, arg1 int) (string, int, error) {

		l := ls.GetState()
		l.Load(strings.NewReader(program), program, "")

		if err := l.ProtectedCall(0, 0, 0); err != nil {
			err = fmt.Errorf("error intaking lua program: %w", err)
			return *new(string), *new(int), err
		}

		l.Global("_ShouldKill")

		l.PushString(arg0)

		l.PushInteger(arg1)

		if err := l.ProtectedCall(2, 2, 0); err != nil {
			err = fmt.Errorf("error doing lua call: %w", err)
			return *new(string), *new(int), err
		}

		res1, isTyped := l.ToInteger(-1)

		if !isTyped {
			err := fmt.Errorf("lua call returned incorrect type, wanted int")
			return *new(string), *new(int), err
		}

		res0, isTyped := l.ToString(-2)

		if !isTyped {
			err := fmt.Errorf("lua call returned incorrect type, wanted string")
			return *new(string), *new(int), err
		}

		return res0, res1, nil

	}, nil

}
