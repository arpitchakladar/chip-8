package cpu

import "fmt"

// Execute decodes and performs the operation specified by the opcode.
func (c *CentralProcessingUnit) Execute(opcode uint16) error {
	// Extract common variables from the 16-bit opcode (0x F X Y N)
	x := (opcode & 0x0F00) >> 8      // The second nibble: used for Register index
	y := (opcode & 0x00F0) >> 4      // The third nibble: used for Register index
	nnn := opcode & 0x0FFF           // The last 12 bits: used for Memory Addresses
	kk := byte(opcode & 0x00FF)      // The last 8 bits: used for Constants
	n := byte(opcode & 0x000F)       // The last 4 bits: used for 4-bit Constants

	switch opcode & 0xF000 {
	case 0x0000:
		switch opcode {
		case 0x00E0: // CLS: Clear the screen
			// We will call display.Clear() from the System struct later
		case 0x00EE: // RET: Return from subroutine
			c.StackPointer--
			c.ProgramCounter = c.Stack[c.StackPointer]
		}

	case 0x1000: // 1NNN: JP address (Jump)
		c.ProgramCounter = nnn

	case 0x6000: // 6XKK: LD Vx, byte (Set Register Vx to KK)
		c.Registers[x] = kk

	case 0x7000: // 7XKK: ADD Vx, byte (Add KK to Vx, no carry)
		c.Registers[x] += kk

	case 0xA000: // ANNN: LD I, address (Set Index Register to NNN)
		c.IndexRegister = nnn

	case 0xD000: // DXYN: DRW Vx, Vy, nibble (Draw sprite)
		// This is where the magic happens.
		// It uses Registers[x], Registers[y], and n.
		// TODO: Replace with actual drawing
		fmt.Printf("Drawing at x: %d, y: %d, height: %d\n", c.Registers[x], c.Registers[y], n)

	default:
		// TODO: Replace with error handling system
		return fmt.Errorf("unsupported opcode: %04X", opcode)
	}

	return nil
}
