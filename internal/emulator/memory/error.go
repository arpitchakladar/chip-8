package memory

import "fmt"

// BoundsError occurs when attempting to access a memory address beyond the 4KB limit.
// CHIP-8 has 4096 bytes of memory (addresses 0x000 to 0xFFF).
type BoundsError struct {
	// Address is the out-of-bounds address that was accessed.
	Address uint16
	// Max is the maximum valid address (4095 = 0xFFF).
	Max uint16
}

func (e *BoundsError) Error() string {
	return fmt.Sprintf("out of bounds: address 0x%03X exceeds maximum 0x%03X", e.Address, e.Max)
}

// WriteProtectedError occurs when attempting to write to the system/interpreter area.
// The first 512 bytes (0x000-0x1FF) are reserved for the font set and should not be
// overwritten by user programs.
type WriteProtectedError struct {
	// Address is the protected address that was write attempted.
	Address uint16
}

func (e *WriteProtectedError) Error() string {
	return fmt.Sprintf("write to protected address 0x%03X (reserved for font set)", e.Address)
}
