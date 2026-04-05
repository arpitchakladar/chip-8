//go:build wasm && js

package display

// WithWASM creates a new display implementation for WebAssembly.
func WithWASM() Display {
	return &WASMDisplay{}
}

// WASMDisplay is a placeholder for WASM display implementation.
type WASMDisplay struct{}

func (d *WASMDisplay) Init() error { return nil }
func (d *WASMDisplay) Clear()      {}
func (d *WASMDisplay) SetPixel(x, y uint8) (bool, error) {
	return false, nil
}
func (d *WASMDisplay) Present() error { return nil }
func (d *WASMDisplay) Close() error   { return nil }
