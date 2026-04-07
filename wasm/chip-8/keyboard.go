//go:build wasm && js

package main

import (
	"strings"
	"syscall/js"

	"github.com/arpitchakladar/chip-8/internal/emulator"
)

type KeyboardHandlers struct {
	KeyDown    js.Func
	KeyUp      js.Func
	Click      js.Func
	MouseEnter js.Func
	MouseLeave js.Func
	Blur       js.Func
	WindowBlur js.Func
}

func setupKeyboardListeners(
	vm *emulator.Emulator,
	canvas js.Value,
) *KeyboardHandlers {
	canvas.Set("tabIndex", 0)

	handlers := &KeyboardHandlers{}

	handlers.Click = createClickHandler(canvas)
	handlers.MouseEnter = createFocusHandler(canvas)
	handlers.MouseLeave = createBlurHandler(canvas)
	handlers.Blur = createClearKeysHandler(vm)
	handlers.WindowBlur = createClearKeysHandler(vm)
	handlers.KeyDown = createKeyHandler(vm, true)
	handlers.KeyUp = createKeyHandler(vm, false)

	canvas.Call("addEventListener", "click", handlers.Click)
	canvas.Call("addEventListener", "mouseenter", handlers.MouseEnter)
	canvas.Call("addEventListener", "mouseleave", handlers.MouseLeave)
	canvas.Call("addEventListener", "blur", handlers.Blur)
	canvas.Call("addEventListener", "keydown", handlers.KeyDown)
	canvas.Call("addEventListener", "keyup", handlers.KeyUp)
	js.Global().Call("addEventListener", "blur", handlers.WindowBlur)

	return handlers
}

func createClickHandler(canvas js.Value) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		canvas.Call("focus")
		return nil
	})
}

func createFocusHandler(canvas js.Value) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		canvas.Call("focus")
		return nil
	})
}

func createBlurHandler(canvas js.Value) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		canvas.Call("blur")
		return nil
	})
}

func createClearKeysHandler(vm *emulator.Emulator) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		for i := range 16 {
			vm.Keyboard.SetKey(byte(i), false)
		}
		return nil
	})
}

func createKeyHandler(vm *emulator.Emulator, pressed bool) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		if !vm.IsRunning() {
			return nil
		}
		event := args[0]
		event.Call("preventDefault")
		key := strings.ToLower(event.Get("key").String())
		if chip8Key := keyToChip8(key); chip8Key != nil {
			vm.Keyboard.SetKey(*chip8Key, pressed)
		}
		return nil
	})
}

func keyToChip8(key string) *byte {
	keyMap := map[string]byte{
		"1": 0x1, "2": 0x2, "3": 0x3, "4": 0xC,
		"q": 0x4, "w": 0x5, "e": 0x6, "r": 0xD,
		"a": 0x7, "s": 0x8, "d": 0x9, "f": 0xE,
		"z": 0xA, "x": 0x0, "c": 0xB, "v": 0xF,
	}
	if ch, ok := keyMap[key]; ok {
		return &ch
	}
	return nil
}
