package lexer

import "fmt"

// MissingStartLabelError occurs when no __START label is found.
type MissingStartLabelError struct{}

func (e *MissingStartLabelError) Error() string {
	return "missing __START label: file must contain a __START marker"
}

// MissingEndLabelError occurs when no __END label is found.
type MissingEndLabelError struct{}

func (e *MissingEndLabelError) Error() string {
	return "missing __END label: file must contain a __END marker"
}

// StartAfterCodeError occurs when __START appears after instructions.
type StartAfterCodeError struct {
	LineNumber uint16
}

func (e *StartAfterCodeError) Error() string {
	return fmt.Sprintf("__START label must be defined before any instructions (line %d)", e.LineNumber)
}

// EndAfterCodeError occurs when instructions appear after __END.
type EndAfterCodeError struct {
	LineNumber uint16
}

func (e *EndAfterCodeError) Error() string {
	return fmt.Sprintf("no instructions allowed after __END (line %d)", e.LineNumber)
}
