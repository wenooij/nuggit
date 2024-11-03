//go:build wasm

package main

import (
	"syscall/js"
)

func main() {
	js.Global().Get("console").Call("log", js.ValueOf("Nuggit was injected into this page and may be collecting data (https://github.com/wenooij/nuggit-chrome-extension)."))
	js.Global().Set("createNuggitAction", js.ValueOf(js.FuncOf(func(_ js.Value, args []js.Value) any {
		config := args[0]
		a, err := CreateAction(config)
		if err != nil {
			js.Global().Get("console").Call("error", js.ValueOf(err.Error()))
			return nil
		}
		return a
	})))

	// Listen for signals.
	// TODO: Handle signals.
	select {}
}
