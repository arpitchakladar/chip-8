//go:build wasm && js

// Package keyboard provides a WebAssembly-compatible keyboard implementation.
package keyboard

// WASMKeyboard implements the Keyboard interface for WebAssembly/JS environments.
// Key state is managed through SetKey() which should be connected to JavaScript
// keydown/keyup event handlers. Supports CHIP-8's standard 16-key layout (0-F).
type WASMKeyboard struct {
	Keys [16]bool // State of each of the 16 keys (true = pressed)
}

// WithWASM creates a new Keyboard implementation for WebAssembly.
// Returns a WASMKeyboard that must be connected to DOM event handlers.
func WithWASM() Keyboard {
	return &WASMKeyboard{}
}

// IsKeyPressed returns true if the specified key is currently pressed.
// The key parameter should be a value from 0-15 representing CHIP-8 keys.
func (kb *WASMKeyboard) IsKeyPressed(key byte) bool {
	return key <= 15 && kb.Keys[key]
}

// AnyKeyPressed returns the first pressed key and true if any key is pressed.
// Used for CHIP-8's wait-for-key instruction (FX0A).
// Returns (0, false) if no keys are pressed.
func (kb *WASMKeyboard) AnyKeyPressed() (byte, bool) {
	for i, isPressed := range kb.Keys {
		if isPressed {
			return byte(i), true
		}
	}
	return 0, false
}

// SetKey updates the pressed state of a key.
// The key parameter should be 0-15, pressed should be true for keydown, false for keyup.
func (kb *WASMKeyboard) SetKey(key byte, pressed bool) {
	if key <= 15 {
		kb.Keys[key] = pressed
	}
}

// PollEvents is a no-op for WASM keyboard since events are pushed directly via SetKey.
// Kept for interface compatibility with SDL keyboard implementation.
func (kb *WASMKeyboard) PollEvents() {}
