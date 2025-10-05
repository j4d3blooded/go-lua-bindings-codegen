package test_1

import (
	"iter"

	"github.com/Shopify/go-lua"
)

//lua:caller
func SetX(v1 int, v2 *lua.State, meow iter.Seq2[int, int]) (a string, b bool) {
	return "", true
}

//lua:bind
func UnsetX(value func(any) int) bool {
	return false
}
