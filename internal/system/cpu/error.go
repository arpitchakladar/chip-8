package cpu

import "fmt"

// CPUError represents an error during instruction execution
type CPUError struct {
	Opcode uint16
	PC     uint16
	Err    error
}

func (e *CPUError) Error() string {
	return fmt.Sprintf("CPU Error at PC 0x%03X [Opcode 0x%04X]: %v", e.PC, e.Opcode, e.Err)
}
