package assembler

import (
	"github.com/arpitchakladar/chip-8/internal/assembler/lexer"
	"github.com/arpitchakladar/chip-8/internal/assembler/parser"
)

// Assembler converts CHIP-8 assembly source code into executable bytecode.
// It uses a two-pass pipeline: lexer scans for labels, then parser generates opcodes.
type Assembler struct {
	// Source contains the raw assembly source code.
	Source string
	// Labels maps label names to their memory addresses.
	Labels map[string]uint16
	// ProgramCounter tracks the current address during assembly.
	ProgramCounter uint16
}

// New creates a new Assembler with the given source code.
// The ProgramCounter starts at 0x200 (CHIP-8 program start address).
func New(source string) *Assembler {
	return &Assembler{
		Source:         source,
		Labels:         make(map[string]uint16),
		ProgramCounter: 0x200, // Program start address
	}
}

// Assemble processes the source code and returns the compiled bytecode.
// It performs a two-pass assembly: first scanning labels, then generating opcodes.
func (a *Assembler) Assemble() ([]byte, error) {
	lexer := lexer.New(a.Source, a.ProgramCounter)
	labels, lines, err := lexer.ScanLabels()
	if err != nil {
		return nil, err
	}

	var program []byte

	parser := parser.New(labels)

	for _, line := range lines {
		opcode, err := parser.Parse(line.Mnemonic, line.Args, line.LineNumber)
		if err != nil {
			return nil, err
		}
		program = append(program, opcode...)
	}

	return program, nil
}
