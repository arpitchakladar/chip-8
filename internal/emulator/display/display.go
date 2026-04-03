package display

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	Width  = 64
	Height = 32
	Scale  = 15 // Each Chip-8 pixel will be 15x15 on screen
)

type Display struct {
	Pixels   [Width * Height]byte
	window   *sdl.Window
	renderer *sdl.Renderer
}

// New creates a new, cleared display instance.
func New() *Display {
	return new(Display)
}

// InitSDL sets up the window and renderer
func (d *Display) Init() error {
	if err := sdl.Init(uint32(sdl.INIT_EVERYTHING)); err != nil {
		return &SDLError{Subsystem: "Initialization", Err: err}
	}

	window, err := sdl.CreateWindow(
		"Chip-8 Emulator",
		int32(sdl.WINDOWPOS_CENTERED), int32(sdl.WINDOWPOS_CENTERED),
		int32(Width*Scale), int32(Height*Scale),
		uint32(sdl.WINDOW_SHOWN),
	)
	if err != nil {
		return &SDLError{Subsystem: "Window Creation", Err: err}
	}

	dr, err := sdl.CreateRenderer(window, -1, uint32(sdl.RENDERER_ACCELERATED))
	if err != nil {
		return &SDLError{Subsystem: "Renderer Creation", Err: err}
	}

	d.window = window
	d.renderer = dr
	return nil
}

// Reset clears the entire pixel buffer to black (0).
func (d *Display) Reset() {
	d.Pixels = [Width * Height]byte{}
}

// Clear is an alias for Reset, often called by the 00E0 opcode.
func (d *Display) Clear() {
	d.Reset()
}

// SetPixel toggles a pixel at (x, y) and returns true if a collision occurred.
// Chip-8 uses XOR drawing: if a pixel is already on and we draw it again, it turns off.
func (d *Display) SetPixel(x, y uint8) (bool, error) {
	var err *CoordinateError

	// Check if the original coordinates were out of bounds
	// even though we will wrap them below.
	if x >= Width || y >= Height {
		err = &CoordinateError{X: x, Y: y}
	}

	// Wrap coordinates (standard Chip-8 behavior)
	x %= Width
	y %= Height

	index := uint16(x) + (uint16(y) * Width)

	// Check for collision (pixel was 1, now will be 0)
	collision := d.Pixels[index] == 1

	// XOR the pixel
	d.Pixels[index] ^= 1

	// Returns collision status AND the error (which is nil if in-bounds)
	return collision, err
}

// Present draws the current Pixels buffer to the SDL window
func (d *Display) Present() error {
	if d.renderer == nil {
		return &SDLError{Subsystem: "Renderer", Err: fmt.Errorf("renderer not initialized")}
	}

	// 1. Set Background to Black
	if err := d.renderer.SetDrawColor(0, 0, 0, 255); err != nil {
		return &SDLError{Subsystem: "SetDrawColor (Background)", Err: err}
	}

	if err := d.renderer.Clear(); err != nil {
		return &SDLError{Subsystem: "Clear", Err: err}
	}

	// 2. Set Pixel Color to White
	if err := d.renderer.SetDrawColor(255, 255, 255, 255); err != nil {
		return &SDLError{Subsystem: "SetDrawColor (Pixel)", Err: err}
	}

	// 3. Draw the active pixels
	for i, val := range d.Pixels {
		if val == 1 {
			x := int32(i % Width)
			y := int32(i / Width)

			rect := sdl.Rect{
				X: x * Scale,
				Y: y * Scale,
				W: Scale,
				H: Scale,
			}

			if err := d.renderer.FillRect(&rect); err != nil {
				return &SDLError{Subsystem: "FillRect", Err: err}
			}
		}
	}

	// 4. Update screen
	d.renderer.Present()

	return nil
}

func (d *Display) Close() error {
	lastErr := error(nil)

	// 1. Attempt to destroy the renderer
	if d.renderer != nil {
		if err := d.renderer.Destroy(); err != nil {
			lastErr = &SDLError{Subsystem: "Renderer Destruction", Err: err}
		}
	}

	// 2. Attempt to destroy the window (even if renderer failed)
	if d.window != nil {
		if err := d.window.Destroy(); err != nil {
			// We wrap the error, but if there was a previous error,
			// you might want to log it or concatenate it.
			lastErr = &SDLError{Subsystem: "Window Destruction", Err: err}
		}
	}

	// 3. Global SDL Cleanup
	sdl.Quit()

	return lastErr
}
