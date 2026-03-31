# CHIP-8

A Chip-8 virtual machine and assembler written in Go with SDL2 for graphics and audio.

## Demo
![Demo](https://github.com/user-attachments/assets/c48a8e14-bb22-4dd9-ab96-d436de5da920)

This is a demo of the Snake game running from `games/snake.asm`.

## Prerequisites

- **Go** 1.25 or later
- **SDL2** development libraries

### Installing SDL2

**Ubuntu/Debian:**
```bash
sudo apt-get install libsdl2-dev
```

**macOS:**
```bash
brew install sdl2
```

**Fedora:**
```bash
sudo dnf install SDL2-devel
```

## Installation

1. Clone the repository:
```bash
git clone https://github.com/arpitchakladar/chip-8.git
cd chip-8
```

2. Download dependencies:
```bash
go mod download
```

3. Build the binary:
```bash
go build -o chip-8 ./cmd/chip-8
```

## Running the Emulator

To run a Chip-8 ROM (.ch8 file):
```bash
./chip-8 run <path-to-rom>
```

Example:
```bash
./chip-8 run roms/1dcell.ch8
```

## Using the Assembler

To assemble a .asm file into a .ch8 file:
```bash
./chip-8 compile <path-to-asm>
```

Example:
```bash
./chip-8 compile examples/hello_world.asm
```

This will output `examples/hello_world.ch8`.

## Assembly Language Syntax

### Registers
- `V0` through `VF` - General purpose registers (V0-VF, VF used for carry/borrow)
- `I` - Index register (points to memory locations)
- `DT` - Delay timer
- `ST` - Sound timer

### Instructions

| Opcode | Syntax | Description |
|--------|--------|-------------|
| 00E0 | `CLS` | Clear the screen |
| 00EE | `RET` | Return from subroutine |
| 0nnn | `SYS addr` | Jump to machine code (unused) |
| 1nnn | `JP addr` | Jump to address nnn |
| 2nnn | `CALL addr` | Call subroutine at nnn |
| 3xkk | `SE Vx, byte` | Skip next if Vx == byte |
| 4xkk | `SNE Vx, byte` | Skip next if Vx != byte |
| 5xy0 | `SE Vx, Vy` | Skip next if Vx == Vy |
| 6xkk | `LD Vx, byte` | Set Vx = byte |
| 7xkk | `ADD Vx, byte` | Set Vx = Vx + byte |
| 8xy0 | `LD Vx, Vy` | Set Vx = Vy |
| 8xy1 | `OR Vx, Vy` | Set Vx = Vx OR Vy |
| 8xy2 | `AND Vx, Vy` | Set Vx = Vx AND Vy |
| 8xy3 | `XOR Vx, Vy` | Set Vx = Vx XOR Vy |
| 8xy4 | `ADD Vx, Vy` | Set Vx = Vx + Vy, set VF = carry |
| 8xy5 | `SUB Vx, Vy` | Set Vx = Vx - Vy, set VF = not borrow |
| 8xy6 | `SHR Vx` | Set Vx = Vx >> 1 |
| 8xy7 | `SUBN Vx, Vy` | Set Vx = Vy - Vx, set VF = not borrow |
| 8xyE | `SHL Vx` | Set Vx = Vx << 1 |
| 9xy0 | `SNE Vx, Vy` | Skip next if Vx != Vy |
| Annn | `LD I, addr` | Set I = nnn |
| Bnnn | `JP V0, addr` | Jump to address nnn + V0 |
| Cxkk | `RND Vx, byte` | Set Vx = random byte AND byte |
| Dxyn | `DRW Vx, Vy, n` | Draw sprite at (Vx, Vy) with n bytes |
| Ex9E | `SKP Vx` | Skip next if key in Vx is pressed |
| ExA1 | `SKNP Vx` | Skip next if key in Vx is not pressed |
| Fx07 | `LD Vx, DT` | Set Vx = delay timer |
| Fx0A | `LD Vx, K` | Wait for key press, store in Vx |
| Fx15 | `LD DT, Vx` | Set delay timer = Vx |
| Fx18 | `LD ST, Vx` | Set sound timer = Vx |
| Fx1E | `ADD I, Vx` | Set I = I + Vx |
| Fx29 | `LD F, Vx` | Set I = location of sprite for digit Vx |
| Fx33 | `LD B, Vx` | Store BCD of Vx at I, I+1, I+2 |
| Fx55 | `LD [I], Vx` | Store V0-Vx in memory starting at I |
| Fx65 | `LD Vx, [I]` | Load V0-Vx from memory starting at I |

### Directives

- `DB byte` - Define byte (data)
- `DW word` - Define word (16-bit data)
- `LABEL:` - Define a label

### Constants

- Hexadecimal values: `0xFF`
- Decimal values: `255`
- Labels can be used as addresses

## Keyboard Controls

The Chip-8 uses a 16-key hexadecimal keypad:

| Key | Chip-8 Key |
|-----|------------|
| 1 | 1 |
| 2 | 2 |
| 3 | 3 |
| 4 | C |
| Q | 4 |
| W | 5 |
| E | 6 |
| R | D |
| A | 7 |
| S | 8 |
| D | 9 |
| F | E |
| Z | A |
| X | 0 |
| C | B |
| V | F |

## Examples

The `examples/` directory contains sample assembly programs:

- `hello_world.asm` - Displays "HELO" on screen
- `audio_input.asm` - Press any key to hear a beep

## ROMs

To test out some other ROMs you may download them from [here](https://johnearnest.github.io/chip8Archive/).

## License

MIT
