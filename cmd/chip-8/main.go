//go:build !wasm || !js

// Package main provides a CHIP-8 emulator and assembler command-line tool.
//
// Usage:
//   - chip-8 run [-c <hz>] <rom>   - Run a .ch8 file
//   - chip-8 assemble [-o <out>] <files...> - Assemble .asm files to .ch8
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// defaultClockSpeed is the default CPU clock speed in Hz.
const defaultClockSpeed = 100000

// main is the entry point for the chip-8 command.
// It dispatches to subcommands: run or compile.
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

	case "assemble":
		assembleCommand := flag.NewFlagSet("assemble", flag.ExitOnError)
		outputPath := assembleCommand.String("o", "", "Output file path")
		if err := assembleCommand.Parse(os.Args[2:]); err != nil {
			fmt.Println("Error: invalid arguments")
			assembleCommand.Usage()
			os.Exit(1)
		}

		if assembleCommand.NArg() < 1 {
			fmt.Println("Error: expected at least one .asm file")
			assembleCommand.Usage()
			os.Exit(1)
		}
		assembleAssembly(assembleCommand.Args(), *outputPath)

	case "help":
		printUsage()

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
		"  chip-8 assemble [-o <out>] <files...> - Assemble .asm files to .ch8",
	)
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  chip-8 run rom.ch8")
	fmt.Println("  chip-8 run -c 500 rom.ch8       # Run at 500 Hz")
	fmt.Println("  chip-8 assemble -o out.ch8 a.asm b.asm")
}
