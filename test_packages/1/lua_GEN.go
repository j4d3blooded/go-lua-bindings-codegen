package test_1
import (
"github.com/Shopify/go-lua"
"fmt"
"iter"
)


const _LIB_NAME = "boundlib"

type _LuaService interface {
	AddLibraryFunction(libName, funcName string, f lua.Function)
	GetState() *lua.State
}

func InstallLuaExtension(ls _LuaService) {
	
		ls.AddLibraryFunction(
			_LIB_NAME,
			"UnsetX",
			_LuaBindingUnsetX,
		)
	
}

func _LuaBindingUnsetX(l *lua.State) int {


if l.Top() != 1 {
	lua.Errorf(l, "incorret number of arguments.")
	return 0
}



arg0Temp := l.ToUserData(0)
arg0, isTyped := arg0Temp.(func(any) int)
if !isTyped {
	lua.Errorf(l, "argument 0 is incorrect type")
	return 0
}

UnsetX(
arg0,
)
return 0
}

func _GetSetXCaller(ls _LuaService, script string) (func(r0 int, r1 *lua.State, r2 iter.Seq2[int, int]) (string, bool), error){


l := ls.GetState()
err := lua.LoadString(l, script)
if err != nil {
	return nil, fmt.Errorf("error parsing lua script: %w", err)
}

return func(r0 int, r1 *lua.State, r2 iter.Seq2[int, int]) (string, bool){


l := ls.GetState()
lua.LoadString(l, script)



l.PushLightUserData(r0)



l.PushLightUserData(r1)



l.PushLightUserData(r2)

l.ProtectedCall(0, lua.MultipleReturns, 0)
l.Global(a)


arg0Temp := l.ToUserData(0)
arg0, isTyped := arg0Temp.(string)
if !isTyped {
	lua.Errorf(l, "argument 0 is incorrect type")
	return 0
}

l.Global(b)


arg1Temp := l.ToUserData(1)
arg1, isTyped := arg1Temp.(bool)
if !isTyped {
	lua.Errorf(l, "argument 1 is incorrect type")
	return 0
}

return arg0, arg1}, nil}