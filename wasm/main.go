//go:build wasm && js

package main

import (
	"syscall/js"
	// "github.com/arpitchakladar/chip-8/internal/emulator".
)

func main() {
	registerCallbacks()
	<-make(chan struct{})
}

func registerCallbacks() {
	js.Global().Set("chip8Init", js.FuncOf(chip8Init))
	js.Global().Set("chip8LoadROM", js.FuncOf(chip8LoadROM))
	js.Global().Set("chip8Run", js.FuncOf(chip8Run))
	js.Global().Set("chip8Stop", js.FuncOf(chip8Stop))
	js.Global().Set("chip8Destroy", js.FuncOf(chip8Destroy))
}

func chip8Init(this js.Value, args []js.Value) any {
	// clockSpeed := uint32(100000)
	// if len(args) > 0 {
	// 	clockSpeed = uint32(args[0].Int())
	// }

	// vm = emulator.WithWASM(clockSpeed)
	return nil
}

func chip8LoadROM(this js.Value, args []js.Value) any {
	// if len(args) < 1 {
	// 	return map[string]string{"error": "ROM data required"}
	// }
	//
	// jsData := args[0]
	// romData := make([]byte, jsData.Length())
	// for i := 0; i < jsData.Length(); i++ {
	// 	romData[i] = uint8(jsData.Index(i).Int())
	// }
	//
	// if err := vm.LoadROM(romData); err != nil {
	// 	return map[string]string{"error": err.Error()}
	// }
	return nil
}

func chip8Run(this js.Value, args []js.Value) any {
	// if vm == nil {
	// 	return map[string]string{"error": "emulator not initialized"}
	// }
	//
	// errChan = make(chan error, 1)
	//
	// go func() {
	// 	if err := vm.Run(); err != nil {
	// 		errChan <- err
	// 	}
	// }()

	return nil
}

func chip8Stop(this js.Value, args []js.Value) any {
	// if cancel != nil {
	// 	cancel()
	// }
	return nil
}

func chip8Destroy(this js.Value, args []js.Value) any {
	// if vm != nil {
	// 	vm.Destroy()
	// }
	return nil
}
