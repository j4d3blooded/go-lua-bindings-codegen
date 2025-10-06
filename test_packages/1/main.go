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

// This function lets you check if you should kill someone and when to do it if you should
//
//lua:caller
//lua:param:name - Name of person to check if we should kill
//lua:param:value - Chance we should kill them
//lua:retrn:0 - What to do
//lua:retrn:1 - When to kill them
func _ShouldKill(name string, value int) (string, int) {
	return "", 0
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
