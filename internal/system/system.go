package system

import (
	"fmt"
	"os"
	"time"

	"github.com/veandco/go-sdl2/sdl"

	"github.com/arpitchakladar/chip-8/internal/cpu"
	"github.com/arpitchakladar/chip-8/internal/display"
	"github.com/arpitchakladar/chip-8/internal/keyboard"
	"github.com/arpitchakladar/chip-8/internal/memory"
)

type System struct {
	CPU      *cpu.CentralProcessingUnit
	Memory   *memory.Memory
	Display  *display.Display
	Keyboard *keyboard.Keyboard
}

func New() *System {
	s := &System{
		CPU:      cpu.New(),
		Memory:   memory.New(),
		Display:  display.New(),
		Keyboard: keyboard.New(),
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
		s.Memory.Write(uint16(0x200+i), b)
	}
	return nil
}

// Step performs one CPU cycle
func (s *System) Step() error {
	// 1. Fetch
	hi := s.Memory.Read(s.CPU.ProgramCounter)
	lo := s.Memory.Read(s.CPU.ProgramCounter + 1)
	opcode := uint16(hi)<<8 | uint16(lo)

	// 2. Increment PC before execution
	s.CPU.ProgramCounter += 2

	// 3. Execute
	return s.CPU.Execute(opcode, s.Memory, s.Display, s.Keyboard)
}

func (s *System) HandleKeyboard(event *sdl.KeyboardEvent) {
	keyCode := event.Keysym.Sym
	// Type 0x300 is KeyDown, 0x301 is KeyUp in SDL
	isPressed := event.Type == sdl.KEYDOWN

	// Explicitly cast constants to sdl.Keycode to satisfy the map type
	mapping := map[sdl.Keycode]byte{
		sdl.Keycode(sdl.K_1): 0x1, sdl.Keycode(sdl.K_2): 0x2, sdl.Keycode(sdl.K_3): 0x3, sdl.Keycode(sdl.K_4): 0xC,
		sdl.Keycode(sdl.K_q): 0x4, sdl.Keycode(sdl.K_w): 0x5, sdl.Keycode(sdl.K_e): 0x6, sdl.Keycode(sdl.K_r): 0xD,
		sdl.Keycode(sdl.K_a): 0x7, sdl.Keycode(sdl.K_s): 0x8, sdl.Keycode(sdl.K_d): 0x9, sdl.Keycode(sdl.K_f): 0xE,
		sdl.Keycode(sdl.K_z): 0xA, sdl.Keycode(sdl.K_x): 0x0, sdl.Keycode(sdl.K_c): 0xB, sdl.Keycode(sdl.K_v): 0xF,
	}

	if chipKey, ok := mapping[keyCode]; ok {
		// Ensure your keyboard package uses a slice or array of bools
		s.Keyboard.Keys[chipKey] = isPressed
	}
}

func (s *System) Run(romPath string) error {
	// 1. Setup
	if err := s.Display.InitSDL(); err != nil {
		return fmt.Errorf("failed to init display: %w", err)
	}
	defer func() {
		if err := s.Display.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing display: %v\n", err)
		}
	}()

	if err := s.LoadROM(romPath); err != nil {
		return fmt.Errorf("failed to load ROM: %w", err)
	}

	// 2. Timing logic
	// We want the CPU to run fast, but Timers/Graphics at 60Hz.
	cpuHz := 700
	cpuInterval := time.Second / time.Duration(cpuHz)
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
				s.HandleKeyboard(t)
			}
		}

		// B. Step the CPU
		if err := s.Step(); err != nil {
			return err
		}

		// C. Sync Timers and Display to 60Hz
		if time.Since(lastTimerUpdate) >= timerInterval {
			s.CPU.UpdateTimers()
			if err := s.Display.Present(); err != nil {
				return err
			}
			lastTimerUpdate = time.Now()
		}
	}

	return nil
}
