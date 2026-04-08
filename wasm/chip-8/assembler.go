//go:build wasm && js

package main

import (
	"fmt"
	"syscall/js"

	"github.com/arpitchakladar/chip-8/internal/assembler"
)

// NewAssembler is a standard JavaScript function that returns an assembler object.
//
// Parameters:
//   - sourceCode: CH8 assembly source code (required)
//
// Returns: A JS object { assemble: function }.
// Usage: const asm = await chip_8.Assembler(source).
func NewAssembler(args []js.Value) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("assembly code string is required")
	}

	assemblyCode := args[0].String()
	asm := assembler.New(assemblyCode)

	// Create the methods for our object
	methods := map[string]any{
		"assemble": js.FuncOf(assembleHandler(asm)),
	}

	// Return the map as a JS object
	return js.ValueOf(methods), nil
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
