package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arpitchakladar/chip-8/internal/assembler"
	"github.com/arpitchakladar/chip-8/internal/emulator"
)

// chip-8 is a CHIP-8 emulator and assembler written in Go with SDL2.
// It provides a command-line interface to run CHIP-8 ROMs or assemble .asm files.
//
// Usage:
//
//	chip-8 run <path-to-rom>             - Run a .ch8 file
//	chip-8 compile [-o <output>] <file>  - Assemble .asm file(s) to .ch8
func main() {
	// Validate minimum arguments: command + at least one argument
	if len(os.Args) < 3 {
		printUsage()
		return
	}

	command := strings.ToLower(os.Args[1])
	args := os.Args[2:]

	// Parse optional -o flag for output path
	outputPath := ""

	for i := 0; i < len(args); i++ {
		if args[i] == "-o" && i+1 < len(args) {
			outputPath = args[i+1]
			args = append(args[:i], args[i+2:]...)
			i--
		}
	}

	// Dispatch to appropriate command handler
	switch command {
	case "run":
		if len(args) != 1 {
			printUsage()
			return
		}
		runEmulator(args[0])
	case "compile":
		if len(args) < 1 {
			printUsage()
			return
		}
		compileAssembly(args, outputPath)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}

// printUsage displays the command-line usage information.
func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("run <path-to-rom>             - Runs a .ch8 file")
	fmt.Println("compile [-o <output>] <path-to-asm> ... - Compiles one or more .asm files to .ch8")
}

// runEmulator loads and runs a CHIP-8 ROM file.
// It creates a new emulator with 100kHz clock speed, loads the ROM into memory,
// and starts the main emulation loop which handles display, audio, and input.
func runEmulator(path string) {
	// Create emulator with 100kHz clock speed (100,000 instructions per second)
	vm := emulator.WithSDL(100000)
	fmt.Printf("Starting emulator with: %s\n", path)

	// Read the ROM file into memory
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Runtime Error: %v\n", err)
		os.Exit(1)
	}

	// Load ROM data into memory starting at ProgramStart (0x200)
	if err := vm.LoadROM(content); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load ROM: %v\n", err)
	}

	// Run the emulator main loop (blocks until window is closed or error occurs)
	if err := vm.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime Error: %v\n", err)
		os.Exit(1)
	}
}

// compileAssembly assembles one or more .asm files into a .ch8 bytecode file.
//
// The assembler requires exactly one file with __START and one with __END markers.
// Files are ordered: start files (with __START) -> regular files -> end files (with __END).
//
// If -o is not specified:
//   - Single file: outputs to <input>.ch8
//   - Multiple files: outputs to combined.ch8
func compileAssembly(filePaths []string, outputPath string) {
	// Categorize files by their marker content
	var startFiles, endFiles, regularFiles []string

	hasStartMarker := false
	hasEndMarker := false

	// First pass: read each file and check for __START and __END markers
	for _, path := range filePaths {
		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "File Error: %v\n", err)
			os.Exit(1)
		}
		hasStart := strings.Contains(string(content), "__START")
		hasEnd := strings.Contains(string(content), "__END")

		if hasStart && hasEnd {
			startFiles = append(startFiles, path)
			hasStartMarker = true
			hasEndMarker = true
			continue
		}

		if hasStart {
			startFiles = append(startFiles, path)
			hasStartMarker = true
		}
		if hasEnd {
			endFiles = append(endFiles, path)
			hasEndMarker = true
		}
		if !hasStart && !hasEnd {
			regularFiles = append(regularFiles, path)
		}
	}

	// Validate that required markers exist
	if !hasStartMarker {
		fmt.Fprintf(os.Stderr, "Error: No file contains __START marker\n")
		os.Exit(1)
	}

	if !hasEndMarker {
		fmt.Fprintf(os.Stderr, "Error: No file contains __END marker\n")
		os.Exit(1)
	}

	// Order files: start -> regular -> end
	filePaths = append(startFiles, append(regularFiles, endFiles...)...)

	// Concatenate all file contents
	var allContent strings.Builder

	for _, path := range filePaths {
		fmt.Printf("Reading %s...\n", path)
		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "File Error: %v\n", err)
			os.Exit(1)
		}
		allContent.Write(content)
		allContent.WriteString("\n")
	}

	// Determine output path if not specified
	if outputPath == "" {
		outputPath = strings.TrimSuffix(filePaths[0], filepath.Ext(filePaths[0])) + ".ch8"
		if len(filePaths) > 1 {
			outputPath = "combined.ch8"
		}
	}

	// Assemble the combined source
	fmt.Printf("Assembling...\n")

	asm := assembler.New(allContent.String())
	binary, err := asm.Assemble()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Assembly Error: %v\n", err)
		os.Exit(1)
	}

	// Write the compiled bytecode to file
	if err := os.WriteFile(outputPath, binary, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Write Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Success! Output saved to: %s\n", outputPath)
}
