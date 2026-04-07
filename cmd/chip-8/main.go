//go:build !wasm || !js

// Package main provides a CHIP-8 emulator and assembler command-line tool.
//
// The chip-8 command supports two subcommands:
//   - run: Execute a CHIP-8 ROM file
//   - compile: Assemble .asm files into a CHIP-8 ROM file
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arpitchakladar/chip-8/internal/assembler"
	"github.com/arpitchakladar/chip-8/internal/emulator"
)

// defaultClockSpeed is the default CPU clock speed in Hz.
const defaultClockSpeed = 100000

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := strings.ToLower(os.Args[1])

	switch command {
	case "run":
		runCommand := flag.NewFlagSet("run", flag.ExitOnError)
		clockSpeed := runCommand.Uint(
			"c",
			defaultClockSpeed,
			"CPU clock speed in Hz",
		)
		if err := runCommand.Parse(os.Args[2:]); err != nil {
			fmt.Println("Error: invalid arguments")
			runCommand.Usage()
			os.Exit(1)
		}

		if runCommand.NArg() != 1 {
			fmt.Println("Error: expected exactly one ROM file")
			runCommand.Usage()
			os.Exit(1)
		}
		statusCode := runEmulator(runCommand.Arg(0), uint32(*clockSpeed))
		os.Exit(statusCode)

	case "compile":
		compileCommand := flag.NewFlagSet("compile", flag.ExitOnError)
		outputPath := compileCommand.String("o", "", "Output file path")
		if err := compileCommand.Parse(os.Args[2:]); err != nil {
			fmt.Println("Error: invalid arguments")
			compileCommand.Usage()
			os.Exit(1)
		}

		if compileCommand.NArg() < 1 {
			fmt.Println("Error: expected at least one .asm file")
			compileCommand.Usage()
			os.Exit(1)
		}
		compileAssembly(compileCommand.Args(), *outputPath)

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}

// printUsage prints the usage information for the chip-8 command.
func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  chip-8 run [-c <hz>] <rom>   - Run a .ch8 file")
	fmt.Println(
		"  chip-8 compile [-o <out>] <files...> - Assemble .asm files to .ch8",
	)
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  chip-8 run rom.ch8")
	fmt.Println("  chip-8 run -c 500 rom.ch8       # Run at 500 Hz")
	fmt.Println("  chip-8 compile -o out.ch8 a.asm b.asm")
}

// runEmulator loads and runs a CHIP-8 ROM file with the specified clock speed.
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

// compileAssembly assembles one or more .asm files into a CHIP-8 ROM file.
// Files containing __START marker are placed at the beginning, files with __END
// marker are placed at the end, and remaining files are placed in between.
// If no output path is specified, it defaults to the first input filename with
// .ch8 extension, or "combined.ch8" if multiple files are provided.
func compileAssembly(filePaths []string, outputPath string) {
	orderedPaths := orderByMarkers(filePaths)
	allContent := readAllFiles(orderedPaths)

	if outputPath == "" {
		outputPath = determineOutputPath(orderedPaths)
	}

	assembleAndWrite(allContent, outputPath)
}

// orderByMarkers reads all input files, categorizes them by __START and __END
// markers, and returns them in the correct order for assembly.
func orderByMarkers(filePaths []string) []string {
	var startFiles, endFiles, regularFiles []string
	hasStartMarker := false
	hasEndMarker := false

	for _, path := range filePaths {
		kind := categorizeFile(path)

		switch kind {
		case "both":
			startFiles = append(startFiles, path)
			hasStartMarker = true
			hasEndMarker = true
		case "start":
			startFiles = append(startFiles, path)
			hasStartMarker = true
		case "end":
			endFiles = append(endFiles, path)
			hasEndMarker = true
		case "regular":
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

	return append(startFiles, append(regularFiles, endFiles...)...)
}

func categorizeFile(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "File Error: %v\n", err)
		os.Exit(1)
	}
	hasStart := strings.Contains(string(content), "__START")
	hasEnd := strings.Contains(string(content), "__END")

	if hasStart && hasEnd {
		return "both"
	}
	if hasStart {
		return "start"
	}
	if hasEnd {
		return "end"
	}
	return "regular"
}

// readAllFiles reads all files from the given paths and concatenates their contents.
func readAllFiles(filePaths []string) strings.Builder {
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

	return allContent
}

// determineOutputPath determines the output path for the compiled ROM file.
func determineOutputPath(filePaths []string) string {
	outputPath := strings.TrimSuffix(
		filePaths[0],
		filepath.Ext(filePaths[0]),
	) + ".ch8"
	if len(filePaths) > 1 {
		outputPath = "combined.ch8"
	}
	return outputPath
}

// assembleAndWrite assembles the combined source and writes the result to the output path.
func assembleAndWrite(source strings.Builder, outputPath string) {
	fmt.Printf("Assembling...\n")

	asm := assembler.New(source.String())
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
