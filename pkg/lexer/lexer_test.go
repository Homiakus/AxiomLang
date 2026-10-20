package lexer_test

import (
	"testing"

	"AxiomLang/pkg/lexer"
)

func TestNextToken_Basic(t *testing.T) {
	input := `
require cap[Console.Out]
intent "Test" {
    let mut x: int = 42;
    if x > 10 {
        print("x is large");
    }
}
`
	tests := []struct {
		expectedType    lexer.TokenType
		expectedLiteral string
	}{
		{lexer.REQUIRE, "require"},
		{lexer.CAP, "cap"},
		{lexer.LBRACKET, "["},
		{lexer.IDENT, "Console"},
		{lexer.DOT, "."},
		{lexer.IDENT, "Out"},
		{lexer.RBRACKET, "]"},
		{lexer.INTENT, "intent"},
		{lexer.STRING, "Test"},
		{lexer.LBRACE, "{"},
		{lexer.LET, "let"},
		{lexer.MUT, "mut"},
		{lexer.IDENT, "x"},
		{lexer.COLON, ":"},
		{lexer.IDENT, "int"},
		{lexer.ASSIGN, "="},
		{lexer.INT, "42"},
		{lexer.SEMICOLON, ";"},
		{lexer.IF, "if"},
		{lexer.IDENT, "x"},
		{lexer.GT, ">"},
		{lexer.INT, "10"},
		{lexer.LBRACE, "{"},
		{lexer.IDENT, "print"},
		{lexer.LPAREN, "("},
		{lexer.STRING, "x is large"},
		{lexer.RPAREN, ")"},
		{lexer.SEMICOLON, ";"},
		{lexer.RBRACE, "}"},
		{lexer.RBRACE, "}"},
		{lexer.EOF, ""},
	}

	l := lexer.New(input)

	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d] - tokentype wrong. expected=%q, got=%q (literal=%q)",
				i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}
