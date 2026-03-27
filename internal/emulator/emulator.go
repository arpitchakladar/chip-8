package emulator

import (
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

type Emulator struct {
	CPU        *cpu.CentralProcessingUnit
	Memory     *memory.Memory
	Display    *display.Display
	Keyboard   *keyboard.Keyboard
	Audio      *audio.Audio
	ClockSpeed uint32 // Instructions per second (in Hz)
	MemoryLock sync.Mutex
}

func WithClockSpeed(clockSpeed uint32) *Emulator {
	e := &Emulator{
		CPU:        cpu.New(),
		Memory:     memory.New(),
		Display:    display.New(),
		Keyboard:   keyboard.New(),
		Audio:      audio.New(),
		MemoryLock: sync.Mutex{},
		ClockSpeed: clockSpeed,
	}

	e.Memory.LoadFontSet()       // Load fonts into 0x000-0x050
	e.CPU.ProgramCounter = 0x200 // Set PC to 0x200
	return e
}

// LoadROM reads a .ch8 file and writes it into memory starting at 0x200
func (e *Emulator) LoadROM(romData []byte) error { // Chip-8 programs start at 0x200
	for i, b := range romData {
		if err := e.Memory.Write(uint16(0x200+i), b); err != nil {
			return err
		}
	}
	return nil
}

// Step performs one CPU cycle
func (e *Emulator) Step() error {
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

func (e *Emulator) UpdateTimers() error {
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

func (e *Emulator) Run(romData []byte) error {
	// 1. Setup
	if err := e.Display.Init(); err != nil {
		return fmt.Errorf("failed to init display: %w", err)
	}

	if err := e.Audio.Init(); err != nil {
		// NOTE: Log error but maybe don't crash? Some systems don't have speakere.
		fmt.Printf("Warning: Audio failed to init: %v\n", err)
	}

	defer func() {
		if err := e.Display.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing display: %v\n", err)
		}

		e.Audio.Close()
	}()

	if err := e.LoadROM(romData); err != nil {
		return fmt.Errorf("failed to load ROM: %w", err)
	}

	// Channels to communicate between the main routine (display) and
	// the cpu execution goroutine
	stop := make(chan struct{})
	defer close(stop)
	errChan := make(chan error, 1)

	go func() {
		cpuClock := time.NewTicker(time.Second / time.Duration(e.ClockSpeed))
		defer cpuClock.Stop()

		for {
			select {
			case <-stop:
				return
			case <-cpuClock.C:
				e.MemoryLock.Lock()
				err := e.Step()
				e.MemoryLock.Unlock()
				if err != nil {
					errChan <- err
					return
				}
			}
		}
	}()

	uiClock := time.NewTicker(time.Second / 60)
	defer uiClock.Stop()

	for {
		select {
		case err := <-errChan:
			return err
		case <-uiClock.C:
			// A. Handle Events (Must be main thread)
			for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
				switch t := event.(type) {
				case *sdl.QuitEvent:
					return nil
				case *sdl.KeyboardEvent:
					e.Keyboard.HandleKeyboard(t)
				}
			}

			e.MemoryLock.Lock()
			// B. Update Timers (60Hz)
			timerErr := e.UpdateTimers()
			// C. Display buffer (60Hz)
			displayErr := e.Display.Present()
			e.MemoryLock.Unlock()

			if timerErr != nil {
				return timerErr
			}
			if displayErr != nil {
				return displayErr
			}
		}
	}
}
