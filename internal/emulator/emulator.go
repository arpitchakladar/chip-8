package emulator

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/arpitchakladar/chip-8/internal/emulator/audio"
	"github.com/arpitchakladar/chip-8/internal/emulator/cpu"
	"github.com/arpitchakladar/chip-8/internal/emulator/display"
	"github.com/arpitchakladar/chip-8/internal/emulator/keyboard"
	"github.com/arpitchakladar/chip-8/internal/emulator/memory"
)

// ProgramStart is the memory address where CHIP-8 programs begin (0x200).
const ProgramStart = 0x200

// Emulator represents a complete CHIP-8 virtual machine.
// It coordinates the CPU, memory, display, keyboard, and audio subsystems.
type Emulator struct {
	CPU        *cpu.CPU
	Memory     *memory.Memory
	Display    display.Display
	Keyboard   keyboard.Keyboard
	Audio      audio.Audio
	ClockSpeed uint32
	MemoryLock sync.Mutex
}

// LoadROM loads a CHIP-8 ROM into memory starting at ProgramStart (0x200).
func (e *Emulator) LoadROM(romData []byte) error {
	for i, b := range romData {
		if err := e.Memory.Write(ProgramStart+uint16(i), b); err != nil {
			return err
		}
	}
	return nil
}

// Run starts the emulator main loop.
// It initializes the display and audio subsystems, then runs the CPU and display loops.
// The function blocks until the emulator is closed or an error occurs.
func (e *Emulator) Run(parentContext context.Context) error {
	if err := e.Display.Init(); err != nil {
		return fmt.Errorf("failed to init display: %w", err)
	}

	if err := e.Audio.Init(); err != nil {
		fmt.Printf("Warning: Audio failed to init: %v\n", err)
	}

	defer func() {
		if err := e.Display.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing display: %v\n", err)
		}
		e.Audio.Close()
	}()

	runEmulatorContext, cancelRunEmulatorContext := context.WithCancel(
		parentContext,
	)
	errChan := make(chan error, 1)
	defer cancelRunEmulatorContext()

	go e.runCPU(runEmulatorContext, errChan)
	e.runDisplay(runEmulatorContext, errChan)

	return <-errChan
}

// runDisplay handles the display update loop at 60Hz.
// It polls for keyboard events, updates timers, and renders the display.
func (e *Emulator) runDisplay(
	runEmulatorContext context.Context,
	errChan chan<- error,
) {
	uiClock := time.NewTicker(time.Second / 60)
	defer uiClock.Stop()

	for {
		select {
		case <-runEmulatorContext.Done():
			return
		case <-uiClock.C:
			e.Keyboard.PollEvents()

			e.MemoryLock.Lock()
			timerErr := e.updateTimers()
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

// runCPU runs the CPU execution loop at the configured ClockSpeed.
func (e *Emulator) runCPU(
	runEmulatorContext context.Context,
	errChan chan<- error,
) {
	cpuClock := time.NewTicker(time.Second / time.Duration(e.ClockSpeed))
	defer cpuClock.Stop()

	for {
		select {
		case <-runEmulatorContext.Done():
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
	hi, err := e.Memory.Read(e.CPU.ProgramCounter)
	if err != nil {
		return err
	}
	lo, err := e.Memory.Read(e.CPU.ProgramCounter + 1)
	if err != nil {
		return err
	}
	opcode := uint16(hi)<<8 | uint16(lo)

	e.CPU.ProgramCounter += 2

	return e.CPU.Execute(opcode, e.Memory, e.Display, e.Keyboard)
}

// updateTimers decrements the delay and sound timers at 60Hz.
func (e *Emulator) updateTimers() error {
	if e.CPU.SoundTimer > 0 {
		if err := e.Audio.GenerateBeep(); err != nil {
			return err
		}
		e.Audio.Play()
		e.CPU.SoundTimer--
	} else {
		e.Audio.Pause()
	}

	if e.CPU.DelayTimer > 0 {
		e.CPU.DelayTimer--
	}

	return nil
}

// Destroy releases emulator resources.
func (e *Emulator) Destroy() {
	if err := e.Display.Close(); err != nil {
		// TODO: Handle this error
		// Just pass for now
		fmt.Printf("Error: Failed to close display")
	}
	e.Audio.Close()
}
