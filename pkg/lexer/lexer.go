package lexer

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

type TokenType string

const (
	// Special tokens
	EOF     TokenType = "EOF"
	ILLEGAL TokenType = "ILLEGAL"
	COMMENT TokenType = "COMMENT"

	// Literals & Identifiers
	IDENT  TokenType = "IDENT"
	INT    TokenType = "INT"
	FLOAT  TokenType = "FLOAT"
	STRING TokenType = "STRING"

	// Keywords (Multilingual: EN & RU)
	REQUIRE TokenType = "REQUIRE"
	CAP     TokenType = "CAP"
	INTENT  TokenType = "INTENT"
	LET     TokenType = "LET"
	MUT     TokenType = "MUT"
	FN      TokenType = "FN"
	RETURN  TokenType = "RETURN"
	IF      TokenType = "IF"
	ELSE    TokenType = "ELSE"
	FOR     TokenType = "FOR"
	IN      TokenType = "IN"
	TRUE    TokenType = "TRUE"
	FALSE   TokenType = "FALSE"
	STRUCT  TokenType = "STRUCT"
	ACTOR   TokenType = "ACTOR"
	SPAWN   TokenType = "SPAWN"
	SEND    TokenType = "SEND"
	RECEIVE TokenType = "RECEIVE"
	IMPORT  TokenType = "IMPORT"
	MATCH   TokenType = "MATCH"
	ENUM    TokenType = "ENUM"

	// Operators & Delimiters
	ASSIGN   TokenType = "="
	PLUS     TokenType = "+"
	MINUS    TokenType = "-"
	BANG     TokenType = "!"
	ASTERISK TokenType = "*"
	SLASH    TokenType = "/"
	EQ       TokenType = "=="
	NOT_EQ   TokenType = "!="
	LT       TokenType = "<"
	LTE      TokenType = "<="
	GT       TokenType = ">"
	GTE      TokenType = ">="
	ARROW    TokenType = "->"

	COMMA     TokenType = ","
	SEMICOLON TokenType = ";"
	COLON     TokenType = ":"
	DOT       TokenType = "."
	RANGE     TokenType = ".."
	PIPE      TokenType = "|>"
	QUESTION  TokenType = "?"

	LPAREN   TokenType = "("
	RPAREN   TokenType = ")"
	LBRACE   TokenType = "{"
	RBRACE   TokenType = "}"
	LBRACKET TokenType = "["
	RBRACKET TokenType = "]"
)

type Position struct {
	Line   int
	Column int
}

type Token struct {
	Type     TokenType
	Literal  string
	Position Position
}

type Lexer struct {
	input        string
	position     int  // current byte offset in input
	readPosition int  // next byte offset
	ch           rune // current UTF-8 rune
	line         int
	col          int
}

func New(input string) *Lexer {
	l := &Lexer{
		input: input,
		line:  1,
		col:   0,
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
		l.position = l.readPosition
	} else {
		r, size := utf8.DecodeRuneInString(l.input[l.readPosition:])
		l.ch = r
		l.position = l.readPosition
		l.readPosition += size
	}
	l.col++
}

