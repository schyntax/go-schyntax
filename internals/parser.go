package internals

import (
	"strconv"
)

type Parser struct {
	lexer *Lexer
}

func NewParser(input string) *Parser {
	p := &Parser{}
	p.lexer = NewLexer(input)
	return p
}

func (p *Parser) Input() string {
	return p.lexer.input
}

func (p *Parser) peek() *Token {
	return p.lexer.Peek()
}

func (p *Parser) advance() *Token {
	return p.lexer.Advance()
}

func (p *Parser) expect(tokType TokenType) *Token {
	if !p.isNext(tokType) {
		panic(p.wrongToken(tokType))
	}

	return p.advance()
}

func (p *Parser) isNext(tokType TokenType) bool {
	return p.peek().Type == tokType
}

func (p *Parser) wrongToken(expectedTokenTypes ...TokenType) error {
	next := p.peek()

	msg := `Unexpected token type ` + next.Type.Name() + ` at index ` + strconv.Itoa(next.Index) + `. Was expecting `
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

	return newParseError(msg, p.Input(), next.Index)
}

func (p *Parser) Parse() *ProgramNode {
	return p.parseProgram()
}

func (p *Parser) parseProgram() *ProgramNode {
	program := &ProgramNode{}

	for !p.isNext(TokenTypeEndOfInput) {
		if p.isNext(TokenTypeOpenCurly) {
			program.AddGroup(p.parseGroup())
		} else if p.isNext(TokenTypeExpressionName) {
			program.AddExpression(p.parseExpression())
		} else {
			panic(p.wrongToken(TokenTypeOpenCurly, TokenTypeExpressionName, TokenTypeComma))
		}

		if p.isNext(TokenTypeComma) { // optional comma
			program.AddToken(p.advance())
		}
	}

	program.AddToken(p.expect(TokenTypeEndOfInput))
	return program
}

func (p *Parser) parseGroup() *GroupNode {
	group := &GroupNode{}
	group.AddToken(p.expect(TokenTypeOpenCurly))

	for !p.isNext(TokenTypeCloseCurly) {
		group.AddExpression(p.parseExpression())

		if p.isNext(TokenTypeComma) {
			group.AddToken(p.advance())
		}
	}

	group.AddToken(p.expect(TokenTypeCloseCurly))
	return group
}

func (p *Parser) parseExpression() *ExpressionNode {
	nameTok := p.expect(TokenTypeExpressionName)
	expType := nameTok.ExpressionType
	exp := &ExpressionNode{}
	exp.ExpressionType = expType
	exp.AddToken(p.expect(TokenTypeOpenParen))

	for {
		exp.AddArgument(p.parseArgument(expType))

		if p.isNext(TokenTypeComma) {
			exp.AddToken(p.advance())
		}

		if p.isNext(TokenTypeCloseParen) {
			break
		}
	}

	exp.AddToken(p.expect(TokenTypeCloseParen))
	return exp
}

func (p *Parser) parseArgument(expressionType ExpressionType) *ArgumentNode {
	arg := &ArgumentNode{}

	if p.isNext(TokenTypeNot) {
		arg.IsExclusion = true
		arg.AddToken(p.advance())
	}

	if p.isNext(TokenTypeWildcard) {
		arg.IsWildcard = true
		arg.AddToken(p.advance())
	} else {
		arg.Range = p.parseRange(expressionType)
	}

	if p.isNext(TokenTypeInterval) {
		arg.AddToken(p.advance())
		arg.Interval = p.parseIntegerValue(ExpressionTypeIntervalValue)
	}

	return arg
}

func (p *Parser) parseRange(expressionType ExpressionType) *RangeNode {
	rangeNode := &RangeNode{}
	if expressionType == ExpressionTypeDates {
		rangeNode.Start = p.parseDate()
	} else {
		rangeNode.Start = p.parseIntegerValue(expressionType)
	}

	isRange := false
	if p.isNext(TokenTypeRangeInclusive) {
		isRange = true
	} else if p.isNext(TokenTypeRangeHalfOpen) {
		isRange = true
		rangeNode.IsHalfOpen = true
	}

	if isRange {
		rangeNode.AddToken(p.advance())
		if expressionType == ExpressionTypeDates {
			rangeNode.End = p.parseDate()
		} else {
			rangeNode.End = p.parseIntegerValue(expressionType)
		}
	}

	return rangeNode
}

func (p *Parser) parseIntegerValue(expressionType ExpressionType) *IntegerValueNode {
	val := &IntegerValueNode{}

	if p.isNext(TokenTypePositiveInteger) {
		// positive integer is valid for anything
		tok := p.advance()
		val.AddToken(tok)
		val.Value = p.parseInt(tok)
	} else if p.isNext(TokenTypeNegativeInteger) {
		if expressionType != ExpressionTypeDaysOfMonth && expressionType != ExpressionTypeDaysOfYear {
			panic(newParseError("Negative values are only allowed in dayofmonth and dayofyear expressions.", p.Input(), p.peek().Index))
		}

		tok := p.advance()
		val.AddToken(tok)
		val.Value = p.parseInt(tok)
	} else if p.isNext(TokenTypeDayLiteral) {
		if expressionType != ExpressionTypeDaysOfWeek {
			panic(newParseError("Unexpected day literal. Day literals are only allowed in daysOfWeek expressions.", p.Input(), p.peek().Index))
		}

		tok := p.advance()
		val.AddToken(tok)
		val.Value = dayToInteger(tok.Value)
	} else {
		switch expressionType {
		case ExpressionTypeDaysOfMonth, ExpressionTypeDaysOfYear:
			panic(p.wrongToken(TokenTypePositiveInteger, TokenTypeNegativeInteger))
		case ExpressionTypeDaysOfWeek:
			panic(p.wrongToken(TokenTypePositiveInteger, TokenTypeDayLiteral))
		default:
			panic(p.wrongToken(TokenTypePositiveInteger))
		}
	}

	return val
}

func (p *Parser) parseDate() *DateValueNode {
	date := &DateValueNode{}

	tok := p.expect(TokenTypePositiveInteger)
	date.AddToken(tok)
	one := p.parseInt(tok)

	date.AddToken(p.expect(TokenTypeForwardSlash))

	tok = p.expect(TokenTypePositiveInteger)
	date.AddToken(tok)
	two := p.parseInt(tok)

	three := -1
	if p.isNext(TokenTypeForwardSlash) {
		date.AddToken(p.advance())

		tok = p.expect(TokenTypePositiveInteger)
		date.AddToken(tok)
		three = p.parseInt(tok)
	}

	if three != -1 {
		// date has a year
		date.HasYear = true
		date.Year = one
		date.Month = two
		date.Day = three
	} else {
		// no year
		date.Month = one
		date.Day = two
	}

	return date
}

func dayToInteger(day string) int {
	switch day {
	case "SUNDAY":
		return 1
	case "MONDAY":
		return 2
	case "TUESDAY":
		return 3
	case "WEDNESDAY":
		return 4
	case "THURSDAY":
		return 5
	case "FRIDAY":
		return 6
	case "SATURDAY":
		return 7
	default:
		panic(day + " is not a day.")
	}
}

func (p *Parser) parseInt(tok *Token) int {
	i, err := strconv.Atoi(tok.Value)
	if err != nil {
		msg := "Integer value is too "
		if i < 0 {
			msg += "small."
		} else {
			msg += "large."
		}

		panic(newParseError(msg, p.Input(), tok.Index))
	}

	return i
}
