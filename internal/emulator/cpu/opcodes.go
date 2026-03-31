package cpu

import (
	"math/rand"

	"github.com/arpitchakladar/chip-8/internal/emulator/display"
	"github.com/arpitchakladar/chip-8/internal/emulator/keyboard"
	"github.com/arpitchakladar/chip-8/internal/emulator/memory"
)

// Execute decodes and performs the operation specified by the 16-bit opcode.
// It returns an error if the opcode is unknown.
func (c *CentralProcessingUnit) Execute(opcode uint16, mem *memory.Memory, disp *display.Display, keyb *keyboard.Keyboard) error {
	// 0x F  X  Y  N
	x := (opcode & 0x0F00) >> 8 // Register index
	y := (opcode & 0x00F0) >> 4 // Register index
	nnn := opcode & 0x0FFF      // 12-bit address
	kk := byte(opcode & 0x00FF) // 8-bit constant
	n := byte(opcode & 0x000F)  // 4-bit constant

	switch opcode & 0xF000 {
	case 0x0000:
		switch opcode {
		case 0x00E0: // CLS: Clear Display
			disp.Clear()
		case 0x00EE: // RET: Return from subroutine
			if c.StackPointer == 0 {
				return &StackError{
					IsOverflow:     false,
					ProgramCounter: c.ProgramCounter - 2,
				}
			}
			c.StackPointer--
			c.ProgramCounter = c.Stack[c.StackPointer]
		default:
			return &InvalidOpcodeError{Opcode: opcode, ProgramCounter: c.ProgramCounter - 2}
		}

	case 0x1000: // 1NNN: JP addr
		c.ProgramCounter = nnn

	case 0x2000: // 2NNN: CALL addr
		if c.StackPointer >= 16 {
			return &StackError{IsOverflow: true, ProgramCounter: c.ProgramCounter - 2}
		}
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
		case 0x0:
			c.Registers[x] = c.Registers[y]
		case 0x1:
			c.Registers[x] |= c.Registers[y]
		case 0x2:
			c.Registers[x] &= c.Registers[y]
		case 0x3:
			c.Registers[x] ^= c.Registers[y]
		case 0x4: // ADD Vx, Vy (With Carry)
			sum := uint16(c.Registers[x]) + uint16(c.Registers[y])
			c.Registers[0xF] = 0
			if sum > 255 {
				c.Registers[0xF] = 1
			}
			c.Registers[x] = byte(sum & 0xFF)
		case 0x5: // SUB Vx, Vy
			c.Registers[0xF] = 1
			if c.Registers[x] < c.Registers[y] {
				c.Registers[0xF] = 0
			}
			c.Registers[x] -= c.Registers[y]
		case 0x6: // SHR Vx (Shift Right)
			c.Registers[0xF] = c.Registers[x] & 0x1
			c.Registers[x] >>= 1
		case 0x7: // SUBN Vx, Vy
			c.Registers[0xF] = 1
			if c.Registers[y] < c.Registers[x] {
				c.Registers[0xF] = 0
			}
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
		c.Registers[0xF] = 0 // Reset collision flag
		for row := range uint16(n) {
			spriteByte, err := mem.Read(c.IndexRegister + row)
			if err != nil {
				// Wrap the memory error into a CPU context
				return &MemorySyncError{Opcode: opcode, ProgramCounter: c.ProgramCounter - 2, Child: err}
			}
			for col := range uint16(8) {
				// Check if the specific bit in the sprite byte is 1
				if (spriteByte & (0x80 >> col)) != 0 {
					posX := (c.Registers[x] + uint8(col)) % 64
					posY := (c.Registers[y] + uint8(row)) % 32
					// TODO: Handle the non-blocking error here
					collision, _ := disp.SetPixel(posX, posY)
					if collision {
						c.Registers[0xF] = 1
					}
				}
			}
		}

	case 0xE000: // Keyboard Inputs
		switch kk {
		case 0x9E: // SKP Vx: Skip next instruction if key with the value of Vx is pressed
			if keyb.IsKeyPressed(c.Registers[x]) {
				c.ProgramCounter += 2
			}
		case 0xA1: // SKNP Vx: Skip next instruction if key with the value of Vx is NOT pressed
			if !keyb.IsKeyPressed(c.Registers[x]) {
				c.ProgramCounter += 2
			}
		}

	case 0xF000:
		switch kk {
		case 0x07: // LD Vx, DT: Set Vx = delay timer value
			c.Registers[x] = c.DelayTimer

		case 0x0A: // LD Vx, K: Wait for a key press, store the value of the key in Vx
			// Thisis a "blocking" opcode. Usually implemented by
			// decrementing PC by 2 if no key is pressed, effectively
			// pausing the CPU on this instruction.
			if key, pressed := keyb.AnyKeyPressed(); pressed {
				c.Registers[x] = key
			} else {
				// No key pressed? Repeat this instruction on the next cycle.
				// We subtract 2 because the Step() function already incremented it.
				c.ProgramCounter -= 2
			}

		case 0x15: // LD DT, Vx: Set delay timer = Vx
			c.DelayTimer = c.Registers[x]

		case 0x18: // LD ST, Vx: Set sound timer = Vx
			c.SoundTimer = c.Registers[x]

		case 0x1E: // ADD I, Vx: Set I = I + Vx
			c.IndexRegister += uint16(c.Registers[x])

		case 0x29: // LD F, Vx: Set I = location of sprite for digit Vx
			// Characters 0-F are 5 bytes high.
			// Since we load the font at 0x000, the address is Vx * 5.
			c.IndexRegister = uint16(c.Registers[x]) * 5

		case 0x33: // BCD: Store BCD representation of Vx in memory locations I, I+1, and I+2
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

		case 0x55: // LD [I], Vx: Store registers V0 through Vx in memory starting at location I
			for i := range x + 1 {
				if err := mem.Write(c.IndexRegister+uint16(i), c.Registers[i]); err != nil {
					return &MemorySyncError{Opcode: opcode, ProgramCounter: c.ProgramCounter - 2, Child: err}
				}
			}

		case 0x65: // LD Vx, [I]: Read registers V0 through Vx from memory starting at location I
			for i := range x + 1 {
				val, err := mem.Read(c.IndexRegister + uint16(i))
				if err != nil {
					return &MemorySyncError{Opcode: opcode, ProgramCounter: c.ProgramCounter - 2, Child: err}
				}
				c.Registers[i] = val
			}
		}

	default:
		return &InvalidOpcodeError{Opcode: opcode, ProgramCounter: c.ProgramCounter - 2}
	}

	return nil
}
