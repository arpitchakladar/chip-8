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
	labels, lines := lexer.ScanLabels(input, a.ProgramCounter)
	// TODO: REMOVE WHEN DEBUGGING IS OVER
	fmt.Printf("Labels: %+v\n", labels)
	fmt.Printf("Lines: %+v\n", lines)
	var program []byte

	for _, line := range lines {
		opcode, err := parser.Parse(line.Mnemonic, line.Args, labels)
		if err != nil {
			return nil, err
		}
		program = append(program, byte(opcode>>8), byte(opcode&0xFF))
	}

	return program, nil
}
