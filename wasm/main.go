//go:build wasm && js

package main

import (
	"context"
	"fmt"
	"sync/atomic"
	"syscall/js"

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
	js.Global().Set("chip8New", js.FuncOf(chip8New))
	js.Global().Set("chip8LoadROM", js.FuncOf(chip8LoadROM))
	js.Global().Set("chip8Run", js.FuncOf(chip8Run))
	js.Global().Set("chip8Destroy", js.FuncOf(chip8Destroy))
	js.Global().Set("chip8SetKey", js.FuncOf(chip8SetKey))
	js.Global().Set("chip8PlayAudio", js.FuncOf(chip8PlayAudio))
}

func chip8New(this js.Value, args []js.Value) any {
	clockSpeed := defaultClockSpeed
	if len(args) > 0 {
		clockSpeed = uint32(args[0].Int())
	}

	atomic.AddUint32(&VMCounter, 1)
	vm := emulator.WithWASM(clockSpeed)
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

func chip8SetKey(this js.Value, args []js.Value) any {
	if len(args) < 3 {
		return map[string]string{
			"error": "VM ID, key, and pressed state are required",
		}
	}

	vm := VMs[args[0].String()]
	if vm == nil {
		return map[string]string{"error": "emulator not initialized"}
	}

	key := byte(args[1].Int())
	pressed := args[2].Bool()
	vm.Keyboard.SetKey(key, pressed)

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
