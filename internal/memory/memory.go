package memory

// Memory represents the 4096 bytes of RAM in a Chip-8 system.
type Memory struct {
	// RAM is the physical storage.
	// 0x000-0x1FF: Reserved for the font set and system.
	// 0x200-0xFFF: Program ROM and work RAM.
	RAM [4096]byte
}

// New creates a blank 4KB memory bank.
func New() *Memory {
	return &Memory{}
}

// Reset clears all memory to zero.
// Note: You will usually call LoadFontSet() immediately after a Reset.
func (m *Memory) Reset() {
	m.RAM = [4096]byte{}
}

// Read returns the byte at the given address.
func (m *Memory) Read(address uint16) byte {
	return m.RAM[address]
}

// Write sets the byte at the given address.
func (m *Memory) Write(address uint16, value byte) {
	m.RAM[address] = value
}
