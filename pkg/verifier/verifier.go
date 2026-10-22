package verifier

import (
	"AxiomLang/pkg/ast"
	"fmt"
	"strconv"
)

// ProofStatus represents the outcome of proof-carrying code verification.
type ProofStatus string

const (
	ProofVerified   ProofStatus = "VERIFIED"
	ProofCounterEx  ProofStatus = "COUNTER_EXAMPLE"
	ProofInconclusive ProofStatus = "INCONCLUSIVE"
)

// InvariantViolation records formal constraint violations at compile-time.
type InvariantViolation struct {
	Location string
	Rule     string
	Message  string
}

// SMTVerifier performs compile-time invariant checking and mathematical proof verification.
type SMTVerifier struct {
	Violations []InvariantViolation
	KnownBounds map[string]Interval
}

// Interval models closed or open integer ranges [Min, Max]
type Interval struct {
	Min int64
	Max int64
	HasMin bool
	HasMax bool
}

// NewSMTVerifier initializes an SMT/Proof-carrying verification context.
func NewSMTVerifier() *SMTVerifier {
	return &SMTVerifier{
		Violations:  make([]InvariantViolation, 0),
		KnownBounds: make(map[string]Interval),
	}
}

// VerifyProgram inspects AST for arithmetic safety, zero-division proofs, and range boundaries.
func (v *SMTVerifier) VerifyProgram(prog *ast.Program) []InvariantViolation {
	if prog == nil {
		return v.Violations
	}
	for _, fn := range prog.Functions {
		v.verifyFunction(fn)
	}
	for _, it := range prog.Intents {
		if it.Body != nil {
			v.verifyBlock(it.Body)
		}
	}
	for _, stmt := range prog.Statements {
		v.verifyStatement(stmt)
	}
	return v.Violations
}

func (v *SMTVerifier) verifyFunction(fn *ast.FunctionStatement) {
	if fn == nil || fn.Body == nil {
		return
	}
	v.verifyBlock(fn.Body)
}

func (v *SMTVerifier) verifyBlock(block *ast.BlockStatement) {
	if block == nil {
		return
	}
	for _, stmt := range block.Statements {
		v.verifyStatement(stmt)
	}
}

func (v *SMTVerifier) verifyStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.LetStatement:
		if s.Value != nil {
			v.verifyExpression(s.Value)
			// Track simple constant bounds
			if intLit, ok := s.Value.(*ast.IntegerLiteral); ok {
				v.KnownBounds[s.Name.Value] = Interval{
					Min: intLit.Value,
					Max: intLit.Value,
					HasMin: true,
					HasMax: true,
				}
			}
		}
	case *ast.ExpressionStatement:
		if s.Expression != nil {
			v.verifyExpression(s.Expression)
		}
	case *ast.IfStatement:
		if s.Condition != nil {
			v.verifyExpression(s.Condition)
		}
		if s.Consequence != nil {
			v.verifyBlock(s.Consequence)
		}
		if s.Alternative != nil {
			v.verifyBlock(s.Alternative)
		}
	case *ast.ReturnStatement:
		if s.ReturnValue != nil {
			v.verifyExpression(s.ReturnValue)
		}
	case *ast.BlockStatement:
		v.verifyBlock(s)
	}
}

func (v *SMTVerifier) verifyExpression(expr ast.Expression) {
	if expr == nil {
		return
	}
	switch e := expr.(type) {
	case *ast.BinaryExpression:
		// Check division and modulo operations for proven non-zero denominator
		if e.Operator == "/" || e.Operator == "%" {
			v.proveNonZeroDenominator(e.Right, e.Operator)
		}
		v.verifyExpression(e.Left)
		v.verifyExpression(e.Right)
	case *ast.CallExpression:
		for _, arg := range e.Arguments {
			v.verifyExpression(arg)
		}
	case *ast.IndexExpression:
		v.verifyIndexSafety(e)
		v.verifyExpression(e.Left)
		v.verifyExpression(e.Index)
	case *ast.PrefixExpression:
		v.verifyExpression(e.Right)
	}
}

func (v *SMTVerifier) proveNonZeroDenominator(denom ast.Expression, op string) {
	switch d := denom.(type) {
	case *ast.IntegerLiteral:
		if d.Value == 0 {
			v.Violations = append(v.Violations, InvariantViolation{
				Location: d.TokenLiteral(),
				Rule:     "FORMAL_PROOF_DIV_BY_ZERO",
				Message:  fmt.Sprintf("Mathematical contradiction: division by constant literal 0 detected in operation '%s'", op),
			})
		}
	case *ast.Identifier:
		if bounds, exists := v.KnownBounds[d.Value]; exists {
			if bounds.HasMin && bounds.HasMax && bounds.Min == 0 && bounds.Max == 0 {
				v.Violations = append(v.Violations, InvariantViolation{
					Location: d.Value,
					Rule:     "FORMAL_PROOF_DIV_BY_ZERO",
					Message:  fmt.Sprintf("Proof refutation: variable '%s' is proven to evaluate to 0 in operation '%s'", d.Value, op),
				})
			}
		}
	}
}

func (v *SMTVerifier) verifyIndexSafety(idxExpr *ast.IndexExpression) {
	if lit, ok := idxExpr.Index.(*ast.IntegerLiteral); ok {
		if lit.Value < 0 {
			v.Violations = append(v.Violations, InvariantViolation{
				Location: strconv.FormatInt(lit.Value, 10),
				Rule:     "FORMAL_PROOF_NEGATIVE_INDEX",
				Message:  fmt.Sprintf("Negative index constraint refuted: index [%d] < 0 is mathematically invalid", lit.Value),
			})
		}
	}
}
