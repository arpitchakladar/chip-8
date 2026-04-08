//go:build wasm && js

// Package display provides a WebAssembly-compatible display implementation
// using the HTML5 Canvas API.
package display

import (
	"fmt"
	"syscall/js"
)

// WASMDisplay implements the Display interface for WebAssembly/JS environments.
// It uses an HTML5 Canvas element to render CHIP-8 graphics through the 2D context.
type WASMDisplay struct {
	buffer    *DisplayBuffer
	Canvas    js.Value
	ctx       js.Value
	imageData js.Value
	data      []byte
}

// WithWASM creates a new Display that uses an HTML5 Canvas for rendering.
// The canvas parameter should be a JavaScript canvas element reference.
func WithWASM(canvas js.Value) Display {
	return &WASMDisplay{
		buffer: NewDisplayBuffer(),
		Canvas: canvas,
	}
}

// Init initializes the WASM display.
// It gets the 2D rendering context, sets the canvas to native 64x32 resolution,
// preserves the original canvas dimensions via CSS, and pre-allocates the pixel buffer.
func (d *WASMDisplay) Init() error {
	if d.Canvas.IsNull() || d.Canvas.IsUndefined() {
		return nil
	}

	d.ctx = d.Canvas.Call("getContext", "2d")
	if d.ctx.IsNull() || d.ctx.IsUndefined() {
		return nil
	}

	canvasWidth := d.Canvas.Get("width").Int()
	canvasHeight := d.Canvas.Get("height").Int()

	d.Canvas.Set("width", Width)
	d.Canvas.Set("height", Height)

	style := d.Canvas.Get("style")
	style.Set("width", js.ValueOf(fmt.Sprintf("%dpx", canvasWidth)))
	style.Set("height", js.ValueOf(fmt.Sprintf("%dpx", canvasHeight)))

	d.imageData = d.ctx.Call("createImageData", Width, Height)
	d.data = make([]byte, Width*Height*4)

	return nil
}

func (d *WASMDisplay) Clear() {
	d.buffer.Clear()
}

// SetPixel delegates to the display buffer for XOR pixel drawing.
// Returns true if the pixel was already on (collision detection for sprites).
func (d *WASMDisplay) SetPixel(x, y uint8) (bool, error) {
	return d.buffer.SetPixel(x, y)
}

// Present renders the current display buffer to the HTML5 Canvas.
// Uses ImageData for fast pixel rendering.
func (d *WASMDisplay) Present() error {
	if d.ctx.IsNull() || d.ctx.IsUndefined() {
		return nil
	}

	pixels := d.buffer.GetPixels()

	for i, val := range pixels {
		offset := i * 4
		if val == 1 {
			d.data[offset] = 255
			d.data[offset+1] = 255
			d.data[offset+2] = 255
			d.data[offset+3] = 255
		} else {
			d.data[offset] = 0
			d.data[offset+1] = 0
			d.data[offset+2] = 0
			d.data[offset+3] = 255
		}
	}

	jsData := d.imageData.Get("data")
	js.CopyBytesToJS(jsData, d.data)

	d.ctx.Call("putImageData", d.imageData, 0, 0)

	return nil
}

func (d *WASMDisplay) Close() error {
	return nil
}

func (d *WASMDisplay) GetPixels() []byte {
	return d.buffer.GetPixels()
}
