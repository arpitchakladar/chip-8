package assembler

import (
	"strings"

	"github.com/arpitchakladar/chip-8/internal/assembler/parser"
)

type Assembler struct {
	Labels         map[string]uint16
	Instructions   []string
	ProgramCounter uint16
}

func WithInstructions(instructions []string) *Assembler {
	return &Assembler{
		Labels:         make(map[string]uint16),
		Instructions:   instructions,
		ProgramCounter: 0x200, // CHIP-8 programs start at 0x200
	}
}

func (a *Assembler) Assemble(input string) ([]byte, error) {
	lines := strings.Split(input, "\n")

	// --- Pass 1: Identify Labels ---
	currAddr := uint16(0x200)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ";") {
			continue
		}

		if label, found := strings.CutSuffix(line, ":"); found {
			a.Labels[label] = currAddr
		} else {
			currAddr += 2
		}
	}

	// --- Pass 2: Parse and Encode ---
	var binary []byte
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ";") || strings.HasSuffix(line, ":") {
			continue
		}

		// Basic split: "LD V1, 0x55" -> mnemonic="LD", args=["V1", "0x55"]
		parts := strings.Fields(strings.ReplaceAll(line, ",", " "))
		mnemonic := parts[0]
		args := parts[1:]

		opcode, err := parser.Parse(mnemonic, args, a.Labels)
		if err != nil {
			return nil, err
		}

		// CHIP-8 is Big-Endian (High byte first)
		binary = append(binary, byte(opcode>>8), byte(opcode&0xFF))
	}

	return binary, nil
}
