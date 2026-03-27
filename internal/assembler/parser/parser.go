package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/arpitchakladar/chip-8/internal/assembler/encoder"
)

type Parser struct {
	Labels map[string]uint16
}

func WithLabels(labels map[string]uint16) *Parser {
	return &Parser{
		Labels: labels,
	}
}


// Parse takes a mnemonic and its arguments, resolving any labels to their addresses.
func (p *Parser) Parse(mnemonic string, args []string) (uint16, error) {
	mnemonic = strings.ToUpper(mnemonic)

	switch mnemonic {
	case "CLS":
		return encoder.OpRaw(encoder.MaskCLS), nil
	case "RET":
		return encoder.OpRaw(encoder.MaskRET), nil

	case "JP":
		if len(args) == 2 && strings.ToUpper(args[0]) == "V0" {
			addr, _ := p.resolveValue(args[1])
			return encoder.OpAddr(encoder.MaskJPV0, addr), nil
		}
		addr, _ := p.resolveValue(args[0])
		return encoder.OpAddr(encoder.MaskJP, addr), nil

	case "CALL":
		addr, _ := p.resolveValue(args[0])
		return encoder.OpAddr(encoder.MaskCALL, addr), nil

	case "SE":
		return p.handleSkip(encoder.MaskSE, encoder.MaskSER, args)

	case "SNE":
		return p.handleSkip(encoder.MaskSNE, encoder.MaskSNER, args)

	case "ADD":
		if p.isRegister(args[1]) {
			return p.handleRegReg(encoder.MaskALU, args, 0x4)
		}
		vx, _ := p.parseReg(args[0])
		val, _ := p.resolveValue(args[1])
		return encoder.OpRegImm(encoder.MaskADD, vx, uint8(val)), nil

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
		return encoder.OpRegImm(encoder.MaskRND, vx, uint8(val)), nil

	case "DRW":
		vx, _ := p.parseReg(args[0])
		vy, _ := p.parseReg(args[1])
		n, _ := p.resolveValue(args[2])
		return encoder.OpRegNibble(encoder.MaskDRW, vx, vy, uint8(n)), nil

	case "SKP":
		vx, _ := p.parseReg(args[0])
		return encoder.OpRegOnly(encoder.MaskKEY, vx, 0x9E), nil

	case "SKNP":
		vx, _ := p.parseReg(args[0])
		return encoder.OpRegOnly(encoder.MaskKEY, vx, 0xA1), nil

	// TODO: DON'T USE THIS. Make sure that the assembler is able to read raw bytes
	case "DW":
		// Resolve the value (e.g., 0x82 -> 130)
		val, err := p.resolveValue(args[0])
		if err != nil {
			return 0, err
		}
		// Return it as a raw 16-bit value.
		// If you only want 8 bits, you'll eventually need to
		// refactor your assembler to handle byte-streams.
		return uint16(val), nil

	default:
		return 0, fmt.Errorf("unknown mnemonic: %s", mnemonic)
	}
}

// --- Helper Handlers ---
func (p *Parser) handleLoad(args []string) (uint16, error) {
	dst, src := args[0], args[1]

	if dst == "I" {
		addr, _ := p.resolveValue(src)
		return encoder.OpAddr(encoder.MaskLDI, addr), nil
	}

	if p.isRegister(dst) {
		vx, _ := p.parseReg(dst)
		switch {
		case src == "DT":
			return encoder.OpRegOnly(encoder.MaskMISC, vx, 0x07), nil
		case src == "K":
			return encoder.OpRegOnly(encoder.MaskMISC, vx, 0x0A), nil
		case p.isRegister(src):
			vy, _ := p.parseReg(src)
			return encoder.OpRegReg(encoder.MaskALU, vx, vy, 0x0), nil
		default:
			val, _ := p.resolveValue(src)
			return encoder.OpRegImm(encoder.MaskLD, vx, uint8(val)), nil
		}
	}

	// LD [Target], Vx (MISC family)
	if p.isRegister(src) {
		vx, _ := p.parseReg(src)
		switch dst {
		case "DT":
			return encoder.OpRegOnly(encoder.MaskMISC, vx, 0x15), nil
		case "ST":
			return encoder.OpRegOnly(encoder.MaskMISC, vx, 0x18), nil
		case "F":
			return encoder.OpRegOnly(encoder.MaskMISC, vx, 0x29), nil
		case "B":
			return encoder.OpRegOnly(encoder.MaskMISC, vx, 0x33), nil
		case "[I]":
			return encoder.OpRegOnly(encoder.MaskMISC, vx, 0x55), nil
		}
	}

	// LD I, Vx (Add to I)
	if dst == "I" && p.isRegister(src) {
		vx, _ := p.parseReg(src)
		return encoder.OpRegOnly(encoder.MaskMISC, vx, 0x1E), nil
	}

	return 0, fmt.Errorf("invalid LD arguments")
}

func (p *Parser) handleSkip(immBase, regBase uint16, args []string) (uint16, error) {
	vx, _ := p.parseReg(args[0])
	if p.isRegister(args[1]) {
		vy, _ := p.parseReg(args[1])
		return encoder.OpRegReg(regBase, vx, vy, 0x0), nil
	}
	val, _ := p.resolveValue(args[1])
	return encoder.OpRegImm(immBase, vx, uint8(val)), nil
}

func (p *Parser) handleRegReg(base uint16, args []string, suffix uint16) (uint16, error) {
	vx, _ := p.parseReg(args[0])
	vy, _ := p.parseReg(args[1])
	return encoder.OpRegReg(base, vx, vy, suffix), nil
}

// --- Utility Functions ---
func (p *Parser) isRegister(s string) bool {
	s = strings.ToUpper(s)
	return len(s) >= 2 && s[0] == 'V'
}

func (p *Parser) parseReg(s string) (uint8, error) {
	val, err := strconv.ParseUint(s[1:], 16, 8)
	return uint8(val), err
}

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
