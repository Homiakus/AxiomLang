package verifier

import (
	"AxiomLang/pkg/ast"
	"AxiomLang/pkg/lexer"
	"testing"
)

func TestSMTVerifierDivisionByZeroLiteral(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ExpressionStatement{
				Expression: &ast.BinaryExpression{
					Left:     &ast.IntegerLiteral{Token: lexer.Token{Literal: "10"}, Value: 10},
					Operator: "/",
					Right:    &ast.IntegerLiteral{Token: lexer.Token{Literal: "0"}, Value: 0},
				},
			},
		},
	}

	verifier := NewSMTVerifier()
	violations := verifier.VerifyProgram(prog)

	if len(violations) == 0 {
		t.Fatalf("expected formal proof violation for division by zero literal, got 0")
	}
	if violations[0].Rule != "FORMAL_PROOF_DIV_BY_ZERO" {
		t.Errorf("expected rule FORMAL_PROOF_DIV_BY_ZERO, got %s", violations[0].Rule)
	}
}

func TestSMTVerifierBoundPropagation(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LetStatement{
				Name:  &ast.Identifier{Token: lexer.Token{Literal: "zero"}, Value: "zero"},
				Value: &ast.IntegerLiteral{Token: lexer.Token{Literal: "0"}, Value: 0},
			},
			&ast.ExpressionStatement{
				Expression: &ast.BinaryExpression{
					Left:     &ast.IntegerLiteral{Token: lexer.Token{Literal: "42"}, Value: 42},
					Operator: "%",
					Right:    &ast.Identifier{Token: lexer.Token{Literal: "zero"}, Value: "zero"},
				},
			},
		},
	}

	verifier := NewSMTVerifier()
	violations := verifier.VerifyProgram(prog)

	if len(violations) == 0 {
		t.Fatalf("expected violation for variable proven to be zero in modulo denominator")
	}
}

func TestSMTVerifierNegativeIndex(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ExpressionStatement{
				Expression: &ast.IndexExpression{
					Left:  &ast.Identifier{Token: lexer.Token{Literal: "arr"}, Value: "arr"},
					Index: &ast.IntegerLiteral{Token: lexer.Token{Literal: "-5"}, Value: -5},
				},
			},
		},
	}

	verifier := NewSMTVerifier()
	violations := verifier.VerifyProgram(prog)

	if len(violations) == 0 {
		t.Fatalf("expected violation for negative constant index")
	}
	if violations[0].Rule != "FORMAL_PROOF_NEGATIVE_INDEX" {
		t.Errorf("expected rule FORMAL_PROOF_NEGATIVE_INDEX, got %s", violations[0].Rule)
	}
}
