//go:build wasm && js

package audio

import (
	"syscall/js"
)

type WASMAudio struct {
	AudioContext js.Value
	Oscillator   js.Value
	GainNode     js.Value
	playing      bool
}

// WithWASM creates a new audio implementation for WebAssembly.
func WithWASM() Audio {
	ctx := js.Global().Get("AudioContext")
	if ctx.IsUndefined() || ctx.IsNull() {
		// Safari fallback
		ctx = js.Global().Get("webkitAudioContext")
	}

	var audioCtx js.Value
	if !ctx.IsUndefined() && !ctx.IsNull() {
		audioCtx = ctx.New()
	}

	return &WASMAudio{
		AudioContext: audioCtx,
	}
}

func (a *WASMAudio) Init() error {
	if a.AudioContext.IsUndefined() || a.AudioContext.IsNull() {
		a.AudioContext = js.Global().Get("AudioContext").New()
	}
	return nil
}

func (a *WASMAudio) Play() error {
	if a.playing {
		return nil
	}

	ctx := a.AudioContext
	if ctx.IsUndefined() || ctx.IsNull() {
		ctx = js.Global().Get("AudioContext").New()
		a.AudioContext = ctx
	}

	// Resume context (must be user-triggered!)
	if ctx.Get("state").String() == "suspended" {
		ctx.Call("resume")
	}

	osc := ctx.Call("createOscillator")
	gain := ctx.Call("createGain")

	// Correct parameter setting
	osc.Set("type", "square")
	osc.Get("frequency").Set("value", 440)

	gain.Get("gain").Set("value", 0.1)

	// Connect nodes
	osc.Call("connect", gain)
	gain.Call("connect", ctx.Get("destination"))

	osc.Call("start")

	a.Oscillator = osc
	a.GainNode = gain
	a.playing = true

	return nil
}

func (a *WASMAudio) Pause() error {
	if !a.playing {
		return nil
	}

	if !a.Oscillator.IsUndefined() && !a.Oscillator.IsNull() {
		a.Oscillator.Call("stop")
	}

	a.Oscillator = js.Value{}
	a.GainNode = js.Value{}
	a.playing = false

	return nil
}

func (a *WASMAudio) Close() error {
	return a.Pause()
}
