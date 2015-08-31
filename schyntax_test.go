package schyntax

import (
	"./internals"
	"fmt"
	"testing"
)

func TestOne(t *testing.T) {
	lex := internals.NewLexer("min(*%5)")
	for tok := lex.Advance(); tok.Type != internals.TokenTypeEndOfInput; tok = lex.Advance() {
		fmt.Println(tok.Type.Name(), tok.RawValue)
	}
}
