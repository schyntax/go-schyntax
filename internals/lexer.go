package internals

import (
	"regexp"
	"strconv"
)

// +gen stringer
type ContextMode int

const (
	ContextModeProgram ContextMode = iota
	ContextModeGroup
	ContextModeExpression
)

type lexMethod func() lexMethod

type Lexer struct {
	contextStack  []ContextMode
	input         string
	index         int
	length        int
	leadingTrivia string
	tokenQueue    TokenQueue
	lexMethod     lexMethod
}

func NewLexer(input string) *Lexer {
	l := &Lexer{}
	l.input = input
	l.length = len(input) // not character count (this is intended)

	l.enterContext(ContextModeProgram)

	l.lexMethod = l.lexList

	return l
}

func (l *Lexer) context() ContextMode {
	return l.contextStack[len(l.contextStack)-1]
}

func (l *Lexer) Input() string {
	return l.input
}

func (l *Lexer) Advance() *Token {
	l.queueNext()
	return l.tokenQueue.Dequeue()
}

func (l *Lexer) Peek() *Token {
	l.queueNext()
	return l.tokenQueue.Peek()
}

func (l *Lexer) queueNext() {
	for l.tokenQueue.IsEmpty() {
		l.consumeWhiteSpace()
		l.lexMethod = l.lexMethod()
	}
}

func (l *Lexer) enterContext(mode ContextMode) {
	l.contextStack = append(l.contextStack, mode)
}

func (l *Lexer) exitContext() {
	length := len(l.contextStack)
	if length == 1 {
		panic("The lexer attempted to exit the last context." + PLEASE_REPORT_BUG_MSG)
	}

	// pop
	l.contextStack = l.contextStack[:length-1]
}

func (l *Lexer) isEndNext() bool {
	return l.index == l.length
}

var s_whitespaceRegex *regexp.Regexp = regexp.MustCompile(`^\s`)

func (l *Lexer) isWhiteSpaceNext() bool {
	return s_whitespaceRegex.MatchString(l.input[l.index:])
}

func (l *Lexer) endOfInput() bool {
	l.consumeWhiteSpace()
	if l.isEndNext() {
		if len(l.contextStack) > 1 {
			panic("Lexer reached the end of the input while in a nested context." + PLEASE_REPORT_BUG_MSG)
		}

		tok := &Token{}
		tok.Type = TokenTypeEndOfInput
		tok.Index = l.index
		tok.RawValue = ""
		tok.Value = ""

		l.consumeToken(tok)
		return true
	}

	return false
}

func (l *Lexer) consumeWhiteSpace() {
	start := l.index
	for !l.isEndNext() && l.isWhiteSpaceNext() {
		l.index++
	}

	l.leadingTrivia += l.input[start:l.index]
}

func (l *Lexer) isNextTerm(term *Terminal) bool {
	l.consumeWhiteSpace()
	return term.GetToken(l.input, l.index) != nil
}

func (l *Lexer) consumeTerm(term *Terminal) {
	l.consumeWhiteSpace()

	tok := term.GetToken(l.input, l.index)
	if tok == nil {
		panic(l.unexpectedText(term.TokenType))
	}

	l.consumeToken(tok)
}

func (l *Lexer) consumeOptionalTerm(term *Terminal) bool {
	l.consumeWhiteSpace()

	tok := term.GetToken(l.input, l.index)
	if tok == nil {
		return false
	}

	l.consumeToken(tok)
	return true
}

func (l *Lexer) consumeToken(tok *Token) {
	l.index += len(tok.RawValue)
	tok.LeadingTrivia = l.leadingTrivia
	l.leadingTrivia = ""
	l.tokenQueue.Enqueue(tok)
}

