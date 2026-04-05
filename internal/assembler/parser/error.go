package parser

import (
	"fmt"
	"strings"
)

// UnknownMnemonicError: The parser encountered an instruction it doesn't recognise.
type UnknownMnemonicError struct {
	Mnemonic   string
	LineNumber uint16
}

func (e *UnknownMnemonicError) Error() string {
	return fmt.Sprintf(
		"UNKNOWN MNEMONIC: \"%s\" at line %d",
		e.Mnemonic,
		e.LineNumber,
	)
}

// WrongArgCountError: The instruction was given too few or too many arguments.
type WrongArgCountError struct {
	Mnemonic   string
	LineNumber uint16
	Expected   int
	Got        int
}

func (e *WrongArgCountError) Error() string {
	return fmt.Sprintf(
		"WRONG ARG COUNT [%s] at line %d: expected %d argument(s), got %d",
		e.Mnemonic, e.LineNumber, e.Expected, e.Got,
	)
}

// InvalidRegisterError: A token that should be a register (e.g. V3) wasn't parseable.
type InvalidRegisterError struct {
	Mnemonic   string
	LineNumber uint16
	Token      string
}

func (e *InvalidRegisterError) Error() string {
	return fmt.Sprintf(
		"INVALID REGISTER [%s] at line %d: \"%s\" is not a valid register (expected V0–VF)",
		e.Mnemonic,
		e.LineNumber,
		e.Token,
	)
}

// InvalidImmediateError: A literal value or label couldn't be resolved to a number.
type InvalidImmediateError struct {
	Mnemonic   string
	LineNumber uint16
	Token      string
	Child      error
}

func (e *InvalidImmediateError) Error() string {
	return fmt.Sprintf(
		"INVALID IMMEDIATE [%s] at line %d: \"%s\" could not be resolved (%v)",
		e.Mnemonic, e.LineNumber, e.Token, e.Child,
	)
}

// ImmediateOutOfRangeError: A value was resolved but exceeds what the instruction can encode.
// For example, a byte field receiving 0x1FF, or a 12-bit address field receiving 0x1000.
type ImmediateOutOfRangeError struct {
	Mnemonic   string
	LineNumber uint16
	Token      string
	Value      uint16
	MaxBits    int
}

func (e *ImmediateOutOfRangeError) Error() string {
	maxValue := (uint16(1) << e.MaxBits) - 1
	return fmt.Sprintf(
		"IMMEDIATE OUT OF RANGE [%s] at line %d: \"%s\" resolves to 0x%X, but field is %d-bit (max 0x%X)",
		e.Mnemonic,
		e.LineNumber,
		e.Token,
		e.Value,
		e.MaxBits,
		maxValue,
	)
}

// UnresolvedLabelError: A label was referenced but never defined anywhere in the source.
type UnresolvedLabelError struct {
	Mnemonic   string
	LineNumber uint16
	Label      string
}

func (e *UnresolvedLabelError) Error() string {
	return fmt.Sprintf(
		"UNRESOLVED LABEL [%s] at line %d: \"%s\" was never defined",
		e.Mnemonic, e.LineNumber, e.Label,
	)
}

// InvalidLoadError: An LD instruction was given a dst/src combination that doesn't map
// to any real CHIP-8 opcode (e.g. LD DT, DT or LD [I], [I]).
type InvalidLoadError struct {
	LineNumber uint16
	Dst        string
	Src        string
}

func (e *InvalidLoadError) Error() string {
	return fmt.Sprintf(
		"INVALID LD at line %d: no encoding exists for \"LD %s, %s\"",
		e.LineNumber, e.Dst, e.Src,
	)
}

// ParseError: Top-level wrapper when a mnemonic fails for any reason.
// Mirrors MemorySyncError in your CPU package — wraps the specific child error
// so callers can either print the full context or unwrap for type-switching.
type ParseError struct {
	Mnemonic   string
	Args       []string
	LineNumber uint16
	Child      error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf(
		"PARSE ERROR at line %d [%s %s]: %v",
		e.LineNumber,
		e.Mnemonic,
		strings.Join(e.Args, ", "),
		e.Child,
	)
}
