package assembler

import (
	"fmt"
	"github.com/arpitchakladar/chip-8/internal/assembler/lexer"
	"github.com/arpitchakladar/chip-8/internal/assembler/parser"
)

type Assembler struct {
	Labels         map[string]uint16
	ProgramCounter uint16
}

func New() *Assembler {
	return &Assembler{
		Labels:         make(map[string]uint16),
		ProgramCounter: 0x200, // CHIP-8 programs start at 0x200
	}
}

func (a *Assembler) Assemble(input string) ([]byte, error) {
	lexer := lexer.New(input, a.ProgramCounter)
	labels, lines := lexer.ScanLabels()
	// TODO: REMOVE WHEN DEBUGGING IS OVER
	fmt.Printf("Labels: %+v\n", labels)
	fmt.Printf("Lines: %+v\n", lines)
	var program []byte

	parser := parser.WithLabels(labels)

	for _, line := range lines {
		opcode, err := parser.Parse(line.Mnemonic, line.Args)
		if err != nil {
			return nil, err
		}
		// Low byte comes before high byte
		program = append(program, byte(opcode>>8), byte(opcode&0xFF))
	}

	return program, nil
}
