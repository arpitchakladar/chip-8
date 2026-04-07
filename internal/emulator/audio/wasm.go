//go:build wasm && js

package audio

import (
	"syscall/js"
)

// WithWASM creates a new audio implementation for WebAssembly.
func WithWASM() Audio {
	return &WASMAudio{
		AudioContext: js.Global().Get("AudioContext"),
	}
}

// WASMAudio implements audio using the Web Audio API.
type WASMAudio struct {
	AudioContext js.Value
	Oscillator   js.Value
	GainNode     js.Value
	playing      bool
}

func (a *WASMAudio) Init() error {
	if a.AudioContext.IsUndefined() || a.AudioContext.IsNull() {
		a.AudioContext = js.Global().Get("AudioContext").New()
	}
	return nil
}

func (a *WASMAudio) GenerateBeep() error {
	if a.AudioContext.IsUndefined() || a.AudioContext.IsNull() {
		a.AudioContext = js.Global().Get("AudioContext").New()
	}

	ctx := a.AudioContext
	if ctx.Get("state").String() == "suspended" {
		ctx.Call("resume")
	}

	osc := ctx.Call("createOscillator")
	gain := ctx.Call("createGain")

	osc.Set("type", "square")
	osc.Set("frequency", 440)

	gain.Set("gain", 0.1)

	osc.Call("connect", gain)
	gain.Call("connect", ctx.Get("destination"))

	osc.Call("start")

	a.Oscillator = osc
	a.GainNode = gain
	a.playing = true

	return nil
}

func (a *WASMAudio) Play() error {
	if a.playing {
		return nil
	}

	if !a.Oscillator.IsUndefined() && !a.Oscillator.IsNull() {
		a.Oscillator.Call("stop")
		a.Oscillator = js.Value{}
		a.GainNode = js.Value{}
	}

	if err := a.GenerateBeep(); err != nil {
		return err
	}

	return nil
}

func (a *WASMAudio) Pause() error {
	if !a.playing {
		return nil
	}
	if !a.Oscillator.IsUndefined() && !a.Oscillator.IsNull() {
		a.Oscillator.Call("stop")
	}
	a.playing = false
	return nil
}

func (a *WASMAudio) Close() error {
	if err := a.Pause(); err != nil {
		return err
	}

	return nil
}
