//go:build wasm && js

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

// WithSDL creates a new Emulator with SDL2-based display, keyboard, and audio.
// The clockSpeed parameter specifies CPU instructions per second (e.g., 100000 for 100kHz).
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
