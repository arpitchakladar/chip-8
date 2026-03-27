package memory

const Size = 4096

// Memory represents the 4096 bytes of RAM in a Chip-8 system.
type Memory struct {
	// RAM is the physical storage.
	// 0x000-0x1FF: Reserved for the font set and system.
	// 0x200-0xFFF: Program ROM and work RAM.
	RAM [Size]byte
}

// New creates a blank 4KB memory bank.
func New() *Memory {
	return new(Memory)
}

// Reset clears all memory to zero.
// Note: You will usually call LoadFontSet() immediately after a Reset.
func (m *Memory) Reset() {
	m.RAM = [Size]byte{}
}

// Read returns the byte at the given address.
func (m *Memory) Read(address uint16) (byte, error) {
	if address >= Size {
		return 0, &BoundsError{Address: address, Max: 4095}
	}
	return m.RAM[address], nil
}

// Write sets the byte at the given address.
func (m *Memory) Write(address uint16, value byte) error {
	// 1. Check physical bounds
	if address >= Size {
		return &BoundsError{Address: address, Max: 4095}
	}

	// 2. Check for "Protected" area (Optional but recommended for debugging)
	// The first 512 bytes are where the Font Set lives.
	// A standard ROM should never overwrite this.
	if address < 0x200 {
		return &WriteProtectedError{Address: address}
	}

	m.RAM[address] = value
	return nil
}

// LoadFontSet populates the first 80 bytes of memory with the standard
// 4x5 pixel characters (0-F).
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

	// copy(destination, source)
	// This writes the 80 bytes of font data into the start of RAM.
	copy(m.RAM[0:80], fontSet)
}
