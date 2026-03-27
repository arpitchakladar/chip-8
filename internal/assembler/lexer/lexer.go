package lexer

import "strings"

type Line struct {
	Mnemonic string
	Args     []string
	Address  uint16
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

	for raw := range strings.SplitSeq(l.Source, "\n") {
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
				program = append(program, Line{
					Mnemonic: strings.ToUpper(parts[0]),
					Args:     parts[1:],
					Address:  l.CurrentAddr,
				})
				l.CurrentAddr += 2
			}
		}
	}
	return labels, program
}
