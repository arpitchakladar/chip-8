//go:build wasm && js

package main

import (
	"context"
	"syscall/js"

	"github.com/arpitchakladar/chip-8/internal/emulator"
)

const defaultClockSpeed = uint32(100000)

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

func destroyHandler(
	vm *emulator.Emulator,
) func(this js.Value, args []js.Value) any {
	return func(this js.Value, args []js.Value) any {
		vm.Destroy()
		return nil
	}
}
