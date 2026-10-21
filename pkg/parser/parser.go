package parser

import (
	"AxiomLang/pkg/ast"
	"AxiomLang/pkg/lexer"
	"fmt"
	"strconv"
)

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l         *lexer.Lexer
	curToken  lexer.Token
	peekToken lexer.Token
	errors    []string
}

const (
	_ int = iota
	LOWEST
	RANGE_PREC  // ..
	EQUALS      // == or !=
	LESSGREATER // > or < or <= or >=
	SUM         // + or -
	PRODUCT     // * or /
	PREFIX      // -X or !X
	CALL        // myFunction(X)
	DOT         // obj.member
)

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: make([]string, 0, 4),
	}

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) parsePrefix() ast.Expression {
	switch p.curToken.Type {
	case lexer.IDENT:
		return p.parseIdentifier()
	case lexer.INT:
		return p.parseIntegerLiteral()
	case lexer.STRING:
		return p.parseStringLiteral()
	case lexer.TRUE, lexer.FALSE:
		return p.parseBooleanLiteral()
	case lexer.LBRACE:
		return p.parseMapLiteral()
	case lexer.LBRACKET:
		return p.parseArrayLiteral()
	case lexer.LPAREN:
		return p.parseGroupedExpression()
	case lexer.MINUS, lexer.BANG:
		return p.parsePrefixExpression()
	default:
		p.errors = append(p.errors, fmt.Sprintf("no prefix parse function for %s", p.curToken.Type))
		return nil
	}
}

func (p *Parser) parseInfix(tokenType lexer.TokenType, left ast.Expression) ast.Expression {
	switch tokenType {
	case lexer.PLUS, lexer.MINUS, lexer.SLASH, lexer.ASTERISK,
		lexer.EQ, lexer.NOT_EQ, lexer.LT, lexer.LTE, lexer.GT, lexer.GTE:
		return p.parseInfixExpression(left)
	case lexer.RANGE:
		return p.parseRangeExpression(left)
	case lexer.LPAREN:
		return p.parseCallExpression(left)
	case lexer.DOT:
		return p.parseMemberAccessExpression(left)
	default:
		return nil
	}
}

func (p *Parser) parseRangeExpression(left ast.Expression) ast.Expression {
	tok := p.curToken
	p.nextToken()
	right := p.parseExpression(RANGE_PREC)
	return &ast.RangeExpression{
		Token: tok,
		Start: left,
		End:   right,
	}
}

