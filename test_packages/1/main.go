package main

import (
	"os"

	lua "github.com/Shopify/go-lua"
)

// Checks if a number is even
//
//lua:bind
//lua:param:a - Value to check
//lua:retrn:0 - true if even, false if odd
func IsEven(a int) bool {
	return a%2 == 0
}

// Modifies values by some rule
//
//lua:caller
//lua:param:values - values to change
//lua:retrn:0 - changed values
func Modify(values []string) []string {
	return nil
}

type _LuaServiceImpl struct {
}

func (ls *_LuaServiceImpl) AddLibraryFunction(libName string, funcName string, f lua.Function) {
}

func (ls *_LuaServiceImpl) GetState() *lua.State {
	l := lua.NewStateEx()
	lua.OpenLibraries(l)
	return l
}

func main() {
	ls := &_LuaServiceImpl{}
	InstallLuaExtension(ls)

	sFile, err := os.Open("test.lua")
	if err != nil {
		panic(err)
	}

	caller, err := _GetLuaCaller_ShouldKill(
		ls, sFile,
	)

	if err != nil {
		panic(err)
	}

	v1, v2, err := caller("charlie", 11)
	if err != nil {
		panic(err)
	}

	println(v1)
	println(v2)

}
