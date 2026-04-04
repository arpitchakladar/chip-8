package audio

import (
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	SampleRate = 44100
	Frequency  = 440.0 // Standard A4 pitch
)

// Audio manages sound output for the CHIP-8 emulator.
// It generates square wave beeps using SDL2 audio devices.
type Audio struct {
	// Device is the SDL audio device ID used for sound output.
	Device sdl.AudioDeviceID
}

// New creates a new Audio instance.
func New() *Audio {
	return new(Audio)
}

// Init opens the default audio device and configures it for sound output.
// It sets up mono audio at 44.1kHz with 16-bit signed samples.
func (a *Audio) Init() error {
	spec := &sdl.AudioSpec{
		Freq:     SampleRate,
		Format:   sdl.AUDIO_S16SYS,
		Channels: 1,
		Samples:  2048,
	}

	// Get audio device
	dev, err := sdl.OpenAudioDevice("", false, spec, nil, 0)
	if err != nil {
		return err
	}
	a.Device = dev

	return nil
}

// GenerateBeep queues a square wave tone to the audio buffer.
// The tone plays at 440Hz (A4) for approximately 1 second.
// It returns early if there's already sufficient audio queued.
func (a *Audio) GenerateBeep() error {
	if sdl.GetQueuedAudioSize(a.Device) >= 4096 {
		return nil
	}

	length := SampleRate
	data := make([]int16, length)
	period := SampleRate / int(Frequency)

	for i := range length {
		// Generating a square wave
		// If we are in the first half of the wave period, stay high
		if i%period < (period / 2) {
			data[i] = 3000
		} else {
			data[i] = -3000
		}
	}

	// 1. Calculate the byte length (each int16 is 2 bytes)
	byteLen := len(data) * 2

	// 2. Convert the int16 slice to a byte slice using unsafe pointers
	byteData := unsafe.Slice((*byte)(unsafe.Pointer(&data[0])), byteLen)

	// 3. Queue the raw bytes
	if err := sdl.QueueAudio(a.Device, byteData); err != nil {
		return err
	}

	return nil
}

// Play unpauses the audio device to resume sound output.
func (a *Audio) Play() {
	// Re-queue the sound if needed, or just Unpause
	sdl.PauseAudioDevice(a.Device, false)
}

// Pause pauses the audio device to stop sound output.
func (a *Audio) Pause() {
	// Re-queue the sound if needed, or just Unpause
	sdl.PauseAudioDevice(a.Device, true)
}

// Close stops and closes the audio device.
// It first silences the device, then releases the audio resources.
func (a *Audio) Close() {
	if a.Device != 0 {
		sdl.PauseAudioDevice(a.Device, true) // Silence it first
		sdl.CloseAudioDevice(a.Device)
		a.Device = 0 // Reset the ID
	}
}
