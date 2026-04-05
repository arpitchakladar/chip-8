package display

// Display represents the display subsystem for the CHIP-8 emulator.
// Implement this interface to provide display output for different platforms
// (e.g., SDL2, JavaScript/Canvas, etc.).
type Display interface {
	// Init initializes the display subsystem and creates any necessary windows or surfaces.
	// Returns an error if initialization fails.
	Init() error

	// Clear clears the entire display buffer to all pixels off.
	Clear()

	// SetPixel toggles a pixel at the specified coordinates using XOR mode.
	// If a pixel is already on, drawing it again turns it off.
	// Coordinates are wrapped to stay within bounds (0-63 for x, 0-31 for y).
	// Returns true if the pixel was already on (collision), false otherwise.
	// Returns an error if coordinates are invalid.
	SetPixel(x, y uint8) (bool, error)

	// Present renders the current display buffer to the screen.
	// This should be called at 60Hz.
	Present() error

	// Reset clears the display buffer without affecting the screen.
	Reset()

	// Close releases display resources.
	// Should be safe to call multiple times.
	Close() error
}
