package lexer

import "strings"

type Line struct {
	Mnemonic string
	Args     []string
	Address  uint16
}

func ScanLabels(source string, startAddr uint16) (map[string]uint16, []Line) {
	labels := make(map[string]uint16)
	var program []Line
	currentAddr := startAddr

	for raw := range strings.SplitSeq(source, "\n") {
		content := strings.Split(raw, ";")[0] // Strip comments
		content = strings.TrimSpace(content)
		if content == "" {
			continue
		}

		if label, found := strings.CutSuffix(content, ":"); found {
			labels[label] = currentAddr
		} else {
			parts := strings.Fields(strings.ReplaceAll(content, ",", " "))

			// Safety check: ensure the line isn't empty before accessing parts[0]
			if len(parts) > 0 {
				program = append(program, Line{
					Mnemonic: strings.ToUpper(parts[0]),
					Args:     parts[1:],
					Address:  currentAddr,
				})
				currentAddr += 2
			}
		}
	}
	return labels, program
}
