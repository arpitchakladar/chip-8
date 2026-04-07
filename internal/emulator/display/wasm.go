//go:build wasm && js

package display

import (
	"syscall/js"
)

type WASMDisplay struct {
	buffer   *DisplayBuffer
	Callback js.Value
}

func WithWASM() Display {
	return &WASMDisplay{buffer: NewDisplayBuffer()}
}

func (d *WASMDisplay) Init() error {
	return nil
}

func (d *WASMDisplay) Clear() {
	d.buffer.Clear()
}

func (d *WASMDisplay) SetPixel(x, y uint8) (bool, error) {
	return d.buffer.SetPixel(x, y)
}

func (d *WASMDisplay) Present() error {
	if !d.Callback.IsNull() && !d.Callback.IsUndefined() {
		d.Callback.Invoke()
	}
	return nil
}

func (d *WASMDisplay) Close() error {
	d.Callback = js.Value{}
	return nil
}

func (d *WASMDisplay) GetPixels() []byte {
	return d.buffer.GetPixels()
}

func (d *WASMDisplay) SetPresentCallback(callback js.Value) {
	d.Callback = callback
}
