//go:build !wasm || !js

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arpitchakladar/chip-8/internal/assembler"
)

func compileAssembly(filePaths []string, outputPath string) {
	orderedPaths := orderByMarkers(filePaths)
	allContent := readAllFiles(orderedPaths)

	if outputPath == "" {
		outputPath = determineOutputPath(orderedPaths)
	}

	assembleAndWrite(allContent, outputPath)
}

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
