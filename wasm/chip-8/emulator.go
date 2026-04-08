//go:build wasm && js

package main

import (
	"context"
	"fmt"
	"syscall/js"

	"github.com/arpitchakladar/chip-8/internal/emulator"
)

const defaultClockSpeed = uint32(100000)

// This key in the canvas element identifies a canvas element
// as a canvas element that is being used by an emulator.
const canvasEmulatorKey = "chip8Emulator"

// NewEmulator is a standard JavaScript function that returns an emulator instance.
//
// Parameters:
//   - canvas: A JavaScript canvas element for rendering (required)
//   - clockSpeed: CPU speed in Hz (optional, defaults to 100000)
//
// Usage: const vm = await chip_8.Emulator(canvas, clockSpeed);
// Returns: A JS object { loadROM, run, destroy, ... }.
func NewEmulator(args []js.Value) (any, error) {
	clockSpeed := defaultClockSpeed

	if len(args) < 1 {
		return nil, fmt.Errorf("a canvas element is required")
	}

	canvas := args[0]

	// Check if the canvas already has an emulator attached
	if canvas.Get(canvasEmulatorKey).Truthy() {
		return nil, fmt.Errorf("an emulator is already attached to this canvas")
	}

	if len(args) > 1 {
		clockSpeed = uint32(args[1].Int())
	}

	// Initialize the Go emulator logic
	vm := emulator.WithWASM(canvas, clockSpeed)
	kh := NewKeyboardHandler(canvas, vm)

	// Mark the canvas as occupied
	canvas.Set(canvasEmulatorKey, js.ValueOf(true))

	// Create the methods map
	methods := map[string]any{
		"loadROM":            js.FuncOf(loadROMHandler(vm)),
		"run":                js.FuncOf(runHandler(vm, canvas)),
		"destroy":            js.FuncOf(destroyHandler(vm, kh, canvas)),
		"handleKeyboard":     js.FuncOf(handleKeyboardHandler(kh)),
		"releaseKeyboard":    js.FuncOf(releaseKeyboardHandler(kh)),
		"sendKey":            js.FuncOf(sendKeyHandler(kh)),
		"isHandlingKeyboard": js.FuncOf(isHandlingKeyboardHandler(kh)),
	}

	// Return the object to JavaScript
	return js.ValueOf(methods), nil
}

// handleKeyboard sets up keyboard handlers for the emulator.
func handleKeyboardHandler(
	kh *KeyboardHandler,
) func(this js.Value, args []js.Value) any {
	return func(this js.Value, args []js.Value) any {
		if err := kh.Setup(); err != nil {
			throw(err.Error())
		}
		return nil
	}
}

// releaseKeyboard removes keyboard handlers from the emulator.
func releaseKeyboardHandler(
	kh *KeyboardHandler,
) func(this js.Value, args []js.Value) any {
	return func(this js.Value, args []js.Value) any {
		if err := kh.Remove(); err != nil {
			throw(err.Error())
		}
		return nil
	}
}

// sendKeyHandler sends a key press/release to the emulator.
func sendKeyHandler(
	kh *KeyboardHandler,
) func(this js.Value, args []js.Value) any {
	return func(this js.Value, args []js.Value) any {
		if len(args) < 2 {
			throw("sendKey requires key and pressed arguments")
		}
		key := uint8(args[0].Int())
		pressed := args[1].Bool()
		kh.SendKey(key, pressed)
		return nil
	}
}

// isHandlingKeyboardHandler returns whether keyboard handling is active.
func isHandlingKeyboardHandler(
	kh *KeyboardHandler,
) func(this js.Value, args []js.Value) any {
	return func(this js.Value, args []js.Value) any {
		return js.ValueOf(kh.IsActive())
	}
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
	_ js.Value,
) func(this js.Value, args []js.Value) any {
	return func(this js.Value, args []js.Value) any {
		go func() {
			if err := vm.Run(context.Background()); err != nil {
				throw(fmt.Sprintf("VM error: %s", err))
			}
		}()
		return nil
	}
}

// destroyHandler creates a function that stops and destroys the emulator.
func destroyHandler(
	vm *emulator.Emulator,
	kh *KeyboardHandler,
	canvas js.Value,
) func(this js.Value, args []js.Value) any {
	return func(this js.Value, args []js.Value) any {
		_ = kh.Remove()
		vm.Destroy()
		canvas.Delete(canvasEmulatorKey)
		return nil
	}
}
