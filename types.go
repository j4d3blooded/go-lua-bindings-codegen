package main

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
