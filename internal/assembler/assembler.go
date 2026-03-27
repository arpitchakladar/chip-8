package assembler

import (
	"github.com/arpitchakladar/chip-8/internal/assembler/lexer"
	"github.com/arpitchakladar/chip-8/internal/assembler/parser"
)

type Assembler struct {
	Source         string
	Labels         map[string]uint16
	ProgramCounter uint16
}

func WithSource(source string) *Assembler {
	return &Assembler{
		Source:         source,
		Labels:         make(map[string]uint16),
		ProgramCounter: 0x200, // CHIP-8 programs start at 0x200
	}
}

func (a *Assembler) Assemble() ([]byte, error) {
	lexer := lexer.New(a.Source, a.ProgramCounter)
	labels, lines := lexer.ScanLabels()

	var program []byte

	parser := parser.WithLabels(labels)

	for _, line := range lines {
		opcode, err := parser.Parse(line.Mnemonic, line.Args)
		if err != nil {
			return nil, err
		}
		// Low byte comes before high byte
		program = append(program, opcode...)
	}

	return program, nil
}
