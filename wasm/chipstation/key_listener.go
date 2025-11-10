//go:build js && wasm

package main

import (
	"syscall/js"
)

var (
	keyMap = map[string]uint8{
		"1": 0x1, "2": 0x2, "3": 0x3, "4": 0xC,
		"q": 0x4, "w": 0x5, "e": 0x6, "r": 0xD,
		"a": 0x7, "s": 0x8, "d": 0x9, "f": 0xE,
		"z": 0xA, "x": 0x0, "c": 0xB, "v": 0xF,
	}

	// Keep references to prevent GC.
	keyDownFunc js.Func
	keyUpFunc   js.Func
	unloadFunc  js.Func
)

func attachKeyListeners() {
	doc := js.Global().Get("document")

	keyDownFunc = js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) == 0 {
			return nil
		}
		event := args[0]
		key := event.Get("key").String()
		if chipKey, ok := keyMap[key]; ok {
			e.SetKeyState(chipKey, 1)
		}
		return nil
	})

	keyUpFunc = js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) == 0 {
			return nil
		}
		event := args[0]
		key := event.Get("key").String()
		if key == "u" {
			e.ToggleUi()
			return nil
		}
		if chipKey, ok := keyMap[key]; ok {
			e.SetKeyState(chipKey, 0)
		}
		return nil
	})

	doc.Call("addEventListener", "keydown", keyDownFunc)
	doc.Call("addEventListener", "keyup", keyUpFunc)

	// Optional: cleanup on page unload.
	unloadFunc = js.FuncOf(func(this js.Value, args []js.Value) any {
		doc.Call("removeEventListener", "keydown", keyDownFunc)
		doc.Call("removeEventListener", "keyup", keyUpFunc)
		keyDownFunc.Release()
		keyUpFunc.Release()
		unloadFunc.Release()
		return nil
	})
	js.Global().Get("window").Call("addEventListener", "beforeunload", unloadFunc)
}
