package parser

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"github.com/arpitchakladar/chip-8/internal/assembler/encoder"
)

type Parser struct {
	Labels  map[string]uint16
	Encoder *encoder.Encoder
}

func WithLabels(labels map[string]uint16) *Parser {
	return &Parser{
		Labels:  labels,
		Encoder: encoder.New(),
	}
}

// Parse takes a mnemonic and its arguments, resolving any labels to their addresses.
func (p *Parser) Parse(mnemonic string, args []string) ([]byte, error) {
	mnemonic = strings.ToUpper(mnemonic)

	switch mnemonic {
	case "CLS":
		return p.toBinary(p.Encoder.Raw(encoder.MaskCLS)), nil
	case "RET":
		return p.toBinary(p.Encoder.Raw(encoder.MaskRET)), nil

	case "JP":
		if len(args) == 2 && strings.ToUpper(args[0]) == "V0" {
			addr, _ := p.resolveValue(args[1])
			return p.toBinary(p.Encoder.Addr(encoder.MaskJPV0, addr)), nil
		}
		addr, _ := p.resolveValue(args[0])
		return p.toBinary(p.Encoder.Addr(encoder.MaskJP, addr)), nil

	case "CALL":
		addr, _ := p.resolveValue(args[0])
		return p.toBinary(p.Encoder.Addr(encoder.MaskCALL, addr)), nil

	case "SE":
		return p.handleSkip(encoder.MaskSE, encoder.MaskSER, args)

	case "SNE":
		return p.handleSkip(encoder.MaskSNE, encoder.MaskSNER, args)

	case "ADD":
		if args[0] == "I" {
			vx, _ := p.parseReg(args[1])
			return p.toBinary(p.Encoder.RegOnly(encoder.MaskMISC, vx, 0x1E)), nil
		}
		if p.isRegister(args[1]) {
			return p.handleRegReg(encoder.MaskALU, args, 0x4)
		}
		vx, _ := p.parseReg(args[0])
		val, _ := p.resolveValue(args[1])
		return p.toBinary(p.Encoder.RegImm(encoder.MaskADD, vx, uint8(val))), nil

	case "OR":
		return p.handleRegReg(encoder.MaskALU, args, 0x1)
	case "AND":
		return p.handleRegReg(encoder.MaskALU, args, 0x2)
	case "XOR":
		return p.handleRegReg(encoder.MaskALU, args, 0x3)
	case "SUB":
		return p.handleRegReg(encoder.MaskALU, args, 0x5)
	case "SHR":
		return p.handleRegReg(encoder.MaskALU, args, 0x6)
	case "SUBN":
		return p.handleRegReg(encoder.MaskALU, args, 0x7)
	case "SHL":
		return p.handleRegReg(encoder.MaskALU, args, 0xE)

	case "LD":
		return p.handleLoad(args)

	case "RND":
		vx, _ := p.parseReg(args[0])
		val, _ := p.resolveValue(args[1])
		return p.toBinary(p.Encoder.RegImm(encoder.MaskRND, vx, uint8(val))), nil

	case "DRW":
		vx, _ := p.parseReg(args[0])
		vy, _ := p.parseReg(args[1])
		n, _ := p.resolveValue(args[2])
		return p.toBinary(p.Encoder.RegNibble(encoder.MaskDRW, vx, vy, uint8(n))), nil

	case "SKP":
		vx, _ := p.parseReg(args[0])
		return p.toBinary(p.Encoder.RegOnly(encoder.MaskKEY, vx, 0x9E)), nil

	case "SKNP":
		vx, _ := p.parseReg(args[0])
		return p.toBinary(p.Encoder.RegOnly(encoder.MaskKEY, vx, 0xA1)), nil

	case "DW":
		// Resolve the value (e.g., 0x82 -> 130)
		val, err := p.resolveValue(args[0])
		if err != nil {
			return []byte{}, err
		}
		return p.toBinary(val), nil

	case "DB":
		// Resolve the value (e.g., 0x82 -> 130)
		val, err := p.resolveValue(args[0])
		if err != nil {
			return []byte{}, err
		}
		return []byte{byte(val)}, nil

	default:
		return []byte{}, fmt.Errorf("unknown mnemonic: %s", mnemonic)
	}
}

// --- Helper Handlers ---
func (p *Parser) toBinary(opcode uint16) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, opcode)
	return buf
}

