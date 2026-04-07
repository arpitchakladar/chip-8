//go:build wasm && js

package main

import (
	"syscall/js"

	"github.com/arpitchakladar/chip-8/internal/assembler"
)

// NewAssembler is a JavaScript constructor for the CHIP-8 assembler.
// It creates an assembler instance from assembly source code.
//
// Parameters:
//   - sourceCode: CH8 assembly source code (required)
//
// Returns: The assembler instance with an assemble method.
func NewAssembler(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		throw("assembly code string is required")
	}

	assemblyCode := args[0].String()
	asm := assembler.New(assemblyCode)

	this.Set("assemble", js.FuncOf(assembleHandler(asm)))

	return nil
}

// assembleHandler creates a function that assembles the source code to bytecode.
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
