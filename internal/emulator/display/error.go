package display

import "fmt"

// RenderError: Failed to communicate with the GPU or SDL.
type RenderError struct {
	Subsystem   string
	Child       error
}

func (e *RenderError) Error() string {
	return fmt.Sprintf("DISPLAY RENDER ERROR [%s]: %v", e.Subsystem, e.Child)
}

// OutOfBoundsError: Attempted to draw outside the 64x32 grid.
// Useful if you want to switch from "Wrapping" mode to "Strict" mode.
type OutOfBoundsError struct {
	X, Y uint8
}

func (e *OutOfBoundsError) Error() string {
	return fmt.Sprintf("DISPLAY BOUNDS VIOLATION: Attempted to draw at (%d, %d)", e.X, e.Y)
}

// SDLError represents a failure in the SDL2 hardware abstraction layer.
type SDLError struct {
	Subsystem string
	Child     error
}

func (e *SDLError) Error() string {
	return fmt.Sprintf("SDL %s Error: %v", e.Subsystem, e.Child)
}

// CoordinateError represents an attempt to draw outside the allowed grid
// (Useful if you ever want to disable auto-wrapping for debugging).
type CoordinateError struct {
	X, Y uint8
}

func (e *CoordinateError) Error() string {
	return fmt.Sprintf("Display out of bounds: (%d, %d)", e.X, e.Y)
}
