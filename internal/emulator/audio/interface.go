// Package audio provides audio output implementations for the CHIP-8 emulator.
// Implement the Audio interface to provide sound for different platforms.
package audio

// Audio represents the audio subsystem for the CHIP-8 emulator.
// Implement this interface to provide audio output for different platforms
// (e.g., SDL2, JavaScript/WebAudio, etc.).
type Audio interface {
	// Init initializes the audio subsystem.
	// Returns an error if initialization fails.
	Init() error

	// Play starts/resumes audio playback.
	Play() error

	// Pause stops audio playback.
	Pause() error

	// Close releases audio resources.
	// Should be safe to call multiple times.
	Close() error
}
