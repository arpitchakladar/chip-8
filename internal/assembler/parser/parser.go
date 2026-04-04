package parser

// Parser converts CHIP-8 assembly instructions into binary opcodes.
// It uses the label map from the lexer to resolve label references.

import (
	"encoding/binary"
	"strconv"
	"strings"

	"github.com/arpitchakladar/chip-8/internal/assembler/encoder"
)

// Parser converts CHIP-8 assembly instructions into binary opcodes.
// It uses the label map from the lexer to resolve label references.
type Parser struct {
	// Labels maps label names to their memory addresses (from lexer).
	Labels map[string]uint16
	// Encoder builds the binary opcode representations.
	Encoder *encoder.Encoder
}

// New creates a new Parser with the given label map.
func New(labels map[string]uint16) *Parser {
	return &Parser{
		Labels:  labels,
		Encoder: encoder.New(),
	}
}

// Parse takes a mnemonic and its arguments, resolving any labels to their addresses.
func (p *Parser) Parse(mnemonic string, args []string, line uint16) ([]byte, error) {
	upperMnemonic := strings.ToUpper(mnemonic)

	var opcode uint16

	switch upperMnemonic {
	case "CLS":
		opcode = p.Encoder.Raw(encoder.MaskCLS)
	case "RET":
		opcode = p.Encoder.Raw(encoder.MaskRET)

	case "JP":
		if len(args) == 2 && strings.ToUpper(args[0]) == "V0" {
			addr, err := p.resolveValue(args[1], upperMnemonic, line, 12)
			if err != nil {
				return nil, p.parseErr(mnemonic, args, line, err)
			}
			opcode = p.Encoder.Addr(encoder.MaskJPV0, addr)
		} else if len(args) == 1 {
			addr, err := p.resolveValue(args[0], upperMnemonic, line, 12)
			if err != nil {
				return nil, p.parseErr(mnemonic, args, line, err)
			}
			opcode = p.Encoder.Addr(encoder.MaskJP, addr)
		} else {
			return nil, p.parseErr(mnemonic, args, line, &WrongArgCountError{upperMnemonic, line, 1, len(args)})
		}

	case "CALL":
		if len(args) != 1 {
			return nil, p.parseErr(mnemonic, args, line, &WrongArgCountError{upperMnemonic, line, 1, len(args)})
		}
		addr, err := p.resolveValue(args[0], upperMnemonic, line, 12)
		if err != nil {
			return nil, p.parseErr(mnemonic, args, line, err)
		}
		opcode = p.Encoder.Addr(encoder.MaskCALL, addr)

	case "SE":
		res, err := p.handleSkip(encoder.MaskSE, encoder.MaskSER, upperMnemonic, args, line)
		if err != nil {
			return nil, p.parseErr(mnemonic, args, line, err)
		}
		return res, nil

	case "SNE":
		res, err := p.handleSkip(encoder.MaskSNE, encoder.MaskSNER, upperMnemonic, args, line)
		if err != nil {
			return nil, p.parseErr(mnemonic, args, line, err)
		}
		return res, nil

	case "ADD":
		if len(args) != 2 {
			return nil, p.parseErr(mnemonic, args, line, &WrongArgCountError{upperMnemonic, line, 2, len(args)})
		}
		if args[0] == "I" {
			vx, err := p.parseReg(args[1], upperMnemonic, line)
			if err != nil {
				return nil, p.parseErr(mnemonic, args, line, err)
			}
			opcode = p.Encoder.RegOnly(encoder.MaskMISC, vx, 0x1E)
		} else if p.isRegister(args[1]) {
			return p.handleRegReg(encoder.MaskALU, upperMnemonic, args, 0x4, line)
		} else {
			vx, err := p.parseReg(args[0], upperMnemonic, line)
			if err != nil {
				return nil, p.parseErr(mnemonic, args, line, err)
			}
			val, err := p.resolveValue(args[1], upperMnemonic, line, 8)
			if err != nil {
				return nil, p.parseErr(mnemonic, args, line, err)
			}
			opcode = p.Encoder.RegImm(encoder.MaskADD, vx, uint8(val))
		}

	case "OR":
		return p.handleRegReg(encoder.MaskALU, upperMnemonic, args, 0x1, line)
	case "AND":
		return p.handleRegReg(encoder.MaskALU, upperMnemonic, args, 0x2, line)
	case "XOR":
		return p.handleRegReg(encoder.MaskALU, upperMnemonic, args, 0x3, line)
	case "SUB":
		return p.handleRegReg(encoder.MaskALU, upperMnemonic, args, 0x5, line)
	case "SHR":
		return p.handleRegReg(encoder.MaskALU, upperMnemonic, args, 0x6, line)
	case "SUBN":
		return p.handleRegReg(encoder.MaskALU, upperMnemonic, args, 0x7, line)
	case "SHL":
		return p.handleRegReg(encoder.MaskALU, upperMnemonic, args, 0xE, line)

	case "LD":
		res, err := p.handleLoad(args, line)
		if err != nil {
			return nil, p.parseErr(mnemonic, args, line, err)
		}
		return res, nil

	case "RND":
		if len(args) != 2 {
			return nil, p.parseErr(mnemonic, args, line, &WrongArgCountError{upperMnemonic, line, 2, len(args)})
		}
		vx, err := p.parseReg(args[0], upperMnemonic, line)
		if err != nil {
			return nil, p.parseErr(mnemonic, args, line, err)
		}
		val, err := p.resolveValue(args[1], upperMnemonic, line, 8)
		if err != nil {
			return nil, p.parseErr(mnemonic, args, line, err)
		}
		opcode = p.Encoder.RegImm(encoder.MaskRND, vx, uint8(val))

	case "DRW":
		if len(args) != 3 {
			return nil, p.parseErr(mnemonic, args, line, &WrongArgCountError{upperMnemonic, line, 3, len(args)})
		}
		vx, err := p.parseReg(args[0], upperMnemonic, line)
		if err != nil {
			return nil, p.parseErr(mnemonic, args, line, err)
		}
		vy, err := p.parseReg(args[1], upperMnemonic, line)
		if err != nil {
			return nil, p.parseErr(mnemonic, args, line, err)
		}
		n, err := p.resolveValue(args[2], upperMnemonic, line, 4)
		if err != nil {
			return nil, p.parseErr(mnemonic, args, line, err)
		}
		opcode = p.Encoder.RegNibble(encoder.MaskDRW, vx, vy, uint8(n))

	case "SKP":
		if len(args) != 1 {
			return nil, p.parseErr(mnemonic, args, line, &WrongArgCountError{upperMnemonic, line, 1, len(args)})
		}
		vx, err := p.parseReg(args[0], upperMnemonic, line)
		if err != nil {
			return nil, p.parseErr(mnemonic, args, line, err)
		}
		opcode = p.Encoder.RegOnly(encoder.MaskKEY, vx, 0x9E)

	case "SKNP":
		if len(args) != 1 {
			return nil, p.parseErr(mnemonic, args, line, &WrongArgCountError{upperMnemonic, line, 1, len(args)})
		}
		vx, err := p.parseReg(args[0], upperMnemonic, line)
		if err != nil {
			return nil, p.parseErr(mnemonic, args, line, err)
		}
		opcode = p.Encoder.RegOnly(encoder.MaskKEY, vx, 0xA1)

	case "DW":
		if len(args) != 1 {
			return nil, p.parseErr(mnemonic, args, line, &WrongArgCountError{upperMnemonic, line, 1, len(args)})
		}
		val, err := p.resolveValue(args[0], upperMnemonic, line, 16)
		if err != nil {
			return nil, p.parseErr(mnemonic, args, line, err)
		}
		return p.toBinary(val), nil

	case "DB":
		if len(args) != 1 {
			return nil, p.parseErr(mnemonic, args, line, &WrongArgCountError{upperMnemonic, line, 1, len(args)})
		}
		val, err := p.resolveValue(args[0], upperMnemonic, line, 8)
		if err != nil {
			return nil, p.parseErr(mnemonic, args, line, err)
		}
		return []byte{byte(val)}, nil

	default:
		return nil, p.parseErr(mnemonic, args, line, &UnknownMnemonicError{upperMnemonic, line})
	}

	return p.toBinary(opcode), nil
}

