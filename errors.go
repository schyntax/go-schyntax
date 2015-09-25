package schyntax

import "github.com/schyntax/go-schyntax/internals"

type SchyntaxError interface {
	Error() string
	Input() string
	Index() int
}

var _ SchyntaxError = &internals.ParseError{}
var _ SchyntaxError = &ValidTimeNotFoundError{}
var _ SchyntaxError = &InternalError{}

type ValidTimeNotFoundError struct {
	input string
}

func (e *ValidTimeNotFoundError) Error() string {
	return "A valid time was not found for the schedule."
}

func (e *ValidTimeNotFoundError) Input() string {
	return e.input
}

func (e *ValidTimeNotFoundError) Index() int {
	return 0
}

type InternalError struct {
	message string
	input   string
}

func newInternalError(msg string, input string) *InternalError {
	msg += " This indicates a bug in Schyntax. Please open an issue on github."
	return &InternalError{msg, input}
}

func (e *InternalError) Error() string {
	return e.message
}

func (e *InternalError) Input() string {
	return e.input
}

func (e *InternalError) Index() int {
	return 0
}
