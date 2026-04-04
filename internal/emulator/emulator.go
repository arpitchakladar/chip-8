package emulator

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/veandco/go-sdl2/sdl"

	"github.com/arpitchakladar/chip-8/internal/emulator/audio"
	"github.com/arpitchakladar/chip-8/internal/emulator/cpu"
	"github.com/arpitchakladar/chip-8/internal/emulator/display"
	"github.com/arpitchakladar/chip-8/internal/emulator/keyboard"
	"github.com/arpitchakladar/chip-8/internal/emulator/memory"
)

// ProgramStart is the memory address where CHIP-8 programs begin.
const ProgramStart = 0x200

// Emulator represents a complete CHIP-8 virtual machine.
type Emulator struct {
	CPU        *cpu.CentralProcessingUnit
	Memory     *memory.Memory
	Display    *display.Display
	Keyboard   *keyboard.Keyboard
	Audio      *audio.Audio
	ClockSpeed uint32 // Instructions per second (in Hz)
	MemoryLock sync.Mutex
}

// New creates a new Emulator with the specified clock speed (in Hz).
// It initializes all subsystems (CPU, Memory, Display, Keyboard, Audio)
// and loads the font set into memory.
func New(clockSpeed uint32) *Emulator {
	e := &Emulator{
		CPU:        cpu.New(),
		Memory:     memory.New(),
		Display:    display.New(),
		Keyboard:   keyboard.New(),
		Audio:      audio.New(),
		MemoryLock: sync.Mutex{},
		ClockSpeed: clockSpeed,
	}

	e.Memory.LoadFontSet()              // Load fonts into 0x000-0x050
	e.CPU.ProgramCounter = ProgramStart // Set PC to program start address
	return e
}

// LoadROM reads a .ch8 file and writes it into memory starting at ProgramStart.
func (e *Emulator) LoadROM(romData []byte) error {
	for i, b := range romData {
		if err := e.Memory.Write(ProgramStart+uint16(i), b); err != nil {
			return err
		}
	}
	return nil
}

// Run starts the emulator main loop.
// It initializes the display and audio subsystems, then runs the CPU and Display loops.
// The function blocks until the emulator is closed or an error occurs.
func (e *Emulator) Run() error {
	// 1. Setup
	if err := e.Display.Init(); err != nil {
		return fmt.Errorf("failed to init display: %w", err)
	}

	if err := e.Audio.Init(); err != nil {
		// Log error but don't crash - some systems don't have audio hardware.
		fmt.Printf("Warning: Audio failed to init: %v\n", err)
	}

	// Close audio and display when finished
	defer func() {
		if err := e.Display.Close(); err != nil {
			// Error in closing display doesn't need to be handled, just mentioned
			fmt.Fprintf(os.Stderr, "Error closing display: %v\n", err)
		}

		e.Audio.Close()
	}()

	// Create cancellable context for goroutine cancellation
	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error, 1)
	defer cancel()

	// Start the CPU loop as goroutine
	go e.runCPU(ctx, errChan)
	// Start the display loop on the main thread (SDL2 needs it)
	e.runDisplay(ctx, errChan)

	// Block until an error occurs
	return <-errChan
}

// runDisplay is the display and timer loop.
// It runs on the main thread (SDL2 requires it), handling events, updating timers, and rendering at 60Hz.
// It communicates errors back through the errChan channel.
func (e *Emulator) runDisplay(ctx context.Context, errChan chan<- error) {
	uiClock := time.NewTicker(time.Second / 60)
	defer uiClock.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-uiClock.C:
			// A. Handle Events (Must be main thread)
			for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
				switch t := event.(type) {
				case *sdl.QuitEvent:
					errChan <- nil
					return
				case *sdl.KeyboardEvent:
					e.Keyboard.HandleKeyboard(t)
				}
			}

			e.MemoryLock.Lock()
			// B. Update Timers (60Hz)
			timerErr := e.updateTimers()
			// C. Display buffer (60Hz)
			displayErr := e.Display.Present()
			e.MemoryLock.Unlock()

			if timerErr != nil {
				errChan <- timerErr
				return
			}
			if displayErr != nil {
				errChan <- displayErr
				return
			}
		}
	}
}

// runCPU is the CPU execution loop.
// It runs as a goroutine, fetching and executing instructions at the configured ClockSpeed.
// It communicates errors back through the errChan channel.
func (e *Emulator) runCPU(ctx context.Context, errChan chan<- error) {
	cpuClock := time.NewTicker(time.Second / time.Duration(e.ClockSpeed))
	defer cpuClock.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cpuClock.C:
			e.MemoryLock.Lock()
			err := e.tick()
			e.MemoryLock.Unlock()
			if err != nil {
				errChan <- err
				return
			}
		}
	}
}

// tick performs one CPU fetch-decode-execute cycle.
func (e *Emulator) tick() error {
	// 1. Fetch
	hi, err := e.Memory.Read(e.CPU.ProgramCounter)
	if err != nil {
		return err
	}
	lo, err := e.Memory.Read(e.CPU.ProgramCounter + 1)
	if err != nil {
		return err
	}
	opcode := uint16(hi)<<8 | uint16(lo)

	// 2. Increment PC before execution
	e.CPU.ProgramCounter += 2

	// 3. Execute
	return e.CPU.Execute(opcode, e.Memory, e.Display, e.Keyboard)
}

// updateTimers decrements the delay and sound timers at 60Hz.
// If the sound timer is greater than zero, it triggers audio playback.
func (e *Emulator) updateTimers() error {
	if e.CPU.SoundTimer > 0 {
		// Unpause the audio device to start the buzz
		if err := e.Audio.GenerateBeep(); err != nil {
			return err
		}
		e.Audio.Play()
		e.CPU.SoundTimer--
	} else {
		// Pause the audio device when timer hits 0
		e.Audio.Pause()
	}

	if e.CPU.DelayTimer > 0 {
		e.CPU.DelayTimer--
	}

	return nil
}
