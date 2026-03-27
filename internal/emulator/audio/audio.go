package audio

import (
	"math"
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

	dev, err := sdl.OpenAudioDevice("", false, spec, nil, 0)
	if err != nil {
		return err
	}
	a.Device = dev

	// --- PRE-GENERATE A SMALL LOOP ---
	// 0.1 seconds is enough to loop smoothly
	loopLength := SampleRate / 10
	data := make([]int16, loopLength)
	for i := range loopLength {
		if math.Sin(2.0*math.Pi*Frequency*float64(i)/SampleRate) > 0 {
			data[i] = 3000
		} else {
			data[i] = -3000
		}
	}

	byteLen := len(data) * 2
	byteData := unsafe.Slice((*byte)(unsafe.Pointer(&data[0])), byteLen)

	// Fill the queue once. SDL will keep playing this data if we don't clear it.
	// However, QueueAudio isn't great for infinite loops.
	// A better way is to just queue a LOT of it once.
	for range 10 { // Queue 1 second total
		if err := sdl.QueueAudio(a.Device, byteData); err != nil {
			return err
		}
	}

	return nil
}

func (a *Audio) GenerateBeep() error {
	if sdl.GetQueuedAudioSize(a.Device) >= 4096 {
		return nil
	}

	length := SampleRate
	data := make([]int16, length)

	for i := range length {
		if math.Sin(2.0*math.Pi*Frequency*float64(i)/SampleRate) > 0 {
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
