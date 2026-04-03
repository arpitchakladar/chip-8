package lexer

import "fmt"

type LexerError struct {
	LineNumber uint16
	Message    string
}

func (e *LexerError) Error() string {
	return fmt.Sprintf("line %d: %s", e.LineNumber, e.Message)
}
