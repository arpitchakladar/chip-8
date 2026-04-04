package lexer

import "strings"

// Lexer tokenizes CHIP-8 assembly source code and performs the first pass of assembly.
// It scans for labels and collects them into a map for the parser to use.
// It holds the state for tokenizing assembly source code.
type Lexer struct {
	// Source is the raw assembly source code.
	Source string
	// CurrentAddr tracks the current address during scanning.
	CurrentAddr uint16
}

// Line represents a single instruction line parsed from the source.
type Line struct {
	// Mnemonic is the instruction name (e.g., "LD", "JP", "ADD").
	Mnemonic string
	// Args contains the instruction arguments.
	Args []string
	// Address is the memory address where this instruction will be placed.
	Address uint16
	// LineNumber is the original source line number (for error reporting).
	LineNumber uint16
}

// New creates a new Lexer with the given source code and starting address.
func New(source string, currentAddr uint16) *Lexer {
	return &Lexer{
		Source:      source,
		CurrentAddr: currentAddr,
	}
}

// ScanLabels performs the first pass of assembly.
// It scans the source code for labels and builds a map of label names to their addresses.
// It also collects all instruction lines for the parser to process in the second pass.
//
// The function enforces the following rules:
//   - __START must be defined before any instructions
//   - __END must be defined after all instructions
//   - Both __START and __END markers are required
//
// Returns:
//   - labels: map of label name -> memory address
//   - program: list of instruction lines to be parsed
//   - error: if any validation fails
func (l *Lexer) ScanLabels() (map[string]uint16, []Line, error) {
	// Initialize label map and program list
	labels := make(map[string]uint16)
	var program []Line

	// Line counter for error reporting
	i := uint16(0)
	// Track whether __START and __END markers have been seen
	seenStart := false
	seenEnd := false

	// Iterate through each line of the source code
	for raw := range strings.SplitSeq(l.Source, "\n") {
		i++
		// Remove comments (everything after semicolon)
		content := strings.Split(raw, ";")[0]
		content = strings.TrimSpace(content)
		// Skip empty lines
		if content == "" {
			continue
		}

		// Check if this line defines a label (ends with colon)
		if label, found := strings.CutSuffix(content, ":"); found {
			switch label {
			case "__START":
				seenStart = true
				labels[label] = l.CurrentAddr
			case "__END":
				seenEnd = true
				labels[label] = l.CurrentAddr
			default:
				// Store user-defined label with current address
				labels[label] = l.CurrentAddr
			}
		} else {
			// This is an instruction, not a label
			// Validate: __START must appear before any instructions
			if !seenStart {
				return nil, nil, &StartAfterCodeError{LineNumber: i}
			}
			// Validate: no instructions allowed after __END
			if seenEnd {
				return nil, nil, &EndAfterCodeError{LineNumber: i}
			}

			// Parse the instruction: split by whitespace or comma
			parts := strings.Fields(strings.ReplaceAll(content, ",", " "))

			// Only process non-empty instruction lines
			if len(parts) > 0 {
				mnemonic := strings.ToUpper(parts[0])
				program = append(program, Line{
					Mnemonic:   mnemonic,
					Args:       parts[1:],
					Address:    l.CurrentAddr,
					LineNumber: i,
				})
				// Update program counter: DB is 1 byte, all other instructions are 2 bytes
				if mnemonic != "DB" {
					l.CurrentAddr += 2
				} else {
					l.CurrentAddr++
				}
			}
		}
	}

	// Validate that required markers are present
	if !seenStart {
		return nil, nil, &MissingStartLabelError{}
	}
	if !seenEnd {
		return nil, nil, &MissingEndLabelError{}
	}

	return labels, program, nil
}
