package keyboard

// SDLKeyboard tracks the state of the 16 CHIP-8 keys.
// It maintains a map of which keys are currently pressed and provides
// methods for checking key state, used by CPU opcodes for input handling.

import "github.com/veandco/go-sdl2/sdl"

// SDLKeyboard tracks the state of the 16 CHIP-8 keys.
type SDLKeyboard struct {
	// Keys stores the pressed state of each CHIP-8 key.
	// Index 0-15 corresponds to CHIP-8 keys 0x0-0xF.
	// true = pressed, false = released
	Keys [16]bool
}

// New creates a new SDLKeyboard instance with all keys initialized to released.
func New() *SDLKeyboard {
	return new(SDLKeyboard)
}

// IsKeyPressed checks if a specific CHIP-8 key is currently pressed.
// It is used by the SKP (0xEx9E) and SKNP (0xExA1) opcodes.
//
// Parameters:
//   - key: the CHIP-8 key index (0-15, corresponds to Vx register values)
//
// Returns:
//   - true if the key is currently pressed
//   - false if the key is released OR if key is out of range (>15)
func (kb *SDLKeyboard) IsKeyPressed(key byte) bool {
	return key <= 15 && kb.Keys[key]
}

// AnyKeyPressed checks if any CHIP-8 key is currently pressed.
// It is used by the LD Vx, K (0xFx0A) opcode to wait for key input.
//
// When no key is pressed, the CPU uses this to implement a blocking wait:
// it repeatedly executes this instruction until a key is pressed.
//
// Returns:
//   - byte: the key index (0-15) of the first pressed key found
//   - bool: true if any key is pressed, false if no keys are pressed
func (kb *SDLKeyboard) AnyKeyPressed() (byte, bool) {
	for i, isPressed := range kb.Keys {
		if isPressed {
			return byte(i), true
		}
	}
	return 0, false
}

// SetKey sets the pressed state of a CHIP-8 key.
func (kb *SDLKeyboard) SetKey(key byte, pressed bool) {
	if key <= 15 {
		kb.Keys[key] = pressed
	}
}

// HandleKeyboard updates the keyboard state based on an SDL keyboard event.
// It maps PC keyboard keys to CHIP-8 hex keys (0-F) and tracks press/release state.
//
// The key mapping follows the standard CHIP-8 layout:
//
//	 PC Key  | CHIP-8
//		--------|--------
//		1 2 3 4 | 1 2 3 C
//		q w e r | 4 5 6 D
//		a s d f | 7 8 9 E
//		z x c v | A 0 B F
//
// Parameters:
//   - event: pointer to an SDL KeyboardEvent (key press or release)
func (kb *SDLKeyboard) HandleKeyboard(event *sdl.KeyboardEvent) {
	keyCode := event.Keysym.Sym
	// Determine if this is a key press (true) or release (false)
	isPressed := event.Type == sdl.KEYDOWN

	// Map PC keyboard keys to CHIP-8 key indices
	mapping := map[sdl.Keycode]byte{
		sdl.Keycode(sdl.K_1): 0x1, sdl.Keycode(sdl.K_2): 0x2, sdl.Keycode(sdl.K_3): 0x3, sdl.Keycode(sdl.K_4): 0xC,
		sdl.Keycode(sdl.K_q): 0x4, sdl.Keycode(sdl.K_w): 0x5, sdl.Keycode(sdl.K_e): 0x6, sdl.Keycode(sdl.K_r): 0xD,
		sdl.Keycode(sdl.K_a): 0x7, sdl.Keycode(sdl.K_s): 0x8, sdl.Keycode(sdl.K_d): 0x9, sdl.Keycode(sdl.K_f): 0xE,
		sdl.Keycode(sdl.K_z): 0xA, sdl.Keycode(sdl.K_x): 0x0, sdl.Keycode(sdl.K_c): 0xB, sdl.Keycode(sdl.K_v): 0xF,
	}

	// Update the key state if the PC key maps to a CHIP-8 key
	if chipKey, ok := mapping[keyCode]; ok {
		kb.Keys[chipKey] = isPressed
	}
}
