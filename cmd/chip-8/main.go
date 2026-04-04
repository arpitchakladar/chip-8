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
func main() {
	if len(os.Args) < 3 {
		printUsage()
		return
	}

	command := strings.ToLower(os.Args[1])
	args := os.Args[2:]

	outputPath := ""

	for i := 0; i < len(args); i++ {
		if args[i] == "-o" && i+1 < len(args) {
			outputPath = args[i+1]
			args = append(args[:i], args[i+2:]...)
			i--
		}
	}

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
// It creates a new emulator instance, loads the ROM, and starts the main loop.
func runEmulator(path string) {
	vm := emulator.New(100000)
	fmt.Printf("Starting emulator with: %s\n", path)

	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Runtime Error: %v\n", err)
		os.Exit(1)
	}

	if err := vm.LoadROM(content); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load ROM: %v\n", err)
	}

	if err := vm.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime Error: %v\n", err)
		os.Exit(1)
	}
}

// compileAssembly assembles one or more .asm files into a .ch8 bytecode file.
// It handles the __START and __END markers to order files correctly.
// If multiple files are provided, they are concatenated in order: start -> regular -> end.
func compileAssembly(filePaths []string, outputPath string) {
	var startFiles, endFiles, regularFiles []string

	hasStartMarker := false
	hasEndMarker := false

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

	if !hasStartMarker {
		fmt.Fprintf(os.Stderr, "Error: No file contains __START marker\n")
		os.Exit(1)
	}

	if !hasEndMarker {
		fmt.Fprintf(os.Stderr, "Error: No file contains __END marker\n")
		os.Exit(1)
	}

	filePaths = append(startFiles, append(regularFiles, endFiles...)...)

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

	if outputPath == "" {
		outputPath = strings.TrimSuffix(filePaths[0], filepath.Ext(filePaths[0])) + ".ch8"
		if len(filePaths) > 1 {
			outputPath = "combined.ch8"
		}
	}

	fmt.Printf("Assembling...\n")

	asm := assembler.New(allContent.String())
	binary, err := asm.Assemble()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Assembly Error: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outputPath, binary, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Write Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Success! Output saved to: %s\n", outputPath)
}
