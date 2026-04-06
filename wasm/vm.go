//go:build wasm && js

package main

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/arpitchakladar/chip-8/internal/emulator"
)

var (
	defaultClockSpeed = uint32(100000)
	VMCounter         = uint32(0)
	VMs               = make(map[string]*WASMVM)
)

type WASMVM struct {
	Id       string
	Emulator *emulator.Emulator
	Cancel   context.CancelFunc
	ErrChan  chan error
}

func CreateVM() string {
	atomic.AddUint32(&VMCounter, 1)
	vm := &WASMVM{
		Id:       fmt.Sprintf("chip-8-vm-%d", atomic.LoadUint32(&VMCounter)),
		Emulator: emulator.WithWASM(defaultClockSpeed),
		ErrChan:  make(chan error, 1),
	}
	VMs[vm.Id] = vm

	return vm.Id
}

func (wvm *WASMVM) LoadROM(romData []byte) error {
	if err := wvm.Emulator.LoadROM(romData); err != nil {
		return err
	}

	return nil
}

func (wvm *WASMVM) Run() error {
	if wvm.Cancel != nil {
		// TODO: ADD ERROR HERE
		fmt.Printf("Error: The vm %s is already running.", wvm.Id)
		return nil
	}
	vmRunningContext, cancelVMRunningContext := context.WithCancel(
		context.Background(),
	)
	wvm.Cancel = cancelVMRunningContext

	go func() {
		defer cancelVMRunningContext()
		if err := wvm.Emulator.Run(vmRunningContext); err != nil {
			wvm.ErrChan <- err
		}
	}()

	return nil
}

func (wvm *WASMVM) Destroy() error {
	if wvm.Cancel == nil {
		// TODO: ADD ERROR HERE
		fmt.Printf("Error: The vm %s is already running.", wvm.Id)
		return nil
	}
	wvm.Cancel()

	return nil
}
