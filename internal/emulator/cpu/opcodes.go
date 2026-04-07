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
type opcodeHandler func(*CPU, uint16, *memory.Memory, display.Display, keyboard.Keyboard, uint8, uint8, uint16, byte, byte) error

var opcodeHandlers = map[uint16]opcodeHandler{
	0x0000: handleGroup0,
	0x1000: handleGroup1,
	0x2000: handleGroup2,
	0x3000: handleGroup3,
	0x4000: handleGroup4,
	0x5000: handleGroup5,
	0x6000: handleGroup6,
	0x7000: handleGroup7,
	0x8000: handleGroup8,
	0x9000: handleGroup9,
	0xA000: handleGroupA,
	0xB000: handleGroupB,
	0xC000: handleGroupC,
	0xD000: handleGroupD,
	0xE000: handleGroupE,
	0xF000: handleGroupF,
}

func (c *CPU) Execute(
	opcode uint16,
	mem *memory.Memory,
	disp display.Display,
	keyb keyboard.Keyboard,
) error {
	x := uint8((opcode & 0x0F00) >> 8)
	y := uint8((opcode & 0x00F0) >> 4)
	nnn := opcode & 0x0FFF
	kk := byte(opcode & 0x00FF)
	n := byte(opcode & 0x000F)

	group := opcode & 0xF000
	handler, ok := opcodeHandlers[group]
	if !ok {
		return c.handleInvalidOpcode(opcode)
	}
	return handler(c, opcode, mem, disp, keyb, x, y, nnn, kk, n)
}

func handleGroup0(
	c *CPU,
	opcode uint16,
	_ *memory.Memory,
	disp display.Display,
	_ keyboard.Keyboard,
	_, _ uint8,
	_ uint16,
	_ byte,
	_ byte,
) error {
	return c.handle0xxx(opcode, disp)
}

func handleGroup1(
	c *CPU,
	_ uint16,
	_ *memory.Memory,
	_ display.Display,
	_ keyboard.Keyboard,
	_, _ uint8,
	nnn uint16,
	_ byte,
	_ byte,
) error {
	c.ProgramCounter = nnn
	return nil
}

func handleGroup2(
	c *CPU,
	_ uint16,
	_ *memory.Memory,
	_ display.Display,
	_ keyboard.Keyboard,
	_, _ uint8,
	nnn uint16,
	_ byte,
	_ byte,
) error {
	return c.handle2xxx(nnn)
}

func handleGroup3(
	c *CPU,
	_ uint16,
	_ *memory.Memory,
	_ display.Display,
	_ keyboard.Keyboard,
	x uint8,
	_ uint8,
	_ uint16,
	kk byte,
	_ byte,
) error {
	if c.Registers[x] == kk {
		c.ProgramCounter += 2
	}
	return nil
}

func handleGroup4(
	c *CPU,
	_ uint16,
	_ *memory.Memory,
	_ display.Display,
	_ keyboard.Keyboard,
	x uint8,
	_ uint8,
	_ uint16,
	kk byte,
	_ byte,
) error {
	if c.Registers[x] != kk {
		c.ProgramCounter += 2
	}
	return nil
}

func handleGroup5(
	c *CPU,
	_ uint16,
	_ *memory.Memory,
	_ display.Display,
	_ keyboard.Keyboard,
	x, y uint8,
	_ uint16,
	_ byte,
	_ byte,
) error {
	if c.Registers[x] == c.Registers[y] {
		c.ProgramCounter += 2
	}
	return nil
}

func handleGroup6(
	c *CPU,
	_ uint16,
	_ *memory.Memory,
	_ display.Display,
	_ keyboard.Keyboard,
	x uint8,
	_ uint8,
	_ uint16,
	kk byte,
	_ byte,
) error {
	c.Registers[x] = kk
	return nil
}

func handleGroup7(
	c *CPU,
	_ uint16,
	_ *memory.Memory,
	_ display.Display,
	_ keyboard.Keyboard,
	x uint8,
	_ uint8,
	_ uint16,
	kk byte,
	_ byte,
) error {
	c.Registers[x] += kk
	return nil
}

func handleGroup8(
	c *CPU,
	_ uint16,
	_ *memory.Memory,
	_ display.Display,
	_ keyboard.Keyboard,
	x, y uint8,
	_ uint16,
	_ byte,
	n byte,
) error {
	return c.handle8xxx(x, y, n)
}

func handleGroup9(
	c *CPU,
	_ uint16,
	_ *memory.Memory,
	_ display.Display,
	_ keyboard.Keyboard,
	x, y uint8,
	_ uint16,
	_ byte,
	_ byte,
) error {
	if c.Registers[x] != c.Registers[y] {
		c.ProgramCounter += 2
	}
	return nil
}

