package cpu

import (
	"fmt"
	"math/rand"
)

// Execute decodes and performs the operation specified by the 16-bit opcode.
// It returns an error if the opcode is unknown.
func (c *CentralProcessingUnit) Execute(opcode uint16) error {
	// 0x F  X  Y  N
	x := (opcode & 0x0F00) >> 8   // Register index
	y := (opcode & 0x00F0) >> 4   // Register index
	nnn := opcode & 0x0FFF        // 12-bit address
	kk := byte(opcode & 0x00FF)   // 8-bit constant
	n := byte(opcode & 0x000F)    // 4-bit constant

	switch opcode & 0xF000 {
	case 0x0000:
		switch opcode {
		case 0x00E0: // CLS: Clear Display
			// Handled by System/Display
		case 0x00EE: // RET: Return from subroutine
			c.StackPointer--
			c.ProgramCounter = c.Stack[c.StackPointer]
		}

	case 0x1000: // 1NNN: JP addr
		c.ProgramCounter = nnn

	case 0x2000: // 2NNN: CALL addr
		c.Stack[c.StackPointer] = c.ProgramCounter
		c.StackPointer++
		c.ProgramCounter = nnn

	case 0x3000: // 3XKK: SE Vx, byte (Skip if Equal)
		if c.Registers[x] == kk {
			c.ProgramCounter += 2
		}

	case 0x4000: // 4XKK: SNE Vx, byte (Skip if Not Equal)
		if c.Registers[x] != kk {
			c.ProgramCounter += 2
		}

	case 0x5000: // 5XY0: SE Vx, Vy (Skip if Vx == Vy)
		if c.Registers[x] == c.Registers[y] {
			c.ProgramCounter += 2
		}

	case 0x6000: // 6XKK: LD Vx, byte
		c.Registers[x] = kk

	case 0x7000: // 7XKK: ADD Vx, byte
		c.Registers[x] += kk

	case 0x8000: // Arithmetic Group
		switch n {
		case 0x0: c.Registers[x] = c.Registers[y]
		case 0x1: c.Registers[x] |= c.Registers[y]
		case 0x2: c.Registers[x] &= c.Registers[y]
		case 0x3: c.Registers[x] ^= c.Registers[y]
		case 0x4: // ADD Vx, Vy (With Carry)
			sum := uint16(c.Registers[x]) + uint16(c.Registers[y])
			c.Registers[0xF] = 0
			if sum > 255 { c.Registers[0xF] = 1 }
			c.Registers[x] = byte(sum & 0xFF)
		case 0x5: // SUB Vx, Vy
			c.Registers[0xF] = 1
			if c.Registers[x] < c.Registers[y] { c.Registers[0xF] = 0 }
			c.Registers[x] -= c.Registers[y]
		case 0x6: // SHR Vx (Shift Right)
			c.Registers[0xF] = c.Registers[x] & 0x1
			c.Registers[x] >>= 1
		case 0x7: // SUBN Vx, Vy
			c.Registers[0xF] = 1
			if c.Registers[y] < c.Registers[x] { c.Registers[0xF] = 0 }
			c.Registers[x] = c.Registers[y] - c.Registers[x]
		case 0xE: // SHL Vx (Shift Left)
			c.Registers[0xF] = (c.Registers[x] & 0x80) >> 7
			c.Registers[x] <<= 1
		}

	case 0x9000: // 9XY0: SNE Vx, Vy
		if c.Registers[x] != c.Registers[y] {
			c.ProgramCounter += 2
		}

	case 0xA000: // ANNN: LD I, addr
		c.IndexRegister = nnn

	case 0xB000: // BNNN: JP V0, addr
		c.ProgramCounter = nnn + uint16(c.Registers[0])

	case 0xC000: // CXKK: RND Vx, byte
		c.Registers[x] = byte(rand.Intn(256)) & kk

	case 0xD000: // DXYN: DRW Vx, Vy, nibble
		// Draw logic triggers here

	case 0xE000: // Keyboard Inputs
		switch kk {
		case 0x9E: // SKP Vx
		case 0xA1: // SKNP Vx
		}

	case 0xF000: // Timers and Memory
		switch kk {
		case 0x07: c.Registers[x] = c.DelayTimer
		case 0x15: c.DelayTimer = c.Registers[x]
		case 0x18: c.SoundTimer = c.Registers[x]
		case 0x1E: c.IndexRegister += uint16(c.Registers[x])
		case 0x29: // LD F, Vx (Set I to font location)
			c.IndexRegister = uint16(c.Registers[x]) * 5
		case 0x33: // BCD (Binary Coded Decimal)
			//TODO: We'll need a Memory.Write helper for these:
			// val := c.Registers[x]
			// Mem[I] = val / 100, etc.
		case 0x55: // LD [I], Vx (Store registers V0..Vx)
		case 0x65: // LD Vx, [I] (Read registers V0..Vx)
		}

	default:
		return fmt.Errorf("unknown opcode: %04X", opcode)
	}

	return nil
}
