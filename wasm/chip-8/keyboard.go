//go:build wasm && js

package main

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"syscall/js"

	"github.com/arpitchakladar/chip-8/internal/emulator"
)

// KeyboardHandler holds JavaScript event handler functions for keyboard input
// and manages the canvas element.
type KeyboardHandler struct {
	Canvas       js.Value
	vm           *emulator.Emulator
	KeyDown      js.Func
	KeyUp        js.Func
	Click        js.Func
	MouseEnter   js.Func
	MouseLeave   js.Func
	Blur         js.Func
	WindowBlur   js.Func
	isActiveLock sync.Mutex
	isActive     atomic.Bool
}

// NewKeyboardHandler creates a new KeyboardHandler with the given canvas and emulator.
func NewKeyboardHandler(
	canvas js.Value,
	vm *emulator.Emulator,
) *KeyboardHandler {
	return &KeyboardHandler{
		Canvas: canvas,
		vm:     vm,
	}
}

// Setup attaches keyboard and focus event handlers to the canvas.
// It handles key input, focus management, and clearing stuck keys on blur.
func (h *KeyboardHandler) Setup() error {
	h.isActiveLock.Lock()
	if h.isActive.Load() {
		h.isActiveLock.Unlock()
		return fmt.Errorf("keyboard events are already being handled")
	}

	h.Canvas.Set("tabIndex", 0)

	h.Click = h.createClickHandler()
	h.MouseEnter = h.createFocusHandler()
	h.MouseLeave = h.createBlurHandler()
	h.Blur = h.createClearKeysHandler()
	h.WindowBlur = h.createClearKeysHandler()
	h.KeyDown = h.createKeyHandler(true)
	h.KeyUp = h.createKeyHandler(false)

	h.Canvas.Call("addEventListener", "click", h.Click)
	h.Canvas.Call("addEventListener", "mouseenter", h.MouseEnter)
	h.Canvas.Call("addEventListener", "mouseleave", h.MouseLeave)
	h.Canvas.Call("addEventListener", "blur", h.Blur)
	h.Canvas.Call("addEventListener", "keydown", h.KeyDown)
	h.Canvas.Call("addEventListener", "keyup", h.KeyUp)
	js.Global().Call("addEventListener", "blur", h.WindowBlur)

	h.isActive.Store(true)
	h.isActiveLock.Unlock()

	return nil
}

// Remove removes all keyboard event handlers from the canvas and window.
func (h *KeyboardHandler) Remove() error {
	h.isActiveLock.Lock()
	if !h.isActive.Load() {
		h.isActiveLock.Unlock()
		return fmt.Errorf("keyboard are not being handled")
	}

	h.Canvas.Call("removeEventListener", "click", h.Click)
	h.Canvas.Call("removeEventListener", "mouseenter", h.MouseEnter)
	h.Canvas.Call("removeEventListener", "mouseleave", h.MouseLeave)
	h.Canvas.Call("removeEventListener", "blur", h.Blur)
	h.Canvas.Call("removeEventListener", "keydown", h.KeyDown)
	h.Canvas.Call("removeEventListener", "keyup", h.KeyUp)
	js.Global().Call("removeEventListener", "blur", h.WindowBlur)

	h.isActive.Store(false)
	h.isActiveLock.Unlock()

	return nil
}

// IsActive returns whether the keyboard handler is currently active.
func (h *KeyboardHandler) IsActive() bool {
	return h.isActive.Load()
}

// SendKey sets the key state for the given CHIP-8 key index.
func (h *KeyboardHandler) SendKey(key uint8, pressed bool) {
	h.vm.Keyboard.SetKey(key, pressed)
}

// createClickHandler creates a click handler that focuses the canvas.
func (h *KeyboardHandler) createClickHandler() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		h.Canvas.Call("focus")
		return nil
	})
}

// createFocusHandler creates a mouseenter handler that focuses the canvas.
func (h *KeyboardHandler) createFocusHandler() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		h.Canvas.Call("focus")
		return nil
	})
}

// createBlurHandler creates a mouseleave handler that blurs the canvas.
func (h *KeyboardHandler) createBlurHandler() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		h.Canvas.Call("blur")
		return nil
	})
}

// createClearKeysHandler creates a blur handler that clears all pressed keys.
func (h *KeyboardHandler) createClearKeysHandler() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		for i := range 16 {
			h.vm.Keyboard.SetKey(byte(i), false)
		}
		return nil
	})
}

// createKeyHandler creates a keydown or keyup handler for CHIP-8 keyboard input.
func (h *KeyboardHandler) createKeyHandler(pressed bool) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		if !h.vm.IsRunning() {
			return nil
		}
		event := args[0]
		event.Call("preventDefault")
		key := strings.ToLower(event.Get("key").String())
		if chip8Key := keyToChip8(key); chip8Key != nil {
			h.vm.Keyboard.SetKey(*chip8Key, pressed)
		}
		return nil
	})
}

// keyToChip8 maps JavaScript key values to CHIP-8 key indices (0-15).
// Returns nil if the key is not mapped.
func keyToChip8(key string) *byte {
	keyMap := map[string]byte{
		"1": 0x1, "2": 0x2, "3": 0x3, "4": 0xC,
		"q": 0x4, "w": 0x5, "e": 0x6, "r": 0xD,
		"a": 0x7, "s": 0x8, "d": 0x9, "f": 0xE,
		"z": 0xA, "x": 0x0, "c": 0xB, "v": 0xF,
	}
	if ch, ok := keyMap[key]; ok {
		return &ch
	}
	return nil
}
