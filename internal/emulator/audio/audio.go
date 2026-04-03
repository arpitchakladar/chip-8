package audio

import (
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	SampleRate = 44100
	Frequency  = 440.0 // Standard A4 pitch
)

type Audio struct {
	Device sdl.AudioDeviceID
}

func New() *Audio {
	return new(Audio)
}

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
		if i % period < (period / 2) {
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

func (a *Audio) Play() {
	// Re-queue the sound if needed, or just Unpause
	sdl.PauseAudioDevice(a.Device, false)
}

func (a *Audio) Pause() {
	// Re-queue the sound if needed, or just Unpause
	sdl.PauseAudioDevice(a.Device, true)
}

func (a *Audio) Close() {
	if a.Device != 0 {
		sdl.PauseAudioDevice(a.Device, true) // Silence it first
		sdl.CloseAudioDevice(a.Device)
		a.Device = 0 // Reset the ID
	}
}
