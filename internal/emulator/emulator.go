package emulator

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/arpitchakladar/chip-8/internal/emulator/audio"
	"github.com/arpitchakladar/chip-8/internal/emulator/cpu"
	"github.com/arpitchakladar/chip-8/internal/emulator/display"
	"github.com/arpitchakladar/chip-8/internal/emulator/keyboard"
	"github.com/arpitchakladar/chip-8/internal/emulator/memory"

	"github.com/veandco/go-sdl2/sdl"
)

const ProgramStart = 0x200

type Emulator struct {
	CPU        *cpu.CPU
	Memory     *memory.Memory
	Display    display.Display
	Keyboard   keyboard.Keyboard
	Audio      audio.Audio
	ClockSpeed uint32
	MemoryLock sync.Mutex
}

func WithSDL(clockSpeed uint32) *Emulator {
	e := &Emulator{
		CPU:        cpu.New(),
		Memory:     memory.New(),
		Display:    display.New(),
		Keyboard:   keyboard.New(),
		Audio:      audio.New(),
		MemoryLock: sync.Mutex{},
		ClockSpeed: clockSpeed,
	}

	e.Memory.LoadFontSet()
	e.CPU.ProgramCounter = ProgramStart
	return e
}

func (e *Emulator) LoadROM(romData []byte) error {
	for i, b := range romData {
		if err := e.Memory.Write(ProgramStart+uint16(i), b); err != nil {
			return err
		}
	}
	return nil
}

func (e *Emulator) Run() error {
	if err := e.Display.Init(); err != nil {
		return fmt.Errorf("failed to init display: %w", err)
	}

	if err := e.Audio.Init(); err != nil {
		fmt.Printf("Warning: Audio failed to init: %v\n", err)
	}

	defer func() {
		if err := e.Display.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing display: %v\n", err)
		}
		e.Audio.Close()
	}()

	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error, 1)
	defer cancel()

	go e.runCPU(ctx, errChan)
	e.runDisplay(ctx, errChan)

	return <-errChan
}

func (e *Emulator) runDisplay(ctx context.Context, errChan chan<- error) {
	uiClock := time.NewTicker(time.Second / 60)
	defer uiClock.Stop()

	sdlKeyboard, isSDL := e.Keyboard.(*keyboard.SDLKeyboard)

	for {
		select {
		case <-ctx.Done():
			return
		case <-uiClock.C:
			if isSDL {
				e.handleSDLEvents(sdlKeyboard, errChan)
			}

			e.MemoryLock.Lock()
			timerErr := e.updateTimers()
			displayErr := e.Display.Present()
			e.MemoryLock.Unlock()

			if timerErr != nil {
				errChan <- timerErr
				return
			}
			if displayErr != nil {
				errChan <- displayErr
				return
			}
		}
	}
}

func (e *Emulator) handleSDLEvents(sdlKeyboard *keyboard.SDLKeyboard, errChan chan<- error) {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch t := event.(type) {
		case *sdl.QuitEvent:
			errChan <- nil
			return
		case *sdl.KeyboardEvent:
			sdlKeyboard.HandleKeyboard(t)
		}
	}
}

func (e *Emulator) runCPU(ctx context.Context, errChan chan<- error) {
	cpuClock := time.NewTicker(time.Second / time.Duration(e.ClockSpeed))
	defer cpuClock.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cpuClock.C:
			e.MemoryLock.Lock()
			err := e.tick()
			e.MemoryLock.Unlock()
			if err != nil {
				errChan <- err
				return
			}
		}
	}
}

func (e *Emulator) tick() error {
	hi, err := e.Memory.Read(e.CPU.ProgramCounter)
	if err != nil {
		return err
	}
	lo, err := e.Memory.Read(e.CPU.ProgramCounter + 1)
	if err != nil {
		return err
	}
	opcode := uint16(hi)<<8 | uint16(lo)

	e.CPU.ProgramCounter += 2

	return e.CPU.Execute(opcode, e.Memory, e.Display, e.Keyboard)
}

func (e *Emulator) updateTimers() error {
	if e.CPU.SoundTimer > 0 {
		if err := e.Audio.GenerateBeep(); err != nil {
			return err
		}
		e.Audio.Play()
		e.CPU.SoundTimer--
	} else {
		e.Audio.Pause()
	}

	if e.CPU.DelayTimer > 0 {
		e.CPU.DelayTimer--
	}

	return nil
}
