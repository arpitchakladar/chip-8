//go:build wasm && js

// Package display provides a WebAssembly-compatible display implementation
// using the HTML5 Canvas API.
package display

import (
	"syscall/js"
)

// WASMDisplay implements the Display interface for WebAssembly/JS environments.
// It uses an HTML5 Canvas element to render CHIP-8 graphics through the 2D context.
type WASMDisplay struct {
	buffer *DisplayBuffer
	Canvas js.Value
}

// WithWASM creates a new Display that uses an HTML5 Canvas for rendering.
// The canvas parameter should be a JavaScript canvas element reference.
func WithWASM(canvas js.Value) Display {
	return &WASMDisplay{
		buffer: NewDisplayBuffer(),
		Canvas: canvas,
	}
}

func (d *WASMDisplay) Init() error {
	// No initialization needed for canvas-based display
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
// Draws each "on" pixel as a white rectangle, scaled to fit the canvas.
func (d *WASMDisplay) Present() error {
	d.render()

	return nil
}

// render draws the current display buffer to the HTML5 Canvas.
// It calculates the appropriate scale factor based on canvas dimensions
// and draws each pixel as a filled rectangle.
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

	ctx.Set("fillStyle", "black")
	ctx.Call("fillRect", 0, 0, width, height)

	ctx.Set("fillStyle", "white")

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
