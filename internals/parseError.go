package internals

import "strings"

var _ error = &ParseError{}

type ParseError struct {
	message string
	input   string
}

func newParseError(msg string, input string, index int) *ParseError {
	msg += getStringSnippet(input, index)
	return &ParseError{msg, input}
}

func (e *ParseError) Error() string {
	return e.message
}

func (e *ParseError) Input() string {
	return e.input
}

func getStringSnippet(input string, index int) string {
	before := []rune(input[0:index])
	after := []rune(input[index:])

	beforeLen := len(before)
	afterLen := len(after)

	if beforeLen > 20 {
		before = before[beforeLen-20:]
		beforeLen = 20
	}

	if afterLen > 50 {
		after = after[0:50]
		afterLen = 50
	}

	return "\n\n" + string(before) + string(after) + "\n" + strings.Repeat(" ", beforeLen) + "^\n"
}
