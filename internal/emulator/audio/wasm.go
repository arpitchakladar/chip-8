//go:build wasm && js

// Package audio provides a WebAssembly-compatible audio implementation
// using the Web Audio API.
package audio

import (
	"syscall/js"
)

// WASMAudio implements the Audio interface for WebAssembly/JS environments.
// It uses the Web Audio API to generate square wave tones for CHIP-8 sound.
type WASMAudio struct {
	AudioContext js.Value // Web Audio API AudioContext
	Oscillator   js.Value // Active oscillator node (if playing)
	GainNode     js.Value // Gain node for volume control
	playing      bool     // Track whether audio is currently playing
}

// WithWASM creates a new Audio implementation for WebAssembly.
// Initializes the Web Audio API AudioContext for generating sound.
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
	// Lazily initialize AudioContext on first use (browser requires user gesture)
	if a.AudioContext.IsUndefined() || a.AudioContext.IsNull() {
		a.AudioContext = js.Global().Get("AudioContext").New()
	}
	return nil
}

// Play starts playing a square wave tone using the Web Audio API.
// Creates an oscillator and gain node, connects them to the audio destination.
// Safe to call multiple times; returns early if already playing.
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

// Pause stops the currently playing sound by stopping the oscillator.
// Safe to call multiple times; returns early if not playing.
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

// Close stops any playing audio and releases resources.
// Implements the io.Closer interface.
func (a *WASMAudio) Close() error {
	return a.Pause()
}