func handleGroupA(
	c *CPU,
	_ uint16,
	_ *memory.Memory,
	_ display.Display,
	_ keyboard.Keyboard,
	_, _ uint8,
	nnn uint16,
	_ byte,
	_ byte,
) error {
	c.IndexRegister = nnn
	return nil
}

func handleGroupB(
	c *CPU,
	_ uint16,
	_ *memory.Memory,
	_ display.Display,
	_ keyboard.Keyboard,
	_, _ uint8,
	nnn uint16,
	_ byte,
	_ byte,
) error {
	c.ProgramCounter = nnn + uint16(c.Registers[0])
	return nil
}

func handleGroupC(
	c *CPU,
	_ uint16,
	_ *memory.Memory,
	_ display.Display,
	_ keyboard.Keyboard,
	x uint8,
	_ uint8,
	_ uint16,
	kk byte,
	_ byte,
) error {
	c.Registers[x] = byte(rand.Intn(256)) & kk
	return nil
}

func handleGroupD(
	c *CPU,
	_ uint16,
	mem *memory.Memory,
	disp display.Display,
	_ keyboard.Keyboard,
	x, y uint8,
	_ uint16,
	_ byte,
	n byte,
) error {
	return c.handleDxxx(x, y, n, mem, disp)
}

func handleGroupE(
	c *CPU,
	_ uint16,
	_ *memory.Memory,
	_ display.Display,
	keyb keyboard.Keyboard,
	x uint8,
	_ uint8,
	_ uint16,
	kk byte,
	_ byte,
) error {
	return c.handleExxx(x, kk, keyb)
}

func handleGroupF(
	c *CPU,
	_ uint16,
	mem *memory.Memory,
	_ display.Display,
	keyb keyboard.Keyboard,
	x uint8,
	_ uint8,
	_ uint16,
	kk byte,
	_ byte,
) error {
	return c.handleFxxx(x, kk, mem, keyb)
}

func (c *CPU) handleInvalidOpcode(opcode uint16) error {
	return &InvalidOpcodeError{
		Opcode:         opcode,
		ProgramCounter: c.ProgramCounter - 2,
	}
}

func (c *CPU) handle0xxx(opcode uint16, disp display.Display) error {
	switch opcode {
	case 0x00E0:
		disp.Clear()
	case 0x00EE:
		if c.StackPointer == 0 {
			return &StackError{
				IsOverflow:     false,
				ProgramCounter: c.ProgramCounter - 2,
			}
		}
		c.StackPointer--
		c.ProgramCounter = c.Stack[c.StackPointer]
	default:
		return &InvalidOpcodeError{
			Opcode:         opcode,
			ProgramCounter: c.ProgramCounter - 2,
		}
	}
	return nil
}

func (c *CPU) handle2xxx(nnn uint16) error {
	if c.StackPointer >= 16 {
		return &StackError{
			IsOverflow:     true,
			ProgramCounter: c.ProgramCounter - 2,
		}
	}
	c.Stack[c.StackPointer] = c.ProgramCounter
	c.StackPointer++
	c.ProgramCounter = nnn
	return nil
}

func (c *CPU) handle8xxx(x, y uint8, n byte) error {
	switch n {
	case 0x0:
		c.Registers[x] = c.Registers[y]
	case 0x1:
		c.Registers[x] |= c.Registers[y]
	case 0x2:
		c.Registers[x] &= c.Registers[y]
	case 0x3:
		c.Registers[x] ^= c.Registers[y]
	case 0x4:
		return c.handleAddCarry(x, y)
	case 0x5:
		return c.handleSub(x, y, true)
	case 0x6:
		return c.handleShiftRight(x)
	case 0x7:
		return c.handleSubReverse(x, y)
	case 0xE:
		return c.handleShiftLeft(x)
	}
	return nil
}

func (c *CPU) handleAddCarry(x, y uint8) error {
	sum := uint16(c.Registers[x]) + uint16(c.Registers[y])
	c.Registers[0xF] = 0
	if sum > 255 {
		c.Registers[0xF] = 1
	}
	c.Registers[x] = byte(sum & 0xFF)
	return nil
}

func (c *CPU) handleSub(x, y uint8, normal bool) error {
	c.Registers[0xF] = 1
	if c.Registers[x] < c.Registers[y] {
		c.Registers[0xF] = 0
	}
	if normal {
		c.Registers[x] -= c.Registers[y]
	} else {
		c.Registers[x] = c.Registers[y] - c.Registers[x]
	}
	return nil
}

func (c *CPU) handleSubReverse(x, y uint8) error {
	c.Registers[0xF] = 1
	if c.Registers[y] < c.Registers[x] {
		c.Registers[0xF] = 0
	}
	c.Registers[x] = c.Registers[y] - c.Registers[x]
	return nil
}

