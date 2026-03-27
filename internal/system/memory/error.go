package memory

import "fmt"

// BoundsError: Attempted to access address > 4095
type BoundsError struct {
	Address uint16
	Max     uint16
}

func (e *BoundsError) Error() string {
	return fmt.Sprintf("OUT OF BOUNDS: Address 0x%03X exceeds physical RAM limit 0x%03X", e.Address, e.Max)
}

// WriteProtectedError: Attempted to write to the Interpreter/Font area (0x000 - 0x1FF)
type WriteProtectedError struct {
	Address uint16
}

func (e *WriteProtectedError) Error() string {
	return fmt.Sprintf("WRITE VIOLATION: Address 0x%03X is reserved for System/Fonts", e.Address)
}

// MemoryCorruptionError: Used if the ROM data is larger than available space
type MemoryLoadError struct {
	Size uint16
}

func (e *MemoryLoadError) Error() string {
	return fmt.Sprintf("LOAD ERROR: ROM size (%d bytes) exceeds available user memory (3584 bytes)", e.Size)
}