// --- Helper Handlers ---

// handleLoad processes LD (load) instructions with various source/destination combinations.
func (p *Parser) handleLoad(args []string, line uint16) ([]byte, error) {
	if len(args) != 2 {
		return nil, &WrongArgCountError{"LD", line, 2, len(args)}
	}
	dst, src := args[0], args[1]

	// LD I, addr
	if dst == "I" {
		addr, err := p.resolveValue(src, "LD", line, 12)
		if err != nil {
			return nil, err
		}
		return p.toBinary(p.Encoder.Addr(encoder.MaskLDI, addr)), nil
	}

	// LD Vx, [Source]
	if p.isRegister(dst) {
		vx, err := p.parseReg(dst, "LD", line)
		if err != nil {
			return nil, err
		}
		switch {
		case src == "DT":
			return p.toBinary(p.Encoder.RegOnly(encoder.MaskMISC, vx, 0x07)), nil
		case src == "K":
			return p.toBinary(p.Encoder.RegOnly(encoder.MaskMISC, vx, 0x0A)), nil
		case p.isRegister(src):
			vy, err := p.parseReg(src, "LD", line)
			if err != nil {
				return nil, err
			}
			return p.toBinary(p.Encoder.RegReg(encoder.MaskALU, vx, vy, 0x0)), nil
		case src == "[I]":
			return p.toBinary(p.Encoder.RegOnly(encoder.MaskMISC, vx, 0x65)), nil
		default:
			val, err := p.resolveValue(src, "LD", line, 8)
			if err != nil {
				return nil, err
			}
			return p.toBinary(p.Encoder.RegImm(encoder.MaskLD, vx, uint8(val))), nil
		}
	}

	// LD [Target], Vx
	if p.isRegister(src) {
		vx, err := p.parseReg(src, "LD", line)
		if err != nil {
			return nil, err
		}
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

	return nil, &InvalidLoadError{line, dst, src}
}

// handleSkip processes SE (skip if equal) and SNE (skip if not equal) instructions.
// It handles both register-to-immediate and register-to-register comparisons.
func (p *Parser) handleSkip(immBase, regBase uint16, mnemonic string, args []string, line uint16) ([]byte, error) {
	if len(args) != 2 {
		return nil, &WrongArgCountError{mnemonic, line, 2, len(args)}
	}
	vx, err := p.parseReg(args[0], mnemonic, line)
	if err != nil {
		return nil, err
	}

	if p.isRegister(args[1]) {
		vy, err := p.parseReg(args[1], mnemonic, line)
		if err != nil {
			return nil, err
		}
		return p.toBinary(p.Encoder.RegReg(regBase, vx, vy, 0x0)), nil
	}

	val, err := p.resolveValue(args[1], mnemonic, line, 8)
	if err != nil {
		return nil, err
	}
	return p.toBinary(p.Encoder.RegImm(immBase, vx, uint8(val))), nil
}

// handleRegReg processes two-register arithmetic/logic instructions (OR, AND, XOR, SUB, etc.).
func (p *Parser) handleRegReg(base uint16, mnemonic string, args []string, suffix uint16, line uint16) ([]byte, error) {
	if len(args) != 2 {
		return nil, &WrongArgCountError{mnemonic, line, 2, len(args)}
	}
	vx, err := p.parseReg(args[0], mnemonic, line)
	if err != nil {
		return nil, err
	}
	vy, err := p.parseReg(args[1], mnemonic, line)
	if err != nil {
		return nil, err
	}
	return p.toBinary(p.Encoder.RegReg(base, vx, vy, suffix)), nil
}

// --- Utility Functions ---

// toBinary converts a 16-bit opcode to a 2-byte slice in big-endian format.
func (p *Parser) toBinary(opcode uint16) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, opcode)
	return buf
}

