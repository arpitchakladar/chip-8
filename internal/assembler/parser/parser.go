package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/arpitchakladar/chip-8/internal/assembler/encoder"
)

// Parse takes a mnemonic and its arguments, resolving any labels to their addresses.
func Parse(mnemonic string, args []string, labels map[string]uint16) (uint16, error) {
	mnemonic = strings.ToUpper(mnemonic)

	switch mnemonic {
	case "CLS":
		return encoder.OpRaw(encoder.MaskCLS), nil
	case "RET":
		return encoder.OpRaw(encoder.MaskRET), nil

	case "JP":
		if len(args) == 2 && strings.ToUpper(args[0]) == "V0" {
			addr, _ := resolveValue(args[1], labels)
			return encoder.OpAddr(encoder.MaskJPV0, addr), nil
		}
		addr, _ := resolveValue(args[0], labels)
		return encoder.OpAddr(encoder.MaskJP, addr), nil

	case "CALL":
		addr, _ := resolveValue(args[0], labels)
		return encoder.OpAddr(encoder.MaskCALL, addr), nil

	case "SE":
		return handleSkip(encoder.MaskSE, encoder.MaskSER, args, labels)

	case "SNE":
		return handleSkip(encoder.MaskSNE, encoder.MaskSNER, args, labels)

	case "ADD":
		if isRegister(args[1]) {
			vx, _ := parseReg(args[0])
			vy, _ := parseReg(args[1])
			return encoder.OpRegReg(encoder.MaskALU, vx, vy, 0x4), nil
		}
		vx, _ := parseReg(args[0])
		val, _ := resolveValue(args[1], labels)
		return encoder.OpRegImm(encoder.MaskADD, vx, uint8(val)), nil

	case "OR":
		return handleRegReg(encoder.MaskALU, args, 0x1)
	case "AND":
		return handleRegReg(encoder.MaskALU, args, 0x2)
	case "XOR":
		return handleRegReg(encoder.MaskALU, args, 0x3)
	case "SUB":
		return handleRegReg(encoder.MaskALU, args, 0x5)
	case "SHR":
		return handleRegReg(encoder.MaskALU, args, 0x6)
	case "SUBN":
		return handleRegReg(encoder.MaskALU, args, 0x7)
	case "SHL":
		return handleRegReg(encoder.MaskALU, args, 0xE)

	case "LD":
		return handleLoad(args, labels)

	case "RND":
		vx, _ := parseReg(args[0])
		val, _ := resolveValue(args[1], labels)
		return encoder.OpRegImm(encoder.MaskRND, vx, uint8(val)), nil

	case "DRW":
		vx, _ := parseReg(args[0])
		vy, _ := parseReg(args[1])
		n, _ := resolveValue(args[2], labels)
		return encoder.OpRegNibble(encoder.MaskDRW, vx, vy, uint8(n)), nil

	case "SKP":
		vx, _ := parseReg(args[0])
		return encoder.OpRegOnly(encoder.MaskKEY, vx, 0x9E), nil

	case "SKNP":
		vx, _ := parseReg(args[0])
		return encoder.OpRegOnly(encoder.MaskKEY, vx, 0xA1), nil

	default:
		return 0, fmt.Errorf("unknown mnemonic: %s", mnemonic)
	}
}

// --- Helper Handlers ---
func handleLoad(args []string, labels map[string]uint16) (uint16, error) {
	dst, src := strings.ToUpper(args[0]), strings.ToUpper(args[1])

	if dst == "I" {
		addr, _ := resolveValue(src, labels)
		return encoder.OpAddr(encoder.MaskLDI, addr), nil
	}

	if isRegister(dst) {
		vx, _ := parseReg(dst)
		switch {
		case src == "DT":
			return encoder.OpRegOnly(encoder.MaskMISC, vx, 0x07), nil
		case src == "K":
			return encoder.OpRegOnly(encoder.MaskMISC, vx, 0x0A), nil
		case isRegister(src):
			vy, _ := parseReg(src)
			return encoder.OpRegReg(encoder.MaskALU, vx, vy, 0x0), nil
		default:
			val, _ := resolveValue(src, labels)
			return encoder.OpRegImm(encoder.MaskLD, vx, uint8(val)), nil
		}
	}

	// LD [Target], Vx (MISC family)
	if isRegister(src) {
		vx, _ := parseReg(src)
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
	if dst == "I" && isRegister(src) {
		vx, _ := parseReg(src)
		return encoder.OpRegOnly(encoder.MaskMISC, vx, 0x1E), nil
	}

	return 0, fmt.Errorf("invalid LD arguments")
}

func handleSkip(immBase, regBase uint16, args []string, labels map[string]uint16) (uint16, error) {
	vx, _ := parseReg(args[0])
	if isRegister(args[1]) {
		vy, _ := parseReg(args[1])
		return encoder.OpRegReg(regBase, vx, vy, 0x0), nil
	}
	val, _ := resolveValue(args[1], labels)
	return encoder.OpRegImm(immBase, vx, uint8(val)), nil
}

func handleRegReg(base uint16, args []string, suffix uint16) (uint16, error) {
	vx, _ := parseReg(args[0])
	vy, _ := parseReg(args[1])
	return encoder.OpRegReg(base, vx, vy, suffix), nil
}

// --- Utility Functions ---
func isRegister(s string) bool {
	s = strings.ToUpper(s)
	return len(s) >= 2 && s[0] == 'V'
}

func parseReg(s string) (uint8, error) {
	val, err := strconv.ParseUint(s[1:], 16, 8)
	return uint8(val), err
}

func resolveValue(s string, labels map[string]uint16) (uint16, error) {
	if val, ok := labels[s]; ok {
		return val, nil
	}

	// Support hex (0x or $) and decimal
	clean := strings.ReplaceAll(strings.ReplaceAll(s, "0x", ""), "$", "")
	base := 10
	if strings.Contains(s, "0x") || strings.Contains(s, "$") {
		base = 16
	}

	v, err := strconv.ParseUint(clean, base, 16)
	return uint16(v), err
}
