//go:build !wasm || !js

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/arpitchakladar/chip-8/internal/emulator"
)

// runEmulator loads and runs a CHIP-8 ROM file with the specified clock speed.
// Returns the exit status code (0 for success, 1 for error).
func runEmulator(path string, clockSpeed uint32) int {
	vm := emulator.WithSDL(clockSpeed)
	fmt.Printf("Starting emulator with: %s (clock: %d Hz)\n", path, clockSpeed)

	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Runtime Error: %v\n", err)
		return 1
	}

	if err := vm.LoadROM(content); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load ROM: %v\n", err)
		return 1
	}

	emulatorRunningContext, cancelEmulatorRunningContext := context.WithCancel(
		context.Background(),
	)
	defer cancelEmulatorRunningContext()

	if err := vm.Run(emulatorRunningContext); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime Error: %v\n", err)
		return 1
	}

	return 0
}
