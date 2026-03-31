package lexer

import "strings"

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

func (l *Lexer) ScanLabels() (map[string]uint16, []Line) {
	labels := make(map[string]uint16)
	var program []Line

	i := uint16(0)
	for raw := range strings.SplitSeq(l.Source, "\n") {
		i++
		content := strings.Split(raw, ";")[0] // Strip comments
		content = strings.TrimSpace(content)
		if content == "" {
			continue
		}

		if label, found := strings.CutSuffix(content, ":"); found {
			labels[label] = l.CurrentAddr
		} else {
			parts := strings.Fields(strings.ReplaceAll(content, ",", " "))

			// Safety check: ensure the line isn't empty before accessing parts[0]
			if len(parts) > 0 {
				mnemonic := strings.ToUpper(parts[0])
				program = append(program, Line{
					Mnemonic:   mnemonic,
					Args:       parts[1:],
					Address:    l.CurrentAddr,
					LineNumber: i,
				})
				// Only DB can create a value of a single binary
				// every other mnemonic is 2 bytes
				if mnemonic != "DB" {
					l.CurrentAddr += 2
				} else {
					l.CurrentAddr++
				}
			}
		}
	}
	return labels, program
}
