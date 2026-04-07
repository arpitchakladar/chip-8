//go:build !wasm || !js

// Package main provides a CHIP-8 emulator and assembler command-line tool.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

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