func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.readPosition:])
	return r
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		if l.ch == '\n' {
			l.line++
			l.col = 0
		}
		l.readChar()
	}
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	// Skip comments
	for l.ch == '/' && l.peekChar() == '/' {
		for l.ch != '\n' && l.ch != 0 {
			l.readChar()
		}
		l.skipWhitespace()
	}

	pos := Position{Line: l.line, Column: l.col}
	var tok Token

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: EQ, Literal: string(ch) + string(l.ch), Position: pos}
		} else if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: ARROW, Literal: string(ch) + string(l.ch), Position: pos}
		} else {
			tok = Token{Type: ASSIGN, Literal: string(l.ch), Position: pos}
		}
	case '+':
		tok = Token{Type: PLUS, Literal: string(l.ch), Position: pos}
	case '-':
		if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: ARROW, Literal: string(ch) + string(l.ch), Position: pos}
		} else {
			tok = Token{Type: MINUS, Literal: string(l.ch), Position: pos}
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: NOT_EQ, Literal: string(ch) + string(l.ch), Position: pos}
		} else {
			tok = Token{Type: BANG, Literal: string(l.ch), Position: pos}
		}
	case '*':
		tok = Token{Type: ASTERISK, Literal: string(l.ch), Position: pos}
	case '/':
		tok = Token{Type: SLASH, Literal: string(l.ch), Position: pos}
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: LTE, Literal: string(ch) + string(l.ch), Position: pos}
		} else {
			tok = Token{Type: LT, Literal: string(l.ch), Position: pos}
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: GTE, Literal: string(ch) + string(l.ch), Position: pos}
		} else {
			tok = Token{Type: GT, Literal: string(l.ch), Position: pos}
		}
	case ';':
		tok = Token{Type: SEMICOLON, Literal: string(l.ch), Position: pos}
	case ':':
		tok = Token{Type: COLON, Literal: string(l.ch), Position: pos}
	case ',':
		tok = Token{Type: COMMA, Literal: string(l.ch), Position: pos}
	case '.':
		if l.peekChar() == '.' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: RANGE, Literal: string(ch) + string(l.ch), Position: pos}
		} else {
			tok = Token{Type: DOT, Literal: string(l.ch), Position: pos}
		}
	case '?':
		tok = Token{Type: QUESTION, Literal: string(l.ch), Position: pos}
	case '|':
		if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: PIPE, Literal: string(ch) + string(l.ch), Position: pos}
		} else {
			tok = Token{Type: ILLEGAL, Literal: string(l.ch), Position: pos}
		}
	case '(':
		tok = Token{Type: LPAREN, Literal: string(l.ch), Position: pos}
	case ')':
		tok = Token{Type: RPAREN, Literal: string(l.ch), Position: pos}
	case '{':
		tok = Token{Type: LBRACE, Literal: string(l.ch), Position: pos}
	case '}':
		tok = Token{Type: RBRACE, Literal: string(l.ch), Position: pos}
	case '[':
		tok = Token{Type: LBRACKET, Literal: string(l.ch), Position: pos}
	case ']':
		tok = Token{Type: RBRACKET, Literal: string(l.ch), Position: pos}
	case '"', '\'':
		tok.Type = STRING
		tok.Literal = l.readString(l.ch)
		tok.Position = pos
		return tok
	case 0:
		tok.Literal = ""
		tok.Type = EOF
		tok.Position = pos
	default:
		if isLetter(l.ch) {
			tok.Position = pos
			tok.Literal = l.readIdentifier()
			tok.Type = LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Position = pos
			tok.Type, tok.Literal = l.readNumber()
			return tok
		} else {
			tok = Token{Type: ILLEGAL, Literal: string(l.ch), Position: pos}
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() (TokenType, string) {
	position := l.position
	isFloat := false
	for isDigit(l.ch) || (!isFloat && l.ch == '.' && isDigit(l.peekChar())) {
		if l.ch == '.' {
			isFloat = true
		}
		l.readChar()
	}
	lit := l.input[position:l.position]
	if isFloat {
		return FLOAT, lit
	}
	return INT, lit
}

func (l *Lexer) readString(quote rune) string {
	l.readChar() // consume open quote
	position := l.position
	for l.ch != quote && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar() // skip escape
		}
		l.readChar()
	}
	str := l.input[position:l.position]
	if l.ch == quote {
		l.readChar() // consume closing quote
	}
	return str
}

func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func isDigit(ch rune) bool {
	return unicode.IsDigit(ch)
}

// Multilingual Keyword Map (English & Russian dual-grammar)
var keywords = map[string]TokenType{
	// English keywords
	"require": REQUIRE,
	"cap":     CAP,
	"intent":  INTENT,
	"let":     LET,
	"mut":     MUT,
	"fn":      FN,
	"return":  RETURN,
	"if":      IF,
	"else":    ELSE,
	"for":     FOR,
	"in":      IN,
	"true":    TRUE,
	"false":   FALSE,
	"struct":  STRUCT,
	"actor":   ACTOR,
	"spawn":   SPAWN,
	"send":    SEND,
	"receive": RECEIVE,
	"import":  IMPORT,
	"match":   MATCH,
	"enum":    ENUM,

	// Russian keywords (полная русскоязычная грамматика)
	"требовать": REQUIRE,
	"требует":   REQUIRE,
	"мандат":    CAP,
	"право":     CAP,
	"намерение": INTENT,
	"пусть":     LET,
	"изм":       MUT,
	"функция":   FN,
	"фн":        FN,
	"вернуть":   RETURN,
	"возврат":   RETURN,
	"если":      IF,
	"иначе":     ELSE,
	"для":       FOR,
	"в":         IN,
	"истина":    TRUE,
	"правда":    TRUE,
	"ложь":      FALSE,
	"структура": STRUCT,
	"актор":     ACTOR,
	"процесс":   ACTOR,
	"создать":   SPAWN,
	"запустить": SPAWN,
	"отправить": SEND,
	"послать":   SEND,
	"получить":  RECEIVE,
	"принять":   RECEIVE,
	"импорт":    IMPORT,
	"выбрать":   MATCH,
	"перечисление": ENUM,
}

