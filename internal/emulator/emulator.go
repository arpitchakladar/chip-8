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
	s := &Emulator{
		CPU:        cpu.New(),
		Memory:     memory.New(),
		Display:    display.New(),
		Keyboard:   keyboard.New(),
		Audio:      audio.New(),
		MemoryLock: sync.Mutex{},
		ClockSpeed: clockSpeed,
	}

	s.Memory.LoadFontSet()       // Load fonts into 0x000-0x050
	s.CPU.ProgramCounter = 0x200 // Set PC to 0x200
	return s
}

// LoadROM reads a .ch8 file and writes it into memory starting at 0x200
func (s *Emulator) LoadROM(romData []byte) error { // Chip-8 programs start at 0x200
	for i, b := range romData {
		if err := s.Memory.Write(uint16(0x200+i), b); err != nil {
			return err
		}
	}
	return nil
}

// Step performs one CPU cycle
func (s *Emulator) Step() error {
	// 1. Fetch
	hi, err := s.Memory.Read(s.CPU.ProgramCounter)
	if err != nil {
		return err
	}
	lo, err := s.Memory.Read(s.CPU.ProgramCounter + 1)
	if err != nil {
		return err
	}
	opcode := uint16(hi)<<8 | uint16(lo)

	// 2. Increment PC before execution
	s.CPU.ProgramCounter += 2

	// 3. Execute
	return s.CPU.Execute(opcode, s.Memory, s.Display, s.Keyboard)
}

func (s *Emulator) UpdateTimers() {
	if s.CPU.SoundTimer > 0 {
		// Unpause the audio device to start the buzz
		sdl.PauseAudioDevice(s.Audio.Device, false)
		s.CPU.SoundTimer--
	} else {
		// Pause the audio device when timer hits 0
		sdl.PauseAudioDevice(s.Audio.Device, true)
	}

	if s.CPU.DelayTimer > 0 {
		s.CPU.DelayTimer--
	}
}

func (s *Emulator) Run(romData []byte) error {
	// 1. Setup
	if err := s.Display.Init(); err != nil {
		return fmt.Errorf("failed to init display: %w", err)
	}

	if err := s.Audio.Init(); err != nil {
		// NOTE: Log error but maybe don't crash? Some systems don't have speakers.
		fmt.Printf("Warning: Audio failed to init: %v\n", err)
	} else if err := s.Audio.GenerateBeep(); err != nil {
		return err
	}

	defer func() {
		if err := s.Display.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing display: %v\n", err)
		}

		s.Audio.Close()
	}()

	if err := s.LoadROM(romData); err != nil {
		return fmt.Errorf("failed to load ROM: %w", err)
	}

	// Channels to communicate between the main routine (display) and
	// the cpu execution goroutine
	stop := make(chan bool)
	errChan := make(chan error)

	go func() {
		cpuClock := time.NewTicker(time.Second / time.Duration(s.ClockSpeed))
		defer cpuClock.Stop()

		for {
			select {
			case <-stop:
				return
			case <-cpuClock.C:
				s.MemoryLock.Lock()
				err := s.Step()
				s.MemoryLock.Unlock()
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
					close(stop)
					return nil
				case *sdl.KeyboardEvent:
					s.Keyboard.HandleKeyboard(t)
				}
			}

			// B. Update Timers (60Hz)
			s.MemoryLock.Lock()
			s.UpdateTimers()
			err := s.Display.Present()
			s.MemoryLock.Unlock()

			// C. Render (60Hz)
			if err != nil {
				return err
			}
		}
	}
}
