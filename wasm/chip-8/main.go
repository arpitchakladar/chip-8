//go:build wasm && js

// Package main provides WebAssembly bindings for the CHIP-8 emulator.
// This package exposes the emulator functionality to JavaScript, allowing
// web applications to run CHIP-8 ROMs directly in the browser.
//
// Exposed JavaScript object:
//   - CHIP8: Object with methods for creating and managing emulator instances
//   - new CHIP8(canvasElement): Creates a new emulator instance
//   - .compile(assemblyCode): Compiles CHIP-8 assembly to bytecode
//   - .loadROM(romData): Loads a ROM into the VM
//   - .run(): Starts the emulator execution
//   - .destroy(): Stops and destroys the VM instance
package main

import (
	"context"
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
	chip8 := js.FuncOf(chip8New)
	chip8.Set("compile", js.FuncOf(chip8Compile))
	js.Global().Set("CHIP8", chip8)
}

// chip8Compile compiles CHIP-8 assembly code to bytecode.
// Returns a Uint8Array containing the compiled ROM.
// Parameter: assemblyCode (string) - The CHIP-8 assembly source code.
func chip8Compile(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		throw("the assembly code string is required")
	}

	assemblyCode := args[0].String()
	asm := assembler.New(assemblyCode)

	compiled, err := asm.Assemble()
	if err != nil {
		throw(err.Error())
	}

	uint8Array := js.Global().Get("Uint8Array").New(len(compiled))
	js.CopyBytesToJS(uint8Array, compiled)

	return uint8Array
}

// chip8New acts as a JavaScript constructor for the CHIP-8 emulator.
// It is intended to be used with `new CHIP8(...)`.
//
// Parameters:
//   - canvasElement: A JavaScript canvas element for rendering (required)
//   - clockSpeed: (optional) CPU speed in instructions per second (defaults to 100000)
//
// JavaScript usage:
//
//	const chip = new CHIP8(canvas, 500);
func chip8New(this js.Value, args []js.Value) any {
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

	setupKeyboardListeners(vm)

	// IMPORTANT: return nil for constructor usage
	return nil
}

func setupKeyboardListeners(vm *emulator.Emulator) any {
	document := js.Global().Get("document")

	document.Call(
		"addEventListener",
		"keydown",
		js.FuncOf(func(this js.Value, args []js.Value) any {
			if !vm.IsRunning() {
				return nil
			}
			event := args[0]
			key := event.Get("key").String()
			chip8Key := keyToChip8(key)
			if chip8Key != nil {
				vm.Keyboard.SetKey(*chip8Key, true)
			}
			return nil
		}),
	)

	document.Call(
		"addEventListener",
		"keyup",
		js.FuncOf(func(this js.Value, args []js.Value) any {
			if !vm.IsRunning() {
				return nil
			}
			event := args[0]
			key := event.Get("key").String()
			chip8Key := keyToChip8(key)
			if chip8Key != nil {
				vm.Keyboard.SetKey(*chip8Key, false)
			}
			return nil
		}),
	)

	return nil
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
