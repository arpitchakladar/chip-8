package encoder

// Encoder builds CHIP-8 opcodes from instruction components.
// It provides methods for encoding different instruction formats using bit masks.

// Encoder builds CHIP-8 opcodes from instruction components.
// It provides methods for encoding different instruction formats using bit masks.

// Instruction represents a 16-bit CHIP-8 opcode.
type Instruction uint16

// Encoder builds CHIP-8 opcodes from their component parts.
type Encoder struct{}

// New creates a new instance of the Encoder.
func New() *Encoder {
	return new(Encoder)
}

// Instruction Masks (Constants associated with the Encoder logic)
const (
	MaskCLS  uint16 = 0x00E0
	MaskRET  uint16 = 0x00EE
	MaskJP   uint16 = 0x1000
	MaskCALL uint16 = 0x2000
	MaskSE   uint16 = 0x3000
	MaskSNE  uint16 = 0x4000
	MaskSER  uint16 = 0x5000
	MaskLD   uint16 = 0x6000
	MaskADD  uint16 = 0x7000
	MaskALU  uint16 = 0x8000
	MaskSNER uint16 = 0x9000
	MaskLDI  uint16 = 0xA000
	MaskJPV0 uint16 = 0xB000
	MaskRND  uint16 = 0xC000
	MaskDRW  uint16 = 0xD000
	MaskKEY  uint16 = 0xE000
	MaskMISC uint16 = 0xF000
)

// Addr handles instructions with a 12-bit address (nnn).
// Used by: JP (1nnn), CALL (2nnn), LD I (Annn), JP V0 (Bnnn)
func (e *Encoder) Addr(prefix uint16, addr uint16) uint16 {
	return prefix | (addr & 0x0FFF)
}

// RegImm handles instructions with a register and an 8-bit immediate (xkk).
// Used by: SE (3xkk), SNE (4xkk), LD (6xkk), ADD (7xkk)
func (e *Encoder) RegImm(prefix uint16, vx uint8, byte uint8) uint16 {
	return prefix | (uint16(vx&0xF) << 8) | uint16(byte)
}

// RegReg handles instructions with two registers (xy).
// Used by: SE (5xy0), LD (8xy0), ALU operations (8xy1-E), SNE (9xy0)
func (e *Encoder) RegReg(prefix uint16, vx uint8, vy uint8, suffix uint16) uint16 {
	return prefix | (uint16(vx&0xF) << 8) | (uint16(vy&0xF) << 4) | (suffix & 0xF)
}

// RegNibble handles instructions with two registers and a 4-bit nibble (xyn).
// Used by: DRW (Dxyn)
func (e *Encoder) RegNibble(prefix uint16, vx uint8, vy uint8, n uint8) uint16 {
	return prefix | (uint16(vx&0xF) << 8) | (uint16(vy&0xF) << 4) | uint16(n&0xF)
}

// RegOnly handles instructions that only specify one register (x).
// Used by: SKP (Ex9E), SKNP (ExA1), etc.
func (e *Encoder) RegOnly(prefix uint16, vx uint8, suffix uint16) uint16 {
	return prefix | (uint16(vx&0xF) << 8) | (suffix & 0xFF)
}

// Raw returns instructions that have no variables.
// Used by: CLS (00E0), RET (00EE)
func (e *Encoder) Raw(opcode uint16) uint16 {
	return opcode
}