func (l *Lexer) unexpectedText(expectedTokenTypes ...TokenType) string {
	msg := `Unexpected input at index ` + strconv.Itoa(l.index) + `. Was expecting `
	if len(expectedTokenTypes) == 1 {
		msg += expectedTokenTypes[0].Name()
	} else {
		msg += `one of: `
		for i, tokenType := range expectedTokenTypes {
			if i > 0 {
				msg += `, `
			}
			msg += tokenType.Name()
		}
	}

	return msg + "\n\n" + getStringSnippet(l.input, l.index)
}

func (l *Lexer) lexPastEndOfInput() lexMethod {
	panic("Lexer was advanced past the end of the input." + PLEASE_REPORT_BUG_MSG)
}

func (l *Lexer) lexList() lexMethod {
	l.consumeOptionalTerm(TermsComma)

	if l.endOfInput() {
		return l.lexPastEndOfInput
	}

	if l.context() == ContextModeProgram {
		if l.isNextTerm(TermsOpenCurly) {
			return l.lexGroup
		}
	} else if l.context() == ContextModeGroup {
		if l.consumeOptionalTerm(TermsCloseCurly) {
			l.exitContext()
			return l.lexList
		}
	} else if l.context() == ContextModeExpression {
		if l.consumeOptionalTerm(TermsCloseParen) {
			l.exitContext()
			return l.lexList
		}
	}

	if l.context() == ContextModeExpression {
		return l.lexExpressionArgument
	}

	return l.lexExpression
}

func (l *Lexer) lexGroup() lexMethod {
	l.consumeTerm(TermsOpenCurly)
	l.enterContext(ContextModeGroup)
	return l.lexList
}

func (l *Lexer) lexExpression() lexMethod {
	consumedExpName := l.consumeOptionalTerm(TermsSeconds) ||
		l.consumeOptionalTerm(TermsMinutes) ||
		l.consumeOptionalTerm(TermsHours) ||
		l.consumeOptionalTerm(TermsDaysOfWeek) ||
		l.consumeOptionalTerm(TermsDaysOfMonth) ||
		l.consumeOptionalTerm(TermsDates)

	if consumedExpName {
		l.consumeTerm(TermsOpenParen)
		l.enterContext(ContextModeExpression)

		return l.lexList
	}

	panic(l.unexpectedText(TokenTypeExpressionName))
}

func (l *Lexer) lexExpressionArgument() lexMethod {
	l.consumeOptionalTerm(TermsNot)

	if !l.consumeOptionalTerm(TermsWildcard) {
		if l.consumeNumberDayOrDate(false) {
			// might be a range
			if l.consumeOptionalTerm(TermsRangeHalfOpen) || l.consumeOptionalTerm(TermsRangeInclusive) {
				l.consumeNumberDayOrDate(true)
			}
		}
	}

	if l.consumeOptionalTerm(TermsInterval) {
		l.consumeTerm(TermsPositiveInteger)
	}

	return l.lexList
}

func (l *Lexer) consumeNumberDayOrDate(required bool) bool {
	if l.consumeOptionalTerm(TermsPositiveInteger) {
		// this might be a date - check for slashes
		if l.consumeOptionalTerm(TermsForwardSlash) {
			l.consumeTerm(TermsPositiveInteger)

			// might have a year... one more check
			if l.consumeOptionalTerm(TermsForwardSlash) {
				l.consumeTerm(TermsPositiveInteger)
			}
		}

		return true
	}

	if l.consumeOptionalTerm(TermsNegativeInteger) ||
		l.consumeOptionalTerm(TermsSunday) ||
		l.consumeOptionalTerm(TermsMonday) ||
		l.consumeOptionalTerm(TermsTuesday) ||
		l.consumeOptionalTerm(TermsWednesday) ||
		l.consumeOptionalTerm(TermsThursday) ||
		l.consumeOptionalTerm(TermsFriday) ||
		l.consumeOptionalTerm(TermsSaturday) {
		return true
	}

	if required {
		panic(l.unexpectedText(TokenTypePositiveInteger, TokenTypeNegativeInteger, TokenTypeDayLiteral))
	}

	return false
}
