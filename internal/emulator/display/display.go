package display

// SDLDisplay manages the display output for the CHIP-8 emulator.
// It maintains a pixel buffer and renders it to an SDL2 window.

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	// Width is the display width in pixels (standard CHIP-8).
	Width = 64
	// Height is the display height in pixels (standard CHIP-8).
	Height = 32
	// Scale is the scaling factor for rendering pixels to screen.
	// Each CHIP-8 pixel will be rendered as Scale x Scale on screen.
	Scale = 15
)

// SDLDisplay maintains the CHIP-8 display state and SDL2 rendering resources.
type SDLDisplay struct {
	// Pixels is the display buffer (2048 bytes for 64x32 display).
	// Each byte represents one pixel: 0 = off, 1 = on.
	Pixels [Width * Height]byte
	// window is the SDL2 window handle.
	window *sdl.Window
	// renderer is the SDL2 renderer for drawing to the window.
	renderer *sdl.Renderer
}

// New creates a new, cleared SDLDisplay instance.
// The pixel buffer is initialized to all zeros (black).
// Call Init() before use to create the SDL window.
func New() *SDLDisplay {
	return new(SDLDisplay)
}

// Init initializes the SDL2 subsystem and creates the window and renderer.
// It initializes all SDL2 subsystems, creates a centered window at 64*Scale x 32*Scale
// pixels, and creates an accelerated renderer for the window.
//
// Returns:
//   - nil on success
//   - *SDLError if SDL initialization, window creation, or renderer creation fails
func (d *SDLDisplay) Init() error {
	// Initialize all SDL2 subsystems
	if err := sdl.Init(uint32(sdl.INIT_EVERYTHING)); err != nil {
		return &SDLError{Subsystem: "Initialization", Child: err}
	}

	// Create the emulator window
	window, err := sdl.CreateWindow(
		"Chip-8 Emulator",
		int32(sdl.WINDOWPOS_CENTERED), int32(sdl.WINDOWPOS_CENTERED),
		int32(Width*Scale), int32(Height*Scale),
		uint32(sdl.WINDOW_SHOWN),
	)
	if err != nil {
		return &SDLError{Subsystem: "Window Creation", Child: err}
	}

	// Create the renderer (hardware accelerated preferred)
	dr, err := sdl.CreateRenderer(window, -1, uint32(sdl.RENDERER_ACCELERATED))
	if err != nil {
		return &SDLError{Subsystem: "Renderer Creation", Child: err}
	}

	d.window = window
	d.renderer = dr
	return nil
}

// Reset clears the entire pixel buffer to black (all zeros).
// This is equivalent to turning off all pixels.
// Note: This only clears the in-memory buffer, not the actual screen.
func (d *SDLDisplay) Reset() {
	d.Pixels = [Width * Height]byte{}
}

// Clear is an alias for Reset.
// It is called by the CLS (0x00E0) opcode to clear the display.
func (d *SDLDisplay) Clear() {
	d.Reset()
}

// SetPixel toggles a pixel at the specified coordinates using XOR mode.
// CHIP-8 uses XOR drawing: if a pixel is already on, drawing it again turns it off.
//
// Coordinates are wrapped to stay within bounds (standard CHIP-8 behavior):
//   - x wraps to 0-63 (64 pixels wide)
//   - y wraps to 0-31 (32 pixels tall)
//
// Parameters:
//   - x: X coordinate (0-63)
//   - y: Y coordinate (0-31)
//
// Returns:
//   - bool: true if the pixel was already on (collision), false otherwise
//   - error: *CoordinateError if coordinates are out of bounds (before wrapping),
//     or nil if coordinates are valid (even after wrapping)
func (d *SDLDisplay) SetPixel(x, y uint8) (bool, error) {
	var err *CoordinateError

	// Check if coordinates were out of bounds (before wrapping)
	// This allows detection of "strict mode" errors if needed
	if x >= Width || y >= Height {
		err = &CoordinateError{X: x, Y: y}
	}

	// Wrap coordinates to display bounds
	x %= Width
	y %= Height

	// Calculate pixel index in buffer
	index := uint16(x) + (uint16(y) * Width)

	// Check for collision: pixel was on (1) before XOR
	collision := d.Pixels[index] == 1

	// XOR the pixel: 0 -> 1, or 1 -> 0
	d.Pixels[index] ^= 1

	return collision, err
}

// Present renders the current pixel buffer to the SDL window.
// It clears the screen to black, then draws all pixels that are set (value 1)
// as white rectangles. Each pixel is scaled according to the Scale constant.
//
// The rendering order is:
//  1. Clear screen to black
//  2. Set draw color to white
//  3. Draw all "on" pixels as rectangles
//  4. Present (flip) the screen
//
// Returns:
//   - nil on success
//   - *SDLError if renderer is not initialized or drawing fails
func (d *SDLDisplay) Present() error {
	if d.renderer == nil {
		return &SDLError{Subsystem: "Renderer", Child: fmt.Errorf("renderer not initialized")}
	}

	// Clear screen to black
	if err := d.renderer.SetDrawColor(0, 0, 0, 255); err != nil {
		return &SDLError{Subsystem: "SetDrawColor (Background)", Child: err}
	}
	if err := d.renderer.Clear(); err != nil {
		return &SDLError{Subsystem: "Clear", Child: err}
	}

	// Set pixel color to white
	if err := d.renderer.SetDrawColor(255, 255, 255, 255); err != nil {
		return &SDLError{Subsystem: "SetDrawColor (Pixel)", Child: err}
	}

	// Draw each "on" pixel as a scaled rectangle
	for i, val := range d.Pixels {
		if val == 1 {
			// Calculate pixel position
			x := int32(i % Width)
			y := int32(i / Width)

			// Create scaled rectangle for pixel
			rect := sdl.Rect{
				X: x * Scale,
				Y: y * Scale,
				W: Scale,
				H: Scale,
			}

			if err := d.renderer.FillRect(&rect); err != nil {
				return &SDLError{Subsystem: "FillRect", Child: err}
			}
		}
	}

	// Present the rendered frame to the screen
	d.renderer.Present()

	return nil
}

// Close releases all display resources.
// It destroys the renderer first, then the window, and finally calls sdl.Quit().
// If either destroy operation fails, the error is returned (preferring the renderer error).
//
// This function is safe to call multiple times (subsequent calls will be no-ops).
//
// Returns:
//   - nil on success
//   - *SDLError containing the last error encountered during cleanup
func (d *SDLDisplay) Close() error {
	lastErr := error(nil)

	// Destroy renderer first
	if d.renderer != nil {
		if err := d.renderer.Destroy(); err != nil {
			lastErr = &SDLError{Subsystem: "Renderer Destruction", Child: err}
		}
	}

	// Then destroy window
	if d.window != nil {
		if err := d.window.Destroy(); err != nil {
			lastErr = &SDLError{Subsystem: "Window Destruction", Child: err}
		}
	}

	// Clean up SDL subsystems
	sdl.Quit()

	return lastErr
}
