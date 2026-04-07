//go:build !wasm || !js

package audio

import (
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	// SampleRate is the audio sample rate in Hz.
	SampleRate = 44100
	// Frequency is the beep frequency in Hz (440Hz = standard A4 pitch).
	Frequency = 440.0
)

// SDLAudio manages sound output for the CHIP-8 emulator.
// It generates square wave beeps using SDL2 audio devices.
type SDLAudio struct {
	// Device is the SDL audio device ID used for sound output.
	Device sdl.AudioDeviceID
}

// New creates a new SDLAudio instance with an uninitialized audio device.
// Call Init() before use to open the audio device.
func WithSDL() *SDLAudio {
	return new(SDLAudio)
}

// Init opens the default audio device and configures it for sound output.
// It sets up mono audio at 44.1kHz with 16-bit signed samples and 2048 sample buffer.
// Returns an error if the audio device cannot be opened.
//
// Note: On systems without audio hardware, this may return an error but the
// emulator should continue to run without sound.
func (a *SDLAudio) Init() error {
	spec := &sdl.AudioSpec{
		Freq:     SampleRate,
		Format:   sdl.AUDIO_S16SYS,
		Channels: 1,
		Samples:  2048,
	}

	// Open the default audio device
	dev, err := sdl.OpenAudioDevice("", false, spec, nil, 0)
	if err != nil {
		return err
	}
	a.Device = dev

	return nil
}

// GenerateBeep generates a 440Hz square wave tone and queues it to the audio buffer.
// The tone plays for approximately 1 second (one full cycle at SampleRate).
// It returns early if there's already sufficient audio queued to avoid buffering overflow.
//
// A square wave alternates between positive and negative amplitude:
//   - First half of period: +3000 (high)
//   - Second half of period: -3000 (low)
func (a *SDLAudio) GenerateBeep() error {
	// Check if sufficient audio is already queued
	if sdl.GetQueuedAudioSize(a.Device) >= 4096 {
		return nil
	}

	// Generate samples for 1 second of audio
	length := SampleRate
	data := make([]int16, length)
	period := SampleRate / int(Frequency)

	// Generate square wave: alternate between +3000 and -3000
	for i := range length {
		if i%period < (period / 2) {
			data[i] = 3000
		} else {
			data[i] = -3000
		}
	}

	// Convert int16 samples to bytes for SDL
	byteLen := len(data) * 2
	byteData := unsafe.Slice((*byte)(unsafe.Pointer(&data[0])), byteLen)

	// Queue the audio data for playback
	if err := sdl.QueueAudio(a.Device, byteData); err != nil {
		return err
	}

	return nil
}

// Play unpauses the audio device to resume sound output.
// Use this when the sound timer is greater than 0 to start/beep.
func (a *SDLAudio) Play() error {
	sdl.PauseAudioDevice(a.Device, false)
	return nil
}

// Pause pauses the audio device to stop sound output.
// Use this when the sound timer reaches 0 to silence the beep.
func (a *SDLAudio) Pause() error {
	sdl.PauseAudioDevice(a.Device, true)
	return nil
}

// Close stops and releases the audio device resources.
// It first silences the device by pausing, then closes the SDL audio device,
// and finally resets the device ID to 0. Safe to call multiple times.
func (a *SDLAudio) Close() error {
	if a.Device != 0 {
		sdl.PauseAudioDevice(a.Device, true) // Silence first
		sdl.CloseAudioDevice(a.Device)
		a.Device = 0 // Reset ID
	}

	return nil
}
