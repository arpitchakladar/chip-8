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
	CPU          *cpu.CPU
	Memory       *memory.Memory
	Display      display.Display
	Keyboard     keyboard.Keyboard
	Audio        audio.Audio
	ClockSpeed   uint32
	memoryLock   sync.Mutex
	running      bool
	runLock      sync.Mutex
	cancelRunner context.CancelFunc
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
	e.runLock.Lock()
	if e.running {
		e.runLock.Unlock()
		return fmt.Errorf("emulator is already running")
	}
	e.running = true
	e.runLock.Unlock()

	if err := e.Display.Init(); err != nil {
		e.runLock.Lock()
		e.running = false
		e.runLock.Unlock()
		return fmt.Errorf("failed to init display: %w", err)
	}

	if err := e.Audio.Init(); err != nil {
		fmt.Printf("Warning: Audio failed to init: %v\n", err)
	}

	defer func() {
		if err := e.Display.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing display: %v\n", err)
		}
		if err := e.Audio.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing audio device: %v\n", err)
		}
	}()

	runEmulatorContext, cancelRunEmulatorContext := context.WithCancel(
		parentContext,
	)
	e.cancelRunner = cancelRunEmulatorContext
	errChan := make(chan error, 1)
	defer cancelRunEmulatorContext()

	go e.runCPU(runEmulatorContext, errChan)
	// Cannot be goroutine as SDL2 wants* to be on the main thread
	e.runDisplay(runEmulatorContext, errChan)

	err := <-errChan
	e.runLock.Lock()
	e.running = false
	e.runLock.Unlock()
	return err
}

func (e *Emulator) IsRunning() bool {
	e.runLock.Lock()
	isRunning := e.running
	e.runLock.Unlock()
	return isRunning
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

			e.memoryLock.Lock()
			timerErr := e.updateTimers()
			displayErr := e.Display.Present()
			e.memoryLock.Unlock()

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
			e.memoryLock.Lock()
			err := e.tick()
			e.memoryLock.Unlock()
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
		if err := e.Audio.Play(); err != nil {
			return err
		}
		e.CPU.SoundTimer--
	} else {
		if err := e.Audio.Pause(); err != nil {
			return err
		}
	}

	if e.CPU.DelayTimer > 0 {
		e.CPU.DelayTimer--
	}

	return nil
}

// Stops the runner (CPU and Display goroutine) threads.
func (e *Emulator) Destroy() {
	e.runLock.Lock()
	defer e.runLock.Unlock()
	e.running = false
	e.cancelRunner()
}
