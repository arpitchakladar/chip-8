// Package parser provides opcode generation for the CHIP-8 assembler.
// It performs the second pass of assembly, converting instructions to binary opcodes.
package parser

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

// Parse converts a single assembly instruction into its binary opcode representation.
// It resolves labels to their addresses and validates arguments.
//
// The function handles all CHIP-8 opcodes:
//   - Flow control: CLS, RET, JP, CALL
//   - Conditional: SE, SNE
//   - Arithmetic: ADD, SUB, SUBN, AND, OR, XOR, SHR, SHL
//   - Memory: LD (various forms)
//   - Display: DRW
//   - Input: SKP, SKNP
//   - Random: RND
//   - Data: DB (1-byte), DW (2-byte)
//
// Arguments:
//   - mnemonic: the instruction name (e.g., "LD", "JP")
//   - args: the instruction arguments (e.g., ["V0", "0x10"])
//   - line: the source line number for error reporting
//
// Returns:
//   - []byte: the binary opcode (2 bytes for most instructions, 1 for DB)
//   - error: if the mnemonic is unknown, arguments are invalid, or values are out of range
func (p *Parser) Parse(
	mnemonic string,
	args []string,
	line uint16,
) ([]byte, error) {
	upperMnemonic := strings.ToUpper(mnemonic)

	if mask, ok := simpleHandlers[upperMnemonic]; ok {
		return p.toBinary(p.Encoder.Raw(mask)), nil
	}

	return p.dispatchParse(upperMnemonic, args, line)
}

var simpleHandlers = map[string]uint16{
	"CLS": encoder.MaskCLS,
	"RET": encoder.MaskRET,
}

type parseHandler func(*Parser, string, []string, uint16) ([]byte, error)

var mnemonicHandlers = map[string]parseHandler{
	"JP":   (*Parser).handleJP,
	"CALL": (*Parser).handleCall,
	"SE":   (*Parser).handleSkipOp,
	"SNE":  (*Parser).handleSkipOp,
	"ADD":  (*Parser).handleAdd,
	"OR":   (*Parser).handleALU,
	"AND":  (*Parser).handleALU,
	"XOR":  (*Parser).handleALU,
	"SUB":  (*Parser).handleALU,
	"SHR":  (*Parser).handleALU,
	"SUBN": (*Parser).handleALU,
	"SHL":  (*Parser).handleALU,
	"LD":   wrapHandler((*Parser).handleLoad),
	"RND":  (*Parser).handleRND,
	"DRW":  (*Parser).handleDRW,
	"SKP":  (*Parser).handleKeySkip,
	"SKNP": (*Parser).handleKeySkip,
	"DW":   (*Parser).handleData,
	"DB":   (*Parser).handleData,
}

func wrapHandler(
	h func(*Parser, []string, uint16) ([]byte, error),
) parseHandler {
	return func(p *Parser, _ string, args []string, line uint16) ([]byte, error) {
		return h(p, args, line)
	}
}

func (p *Parser) dispatchParse(
	upperMnemonic string,
	args []string,
	line uint16,
) ([]byte, error) {
	if handler, ok := mnemonicHandlers[upperMnemonic]; ok {
		return handler(p, upperMnemonic, args, line)
	}
	return nil, p.parseErr(
		upperMnemonic,
		args,
		line,
		&UnknownMnemonicError{upperMnemonic, line},
	)
}

func (p *Parser) handleJP(
	mnemonic string,
	args []string,
	line uint16,
) ([]byte, error) {
	switch len(args) {
	case 2:
		if strings.ToUpper(args[0]) == "V0" {
			addr, err := p.resolveValue(args[1], mnemonic, line, 12)
			if err != nil {
				return nil, p.parseErr(mnemonic, args, line, err)
			}
			return p.toBinary(p.Encoder.Addr(encoder.MaskJPV0, addr)), nil
		}
		fallthrough
	case 1:
		addr, err := p.resolveValue(args[0], mnemonic, line, 12)
		if err != nil {
			return nil, p.parseErr(mnemonic, args, line, err)
		}
		return p.toBinary(p.Encoder.Addr(encoder.MaskJP, addr)), nil
	default:
		return nil, p.parseErr(
			mnemonic,
			args,
			line,
			&WrongArgCountError{mnemonic, line, 1, len(args)},
		)
	}
}

