//go:build wasm && js

// Package emulator provides the core CHIP-8 emulator functionality.
package emulator

import (
	"sync"
	"syscall/js"

	"github.com/arpitchakladar/chip-8/internal/emulator/audio"
	"github.com/arpitchakladar/chip-8/internal/emulator/cpu"
	"github.com/arpitchakladar/chip-8/internal/emulator/display"
	"github.com/arpitchakladar/chip-8/internal/emulator/keyboard"
	"github.com/arpitchakladar/chip-8/internal/emulator/memory"
)

// WithWASM creates a new Emulator configured for WebAssembly/JavaScript execution.
// It uses HTML5 Canvas for display and Web Audio API for sound via the WASM implementations.
//
// Parameters:
//   - canvas: A JavaScript canvas element reference for rendering graphics
//   - clockSpeed: CPU instructions per second (e.g., 100000 for 100kHz)
//
// Returns a configured Emulator ready to load and run CHIP-8 ROMs.
func WithWASM(canvas js.Value, clockSpeed uint32) *Emulator {
	e := &Emulator{
		CPU:        cpu.New(),
		Memory:     memory.New(),
		Display:    display.WithWASM(canvas),
		Keyboard:   keyboard.WithWASM(),
		Audio:      audio.WithWASM(),
		memoryLock: sync.Mutex{},
		ClockSpeed: clockSpeed,
	}

	e.Memory.LoadFontSet()
	e.CPU.ProgramCounter = ProgramStart
	return e
}
