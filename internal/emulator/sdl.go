//go:build !wasm || !js

// Package emulator provides the core CHIP-8 emulator functionality for native platforms.
package emulator

import (
	"sync"

	"github.com/arpitchakladar/chip-8/internal/emulator/audio"
	"github.com/arpitchakladar/chip-8/internal/emulator/cpu"
	"github.com/arpitchakladar/chip-8/internal/emulator/display"
	"github.com/arpitchakladar/chip-8/internal/emulator/keyboard"
	"github.com/arpitchakladar/chip-8/internal/emulator/memory"
)

// WithSDL creates a new Emulator with SDL2-based display, keyboard, and audio.
// The clockSpeed parameter specifies CPU instructions per second (e.g., 100000 for 100kHz).
func WithSDL(clockSpeed uint32) *Emulator {
	e := &Emulator{
		CPU:        cpu.New(),
		Memory:     memory.New(),
		Display:    display.WithSDL(),
		Keyboard:   keyboard.WithSDL(),
		Audio:      audio.WithSDL(),
		memoryLock: sync.Mutex{},
		ClockSpeed: clockSpeed,
	}

	e.Memory.LoadFontSet()
	e.CPU.ProgramCounter = ProgramStart
	return e
}
