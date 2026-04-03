package lexer

import (
	"strings"
)

type Line struct {
	Mnemonic   string
	Args       []string
	Address    uint16
	LineNumber uint16
}

type Lexer struct {
	Source      string
	CurrentAddr uint16
}

func New(source string, currentAddr uint16) *Lexer {
	return &Lexer{
		Source:      source,
		CurrentAddr: currentAddr,
	}
}

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
				return nil, nil, &LexerError{
					LineNumber: i,
					Message:    "__START label must be defined before any instructions",
				}
			}
			// __END should be after everything else
			if seenEnd {
				return nil, nil, &LexerError{
					LineNumber: i,
					Message:    "No instructions allowed after __END label",
				}
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
	return labels, program, nil
}
