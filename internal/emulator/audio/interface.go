package audio

// Audio represents the audio subsystem for the CHIP-8 emulator.
// Implement this interface to provide audio output for different platforms
// (e.g., SDL2, JavaScript/WebAudio, etc.).
type Audio interface {
	// Init initializes the audio subsystem.
	// Returns an error if initialization fails.
	Init() error

	// GenerateBeep generates a beep sound and queues it for playback.
	// The beep should play for approximately 1 second at 440Hz (A4 pitch).
	GenerateBeep() error

	// Play starts/resumes audio playback.
	Play()

	// Pause stops audio playback.
	Pause()

	// Close releases audio resources.
	// Should be safe to call multiple times.
	Close()
}
