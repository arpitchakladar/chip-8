//go:build wasm && js

package main

import (
	"context"
	"syscall/js"

	"github.com/arpitchakladar/chip-8/internal/emulator"
)

// defaultClockSpeed is the default CPU clock speed in Hz.
const defaultClockSpeed = uint32(100000)

// NewEmulator is a JavaScript constructor for the CHIP-8 emulator.
// It creates a new emulator instance attached to a canvas element.
//
// Parameters:
//   - canvas: A JavaScript canvas element for rendering (required)
//   - clockSpeed: CPU speed in Hz (optional, defaults to 100000)
//
// Returns: The emulator instance with loadROM, run, and destroy methods.
func NewEmulator(this js.Value, args []js.Value) any {
	clockSpeed := defaultClockSpeed

	if len(args) < 1 {
		throw("a canvas element is required")
	}

	if len(args) > 1 {
		clockSpeed = uint32(args[1].Int())
	}

	canvas := args[0]
	vm := emulator.WithWASM(canvas, clockSpeed)

	this.Set("loadROM", js.FuncOf(loadROMHandler(vm)))
	this.Set("run", js.FuncOf(runHandler(vm)))
	this.Set("destroy", js.FuncOf(destroyHandler(vm)))

	setupKeyboardListeners(vm, canvas)

	return nil
}

// loadROMHandler creates a function that loads ROM data into the emulator.
// Parameter: romData (Uint8Array) - The ROM bytecode to load.
func loadROMHandler(
	vm *emulator.Emulator,
) func(this js.Value, args []js.Value) any {
	return func(this js.Value, args []js.Value) any {
		jsData := args[0]
		romData := make([]byte, jsData.Length())
		for i := 0; i < jsData.Length(); i++ {
			romData[i] = uint8(jsData.Index(i).Int())
		}
		if err := vm.LoadROM(romData); err != nil {
			throw(err.Error())
		}
		return nil
	}
}

// runHandler creates a function that starts the emulator execution.
func runHandler(
	vm *emulator.Emulator,
) func(this js.Value, args []js.Value) any {
	return func(this js.Value, args []js.Value) any {
		go func() {
			if err := vm.Run(context.Background()); err != nil {
				println("VM error:", err.Error())
			}
		}()
		return nil
	}
}

// destroyHandler creates a function that stops and destroys the emulator.
func destroyHandler(
	vm *emulator.Emulator,
) func(this js.Value, args []js.Value) any {
	return func(this js.Value, args []js.Value) any {
		vm.Destroy()
		return nil
	}
}
