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
// Usage: const asm = await chip_8.Assembler(source);
// Returns: A JS object { assemble: function }.
func NewAssembler(args []js.Value) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("assembly code string is required")
	}

	assemblyCode := args[0].String()
	asm := assembler.New(assemblyCode)

	// Create the methods for our object
	methods := map[string]any{
		"assemble": AsyncWrapper(assembleHandler(asm)),
	}

	// Return the map as a JS object
	return js.ValueOf(methods), nil
}

// assembleHandler creates a function that assembles the source code to bytecode.
//
// Usage: const rom = await asm.assemble();
// Returns: Uint8Array with the ROM data from the assembly file.
func assembleHandler(
	asm *assembler.Assembler,
) func(args []js.Value) (any, error) {
	return func(args []js.Value) (any, error) {
		compiled, err := asm.Assemble()
		if err != nil {
			return nil, err
		}
		uint8Array := js.Global().Get("Uint8Array").New(len(compiled))
		js.CopyBytesToJS(uint8Array, compiled)
		return uint8Array, nil
	}
}
