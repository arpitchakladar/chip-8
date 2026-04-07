// Package keyboard provides keyboard input implementations for the CHIP-8 emulator.
// Implement the Keyboard interface to provide input handling for different platforms.
package keyboard

// Keyboard represents the keyboard input subsystem for the CHIP-8 emulator.
// CHIP-8 has 16 keys (0-F). Implement this interface to provide input handling
// for different platforms (e.g., SDL2, JavaScript, etc.).
type Keyboard interface {
	// IsKeyPressed checks if a specific CHIP-8 key is currently pressed.
	// Key should be 0-15. Returns true if pressed, false otherwise.
	IsKeyPressed(key byte) bool

	// AnyKeyPressed checks if any CHIP-8 key is currently pressed.
	// Returns the key index (0-15) and true if any key is pressed,
	// or 0 and false if no keys are pressed.
	AnyKeyPressed() (byte, bool)

	// SetKey sets the pressed state of a CHIP-8 key.
	// Key should be 0-15.
	SetKey(key byte, pressed bool)

	// PollEvents processes platform-specific input events.
	// This should be called periodically (e.g., at 60Hz).
	// For SDL: polls and processes SDL keyboard events.
	// For Web: can be a no-op if using event-driven updates.
	PollEvents()
}
