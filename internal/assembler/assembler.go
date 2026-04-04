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
// The Labels map is initialized empty and will be populated by the lexer during Assemble().
func New(source string) *Assembler {
	return &Assembler{
		Source:         source,
		Labels:         make(map[string]uint16),
		ProgramCounter: 0x200, // Program start address
	}
}

// Assemble processes the source code and returns the compiled bytecode.
// It performs a two-pass assembly:
//
// First pass (lexer):
//   - Scans for labels and builds a label-to-address map
//   - Collects all instruction lines
//   - Validates that __START and __END markers are present
//
// Second pass (parser):
//   - Converts each instruction to its binary opcode
//   - Resolves label references to their addresses
//   - Validates register indices and immediate values
//
// Returns:
//   - []byte: the compiled bytecode ready to be written to a .ch8 file
//   - error: if either pass fails (lexer or parser error)
func (a *Assembler) Assemble() ([]byte, error) {
	// First pass: Lexer scans for labels
	lexer := lexer.New(a.Source, a.ProgramCounter)
	labels, lines, err := lexer.ScanLabels()
	if err != nil {
		return nil, err
	}

	// Initialize the program bytecode
	var program []byte

	// Second pass: Parser converts instructions to opcodes
	parser := parser.New(labels)

	// Process each instruction line from the lexer
	for _, line := range lines {
		opcode, err := parser.Parse(line.Mnemonic, line.Args, line.LineNumber)
		if err != nil {
			return nil, err
		}
		program = append(program, opcode...)
	}

	return program, nil
}
