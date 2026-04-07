//go:build wasm && js

package main

import "syscall/js"

func throw(message string) {
	panic(js.Global().Get("Error").New(message))
}

func main() {
	registerCallbacks()
	<-make(chan struct{})
}

func registerCallbacks() {
	chip8 := js.Global().Get("Object").New()
	chip8.Set("Emulator", js.FuncOf(NewEmulator))
	chip8.Set("Assembler", js.FuncOf(NewAssembler))
	js.Global().Set("chip_8", chip8)
}
