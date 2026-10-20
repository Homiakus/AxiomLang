package parser_test

import (
	"testing"

	"AxiomLang/pkg/ast"
	"AxiomLang/pkg/lexer"
	"AxiomLang/pkg/parser"
)

func TestParser_CapabilitiesAndIntents(t *testing.T) {
	input := `
require cap[FileSystem.Write("/data/user.json")]
require cap[Console.Out]

intent "Сохранение настроек" {
    let payload = "admin_settings";
    print(payload);
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser had %d errors: %v", len(p.Errors()), p.Errors())
	}

	if len(prog.Capabilities) != 2 {
		t.Fatalf("expected 2 capabilities, got %d", len(prog.Capabilities))
	}

	if prog.Capabilities[0].Resource != "FileSystem.Write" {
		t.Fatalf("expected FileSystem.Write, got %s", prog.Capabilities[0].Resource)
	}

	if len(prog.Intents) != 1 {
		t.Fatalf("expected 1 intent, got %d", len(prog.Intents))
	}

	intent := prog.Intents[0]
	if intent.Description != "Сохранение настроек" {
		t.Fatalf("expected 'Сохранение настроек', got %s", intent.Description)
	}

	if len(intent.Body.Statements) != 2 {
		t.Fatalf("expected 2 statements in intent body, got %d", len(intent.Body.Statements))
	}

	letStmt, ok := intent.Body.Statements[0].(*ast.LetStatement)
	if !ok {
		t.Fatalf("expected LetStatement, got %T", intent.Body.Statements[0])
	}
	if letStmt.Name.Value != "payload" {
		t.Fatalf("expected let variable name 'payload', got %s", letStmt.Name.Value)
	}
}
