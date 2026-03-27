package system

import (
	"os"

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
