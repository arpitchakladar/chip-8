//go:build wasm && js

// Package main provides WebAssembly bindings for the CHIP-8 emulator.
// This package exposes the emulator functionality to JavaScript, allowing
// web applications to run CHIP-8 ROMs directly in the browser.
//
// Exposed JavaScript object:
//   - chip_8.Emulator: Constructor for creating emulator instances
//   - chip_8.Assembler: Constructor for assembling CH8 assembly code
package main

import "syscall/js"

// throw throws a JavaScript error with the given message.
func throw(message string) {
	panic(js.Global().Get("Error").New(message))
}

// main is the entry point for the WebAssembly module.
// It registers the JavaScript callbacks and blocks forever.
func main() {
	registerCallbacks()
	<-make(chan struct{})
}

// registerCallbacks exposes emulator and assembler constructors to JavaScript.
// Creates a chip_8 object with Emulator and Assembler constructors.
func registerCallbacks() {
	chip8 := js.Global().Get("Object").New()
	chip8.Set("Emulator", js.FuncOf(NewEmulator))
	chip8.Set("Assembler", js.FuncOf(NewAssembler))
	js.Global().Set("chip_8", chip8)
}
