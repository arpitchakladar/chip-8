//go:build wasm && js

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
	defaultClockSpeed = uint32(100000)
	VMCounter         = uint32(0)
	VMs               = make(map[string]*emulator.Emulator)
)

func main() {
	registerCallbacks()
	<-make(chan struct{})
}

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

func chip8Compile(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return map[string]string{
			"error": "the assembly code string is required",
		}
	}

	assemblyCode := args[0].String()
	asm := assembler.New(assemblyCode)
	compiled, err := asm.Assemble()
	if err != nil {
		return map[string]string{"error": err.Error()}
	}
	uint8Array := js.Global().Get("Uint8Array").New(len(compiled))
	js.CopyBytesToJS(uint8Array, compiled)

	return uint8Array
}

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

func chip8Destroy(this js.Value, args []js.Value) any {
	vm := VMs[args[0].String()]
	if vm == nil {
		return map[string]string{"error": "no VM was found."}
	}

	vm.Destroy()

	return nil
}

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
