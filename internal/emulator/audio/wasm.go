//go:build wasm && js

package audio

// WithWASM creates a new audio implementation for WebAssembly.
func WithWASM() Audio {
	return &WASMAudio{}
}

// WASMAudio is a placeholder for WASM audio implementation.
type WASMAudio struct{}

func (a *WASMAudio) Init() error         { return nil }
func (a *WASMAudio) GenerateBeep() error { return nil }
func (a *WASMAudio) Play()               {}
func (a *WASMAudio) Pause()              {}
func (a *WASMAudio) Close()              {}