func (p *Parser) handleCall(
	mnemonic string,
	args []string,
	line uint16,
) ([]byte, error) {
	if len(args) != 1 {
		return nil, p.parseErr(
			mnemonic,
			args,
			line,
			&WrongArgCountError{mnemonic, line, 1, len(args)},
		)
	}
	addr, err := p.resolveValue(args[0], mnemonic, line, 12)
	if err != nil {
		return nil, p.parseErr(mnemonic, args, line, err)
	}
	return p.toBinary(p.Encoder.Addr(encoder.MaskCALL, addr)), nil
}

func (p *Parser) handleSkipOp(
	mnemonic string,
	args []string,
	line uint16,
) ([]byte, error) {
	immBase, regBase := encoder.MaskSE, encoder.MaskSER
	if mnemonic == "SNE" {
		immBase, regBase = encoder.MaskSNE, encoder.MaskSNER
	}
	res, err := p.handleSkip(immBase, regBase, mnemonic, args, line)
	if err != nil {
		return nil, p.parseErr(mnemonic, args, line, err)
	}
	return res, nil
}

func (p *Parser) handleAdd(
	mnemonic string,
	args []string,
	line uint16,
) ([]byte, error) {
	if len(args) != 2 {
		return nil, p.parseErr(
			mnemonic,
			args,
			line,
			&WrongArgCountError{mnemonic, line, 2, len(args)},
		)
	}
	if args[0] == "I" {
		vx, err := p.parseReg(args[1], mnemonic, line)
		if err != nil {
			return nil, p.parseErr(mnemonic, args, line, err)
		}
		return p.toBinary(p.Encoder.RegOnly(encoder.MaskMISC, vx, 0x1E)), nil
	}
	if p.isRegister(args[1]) {
		return p.handleRegReg(
			encoder.MaskALU,
			mnemonic,
			args,
			0x4,
			line,
		)
	}
	vx, err := p.parseReg(args[0], mnemonic, line)
	if err != nil {
		return nil, p.parseErr(mnemonic, args, line, err)
	}
	val, err := p.resolveValue(args[1], mnemonic, line, 8)
	if err != nil {
		return nil, p.parseErr(mnemonic, args, line, err)
	}
	return p.toBinary(p.Encoder.RegImm(encoder.MaskADD, vx, uint8(val))), nil
}

func (p *Parser) handleALU(
	mnemonic string,
	args []string,
	line uint16,
) ([]byte, error) {
	suffixes := map[string]uint16{
		"OR": 0x1, "AND": 0x2, "XOR": 0x3,
		"SUB": 0x5, "SHR": 0x6, "SUBN": 0x7, "SHL": 0xE,
	}
	suffix, ok := suffixes[mnemonic]
	if !ok {
		return nil, p.parseErr(
			mnemonic,
			args,
			line,
			&UnknownMnemonicError{mnemonic, 0},
		)
	}
	return p.handleRegReg(encoder.MaskALU, mnemonic, args, suffix, line)
}

func (p *Parser) handleRND(
	mnemonic string,
	args []string,
	line uint16,
) ([]byte, error) {
	if len(args) != 2 {
		return nil, p.parseErr(
			mnemonic,
			args,
			line,
			&WrongArgCountError{mnemonic, line, 2, len(args)},
		)
	}
	vx, err := p.parseReg(args[0], mnemonic, line)
	if err != nil {
		return nil, p.parseErr(mnemonic, args, line, err)
	}
	val, err := p.resolveValue(args[1], mnemonic, line, 8)
	if err != nil {
		return nil, p.parseErr(mnemonic, args, line, err)
	}
	return p.toBinary(p.Encoder.RegImm(encoder.MaskRND, vx, uint8(val))), nil
}

func (p *Parser) handleDRW(
	mnemonic string,
	args []string,
	line uint16,
) ([]byte, error) {
	if len(args) != 3 {
		return nil, p.parseErr(
			mnemonic,
			args,
			line,
			&WrongArgCountError{mnemonic, line, 3, len(args)},
		)
	}
	vx, err := p.parseReg(args[0], mnemonic, line)
	if err != nil {
		return nil, p.parseErr(mnemonic, args, line, err)
	}
	vy, err := p.parseReg(args[1], mnemonic, line)
	if err != nil {
		return nil, p.parseErr(mnemonic, args, line, err)
	}
	n, err := p.resolveValue(args[2], mnemonic, line, 4)
	if err != nil {
		return nil, p.parseErr(mnemonic, args, line, err)
	}
	return p.toBinary(
		p.Encoder.RegNibble(encoder.MaskDRW, vx, vy, uint8(n)),
	), nil
}

