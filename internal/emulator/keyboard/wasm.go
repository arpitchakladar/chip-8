//go:build wasm && js

package keyboard

// WithWASM creates a new keyboard implementation for WebAssembly.
func WithWASM() Keyboard {
	return &WASMKeyboard{}
}

// WASMKeyboard is a placeholder for WASM keyboard implementation.
type WASMKeyboard struct {
	Keys [16]bool
}

func (kb *WASMKeyboard) IsKeyPressed(key byte) bool {
	return key <= 15 && kb.Keys[key]
}

func (kb *WASMKeyboard) AnyKeyPressed() (byte, bool) {
	for i, isPressed := range kb.Keys {
		if isPressed {
			return byte(i), true
		}
	}
	return 0, false
}

func (kb *WASMKeyboard) SetKey(key byte, pressed bool) {
	if key <= 15 {
		kb.Keys[key] = pressed
	}
}

func (kb *WASMKeyboard) PollEvents() {}
