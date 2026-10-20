package lexer_test

import (
	"testing"

	"AxiomLang/pkg/lexer"
)

func TestLexer_RussianSyntax(t *testing.T) {
	input := `
требовать мандат[Console.Out]

структура Пользователь {
    имя: string,
    возраст: int
}

функция удвоить(х: int) -> int {
    пусть результат = х * 2;
    вернуть результат;
}

намерение "Тест русскоязычного синтаксиса AxiomLang" {
    пусть изм счетчик: int = 10;
    если счетчик > 5 {
        пусть удвоенный = удвоить(счетчик);
        print(удвоенный);
    } иначе {
        print("Счетчик мал");
    }
}
`

	l := lexer.New(input)
	var tokens []lexer.Token

	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == lexer.EOF {
			break
		}
	}

	// Verify that Russian tokens were properly classified to AST TokenTypes
	expectedTokenTypes := []lexer.TokenType{
		lexer.REQUIRE, // требовать
		lexer.CAP,     // мандат
		lexer.LBRACKET,
		lexer.IDENT, // Console
		lexer.DOT,
		lexer.IDENT, // Out
		lexer.RBRACKET,
		lexer.STRUCT, // структура
		lexer.IDENT,  // Пользователь
		lexer.LBRACE,
		lexer.IDENT, // имя
		lexer.COLON,
		lexer.IDENT, // string
		lexer.COMMA,
		lexer.IDENT, // возраст
		lexer.COLON,
		lexer.IDENT, // int
		lexer.RBRACE,
		lexer.FN, // функция
	}

	for i, expected := range expectedTokenTypes {
		if tokens[i].Type != expected {
			t.Fatalf("token[%d] mismatch: expected %s, got %s (%q)", i, expected, tokens[i].Type, tokens[i].Literal)
		}
	}
}
