package keyboard

// BasicKeyboard is a simple keyboard implementation for platforms that don't need
// event polling (e.g., WebAssembly). Key states are set externally via SetKey().
type BasicKeyboard struct {
	Keys [16]bool
}

// NewBasic creates a new BasicKeyboard with all keys initialized to released.
func NewBasic() *BasicKeyboard {
	return new(BasicKeyboard)
}

// IsKeyPressed checks if a specific CHIP-8 key is currently pressed.
func (kb *BasicKeyboard) IsKeyPressed(key byte) bool {
	return key <= 15 && kb.Keys[key]
}

// AnyKeyPressed checks if any CHIP-8 key is currently pressed.
func (kb *BasicKeyboard) AnyKeyPressed() (byte, bool) {
	for i, isPressed := range kb.Keys {
		if isPressed {
			return byte(i), true
		}
	}
	return 0, false
}

// SetKey sets the pressed state of a CHIP-8 key.
func (kb *BasicKeyboard) SetKey(key byte, pressed bool) {
	if key <= 15 {
		kb.Keys[key] = pressed
	}
}

// PollEvents is a no-op for BasicKeyboard.
// In web builds, key events should be handled externally.
func (kb *BasicKeyboard) PollEvents() {}