func (p *Parser) handleLoad(args []string) ([]byte, error) {
	dst, src := args[0], args[1]

	// LD I, addr
	if dst == "I" {
		addr, _ := p.resolveValue(src)
		return p.toBinary(p.Encoder.Addr(encoder.MaskLDI, addr)), nil
	}

	// LD Vx, [Source]
	if p.isRegister(dst) {
		vx, _ := p.parseReg(dst)
		switch {
		case src == "DT":
			return p.toBinary(p.Encoder.RegOnly(encoder.MaskMISC, vx, 0x07)), nil
		case src == "K":
			return p.toBinary(p.Encoder.RegOnly(encoder.MaskMISC, vx, 0x0A)), nil
		case p.isRegister(src):
			vy, _ := p.parseReg(src)
			return p.toBinary(p.Encoder.RegReg(encoder.MaskALU, vx, vy, 0x0)), nil
		case src == "[I]":
			return p.toBinary(p.Encoder.RegOnly(encoder.MaskMISC, vx, 0x65)), nil
		default:
			val, _ := p.resolveValue(src)
			return p.toBinary(p.Encoder.RegImm(encoder.MaskLD, vx, uint8(val))), nil
		}
	}

	// LD [Target], Vx (MISC family)
	if p.isRegister(src) {
		vx, _ := p.parseReg(src)
		switch dst {
		case "DT":
			return p.toBinary(p.Encoder.RegOnly(encoder.MaskMISC, vx, 0x15)), nil
		case "ST":
			return p.toBinary(p.Encoder.RegOnly(encoder.MaskMISC, vx, 0x18)), nil
		case "F":
			return p.toBinary(p.Encoder.RegOnly(encoder.MaskMISC, vx, 0x29)), nil
		case "B":
			return p.toBinary(p.Encoder.RegOnly(encoder.MaskMISC, vx, 0x33)), nil
		case "[I]":
			return p.toBinary(p.Encoder.RegOnly(encoder.MaskMISC, vx, 0x55)), nil
		}
	}

	return []byte{}, fmt.Errorf("invalid LD arguments")
}

func (p *Parser) handleSkip(immBase, regBase uint16, args []string) ([]byte, error) {
	vx, _ := p.parseReg(args[0])
	if p.isRegister(args[1]) {
		vy, _ := p.parseReg(args[1])
		return p.toBinary(p.Encoder.RegReg(regBase, vx, vy, 0x0)), nil
	}
	val, _ := p.resolveValue(args[1])
	return p.toBinary(p.Encoder.RegImm(immBase, vx, uint8(val))), nil
}

func (p *Parser) handleRegReg(base uint16, args []string, suffix uint16) ([]byte, error) {
	vx, err := p.parseReg(args[0])
	if err != nil {
		return nil, err
	}
	vy, err := p.parseReg(args[1])
	if err != nil {
		return nil, err
	}
	return p.toBinary(p.Encoder.RegReg(base, vx, vy, suffix)), nil
}

// --- Utility Functions ---
// Returns true if the given token resembles the syntax for a registor
func (p *Parser) isRegister(s string) bool {
	s = strings.ToUpper(s)
	return len(s) >= 2 && s[0] == 'V'
}

// Returns the registor number from its name
func (p *Parser) parseReg(s string) (uint8, error) {
	// Check if string is long enough and starts with 'v' or 'V'
	if !p.isRegister(s) {
		return 0, fmt.Errorf("invalid register format: %s (expected vX or VX)", s)
	}

	// Parse the remainder of the string as hex (base 16)
	// s[1:] skips the prefix character
	val, err := strconv.ParseUint(s[1:], 16, 8)
	if err != nil {
		return 0, fmt.Errorf("invalid register value: %w", err)
	}

	return uint8(val), nil
}

// Evaluate the value of a value type token (labels, constants)
func (p *Parser) resolveValue(s string) (uint16, error) {
	if val, ok := p.Labels[s]; ok {
		return val, nil
	}

	// Support hex (0x or $) and decimal
	clean := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, "0X", ""), "0x", ""), "$", "")
	base := 10
	if strings.Contains(strings.ToUpper(s), "0X") || strings.Contains(s, "$") {
		base = 16
	}

	v, err := strconv.ParseUint(clean, base, 16)
	return uint16(v), err
}

// Error handler
// func (p *Parser) parseErr(mnemonic string, args []string, line uint16, child error) error {
// 	return &ParseError{
// 		Mnemonic:   mnemonic,
// 		Args:       args,
// 		LineNumber: line,
// 		Child:      child,
// 	}
// }
