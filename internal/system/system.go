package system

import (
	"fmt"
	"os"
	"time"

	"github.com/veandco/go-sdl2/sdl"

	"github.com/arpitchakladar/chip-8/internal/system/audio"
	"github.com/arpitchakladar/chip-8/internal/system/cpu"
	"github.com/arpitchakladar/chip-8/internal/system/display"
	"github.com/arpitchakladar/chip-8/internal/system/keyboard"
	"github.com/arpitchakladar/chip-8/internal/system/memory"
)

type System struct {
	CPU        *cpu.CentralProcessingUnit
	Memory     *memory.Memory
	Display    *display.Display
	Keyboard   *keyboard.Keyboard
	Audio      *audio.Audio
	ClockSpeed uint32 // Instructions per second (in Hz)
}

func WithClockSpeed(clockSpeed uint32) *System {
	s := &System{
		CPU:        cpu.New(),
		Memory:     memory.New(),
		Display:    display.New(),
		Keyboard:   keyboard.New(),
		Audio:      audio.New(),
		ClockSpeed: clockSpeed,
	}

	s.Memory.LoadFontSet()       // Load fonts into 0x000-0x050
	s.CPU.ProgramCounter = 0x200 // Set PC to 0x200
	return s
}

// LoadROM reads a .ch8 file and writes it into memory starting at 0x200
func (s *System) LoadROM(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// Chip-8 programs start at 0x200
	for i, b := range data {
		if err := s.Memory.Write(uint16(0x200+i), b); err != nil {
			return err
		}
	}
	return nil
}

// Step performs one CPU cycle
func (s *System) Step() error {
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

func (s *System) UpdateTimers() {
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

func (s *System) Run(romPath string) error {
	// 1. Setup
	if err := s.Display.Init(); err != nil {
		return fmt.Errorf("failed to init display: %w", err)
	}

	if err := s.Audio.Init(); err != nil {
		// Log error but maybe don't crash? Some systems don't have speakers.
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

	if err := s.LoadROM(romPath); err != nil {
		return fmt.Errorf("failed to load ROM: %w", err)
	}

	// 2. Timing logic
	// We want the CPU to run fast, but Timers/Graphics at 60Hz.
	cpuInterval := time.Second / time.Duration(s.ClockSpeed)
	timerInterval := time.Second / 60

	ticker := time.NewTicker(cpuInterval)
	defer ticker.Stop()

	lastTimerUpdate := time.Now()

	for range ticker.C {
		// A. Handle SDL Events
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				return nil
			case *sdl.KeyboardEvent:
				s.Keyboard.HandleKeyboard(t)
			}
		}

		// B. Step the CPU
		if err := s.Step(); err != nil {
			return err
		}

		// C. Sync Timers and Display to 60Hz
		if time.Since(lastTimerUpdate) >= timerInterval {
			s.UpdateTimers()
			if err := s.Display.Present(); err != nil {
				return err
			}
			lastTimerUpdate = time.Now()
		}
	}

	return nil
}
