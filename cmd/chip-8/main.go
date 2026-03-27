package main

import (
	"fmt"
	"os"

	"github.com/arpitchakladar/chip-8/internal/system"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: <path-to-rom>")
		return
	}

	chip8 := system.WithClockSpeed(1000)

	// Everything is now encapsulated!
	if err := chip8.Run(os.Args[1]); err != nil {
		fmt.Fprintf(os.Stderr, "Emulation stopped: %v\n", err)
		os.Exit(1)
	}
}