func (c *CPU) handleShiftRight(x uint8) error {
	c.Registers[0xF] = c.Registers[x] & 0x1
	c.Registers[x] >>= 1
	return nil
}

func (c *CPU) handleShiftLeft(x uint8) error {
	c.Registers[0xF] = (c.Registers[x] & 0x80) >> 7
	c.Registers[x] <<= 1
	return nil
}

func (c *CPU) handleDxxx(
	x, y uint8,
	n byte,
	mem *memory.Memory,
	disp display.Display,
) error {
	c.Registers[0xF] = 0
	for row := range uint16(n) {
		spriteByte, err := mem.Read(c.IndexRegister + row)
		if err != nil {
			return &MemorySyncError{
				Opcode: 0xD000 | (uint16(x) << 8) | (uint16(y) << 4) | uint16(
					n,
				),
				ProgramCounter: c.ProgramCounter - 2,
				Child:          err,
			}
		}
		for col := range uint16(8) {
			if (spriteByte & (0x80 >> col)) != 0 {
				posX := (c.Registers[x] + uint8(col)) % 64
				posY := (c.Registers[y] + uint8(row)) % 32
				collision, _ := disp.SetPixel(posX, posY)
				if collision {
					c.Registers[0xF] = 1
				}
			}
		}
	}
	return nil
}

func (c *CPU) handleExxx(x uint8, kk byte, keyb keyboard.Keyboard) error {
	switch kk {
	case 0x9E:
		if keyb.IsKeyPressed(c.Registers[x]) {
			c.ProgramCounter += 2
		}
	case 0xA1:
		if !keyb.IsKeyPressed(c.Registers[x]) {
			c.ProgramCounter += 2
		}
	}
	return nil
}

func (c *CPU) handleFxxx(
	x uint8,
	kk byte,
	mem *memory.Memory,
	keyb keyboard.Keyboard,
) error {
	switch kk {
	case 0x07:
		c.Registers[x] = c.DelayTimer
	case 0x0A:
		return c.handleKeyWait(x, keyb)
	case 0x15:
		c.DelayTimer = c.Registers[x]
	case 0x18:
		c.SoundTimer = c.Registers[x]
	case 0x1E:
		c.IndexRegister += uint16(c.Registers[x])
	case 0x29:
		c.IndexRegister = uint16(c.Registers[x]) * 5
	case 0x33:
		return c.handleBCD(x, mem)
	case 0x55:
		return c.handleStoreRegs(x, mem)
	case 0x65:
		return c.handleLoadRegs(x, mem)
	}
	return nil
}

func (c *CPU) handleKeyWait(x uint8, keyb keyboard.Keyboard) error {
	if key, pressed := keyb.AnyKeyPressed(); pressed {
		c.Registers[x] = key
	} else {
		c.ProgramCounter -= 2
	}
	return nil
}

func (c *CPU) handleBCD(x uint8, mem *memory.Memory) error {
	val := c.Registers[x]
	opcode := 0x3000 | (uint16(x) << 8)
	if err := mem.Write(c.IndexRegister, val/100); err != nil {
		return &MemorySyncError{
			Opcode:         opcode,
			ProgramCounter: c.ProgramCounter - 2,
			Child:          err,
		}
	}
	if err := mem.Write(c.IndexRegister+1, (val/10)%10); err != nil {
		return &MemorySyncError{
			Opcode:         opcode,
			ProgramCounter: c.ProgramCounter - 2,
			Child:          err,
		}
	}
	if err := mem.Write(c.IndexRegister+2, val%10); err != nil {
		return &MemorySyncError{
			Opcode:         opcode,
			ProgramCounter: c.ProgramCounter - 2,
			Child:          err,
		}
	}
	return nil
}

func (c *CPU) handleStoreRegs(x uint8, mem *memory.Memory) error {
	opcode := 0xF000 | (uint16(x) << 8) | 0x55
	for i := uint8(0); i <= x; i++ {
		if err := mem.Write(c.IndexRegister+uint16(i), c.Registers[i]); err != nil {
			return &MemorySyncError{
				Opcode:         opcode,
				ProgramCounter: c.ProgramCounter - 2,
				Child:          err,
			}
		}
	}
	return nil
}

func (c *CPU) handleLoadRegs(x uint8, mem *memory.Memory) error {
	opcode := 0xF000 | (uint16(x) << 8) | 0x65
	for i := uint8(0); i <= x; i++ {
		val, err := mem.Read(c.IndexRegister + uint16(i))
		if err != nil {
			return &MemorySyncError{
				Opcode:         opcode,
				ProgramCounter: c.ProgramCounter - 2,
				Child:          err,
			}
		}
		c.Registers[i] = val
	}
	return nil
}
