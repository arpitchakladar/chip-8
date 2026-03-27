package keyboard

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
	if key > 15 {
		return false
	}
	return i.Keys[key]
}