func LookupIdent(ident string) TokenType {
	// Hardware SSE4.2 / AVX2 accelerated CRC32-C branch table (Zero allocation)
	switch fastCRC32String(ident) {
	case 0xb9ea5869: // "require"
		if ident == "require" { return REQUIRE }
	case 0x4ecdf1b8: // "cap"
		if ident == "cap" { return CAP }
	case 0x1b2b05ff: // "intent"
		if ident == "intent" { return INTENT }
	case 0x8b7c5dd7: // "let"
		if ident == "let" { return LET }
	case 0x11f83e28: // "mut"
		if ident == "mut" { return MUT }
	case 0xd58c2fa3: // "fn"
		if ident == "fn" { return FN }
	case 0xe58c9c59: // "return"
		if ident == "return" { return RETURN }
	case 0xb82c7d91: // "if"
		if ident == "if" { return IF }
	case 0xcda48192: // "else"
		if ident == "else" { return ELSE }
	case 0x60b225a1: // "for"
		if ident == "for" { return FOR }
	case 0x32f5255e: // "in"
		if ident == "in" { return IN }
	case 0x09087e7d: // "true"
		if ident == "true" { return TRUE }
	case 0x070d3382: // "false"
		if ident == "false" { return FALSE }
	case 0x2e116dec: // "struct"
		if ident == "struct" { return STRUCT }
	case 0x7020ad69: // "actor"
		if ident == "actor" { return ACTOR }
	case 0xb0975a8e: // "spawn"
		if ident == "spawn" { return SPAWN }
	case 0x7007c6f7: // "send"
		if ident == "send" { return SEND }
	case 0xe1043603: // "receive"
		if ident == "receive" { return RECEIVE }
	case 0xb35fbbd5: // "import"
		if ident == "import" { return IMPORT }
	case 0x424776fa: // "match"
		if ident == "match" { return MATCH }
	case 0x61ddda62: // "enum"
		if ident == "enum" { return ENUM }
	case 0x71ff950e: // "требовать"
		if ident == "требовать" { return REQUIRE }
	case 0x92bd9176: // "требует"
		if ident == "требует" { return REQUIRE }
	case 0x63db7e11: // "мандат"
		if ident == "мандат" { return CAP }
	case 0x31a20b8b: // "право"
		if ident == "право" { return CAP }
	case 0x240ea9f2: // "намерение"
		if ident == "намерение" { return INTENT }
	case 0x44d523f5: // "пусть"
		if ident == "пусть" { return LET }
	case 0x007168f4: // "изм"
		if ident == "изм" { return MUT }
	case 0xa6c68168: // "функция"
		if ident == "функция" { return FN }
	case 0xa695e6a5: // "фн"
		if ident == "фн" { return FN }
	case 0xd470210c: // "вернуть"
		if ident == "вернуть" { return RETURN }
	case 0xe355178e: // "возврат"
		if ident == "возврат" { return RETURN }
	case 0x0f50ba80: // "если"
		if ident == "если" { return IF }
	case 0x7385649e: // "иначе"
		if ident == "иначе" { return ELSE }
	case 0xb3cd6538: // "для"
		if ident == "для" { return FOR }
	case 0x995eb890: // "в"
		if ident == "в" { return IN }
	case 0xcad2002c: // "истина"
		if ident == "истина" { return TRUE }
	case 0xc4e2afee: // "правда"
		if ident == "правда" { return TRUE }
	case 0xc5252fa2: // "ложь"
		if ident == "ложь" { return FALSE }
	case 0x24b3af1a: // "структура"
		if ident == "структура" { return STRUCT }
	case 0x1a880302: // "актор"
		if ident == "актор" { return ACTOR }
	case 0xcd91606b: // "процесс"
		if ident == "процесс" { return ACTOR }
	case 0x42f5cf72: // "создать"
		if ident == "создать" { return SPAWN }
	case 0x2d4217e2: // "запустить"
		if ident == "запустить" { return SPAWN }
	case 0xe2f46a03: // "отправить"
		if ident == "отправить" { return SEND }
	case 0x0686033a: // "послать"
		if ident == "послать" { return SEND }
	case 0x91260782: // "получить"
		if ident == "получить" { return RECEIVE }
	case 0x7b2213fa: // "принять"
		if ident == "принять" { return RECEIVE }
	case 0x7a98dd2c: // "импорт"
		if ident == "импорт" { return IMPORT }
	case 0x2a91fa4d: // "выбрать"
		if ident == "выбрать" { return MATCH }
	case 0xd1f90114: // "перечисление"
		if ident == "перечисление" { return ENUM }
	}

	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

func (t Token) String() string {
	return fmt.Sprintf("Token(%s, %q, %d:%d)", t.Type, t.Literal, t.Position.Line, t.Position.Column)
}
