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
// It identifies all labels and their addresses, then returns the label map
// and a list of instruction lines for the parser.
func (l *Lexer) ScanLabels() (map[string]uint16, []Line, error) {
	labels := make(map[string]uint16)
	var program []Line

	i := uint16(0)
	seenStart := false
	seenEnd := false

	for raw := range strings.SplitSeq(l.Source, "\n") {
		i++
		content := strings.Split(raw, ";")[0]
		content = strings.TrimSpace(content)
		if content == "" {
			continue
		}

		if label, found := strings.CutSuffix(content, ":"); found {
			switch label {
			case "__START":
				seenStart = true
				labels[label] = l.CurrentAddr
			case "__END":
				seenEnd = true
				labels[label] = l.CurrentAddr
			default:
				labels[label] = l.CurrentAddr
			}
		} else {
			// __START should be before anything else
			if !seenStart {
				return nil, nil, &StartAfterCodeError{LineNumber: i}
			}
			// __END should be after everything else
			if seenEnd {
				return nil, nil, &EndAfterCodeError{LineNumber: i}
			}

			parts := strings.Fields(strings.ReplaceAll(content, ",", " "))

			if len(parts) > 0 {
				mnemonic := strings.ToUpper(parts[0])
				program = append(program, Line{
					Mnemonic:   mnemonic,
					Args:       parts[1:],
					Address:    l.CurrentAddr,
					LineNumber: i,
				})
				if mnemonic != "DB" {
					l.CurrentAddr += 2
				} else {
					l.CurrentAddr++
				}
			}
		}
	}

	// Validate required markers are present
	if !seenStart {
		return nil, nil, &MissingStartLabelError{}
	}
	if !seenEnd {
		return nil, nil, &MissingEndLabelError{}
	}

	return labels, program, nil
}
