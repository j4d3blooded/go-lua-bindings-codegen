package main

import (
	"flag"
	"fmt"
	"path/filepath"
)

func _GetLuaBindingFuncs(pkgDir string) (LuaLibraryBinding, error) {

}

func main() {
	var (
		TARGET_DIR  string
		LIB_NAME    string
		OUTPUT_FILE string
		STUB_FILE   string
	)

	flag.StringVar(&TARGET_DIR, "dir", ".", "directory to build binding from")
	flag.StringVar(&LIB_NAME, "name", "boundlib", "name for lua library")
	flag.StringVar(&OUTPUT_FILE, "out", "lua_GEN.go", "output file name")
	flag.StringVar(&STUB_FILE, "stub", "stub.lua", "output stub file name")

	flag.Parse()

	TARGET_DIR, _ = filepath.Abs(TARGET_DIR)
	bindInfo, err := _GetLuaBindingFuncs(TARGET_DIR)
	if err != nil {
		panic(fmt.Errorf("error generating bindings: %w", err))
	}
}
