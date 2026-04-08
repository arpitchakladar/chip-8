# CHIP-8

A Chip-8 virtual machine and assembler written in Go with SDL2 for graphics and audio.

## Demo

![Demo](https://github.com/user-attachments/assets/c48a8e14-bb22-4dd9-ab96-d436de5da920)

This is a demo of the Snake game running from `games/snake.asm`.

## Documentation

For a more detailed explanation of the project, see the [docs](https://arpitchakladar.github.io/chip-8/). Note that it is more of a presentation than traditional documentation.

## Prerequisites

- **Go** 1.25 or later
- **SDL2** development libraries

### Installing SDL2

See the [SDL2 installation guide](https://wiki.libsdl.org/SDL2/Installation) for your platform.

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

To assemble one or more .asm file into a .ch8 file:
```bash
./chip-8 assemble <path-to-asm1> <path-to-asm2> ...
```

Example:
```bash
./chip-8 assemble examples/hello_world.asm
```

To specify a custom output path:
```bash
./chip-8 assemble -o output.ch8 examples/hello_world.asm
```

This will output `examples/hello_world.ch8` by default, or `output.ch8` if `-o` is specified.

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

### Special Labels

- `__START` - Must be defined before any instructions (required)
- `__END` - Must be the last label, no instructions allowed after it (required)

When compiling multiple assembly files, files containing `__START` are placed first, and files containing `__END` are placed last in the concatenation order.

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
[1dcell.ch8](https://johnearnest.github.io/chip8Archive/play.html?p=1dcell)
![1dcell](https://github.com/user-attachments/assets/bdb9391f-c3c9-4ce8-a1af-769320c1b85a)

[octoachip8story.ch8](https://johnearnest.github.io/chip8Archive/play.html?p=octoachip8story)
![octoachip8story](https://github.com/user-attachments/assets/4bff8aa9-e621-4f93-86fb-ca9a79913b55)

[pumpkindressup.ch8](https://johnearnest.github.io/chip8Archive/play.html?p=pumpkindressup)
![pumpkindressup](https://github.com/user-attachments/assets/ec7dfe4f-c5f6-4e92-adad-4d11def91889)

[RPS.ch8](https://johnearnest.github.io/chip8Archive/play.html?p=RPS)
![RPS](https://github.com/user-attachments/assets/5236e161-7796-46b3-8380-737c8c9c6d91)

To test out some other ROMs you may download them from [here](https://johnearnest.github.io/chip8Archive/).

## WebAssembly (WASM)

The CHIP-8 emulator can be compiled to WebAssembly to run in web browsers.

### Building for Web

```bash
GOOS=js GOARCH=wasm go build -o examples/main.wasm ./cmd/chip-8
```

### Running WASM in Browser

You need a web server to serve the WASM files (browsers block WASM from file:// URLs).

```bash
cd examples
python3 -m http.server 8080
```

Then open `http://localhost:8080` in your browser.

### JavaScript API

The WASM module exposes a `chip_8` global object with the following:

#### `chip_8.Emulator(canvas, clockSpeed)`

Creates a new CHIP-8 emulator instance.

**Parameters:**
- `canvas` - A JavaScript canvas element for rendering
- `clockSpeed` - CPU clock speed in Hz (e.g., 500 for 500 Hz)

**Returns:** An emulator instance with the following methods:

- `loadROM(data)` - Load ROM bytecode (Uint8Array)
- `run()` - Start the emulator
- `destroy()` - Stop the emulator and release resources
- `handleKeyboard()` - Attach keyboard event handlers
- `releaseKeyboard()` - Remove keyboard event handlers
- `sendKey(key, pressed)` - Send a key press/release (key: 0-15, pressed: boolean)
- `isHandlingKeyboard()` - Returns whether keyboard handlers are active (boolean)

#### `chip_8.Assembler(source)`

Creates an assembler to compile CHIP-8 assembly code.

**Parameters:**
- `source` - Assembly source code as a string

**Returns:** An assembler with:

- `assemble()` - Compiles the source and returns Uint8Array ROM data

### Example Usage

```javascript
const canvas = document.getElementById("canvas");
const vm = new chip_8.Emulator(canvas, 500);

const response = await fetch("game.asm");
const asmCode = await response.text();
const assembler = new chip_8.Assembler(asmCode);
const romData = assembler.assemble();

vm.loadROM(romData);
vm.run();

// To make the emulator handle keyboard inputs automatically
vm.handleKeyboard();
```

See `examples/index.html` for a complete working example.

## License

MIT