func (p *Parser) handleKeySkip(
	mnemonic string,
	args []string,
	line uint16,
) ([]byte, error) {
	if len(args) != 1 {
		return nil, p.parseErr(
			mnemonic,
			args,
			line,
			&WrongArgCountError{mnemonic, line, 1, len(args)},
		)
	}
	vx, err := p.parseReg(args[0], mnemonic, line)
	if err != nil {
		return nil, p.parseErr(mnemonic, args, line, err)
	}
	var opcode uint16
	if mnemonic == "SKNP" {
		opcode = 0xA1
	} else {
		opcode = 0x9E
	}
	return p.toBinary(p.Encoder.RegOnly(encoder.MaskKEY, vx, opcode)), nil
}

func (p *Parser) handleData(
	mnemonic string,
	args []string,
	line uint16,
) ([]byte, error) {
	if len(args) != 1 {
		return nil, p.parseErr(
			mnemonic,
			args,
			line,
			&WrongArgCountError{mnemonic, line, 1, len(args)},
		)
	}
	bits := 8
	if mnemonic == "DW" {
		bits = 16
	}
	val, err := p.resolveValue(args[0], mnemonic, line, bits)
	if err != nil {
		return nil, p.parseErr(mnemonic, args, line, err)
	}
	if bits == 16 {
		return p.toBinary(val), nil
	}
	return []byte{byte(val)}, nil
}

// --- Helper Handlers ---

// handleLoad processes LD (load) instructions with various source/destination combinations.
func (p *Parser) handleLoad(args []string, line uint16) ([]byte, error) {
	if len(args) != 2 {
		return nil, &WrongArgCountError{"LD", line, 2, len(args)}
	}
	dst, src := args[0], args[1]

	if dst == "I" {
		addr, err := p.resolveValue(src, "LD", line, 12)
		if err != nil {
			return nil, err
		}
		return p.toBinary(p.Encoder.Addr(encoder.MaskLDI, addr)), nil
	}

	if p.isRegister(dst) {
		return p.handleLoadVxSrc(dst, src, line)
	}

	if p.isRegister(src) {
		return p.handleLoadDstVx(dst, src, line)
	}

	return nil, &InvalidLoadError{line, dst, src}
}

func (p *Parser) handleLoadVxSrc(dst, src string, line uint16) ([]byte, error) {
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

func (p *Parser) handleLoadDstVx(dst, src string, line uint16) ([]byte, error) {
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
	return nil, &InvalidLoadError{line, dst, src}
}

// handleSkip processes SE (skip if equal) and SNE (skip if not equal) instructions.
// It handles both register-to-immediate and register-to-register comparisons.
func (p *Parser) handleSkip(
	immBase, regBase uint16,
	mnemonic string,
	args []string,
	line uint16,
) ([]byte, error) {
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
func (p *Parser) handleRegReg(
	base uint16,
	mnemonic string,
	args []string,
	suffix uint16,
	line uint16,
) ([]byte, error) {
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
func (p *Parser) parseReg(
	s string,
	mnemonic string,
	line uint16,
) (uint8, error) {
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
func (p *Parser) resolveValue(
	s string,
	mnemonic string,
	line uint16,
	bits int,
) (uint16, error) {
	var val uint16
	var ok bool

	// 1. Try Labels
	if val, ok = p.Labels[s]; !ok {
		// 2. Try Literal (Hex or Dec)
		clean := strings.ReplaceAll(
			strings.ReplaceAll(strings.ReplaceAll(s, "0X", ""), "0x", ""),
			"$",
			"",
		)
		base := 10
		if strings.Contains(strings.ToUpper(s), "0X") ||
			strings.Contains(s, "$") {
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
	maxValue := (uint32(1) << bits) - 1
	if uint32(val) > maxValue {
		return 0, &ImmediateOutOfRangeError{mnemonic, line, s, val, bits}
	}

	return val, nil
}

// parseErr wraps a child error into a ParseError with context about the failing instruction.
func (p *Parser) parseErr(
	mnemonic string,
	args []string,
	line uint16,
	child error,
) error {
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
