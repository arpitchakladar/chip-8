package encoder

/**
 * CHIP-8 Instruction Formats:
 * nnn or addr - A 12-bit value, the lowest 12 bits of the instruction
 * n           - A 4-bit value, the lowest 4 bits of the instruction
 * x           - A 4-bit value, the lower 4 bits of the high byte
 * y           - A 4-bit value, the upper 4 bits of the low byte
 * kk or byte  - An 8-bit value, the lowest 8 bits of the instruction
 */

// OpAddr handles instructions with a 12-bit address (nnn).
// Used by: JP (1nnn), CALL (2nnn), LD I (Annn), JP V0 (Bnnn)
func OpAddr(prefix uint16, addr uint16) uint16 {
	return prefix | (addr & 0x0FFF)
}

// OpRegImm handles instructions with a register and an 8-bit immediate (xkk).
// Used by: SE (3xkk), SNE (4xkk), LD (6xkk), ADD (7xkk)
func OpRegImm(prefix uint16, vx uint8, byte uint8) uint16 {
	return prefix | (uint16(vx&0xF) << 8) | uint16(byte)
}

// OpRegReg handles instructions with two registers (xy).
// Used by: SE (5xy0), LD (8xy0), OR (8xy1), AND (8xy2), XOR (8xy3),
// ADD (8xy4), SUB (8xy5), SHR (8xy6), SUBN (8xy7), SHL (8xyE), SNE (9xy0)
func OpRegReg(prefix uint16, vx uint8, vy uint8, suffix uint16) uint16 {
	return prefix | (uint16(vx&0xF) << 8) | (uint16(vy&0xF) << 4) | (suffix & 0xF)
}

// OpRegNibble handles instructions with two registers and a 4-bit nibble (xyn).
// Used by: DRW (Dxyn)
func OpRegNibble(prefix uint16, vx uint8, vy uint8, n uint8) uint16 {
	return prefix | (uint16(vx&0xF) << 8) | (uint16(vy&0xF) << 4) | uint16(n&0xF)
}

// OpRegOnly handles instructions that only specify one register (x).
// Used by: SKP (Ex9E), SKNP (ExA1), LD Vx, DT (Fx07), LD Vx, K (Fx0A), etc.
func OpRegOnly(prefix uint16, vx uint8, suffix uint16) uint16 {
	return prefix | (uint16(vx&0xF) << 8) | (suffix & 0xFF)
}

// OpRaw returns instructions that have no variables.
// Used by: CLS (00E0), RET (00EE)
func OpRaw(opcode uint16) uint16 {
	return opcode
}
