package ai_test

import (
	"testing"

	"AxiomLang/pkg/ai"
)

func TestMicroNeuralHealer_AutoRepairAndDocument(t *testing.T) {
	healer := ai.NewMicroNeuralHealer()

	// Source with typos: "requir capa[console]", "intetn"
	sourceWithTypos := `
requir capa[console]

intetn "Тестовый интент с опечатками" {
    print("Проверка автоисправления")
}
`

	healed, suggestions := healer.HealSource(sourceWithTypos)

	if len(suggestions) == 0 {
		t.Fatalf("expected healer to identify and fix typos, got 0 suggestions")
	}

	// Verify keyword fixes
	if !containsSubstring(healed, "require cap[Console.Out]") {
		t.Fatalf("expected healed source to contain 'require cap[Console.Out]', got:\n%s", healed)
	}

	if !containsSubstring(healed, "intent ") {
		t.Fatalf("expected healed source to contain 'intent', got:\n%s", healed)
	}

	// Verify auto-documentation invariant comments
	if !containsSubstring(healed, "[AI-INVARIANT:") {
		t.Fatalf("expected auto-injected invariant comment, got:\n%s", healed)
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && searchSubstr(s, substr))
}

func searchSubstr(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
