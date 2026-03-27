package cpu

import "fmt"

// InvalidOpcodeError: The CPU encountered an instruction it doesn't recognize.
type InvalidOpcodeError struct {
	Opcode         uint16
	ProgramCounter uint16
}

func (e *InvalidOpcodeError) Error() string {
	return fmt.Sprintf("INVALID OPCODE: 0x%04X at PC 0x%03X", e.Opcode, e.ProgramCounter)
}

// StackError: Attempted to CALL when full (Overflow) or RET when empty (Underflow).
type StackError struct {
	IsOverflow     bool
	ProgramCounter uint16
}

func (e *StackError) Error() string {
	msg := "Stack Underflow (RET without CALL)"
	if e.IsOverflow {
		msg = "Stack Overflow (Exceeded 16 levels)"
	}
	return fmt.Sprintf("STACK ERROR at PC 0x%03X: %s", e.ProgramCounter, msg)
}

// MemorySyncError: High-level error if a CPU instruction triggers a memory failure.
type MemorySyncError struct {
	Opcode         uint16
	ProgramCounter uint16
	Child          error
}

func (e *MemorySyncError) Error() string {
	return fmt.Sprintf("CPU MEMORY ERROR [Opcode 0x%04X at PC 0x%03X]: %v", e.Opcode, e.ProgramCounter, e.Child)
}
