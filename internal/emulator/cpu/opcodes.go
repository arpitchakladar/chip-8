package cpu

import (
	"math/rand"

	"github.com/arpitchakladar/chip-8/internal/emulator/display"
	"github.com/arpitchakladar/chip-8/internal/emulator/keyboard"
	"github.com/arpitchakladar/chip-8/internal/emulator/memory"
)

// Execute decodes and performs the operation specified by the 16-bit opcode.
// It handles all CHIP-8 instructions by decoding the opcode into its component parts
// and executing the appropriate logic.
//
// Opcode format (16 bits): F | X | Y | N
//   - F: 4-bit function code (for 8xyN instructions)
//   - X: 4-bit register index (Vy reference)
//   - Y: 4-bit register index (Vy reference)
//   - N: 4-bit constant (or nibble)
//
// Additional decoded values:
//   - nnn: 12-bit address (lower 12 bits of opcode)
//   - kk: 8-bit constant (lower 8 bits of opcode)
//
// Opcode groups:
//   - 0x0000: System and subroutine operations (CLS, RET)
//   - 0x1000: JP - Jump to address
//   - 0x2000: CALL - Call subroutine
//   - 0x3000: SE Vx, byte - Skip if equal
//   - 0x4000: SNE Vx, byte - Skip if not equal
//   - 0x5000: SE Vx, Vy - Skip if registers equal
//   - 0x6000: LD Vx, byte - Load constant
//   - 0x7000: ADD Vx, byte - Add constant
//   - 0x8000: ALU operations (OR, AND, XOR, SUB, SHR, SUBN, SHL)
//   - 0x9000: SNE Vx, Vy - Skip if registers not equal
//   - 0xA000: LD I, addr - Load address into I
//   - 0xB000: JP V0, addr - Jump with offset
//   - 0xC000: RND - Random number
//   - 0xD000: DRW - Draw sprite
//   - 0xE000: Keyboard operations (SKP, SKNP)
//   - 0xF000: Miscellaneous (timers, memory, BCD)
//
// Returns:
//   - nil: on successful execution
//   - *InvalidOpcodeError: if the opcode is not recognized
//   - *StackError: if stack overflow/underflow occurs
//   - *MemorySyncError: if memory read/write fails
func (c *CentralProcessingUnit) Execute(opcode uint16, mem *memory.Memory, disp *display.Display, keyb *keyboard.Keyboard) error {
	// Decode opcode into component parts
	// Format: 0x FXXX YYYN -> nibble, X, Y, N
	x := (opcode & 0x0F00) >> 8 // Register index (Vx)
	y := (opcode & 0x00F0) >> 4 // Register index (Vy)
	nnn := opcode & 0x0FFF      // 12-bit address constant
	kk := byte(opcode & 0x00FF) // 8-bit constant
	n := byte(opcode & 0x000F)  // 4-bit constant

	// Dispatch to the appropriate opcode handler
	switch opcode & 0xF000 {
	case 0x0000:
		switch opcode {
		case 0x00E0: // CLS: Clear Display
			// Clears the display by resetting all pixels to 0
			disp.Clear()
		case 0x00EE: // RET: Return from subroutine
			// Pop the return address from stack and jump to it
			if c.StackPointer == 0 {
				return &StackError{
					IsOverflow:     false,
					ProgramCounter: c.ProgramCounter - 2,
				}
			}
			c.StackPointer--
			c.ProgramCounter = c.Stack[c.StackPointer]
		default:
			// Unknown 0xxx opcode
			return &InvalidOpcodeError{Opcode: opcode, ProgramCounter: c.ProgramCounter - 2}
		}

	case 0x1000: // 1NNN: JP addr
		// Jump to the 12-bit address nnn
		c.ProgramCounter = nnn

	case 0x2000: // 2NNN: CALL addr
		// Push current PC onto stack and jump to address
		if c.StackPointer >= 16 {
			return &StackError{IsOverflow: true, ProgramCounter: c.ProgramCounter - 2}
		}
		c.Stack[c.StackPointer] = c.ProgramCounter
		c.StackPointer++
		c.ProgramCounter = nnn

	case 0x3000: // 3XKK: SE Vx, byte
		// Skip next instruction if Vx == kk
		if c.Registers[x] == kk {
			c.ProgramCounter += 2
		}

	case 0x4000: // 4XKK: SNE Vx, byte
		// Skip next instruction if Vx != kk
		if c.Registers[x] != kk {
			c.ProgramCounter += 2
		}

	case 0x5000: // 5XY0: SE Vx, Vy
		// Skip next instruction if Vx == Vy
		if c.Registers[x] == c.Registers[y] {
			c.ProgramCounter += 2
		}

	case 0x6000: // 6XKK: LD Vx, byte
		// Load constant kk into register Vx
		c.Registers[x] = kk

	case 0x7000: // 7XKK: ADD Vx, byte
		// Add constant kk to register Vx (no carry flag)
		c.Registers[x] += kk

	case 0x8000: // ALU Operations (8xyN)
		switch n {
		case 0x0: // LD Vx, Vy
			c.Registers[x] = c.Registers[y]
		case 0x1: // OR Vx, Vy
			c.Registers[x] |= c.Registers[y]
		case 0x2: // AND Vx, Vy
			c.Registers[x] &= c.Registers[y]
		case 0x3: // XOR Vx, Vy
			c.Registers[x] ^= c.Registers[y]
		case 0x4: // ADD Vx, Vy (with carry)
			sum := uint16(c.Registers[x]) + uint16(c.Registers[y])
			c.Registers[0xF] = 0
			if sum > 255 {
				c.Registers[0xF] = 1 // Set carry flag
			}
			c.Registers[x] = byte(sum & 0xFF)
		case 0x5: // SUB Vx, Vy
			c.Registers[0xF] = 1
			if c.Registers[x] < c.Registers[y] {
				c.Registers[0xF] = 0 // Clear borrow flag
			}
			c.Registers[x] -= c.Registers[y]
		case 0x6: // SHR Vx (Shift Right)
			// VF = LSB before shift
			c.Registers[0xF] = c.Registers[x] & 0x1
			c.Registers[x] >>= 1
		case 0x7: // SUBN Vx, Vy
			c.Registers[0xF] = 1
			if c.Registers[y] < c.Registers[x] {
				c.Registers[0xF] = 0 // Clear borrow flag
			}
			c.Registers[x] = c.Registers[y] - c.Registers[x]
		case 0xE: // SHL Vx (Shift Left)
			// VF = MSB before shift
			c.Registers[0xF] = (c.Registers[x] & 0x80) >> 7
			c.Registers[x] <<= 1
		}

	case 0x9000: // 9XY0: SNE Vx, Vy
		// Skip next instruction if Vx != Vy
		if c.Registers[x] != c.Registers[y] {
			c.ProgramCounter += 2
		}

	case 0xA000: // ANNN: LD I, addr
		// Load 12-bit address into index register I
		c.IndexRegister = nnn

	case 0xB000: // BNNN: JP V0, addr
		// Jump to address + V0 (for legacy compatibility)
		c.ProgramCounter = nnn + uint16(c.Registers[0])

	case 0xC000: // CXKK: RND Vx, byte
		// Generate random number and AND with kk
		c.Registers[x] = byte(rand.Intn(256)) & kk

	case 0xD000: // DXYN: DRW Vx, Vy, nibble
		// Draw N-byte sprite at (Vx, Vy) using XOR mode
		// Sprite data is read from memory starting at address I
		c.Registers[0xF] = 0 // Reset collision flag
		for row := range uint16(n) {
			spriteByte, err := mem.Read(c.IndexRegister + row)
			if err != nil {
				return &MemorySyncError{Opcode: opcode, ProgramCounter: c.ProgramCounter - 2, Child: err}
			}
			for col := range uint16(8) {
				// Check if bit is set in sprite
				if (spriteByte & (0x80 >> col)) != 0 {
					// Wrap coordinates (standard Chip-8 behavior)
					posX := (c.Registers[x] + uint8(col)) % 64
					posY := (c.Registers[y] + uint8(row)) % 32
					collision, _ := disp.SetPixel(posX, posY)
					if collision {
						c.Registers[0xF] = 1 // Set collision flag
					}
				}
			}
		}

	case 0xE000: // Keyboard Operations
		switch kk {
		case 0x9E: // SKP Vx
			// Skip next instruction if key Vx is pressed
			if keyb.IsKeyPressed(c.Registers[x]) {
				c.ProgramCounter += 2
			}
		case 0xA1: // SKNP Vx
			// Skip next instruction if key Vx is NOT pressed
			if !keyb.IsKeyPressed(c.Registers[x]) {
				c.ProgramCounter += 2
			}
		}

	case 0xF000: // Miscellaneous Operations
		switch kk {
		case 0x07: // LD Vx, DT
			// Load delay timer value into Vx
			c.Registers[x] = c.DelayTimer

		case 0x0A: // LD Vx, K
			// Wait for key press (blocking opcode)
			// If no key is pressed, decrement PC to re-execute this instruction
			if key, pressed := keyb.AnyKeyPressed(); pressed {
				c.Registers[x] = key
			} else {
				c.ProgramCounter -= 2
			}

		case 0x15: // LD DT, Vx
			// Set delay timer from Vx
			c.DelayTimer = c.Registers[x]

		case 0x18: // LD ST, Vx
			// Set sound timer from Vx (triggers beep when > 0)
			c.SoundTimer = c.Registers[x]

		case 0x1E: // ADD I, Vx
			// Add Vx to index register I
			c.IndexRegister += uint16(c.Registers[x])

		case 0x29: // LD F, Vx
			// Set I to the font sprite address for digit Vx
			// Font characters 0-F are 5 bytes each, stored at 0x000-0x050
			c.IndexRegister = uint16(c.Registers[x]) * 5

		case 0x33: // BCD: Vx
			// Store BCD representation of Vx at I, I+1, I+2
			val := c.Registers[x]
			if err := mem.Write(c.IndexRegister, val/100); err != nil {
				return &MemorySyncError{Opcode: opcode, ProgramCounter: c.ProgramCounter - 2, Child: err}
			}
			if err := mem.Write(c.IndexRegister+1, (val/10)%10); err != nil {
				return &MemorySyncError{Opcode: opcode, ProgramCounter: c.ProgramCounter - 2, Child: err}
			}
			if err := mem.Write(c.IndexRegister+2, val%10); err != nil {
				return &MemorySyncError{Opcode: opcode, ProgramCounter: c.ProgramCounter - 2, Child: err}
			}

		case 0x55: // LD [I], Vx
			// Store registers V0 through Vx in memory starting at I
			for i := range x + 1 {
				if err := mem.Write(c.IndexRegister+uint16(i), c.Registers[i]); err != nil {
					return &MemorySyncError{Opcode: opcode, ProgramCounter: c.ProgramCounter - 2, Child: err}
				}
			}

		case 0x65: // LD Vx, [I]
			// Load registers V0 through Vx from memory starting at I
			for i := range x + 1 {
				val, err := mem.Read(c.IndexRegister + uint16(i))
				if err != nil {
					return &MemorySyncError{Opcode: opcode, ProgramCounter: c.ProgramCounter - 2, Child: err}
				}
				c.Registers[i] = val
			}
		}

	default:
		// Unrecognized opcode
		return &InvalidOpcodeError{Opcode: opcode, ProgramCounter: c.ProgramCounter - 2}
	}

	return nil
}