func (p *Parser) isInfixToken(tokenType lexer.TokenType) bool {
	switch tokenType {
	case lexer.PLUS, lexer.MINUS, lexer.SLASH, lexer.ASTERISK,
		lexer.EQ, lexer.NOT_EQ, lexer.LT, lexer.LTE, lexer.GT, lexer.GTE,
		lexer.RANGE,
		lexer.LPAREN, lexer.DOT:
		return true
	default:
		return false
	}
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekPrecedence() int {
	switch p.peekToken.Type {
	case lexer.RANGE:
		return RANGE_PREC
	case lexer.EQ, lexer.NOT_EQ:
		return EQUALS
	case lexer.LT, lexer.GT, lexer.LTE, lexer.GTE:
		return LESSGREATER
	case lexer.PLUS, lexer.MINUS:
		return SUM
	case lexer.SLASH, lexer.ASTERISK:
		return PRODUCT
	case lexer.LPAREN:
		return CALL
	case lexer.DOT:
		return DOT
	default:
		return LOWEST
	}
}

func (p *Parser) curPrecedence() int {
	switch p.curToken.Type {
	case lexer.RANGE:
		return RANGE_PREC
	case lexer.EQ, lexer.NOT_EQ:
		return EQUALS
	case lexer.LT, lexer.GT, lexer.LTE, lexer.GTE:
		return LESSGREATER
	case lexer.PLUS, lexer.MINUS:
		return SUM
	case lexer.SLASH, lexer.ASTERISK:
		return PRODUCT
	case lexer.LPAREN:
		return CALL
	case lexer.DOT:
		return DOT
	default:
		return LOWEST
	}
}

func (p *Parser) ParseProgram() *ast.Program {
	prog := &ast.Program{
		Imports:      make([]*ast.ImportStatement, 0, 8),
		Enums:        make([]*ast.EnumStatement, 0, 8),
		Capabilities: make([]*ast.CapabilityStatement, 0, 8),
		Structs:      make([]*ast.StructStatement, 0, 8),
		Functions:    make([]*ast.FunctionStatement, 0, 16),
		Intents:      make([]*ast.IntentStatement, 0, 8),
		Statements:   make([]ast.Statement, 0, 32),
	}

	for p.curToken.Type != lexer.EOF {
		if p.curToken.Type == lexer.IMPORT {
			imp := p.parseImportStatement()
			if imp != nil {
				prog.Imports = append(prog.Imports, imp)
			}
		} else if p.curToken.Type == lexer.ENUM {
			enumStmt := p.parseEnumStatement()
			if enumStmt != nil {
				prog.Enums = append(prog.Enums, enumStmt)
			}
		} else if p.curToken.Type == lexer.REQUIRE {
			capStmt := p.parseCapabilityStatement()
			if capStmt != nil {
				prog.Capabilities = append(prog.Capabilities, capStmt)
			}
		} else if p.curToken.Type == lexer.STRUCT {
			structStmt := p.parseStructStatement()
			if structStmt != nil {
				prog.Structs = append(prog.Structs, structStmt)
			}
		} else if p.curToken.Type == lexer.FN {
			fnStmt := p.parseFunctionStatement()
			if fnStmt != nil {
				prog.Functions = append(prog.Functions, fnStmt)
			}
		} else if p.curToken.Type == lexer.INTENT {
			intentStmt := p.parseIntentStatement()
			if intentStmt != nil {
				prog.Intents = append(prog.Intents, intentStmt)
			}
		} else {
			stmt := p.parseStatement()
			if stmt != nil {
				prog.Statements = append(prog.Statements, stmt)
			}
		}
		p.nextToken()
	}

	return prog
}

func (p *Parser) parseImportStatement() *ast.ImportStatement {
	stmt := &ast.ImportStatement{Token: p.curToken}
	if p.peekToken.Type != lexer.STRING {
		p.errors = append(p.errors, fmt.Sprintf("expected string path after import at %v", p.peekToken.Position))
		return nil
	}
	p.nextToken()
	stmt.Path = p.curToken.Literal
	if p.peekToken.Type == lexer.SEMICOLON {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseEnumStatement() *ast.EnumStatement {
	stmt := &ast.EnumStatement{Token: p.curToken, Variants: []*ast.EnumVariant{}}
	if p.peekToken.Type != lexer.IDENT {
		p.errors = append(p.errors, fmt.Sprintf("expected enum identifier at %v", p.peekToken.Position))
		return nil
	}
	p.nextToken()
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if p.peekToken.Type != lexer.LBRACE {
		p.errors = append(p.errors, fmt.Sprintf("expected '{' after enum name at %v", p.peekToken.Position))
		return nil
	}
	p.nextToken()

	for p.peekToken.Type != lexer.RBRACE && p.peekToken.Type != lexer.EOF {
		if p.peekToken.Type != lexer.IDENT {
			break
		}
		p.nextToken()
		v := &ast.EnumVariant{Name: &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}}
		if p.peekToken.Type == lexer.LPAREN {
			p.nextToken() // (
			if p.peekToken.Type == lexer.IDENT {
				p.nextToken()
				v.Type = p.curToken.Literal
			}
			if p.peekToken.Type == lexer.RPAREN {
				p.nextToken() // )
			}
		}
		stmt.Variants = append(stmt.Variants, v)
		if p.peekToken.Type == lexer.COMMA {
			p.nextToken()
		}
	}

	if p.peekToken.Type == lexer.RBRACE {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseStructStatement() *ast.StructStatement {
	stmt := &ast.StructStatement{Token: p.curToken, Fields: []*ast.StructField{}}

	if p.peekToken.Type != lexer.IDENT {
		p.errors = append(p.errors, fmt.Sprintf("expected struct name at %v", p.peekToken.Position))
		return nil
	}
	p.nextToken()
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if p.peekToken.Type != lexer.LBRACE {
		p.errors = append(p.errors, fmt.Sprintf("expected '{' after struct name at %v", p.peekToken.Position))
		return nil
	}
	p.nextToken()

	for p.peekToken.Type != lexer.RBRACE && p.peekToken.Type != lexer.EOF {
		if p.peekToken.Type != lexer.IDENT {
			break
		}
		p.nextToken()
		fName := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		fType := "interface{}"

		if p.peekToken.Type == lexer.COLON {
			p.nextToken()
			p.nextToken()
			fType = p.curToken.Literal
		}
		stmt.Fields = append(stmt.Fields, &ast.StructField{Name: fName, Type: fType})

		if p.peekToken.Type == lexer.COMMA || p.peekToken.Type == lexer.SEMICOLON {
			p.nextToken()
		}
	}

	if p.peekToken.Type == lexer.RBRACE {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseFunctionStatement() *ast.FunctionStatement {
	stmt := &ast.FunctionStatement{Token: p.curToken, Parameters: []*ast.Parameter{}}

	if p.peekToken.Type != lexer.IDENT {
		p.errors = append(p.errors, fmt.Sprintf("expected function name at %v", p.peekToken.Position))
		return nil
	}
	p.nextToken()
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if p.peekToken.Type != lexer.LPAREN {
		p.errors = append(p.errors, fmt.Sprintf("expected '(' after function name at %v", p.peekToken.Position))
		return nil
	}
	p.nextToken()

	// Parse parameters: (a: int, b: int)
	if p.peekToken.Type != lexer.RPAREN {
		for {
			if p.peekToken.Type != lexer.IDENT {
				break
			}
			p.nextToken()
			paramName := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
			paramType := ""
			if p.peekToken.Type == lexer.COLON {
				p.nextToken()
				p.nextToken()
				paramType = p.curToken.Literal
			}
			stmt.Parameters = append(stmt.Parameters, &ast.Parameter{Name: paramName, Type: paramType})

			if p.peekToken.Type == lexer.COMMA {
				p.nextToken()
			} else {
				break
			}
		}
	}

	if p.peekToken.Type == lexer.RPAREN {
		p.nextToken()
	}

	if p.peekToken.Type == lexer.ARROW {
		p.nextToken() // ->
		p.nextToken() // return type
		stmt.ReturnType = p.curToken.Literal
	}

	if p.peekToken.Type != lexer.LBRACE {
		p.errors = append(p.errors, fmt.Sprintf("expected '{' for function body at %v", p.peekToken.Position))
		return nil
	}
	p.nextToken()
	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseCapabilityStatement() *ast.CapabilityStatement {
	stmt := &ast.CapabilityStatement{Token: p.curToken}

	if p.peekToken.Type != lexer.CAP {
		p.errors = append(p.errors, fmt.Sprintf("expected 'cap' after 'require' at %v", p.peekToken.Position))
		return nil
	}
	p.nextToken()

	if p.peekToken.Type != lexer.LBRACKET {
		p.errors = append(p.errors, fmt.Sprintf("expected '[' after 'cap' at %v", p.peekToken.Position))
		return nil
	}
	p.nextToken()
	p.nextToken()

	var resName string
	for p.curToken.Type == lexer.IDENT || p.curToken.Type == lexer.DOT {
		resName += p.curToken.Literal
		if p.peekToken.Type == lexer.IDENT || p.peekToken.Type == lexer.DOT {
			p.nextToken()
		} else {
			break
		}
	}
	stmt.Resource = resName

	if p.peekToken.Type == lexer.LPAREN {
		p.nextToken()
		p.nextToken()
		stmt.Arg = p.parseExpression(LOWEST)
		if p.peekToken.Type == lexer.RPAREN {
			p.nextToken()
		}
	}

	if p.peekToken.Type == lexer.RBRACKET {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseIntentStatement() *ast.IntentStatement {
	stmt := &ast.IntentStatement{Token: p.curToken}

	if p.peekToken.Type != lexer.STRING {
		p.errors = append(p.errors, fmt.Sprintf("expected string description for intent at %v", p.peekToken.Position))
		return nil
	}
	p.nextToken()
	stmt.Description = p.curToken.Literal

	if p.peekToken.Type != lexer.LBRACE {
		p.errors = append(p.errors, fmt.Sprintf("expected '{' after intent description at %v", p.peekToken.Position))
		return nil
	}
	p.nextToken()

	stmt.Body = p.parseBlockStatement()
	return stmt
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken, Statements: []ast.Statement{}}
	p.nextToken()

	for p.curToken.Type != lexer.RBRACE && p.curToken.Type != lexer.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}
	return block
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case lexer.LET:
		return p.parseLetStatement()
	case lexer.SPAWN:
		return p.parseSpawnStatement()
	case lexer.SEND:
		return p.parseSendStatement()
	case lexer.RETURN:
		return p.parseReturnStatement()
	case lexer.IF:
		return p.parseIfStatement()
	case lexer.FOR:
		return p.parseForInStatement()
	case lexer.MATCH:
		return p.parseMatchStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseMatchStatement() *ast.MatchStatement {
	stmt := &ast.MatchStatement{Token: p.curToken, Cases: []*ast.MatchCase{}}
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	if p.peekToken.Type != lexer.LBRACE {
		p.errors = append(p.errors, fmt.Sprintf("expected '{' after match expression at %v", p.peekToken.Position))
		return nil
	}
	p.nextToken() // cur at {

	for p.peekToken.Type != lexer.RBRACE && p.peekToken.Type != lexer.EOF {
		p.nextToken()
		pattern := p.parseExpression(LOWEST)
		if p.peekToken.Type == lexer.ARROW || p.peekToken.Literal == "=>" {
			p.nextToken()
		}
		if p.peekToken.Type != lexer.LBRACE {
			p.errors = append(p.errors, fmt.Sprintf("expected '{' in match case at %v", p.peekToken.Position))
			break
		}
		p.nextToken()
		caseBody := p.parseBlockStatement()
		stmt.Cases = append(stmt.Cases, &ast.MatchCase{
			Pattern: pattern,
			Body:    caseBody,
		})
		if p.peekToken.Type == lexer.COMMA {
			p.nextToken()
		}
	}

	if p.peekToken.Type == lexer.RBRACE {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseSpawnStatement() *ast.SpawnStatement {
	stmt := &ast.SpawnStatement{Token: p.curToken}

	if p.peekToken.Type != lexer.IDENT {
		p.errors = append(p.errors, fmt.Sprintf("expected actor identifier after spawn at %v", p.peekToken.Position))
		return nil
	}
	p.nextToken()
	stmt.ActorID = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if p.peekToken.Type == lexer.LPAREN {
		p.nextToken() // at (
		if p.peekToken.Type != lexer.RPAREN {
			p.nextToken()
			stmt.Payload = p.parseExpression(LOWEST)
		}
		if p.peekToken.Type == lexer.RPAREN {
			p.nextToken()
		}
	}

	if p.peekToken.Type == lexer.SEMICOLON {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseSendStatement() *ast.SendStatement {
	stmt := &ast.SendStatement{Token: p.curToken}

	if p.peekToken.Type != lexer.IDENT {
		p.errors = append(p.errors, fmt.Sprintf("expected target identifier after send at %v", p.peekToken.Position))
		return nil
	}
	p.nextToken()
	stmt.Target = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if p.peekToken.Type == lexer.LT || p.peekToken.Type == lexer.ARROW {
		p.nextToken() // consume '<-' or '->'
		if p.peekToken.Type == lexer.MINUS {
			p.nextToken()
		}
	}
	p.nextToken()
	stmt.Message = p.parseExpression(LOWEST)

	if p.peekToken.Type == lexer.SEMICOLON {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	if p.peekToken.Type != lexer.SEMICOLON && p.peekToken.Type != lexer.RBRACE {
		p.nextToken()
		stmt.ReturnValue = p.parseExpression(LOWEST)
	}
	if p.peekToken.Type == lexer.SEMICOLON {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseForInStatement() *ast.ForInStatement {
	stmt := &ast.ForInStatement{Token: p.curToken}

	if p.peekToken.Type != lexer.IDENT {
		p.errors = append(p.errors, fmt.Sprintf("expected identifier after 'for' at %v", p.peekToken.Position))
		return nil
	}
	p.nextToken()
	stmt.Item = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if p.peekToken.Type != lexer.IN {
		p.errors = append(p.errors, fmt.Sprintf("expected 'in' after for variable at %v", p.peekToken.Position))
		return nil
	}
	p.nextToken()
	p.nextToken()

	stmt.Collection = p.parseExpression(LOWEST)

	if p.peekToken.Type != lexer.LBRACE {
		p.errors = append(p.errors, fmt.Sprintf("expected '{' in for loop body at %v", p.peekToken.Position))
		return nil
	}
	p.nextToken()
	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseIfStatement() *ast.IfStatement {
	stmt := &ast.IfStatement{Token: p.curToken}
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if p.peekToken.Type != lexer.LBRACE {
		p.errors = append(p.errors, fmt.Sprintf("expected '{' after if condition at %v", p.peekToken.Position))
		return nil
	}
	p.nextToken()
	stmt.Consequence = p.parseBlockStatement()

	if p.peekToken.Type == lexer.ELSE {
		p.nextToken()
		if p.peekToken.Type != lexer.LBRACE {
			p.errors = append(p.errors, fmt.Sprintf("expected '{' after else at %v", p.peekToken.Position))
			return nil
		}
		p.nextToken()
		stmt.Alternative = p.parseBlockStatement()
	}

	return stmt
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	if p.peekToken.Type == lexer.MUT {
		p.nextToken()
		stmt.Mutable = true
	}

	if p.peekToken.Type != lexer.IDENT {
		p.errors = append(p.errors, fmt.Sprintf("expected identifier in let statement at %v", p.peekToken.Position))
		return nil
	}
	p.nextToken()
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if p.peekToken.Type == lexer.COLON {
		p.nextToken()
		p.nextToken()
		stmt.Type = p.curToken.Literal
	}

	if p.peekToken.Type != lexer.ASSIGN {
		p.errors = append(p.errors, fmt.Sprintf("expected '=' in let statement at %v", p.peekToken.Position))
		return nil
	}
	p.nextToken()
	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekToken.Type == lexer.SEMICOLON {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekToken.Type == lexer.SEMICOLON {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	leftExp := p.parsePrefix()
	if leftExp == nil {
		return nil
	}

	for p.peekToken.Type != lexer.SEMICOLON && precedence < p.peekPrecedence() {
		if !p.isInfixToken(p.peekToken.Type) {
			return leftExp
		}
		p.nextToken()
		leftExp = p.parseInfix(p.curToken.Type, leftExp)
		if leftExp == nil {
			return nil
		}
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	val, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		p.errors = append(p.errors, fmt.Sprintf("could not parse %q as integer", p.curToken.Literal))
		return nil
	}
	return &ast.IntegerLiteral{Token: p.curToken, Value: val}
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{Token: p.curToken, Value: p.curToken.Type == lexer.TRUE}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if p.peekToken.Type != lexer.RPAREN {
		p.errors = append(p.errors, fmt.Sprintf("expected ')' at %v", p.peekToken.Position))
		return nil
	}
	p.nextToken()
	return exp
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	tok := p.curToken
	p.nextToken()
	right := p.parseExpression(PREFIX)
	return &ast.PrefixExpression{
		Token:    tok,
		Operator: tok.Literal,
		Right:    right,
	}
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	arr := &ast.ArrayLiteral{Token: p.curToken}
	arr.Elements = p.parseExpressionList(lexer.RBRACKET)
	return arr
}

func (p *Parser) parseMapLiteral() ast.Expression {
	mapLit := &ast.MapLiteral{Token: p.curToken, Pairs: make(map[ast.Expression]ast.Expression)}

	for p.peekToken.Type != lexer.RBRACE && p.peekToken.Type != lexer.EOF {
		p.nextToken()
		var key ast.Expression
		if p.curToken.Type == lexer.IDENT {
			key = &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
		} else {
			key = p.parseExpression(LOWEST)
		}

		if p.peekToken.Type != lexer.COLON {
			p.errors = append(p.errors, fmt.Sprintf("expected ':' in map literal at %v", p.peekToken.Position))
			return nil
		}
		p.nextToken()
		p.nextToken()
		val := p.parseExpression(LOWEST)

		mapLit.Pairs[key] = val

		if p.peekToken.Type != lexer.RBRACE && p.peekToken.Type != lexer.COMMA {
			return nil
		}
		if p.peekToken.Type == lexer.COMMA {
			p.nextToken()
		}
	}

	if p.peekToken.Type != lexer.RBRACE {
		p.errors = append(p.errors, fmt.Sprintf("expected '}' at %v", p.peekToken.Position))
		return nil
	}
	p.nextToken()
	return mapLit
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	exp := &ast.BinaryExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	exp.Right = p.parseExpression(precedence)
	return exp
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(lexer.RPAREN)
	return exp
}

func (p *Parser) parseExpressionList(end lexer.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekToken.Type == end {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekToken.Type == lexer.COMMA {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if p.peekToken.Type != end {
		p.errors = append(p.errors, fmt.Sprintf("expected %s at %v", end, p.peekToken.Position))
		return nil
	}
	p.nextToken()

	return list
}

func (p *Parser) parseMemberAccessExpression(object ast.Expression) ast.Expression {
	tok := p.curToken
	if p.peekToken.Type != lexer.IDENT {
		p.errors = append(p.errors, fmt.Sprintf("expected identifier after '.' at %v", p.peekToken.Position))
		return nil
	}
	p.nextToken()
	return &ast.MemberAccessExpression{
		Token:  tok,
		Object: object,
		Member: p.curToken.Literal,
	}
}