// isRegister returns true if the string looks like a register (e.g., "V0", "VF").
func (p *Parser) isRegister(s string) bool {
	s = strings.ToUpper(s)
	return len(s) >= 2 && s[0] == 'V'
}

// parseReg parses a register string (Vx) and returns the register index.
// It returns an error if the string is not a valid register.
func (p *Parser) parseReg(s string, mnemonic string, line uint16) (uint8, error) {
	if !p.isRegister(s) {
		return 0, &InvalidRegisterError{mnemonic, line, s}
	}

	val, err := strconv.ParseUint(s[1:], 16, 8)
	if err != nil || val > 0xF {
		return 0, &InvalidRegisterError{mnemonic, line, s}
	}

	return uint8(val), nil
}

// resolveValue resolves a string to a numeric value.
// It first checks the label map, then tries to parse as a literal (hex or decimal).
// The bits parameter specifies the maximum bit width for range checking.
func (p *Parser) resolveValue(s string, mnemonic string, line uint16, bits int) (uint16, error) {
	var val uint16
	var ok bool

	// 1. Try Labels
	if val, ok = p.Labels[s]; !ok {
		// 2. Try Literal (Hex or Dec)
		clean := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, "0X", ""), "0x", ""), "$", "")
		base := 10
		if strings.Contains(strings.ToUpper(s), "0X") || strings.Contains(s, "$") {
			base = 16
		}

		v64, err := strconv.ParseUint(clean, base, 16)
		if err != nil {
			// If it's not a number and not in labels, it's an unresolved label
			// (assuming labels don't start with numbers)
			if base == 10 && (s[0] < '0' || s[0] > '9') {
				return 0, &UnresolvedLabelError{mnemonic, line, s}
			}
			return 0, &InvalidImmediateError{mnemonic, line, s, err}
		}
		val = uint16(v64)
	}

	// 3. Range Check
	max := (uint32(1) << bits) - 1
	if uint32(val) > max {
		return 0, &ImmediateOutOfRangeError{mnemonic, line, s, val, bits}
	}

	return val, nil
}

// parseErr wraps a child error into a ParseError with context about the failing instruction.
func (p *Parser) parseErr(mnemonic string, args []string, line uint16, child error) error {
	// If it's already a ParseError, don't double wrap
	if _, ok := child.(*ParseError); ok {
		return child
	}
	return &ParseError{
		Mnemonic:   mnemonic,
		Args:       args,
		LineNumber: line,
		Child:      child,
	}
}
