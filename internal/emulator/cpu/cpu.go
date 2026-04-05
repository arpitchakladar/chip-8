package cpu

// The CPU is responsible for identifying and running
// each opcode
type CPU struct {
	// Registers are the 16 general-purpose 8-bit registers.
	// Historically referred to as V0 through VF.
	Registers [16]byte

	// IndexRegister stores memory addresses for use in operations.
	// Historically referred to as the 'I' register.
	IndexRegister uint16

	// ProgramCounter stores the memory address of the next instruction to be executed.
	ProgramCounter uint16

	// StackPointer points to the current top of the stack.
	StackPointer uint8

	// Stack is used to store the return addresses when subroutines are called.
	// It allows for up to 16 levels of nested function calls.
	Stack [16]uint16

	// DelayTimer is used for game events; it decrements at a rate of 60Hz.
	DelayTimer byte

	// SoundTimer decrements at 60Hz and triggers a buzz as long as the value is > 0.
	SoundTimer byte
}

// New initializes a CPU with the standard entry point for Chip-8 programs.
func New() *CPU {
	return new(CPU)
}
