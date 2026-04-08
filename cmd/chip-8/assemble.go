//go:build !wasm || !js

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arpitchakladar/chip-8/internal/assembler"
)

// assembleAssembly assembles one or more .asm files into a CHIP-8 ROM file.
// Files are ordered by __START and __END markers: start files first,
// then regular files, then end files.
func assembleAssembly(filePaths []string, outputPath string) {
	orderedPaths := orderByMarkers(filePaths)
	allContent := readAllFiles(orderedPaths)

	if outputPath == "" {
		outputPath = determineOutputPath(orderedPaths)
	}

	assembleAndWrite(allContent, outputPath)
}

// orderByMarkers reads all input files, categorizes them by __START and __END
// markers, and returns them in the correct order for assembly.
// Exits with an error if no file contains __START or __END.
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

// categorizeFile determines whether a file contains __START, __END, both, or neither.
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
// If only one input file, uses that name with .ch8 extension.
// If multiple files, defaults to "combined.ch8".
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
