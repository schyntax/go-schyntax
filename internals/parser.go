package internals
import "strconv"

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

func (p *Parser) wrongToken(expectedTokenTypes ...TokenType) string {
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

	return msg
}

//func (p *Parser) Parse()
