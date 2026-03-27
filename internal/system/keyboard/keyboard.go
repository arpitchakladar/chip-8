package keyboard

import "github.com/veandco/go-sdl2/sdl"

// Keyboard tracks the state of the 16 Chip-8 keys.
type Keyboard struct {
	// Keys stores true if a key is currently pressed.
	// Index 0-15 corresponds to Chip-8 keys 0x0-0xF.
	Keys [16]bool
}

func New() *Keyboard {
	return new(Keyboard)
}

// IsKeyPressed is a helper for the CPU opcodes (EX9E and EXA1).
func (i *Keyboard) IsKeyPressed(key byte) bool {
	return key <= 15 && i.Keys[key]
}

func (in *Keyboard) AnyKeyPressed() (byte, bool) {
	for i, isPressed := range in.Keys {
		if isPressed {
			return byte(i), true
		}
	}
	return 0, false
}

func (kb *Keyboard) HandleKeyboard(event *sdl.KeyboardEvent) {
	keyCode := event.Keysym.Sym
	// Type 0x300 is KeyDown, 0x301 is KeyUp in SDL
	isPressed := event.Type == sdl.KEYDOWN

	// Explicitly cast constants to sdl.Keycode to satisfy the map type
	mapping := map[sdl.Keycode]byte{
		sdl.Keycode(sdl.K_1): 0x1, sdl.Keycode(sdl.K_2): 0x2, sdl.Keycode(sdl.K_3): 0x3, sdl.Keycode(sdl.K_4): 0xC,
		sdl.Keycode(sdl.K_q): 0x4, sdl.Keycode(sdl.K_w): 0x5, sdl.Keycode(sdl.K_e): 0x6, sdl.Keycode(sdl.K_r): 0xD,
		sdl.Keycode(sdl.K_a): 0x7, sdl.Keycode(sdl.K_s): 0x8, sdl.Keycode(sdl.K_d): 0x9, sdl.Keycode(sdl.K_f): 0xE,
		sdl.Keycode(sdl.K_z): 0xA, sdl.Keycode(sdl.K_x): 0x0, sdl.Keycode(sdl.K_c): 0xB, sdl.Keycode(sdl.K_v): 0xF,
	}

	if chipKey, ok := mapping[keyCode]; ok {
		// Ensure your keyboard package uses a slice or array of bools
		kb.Keys[chipKey] = isPressed
	}
}
