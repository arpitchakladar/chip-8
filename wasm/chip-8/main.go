//go:build wasm && js

// Package main provides WebAssembly bindings for the CHIP-8 emulator.
// This package exposes the emulator functionality to JavaScript, allowing
// web applications to run CHIP-8 ROMs directly in the browser.
//
// Exposed JavaScript object:
//   - CHIP8: Acts as a constructor function for creating emulator instances
//   - new CHIP8(canvasElement, clockSpeed?): Creates a new emulator instance
//   - .compile(assemblyCode): Compiles CHIP-8 assembly to bytecode
//   - .loadROM(romData): Loads a ROM into the VM (instance method)
//   - .run(): Starts the emulator execution (instance method)
//   - .destroy(): Stops and destroys the VM instance (instance method)
package main

import (
	"context"
	"strings"
	"syscall/js"

	"github.com/arpitchakladar/chip-8/internal/assembler"
	"github.com/arpitchakladar/chip-8/internal/emulator"
)

var (
	defaultClockSpeed = uint32(
		100000,
	) // Default 100kHz CPU speed
)

func throw(message string) {
	panic(js.Global().Get("Error").New(message))
}

func main() {
	// Register JavaScript callbacks and start the WASM event loop
	registerCallbacks()
	// Block forever - WASM needs the main thread to remain active
	<-make(chan struct{})
}

// registerCallbacks exposes emulator functions to JavaScript as global functions.
// Each function is wrapped in a js.Func to be callable from JS.
func registerCallbacks() {
	chip8 := js.Global().Get("Object").New()

	chip8.Set("Emulator", js.FuncOf(newEmulator))
	chip8.Set("Assembler", js.FuncOf(newAssembler))

	js.Global().Set("chip_8", chip8)
}

// newAssembler acts as a JavaScript constructor for the Assembler.
// It is intended to be used with `new Assembler(sourceCode)`.
//
// Parameters:
//   - sourceCode: CHIP-8 assembly source code (required)
//
// JavaScript usage:
//
//	const asm = new Assembler(source);
//	const rom = asm.assemble();
func newAssembler(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		throw("assembly code string is required")
	}

	assemblyCode := args[0].String()
	asm := assembler.New(assemblyCode)

	// attach method to THIS instance
	this.Set("assemble", js.FuncOf(func(this js.Value, args []js.Value) any {
		compiled, err := asm.Assemble()
		if err != nil {
			throw(err.Error())
		}

		uint8Array := js.Global().Get("Uint8Array").New(len(compiled))
		js.CopyBytesToJS(uint8Array, compiled)

		return uint8Array
	}))

	return nil
}

// newEmulator acts as a JavaScript constructor for the CHIP-8 emulator.
// It is intended to be used with `new chip_8.Emulator(...)`.
//
// Parameters:
//   - canvasElement: A JavaScript canvas element for rendering (required)
//   - clockSpeed: (optional) CPU speed in instructions per second (defaults to 100000)
//
// JavaScript usage:
//
//	const chip = new chip_8.Emulator(canvas, 500);
func newEmulator(this js.Value, args []js.Value) any {
	clockSpeed := defaultClockSpeed

	if len(args) < 1 {
		throw("a canvas element is required")
	}

	if len(args) > 1 {
		clockSpeed = uint32(args[1].Int())
	}

	canvas := args[0]

	vm := emulator.WithWASM(canvas, clockSpeed)

	// Attach methods to THIS (the JS instance)

	this.Set("loadROM", js.FuncOf(func(this js.Value, args []js.Value) any {
		jsData := args[0]
		romData := make([]byte, jsData.Length())

		for i := 0; i < jsData.Length(); i++ {
			romData[i] = uint8(jsData.Index(i).Int())
		}

		if err := vm.LoadROM(romData); err != nil {
			throw(err.Error())
		}
		return nil
	}))

	this.Set("run", js.FuncOf(func(this js.Value, args []js.Value) any {
		go func() {
			if err := vm.Run(context.Background()); err != nil {
				// You may want to surface this to JS later
				println("VM error:", err.Error())
			}
		}()
		return nil
	}))

	this.Set("destroy", js.FuncOf(func(this js.Value, args []js.Value) any {
		vm.Destroy()
		return nil
	}))

	setupKeyboardListeners(vm, canvas)

	// IMPORTANT: return nil for constructor usage
	return nil
}

type KeyboardHandlers struct {
	KeyDown    js.Func
	KeyUp      js.Func
	Click      js.Func
	MouseEnter js.Func
	MouseLeave js.Func
	Blur       js.Func
	WindowBlur js.Func
}

// setupKeyboardListeners attaches keyboard and focus event handlers to the canvas.
// It returns a KeyboardHandlers struct for cleanup if needed.
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

// keyToChip8 maps JavaScript keyboard events to CHIP-8 key codes.
// Returns a pointer to the CHIP-8 key byte (0-15), or nil if the key is not mapped.
// The mapping follows the standard CHIP-8 keyboard layout:
//
//	1 2 3 C -> 1 2 3 4
//	4 5 6 D -> Q W E R
//	7 8 9 E -> A S D F
//	A 0 B F -> Z X C V
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
