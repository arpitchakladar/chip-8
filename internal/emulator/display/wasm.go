//go:build wasm && js

package display

import (
	"syscall/js"
)

type WASMDisplay struct {
	buffer *DisplayBuffer
	Canvas js.Value
}

func WithWASM(canvas js.Value) Display {
	return &WASMDisplay{
		buffer: NewDisplayBuffer(),
		Canvas: canvas,
	}
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
	d.render()

	return nil
}

func (d *WASMDisplay) render() {
	if d.Canvas.IsNull() || d.Canvas.IsUndefined() {
		return
	}

	ctx := d.Canvas.Call("getContext", "2d")
	if ctx.IsNull() || ctx.IsUndefined() {
		return
	}

	width := d.Canvas.Get("width").Int()
	height := d.Canvas.Get("height").Int()
	scale := width / Width
	if newScale := height / Height; newScale < scale {
		scale = newScale
	}

	ctx.Call("fillStyle", "black")
	ctx.Call("fillRect", 0, 0, width, height)

	ctx.Call("setFillStyle", "white")

	pixels := d.buffer.GetPixels()
	for i, val := range pixels {
		if val == 1 {
			x := (i % Width) * scale
			y := (i / Width) * scale
			ctx.Call("fillRect", x, y, scale, scale)
		}
	}
}

func (d *WASMDisplay) Close() error {
	return nil
}

func (d *WASMDisplay) GetPixels() []byte {
	return d.buffer.GetPixels()
}
