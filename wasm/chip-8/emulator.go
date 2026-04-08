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
		"loadROM":            asyncWrapper(loadROMHandler(vm)),
		"run":                asyncWrapper(runHandler(vm, canvas)),
		"destroy":            asyncWrapper(destroyHandler(vm, kh, canvas)),
		"handleKeyboard":     asyncWrapper(handleKeyboardHandler(kh)),
		"releaseKeyboard":    asyncWrapper(releaseKeyboardHandler(kh)),
		"sendKey":            asyncWrapper(sendKeyHandler(kh)),
		"isHandlingKeyboard": asyncWrapper(isHandlingKeyboardHandler(kh)),
	}

	// Return the object to JavaScript
	return js.ValueOf(methods), nil
}

// handleKeyboard sets up keyboard handlers for the emulator.
func handleKeyboardHandler(
	kh *KeyboardHandler,
) func(args []js.Value) (any, error) {
	return func(args []js.Value) (any, error) {
		if err := kh.Setup(); err != nil {
			return nil, err
		}
		return nil, nil
	}
}

// releaseKeyboard removes keyboard handlers from the emulator.
func releaseKeyboardHandler(
	kh *KeyboardHandler,
) func(args []js.Value) (any, error) {
	return func(args []js.Value) (any, error) {
		if err := kh.Remove(); err != nil {
			return nil, err
		}
		return nil, nil
	}
}

// sendKeyHandler sends a key press/release to the emulator.
func sendKeyHandler(
	kh *KeyboardHandler,
) func(args []js.Value) (any, error) {
	return func(args []js.Value) (any, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("sendKey requires key and pressed arguments")
		}
		key := uint8(args[0].Int())
		pressed := args[1].Bool()
		kh.SendKey(key, pressed)
		return nil, nil
	}
}

// isHandlingKeyboardHandler returns whether keyboard handling is active.
func isHandlingKeyboardHandler(
	kh *KeyboardHandler,
) func(args []js.Value) (any, error) {
	return func(args []js.Value) (any, error) {
		return js.ValueOf(kh.IsActive()), nil
	}
}

// loadROMHandler creates a function that loads ROM data into the emulator.
// Parameter: romData (Uint8Array) - The ROM bytecode to load.
func loadROMHandler(
	vm *emulator.Emulator,
) func(args []js.Value) (any, error) {
	return func(args []js.Value) (any, error) {
		jsData := args[0]
		romData := make([]byte, jsData.Length())
		for i := 0; i < jsData.Length(); i++ {
			romData[i] = uint8(jsData.Index(i).Int())
		}
		if err := vm.LoadROM(romData); err != nil {
			return nil, err
		}
		return nil, nil
	}
}

// runHandler creates a function that starts the emulator execution.
func runHandler(
	vm *emulator.Emulator,
	_ js.Value,
) func(args []js.Value) (any, error) {
	return func(args []js.Value) (any, error) {
		if err := vm.Run(context.Background()); err != nil {
			return nil, err
		}

		return nil, nil
	}
}

// destroyHandler creates a function that stops and destroys the emulator.
func destroyHandler(
	vm *emulator.Emulator,
	kh *KeyboardHandler,
	canvas js.Value,
) func(args []js.Value) (any, error) {
	return func(args []js.Value) (any, error) {
		err := kh.Remove()
		vm.Destroy()
		canvas.Delete(canvasEmulatorKey)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
}
