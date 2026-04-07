//go:build wasm && js

package main

import (
	"syscall/js"

	"github.com/arpitchakladar/chip-8/internal/assembler"
)

func NewAssembler(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		throw("assembly code string is required")
	}

	assemblyCode := args[0].String()
	asm := assembler.New(assemblyCode)

	this.Set("assemble", js.FuncOf(assembleHandler(asm)))

	return nil
}

func assembleHandler(
	asm *assembler.Assembler,
) func(this js.Value, args []js.Value) any {
	return func(this js.Value, args []js.Value) any {
		compiled, err := asm.Assemble()
		if err != nil {
			throw(err.Error())
		}
		uint8Array := js.Global().Get("Uint8Array").New(len(compiled))
		js.CopyBytesToJS(uint8Array, compiled)
		return uint8Array
	}
}
