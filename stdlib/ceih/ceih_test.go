package ceih_test

import (
	"testing"

	"AxiomLang/stdlib/ceih"
)

func TestCEIH_VerifyCode_Valid(t *testing.T) {
	inspector := ceih.NewInspector()

	code := `
требовать Capability[StdIO]
намерение Main {
    пусть val = 42
    вернуть val
}
`
	seal, err := inspector.VerifyCode(code)
	if err != nil {
		t.Fatalf("unexpected verification error: %v", err)
	}
	if seal.Adherence != 1.0 {
		t.Fatalf("expected adherence 1.0, got %f", seal.Adherence)
	}
	if len(seal.Hash) != 64 {
		t.Fatalf("invalid seal hash length: %d", len(seal.Hash))
	}
}

func TestCEIH_VerifyCode_Violation(t *testing.T) {
	inspector := ceih.NewInspector()

	maliciousCode := `
намерение Hack {
    system("rm -rf /")
}
`
	seal, err := inspector.VerifyCode(maliciousCode)
	if err == nil {
		t.Fatalf("expected invariant violation, got nil")
	}
	if len(seal.Violations) == 0 {
		t.Fatalf("expected recorded violations in seal")
	}
}
