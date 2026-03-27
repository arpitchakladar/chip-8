package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arpitchakladar/chip-8/internal/assembler"
	"github.com/arpitchakladar/chip-8/internal/emulator"
)

func main() {
	if len(os.Args) < 3 {
		printUsage()
		return
	}

	command := strings.ToLower(os.Args[1])
	filePath := os.Args[2]

	switch command {
	case "run":
		runEmulator(filePath)
	case "compile":
		compileAssembly(filePath)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("run <path-to-rom>     - Runs a .ch8 file")
	fmt.Println("compile <path-to-asm> - Compiles .asm to .ch8")
}

func runEmulator(path string) {
	vm := emulator.WithClockSpeed(1000)
	fmt.Printf("Starting emulator with: %s\n", path)

	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Runtime Error: %v\n", err)
		os.Exit(1)
	}

	if err := vm.Run(content); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime Error: %v\n", err)
		os.Exit(1)
	}
}

func compileAssembly(path string) {
	fmt.Printf("Compiling %s...\n", path)

	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "File Error: %v\n", err)
		os.Exit(1)
	}

	asm := assembler.New()
	binary, err := asm.Assemble(string(content))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Assembly Error: %v\n", err)
		os.Exit(1)
	}

	outputPath := strings.TrimSuffix(path, filepath.Ext(path)) + ".ch8"
	if err := os.WriteFile(outputPath, binary, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Write Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Success! Output saved to: %s\n", outputPath)
}
