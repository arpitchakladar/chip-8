package memory

// Memory provides the read/write interface for CHIP-8 system memory.
// It manages 4096 bytes of RAM with font set loading and write protection.

const (
	// Size is the total memory size in bytes (4KB, standard CHIP-8).
	Size = 4096
)

// Memory represents the 4096 bytes of RAM in a CHIP-8 system.
type Memory struct {
	// RAM is the physical storage for all memory operations.
	// Memory layout:
	//   - 0x000-0x1FF (512 bytes): Reserved for font set and system use
	//   - 0x200-0xFFF (3840 bytes): Program ROM and work RAM
	RAM [Size]byte
}

// New creates a blank 4KB Memory instance with all bytes initialized to zero.
// The font set must be loaded using LoadFontSet() before the emulator can run.
func New() *Memory {
	return new(Memory)
}

// Reset clears all 4096 bytes of memory to zero.
// Note: After resetting, you must call LoadFontSet() to restore the font data,
// otherwise the display will not render characters correctly.
func (m *Memory) Reset() {
	m.RAM = [Size]byte{}
}

// Read returns the byte stored at the given memory address.
//
// Parameters:
//   - address: 16-bit memory address (0x000 to 0xFFF)
//
// Returns:
//   - byte: the value stored at the address
//   - error: *BoundsError if address is out of range (>= 4096)
func (m *Memory) Read(address uint16) (byte, error) {
	if address >= Size {
		return 0, &BoundsError{Address: address, Max: 4095}
	}
	return m.RAM[address], nil
}

// Write stores a byte at the given memory address.
//
// Parameters:
//   - address: 16-bit memory address (0x000 to 0xFFF)
//   - value: the byte to write
//
// Returns:
//   - nil on success
//   - *BoundsError if address is out of range (>= 4096)
//   - *WriteProtectedError if attempting to write to font area (0x000-0x1FF)
//
// The first 512 bytes (0x000-0x1FF) are write-protected because they
// contain the font set. Standard programs should not write to this area.
func (m *Memory) Write(address uint16, value byte) error {
	// Check physical bounds
	if address >= Size {
		return &BoundsError{Address: address, Max: 4095}
	}

	// Check for protected area (font set location)
	// Standard ROMs should never overwrite the font set
	if address < 0x200 {
		return &WriteProtectedError{Address: address}
	}

	m.RAM[address] = value
	return nil
}

// LoadFontSet populates the first 80 bytes of memory (0x000-0x04F)
// with the standard 4x5 pixel font set for hexadecimal characters 0-F.
//
// Each character is encoded as 5 bytes, with each byte representing
// a row of 8 pixels (only the lower 4 bits are used):
//
//	Byte 0: Top row (bits 4-7 used)
//	Byte 1: Second row
//	Byte 2: Third row
//	Byte 3: Fourth row
//	Byte 4: Bottom row
//
// Font data is indexed by character: address = character * 5
// e.g., character '0' at 0x000, '1' at 0x005, 'A' at 0x0050, etc.
func (m *Memory) LoadFontSet() {
	fontSet := []byte{
		0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
		0x20, 0x60, 0x20, 0x20, 0x70, // 1
		0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
		0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
		0x90, 0x90, 0xF0, 0x10, 0x10, // 4
		0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
		0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
		0xF0, 0x10, 0x20, 0x40, 0x40, // 7
		0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
		0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
		0xF0, 0x90, 0xF0, 0x90, 0x90, // A
		0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
		0xF0, 0x80, 0x80, 0x80, 0xF0, // C
		0xE0, 0x90, 0x90, 0x90, 0xE0, // D
		0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
		0xF0, 0x80, 0xF0, 0x80, 0x80, // F
	}

	// Copy the 80 bytes of font data into the start of RAM
	copy(m.RAM[0:80], fontSet)
}
