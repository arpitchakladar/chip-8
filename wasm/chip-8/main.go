//go:build wasm && js

// Package main provides WebAssembly bindings for the CHIP-8 emulator.
// This package exposes the emulator functionality to JavaScript, allowing
// web applications to run CHIP-8 ROMs directly in the browser.
//
// Exposed JavaScript functions:
//   - chip8Compile(assemblyCode): Compiles CHIP-8 assembly to bytecode
//   - chip8New(canvasElement, clockSpeed?): Creates a new emulator instance
//   - chip8LoadROM(vmId, romData): Loads a ROM into the specified VM
//   - chip8Run(vmId): Starts the emulator execution
//   - chip8Destroy(vmId): Stops and destroys a VM instance
//   - chip8PlayAudio(vmId): Manually triggers audio playback
//   - chip8SetKeyboardHandler(vmId): Attaches keyboard event handlers
package main

import (
	"context"
	"fmt"
	"sync/atomic"
	"syscall/js"

	"github.com/arpitchakladar/chip-8/internal/assembler"
	"github.com/arpitchakladar/chip-8/internal/emulator"
)

var (
	defaultClockSpeed = uint32(
		100000,
	) // Default 100kHz CPU speed
	VMCounter uint32 // Atomic counter for generating unique VM IDs
	VMs       = make(
		map[string]*emulator.Emulator,
	) // Active emulator instances
)

func main() {
	// Register JavaScript callbacks and start the WASM event loop
	registerCallbacks()
	// Block forever - WASM needs the main thread to remain active
	<-make(chan struct{})
}

// registerCallbacks exposes emulator functions to JavaScript as global functions.
// Each function is wrapped in a js.Func to be callable from JS.
func registerCallbacks() {
	js.Global().Set("chip8Compile", js.FuncOf(chip8Compile))
	js.Global().Set("chip8New", js.FuncOf(chip8New))
	js.Global().Set("chip8LoadROM", js.FuncOf(chip8LoadROM))
	js.Global().Set("chip8Run", js.FuncOf(chip8Run))
	js.Global().Set("chip8Destroy", js.FuncOf(chip8Destroy))
	js.Global().Set("chip8PlayAudio", js.FuncOf(chip8PlayAudio))
	js.Global().
		Set("chip8SetKeyboardHandler", js.FuncOf(chip8SetKeyboardHandler))
}

// chip8Compile compiles CHIP-8 assembly code to bytecode.
// Returns a Uint8Array containing the compiled ROM, or an error object on failure.
// Parameter: assemblyCode (string) - The CHIP-8 assembly source code.
func chip8Compile(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		errObj := js.Global().Get("Object").New()
		errObj.Set("error", "the assembly code string is required")
		return errObj
	}

	assemblyCode := args[0].String()
	asm := assembler.New(assemblyCode)

	compiled, err := asm.Assemble()
	if err != nil {
		errObj := js.Global().Get("Object").New()
		errObj.Set("error", err.Error())
		return errObj
	}

	uint8Array := js.Global().Get("Uint8Array").New(len(compiled))
	js.CopyBytesToJS(uint8Array, compiled)

	return uint8Array
}

// chip8New creates a new CHIP-8 emulator instance.
// Parameters:
//   - canvasElement: A JavaScript canvas element for rendering
//   - clockSpeed: (optional) CPU speed in instructions per second, defaults to 100000
//
// Returns: A unique VM ID string used for subsequent operations, or error object.
func chip8New(this js.Value, args []js.Value) any {
	clockSpeed := defaultClockSpeed
	if len(args) < 1 {
		return map[string]string{"error": "a canvas element is required"}
	}
	if len(args) > 1 {
		clockSpeed = uint32(args[1].Int())
	}

	canvas := args[0]

	atomic.AddUint32(&VMCounter, 1)
	vm := emulator.WithWASM(canvas, clockSpeed)
	id := fmt.Sprintf("chip-8-vm-%d", atomic.LoadUint32(&VMCounter))
	VMs[id] = vm

	return id
}

// chip8LoadROM loads ROM data into an existing emulator instance.
// Parameters:
//   - vmId: The VM ID returned by chip8New
//   - romData: A Uint8Array containing the ROM bytecode
//
// Returns: null on success, or error object on failure.
func chip8LoadROM(this js.Value, args []js.Value) any {
	if len(args) < 2 {
		return map[string]string{"error": "VM ID and ROM data are required"}
	}

	vm := VMs[args[0].String()]
	jsData := args[1]
	romData := make([]byte, jsData.Length())
	for i := 0; i < jsData.Length(); i++ {
		romData[i] = uint8(jsData.Index(i).Int())
	}

	if err := vm.LoadROM(romData); err != nil {
		return map[string]string{"error": err.Error()}
	}
	return nil
}

// chip8Run starts the emulator execution loop.
// Parameter: vmId - The VM ID returned by chip8New
// Returns: null on success (runs asynchronously), or error object on failure.
func chip8Run(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return map[string]string{"error": "VM ID is required"}
	}

	vm := VMs[args[0].String()]
	if vm == nil {
		return map[string]string{"error": "emulator not initialized"}
	}

	errChan := make(chan error, 1)

	go func() {
		if err := vm.Run(context.Background()); err != nil {
			errChan <- err
		}
	}()

	return nil
}

// chip8Destroy stops and destroys an emulator instance, releasing all resources.
// Parameter: vmId - The VM ID of the instance to destroy
// Returns: null on success, or error object if VM not found.
func chip8Destroy(this js.Value, args []js.Value) any {
	vm := VMs[args[0].String()]
	if vm == nil {
		return map[string]string{"error": "no VM was found."}
	}

	vm.Destroy()

	return nil
}

// chip8PlayAudio manually triggers audio playback for the emulator.
// Useful for testing or when user interaction is needed to start audio (browser policy).
// Parameter: vmId - The VM ID
// Returns: null on success, or error object on failure.
func chip8PlayAudio(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return map[string]string{"error": "VM ID is required"}
	}

	vm := VMs[args[0].String()]
	if vm == nil {
		return map[string]string{"error": "emulator not initialized"}
	}

	if err := vm.Audio.Play(); err != nil {
		return map[string]string{
			"error": fmt.Sprintf("audio device error: %s", err),
		}
	}

	return nil
}

// chip8SetKeyboardHandler attaches keyboard event listeners to the document.
// Maps JavaScript keyboard events to CHIP-8 key presses for the specified VM.
// Parameter: vmId - The VM ID
// Returns: null on success, or error object on failure.
func chip8SetKeyboardHandler(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return map[string]string{"error": "VM ID is required"}
	}

	vm := VMs[args[0].String()]
	if vm == nil {
		return map[string]string{"error": "emulator not initialized"}
	}

	document := js.Global().Get("document")
	document.Call(
		"addEventListener",
		"keydown",
		js.FuncOf(func(this js.Value, args []js.Value) any {
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
