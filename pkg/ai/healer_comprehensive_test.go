package ai_test

import (
	"strings"
	"testing"

	"AxiomLang/pkg/ai"
)

func TestMicroNeuralHealer_ComprehensiveAutoRepairAndDoc(t *testing.T) {
	healer := ai.NewMicroNeuralHealer()

	// 1. Stress test with multiple typos: 'requir capa[stdout]', 'intnet', 'funct', 'spwn', 'snd'
	malformedSource := `
requir capa[stdout]
requir capa[file]

funct compute_score(x: int) -> int {
    let base = 50;
    return base;
}

intnet "Главный сценарий обработки" {
    let score = compute_score(10);
    print(score);
    spwn WorkerNode();
    snd WorkerNode <- "START";
}
`

	healed, suggestions := healer.HealSource(malformedSource)

	if len(suggestions) < 3 {
		t.Fatalf("expected at least 3 auto-healed suggestions, got %d", len(suggestions))
	}

	// Verify keyword repairs
	if !strings.Contains(healed, "require cap[Console.Out]") {
		t.Fatalf("expected healed 'require cap[Console.Out]', got:\n%s", healed)
	}

	if !strings.Contains(healed, "require cap[FileSystem.Write]") {
		t.Fatalf("expected healed 'require cap[FileSystem.Write]', got:\n%s", healed)
	}

	if !strings.Contains(healed, "intent ") {
		t.Fatalf("expected healed 'intent', got:\n%s", healed)
	}

	if !strings.Contains(healed, "fn compute_score") {
		t.Fatalf("expected healed 'fn compute_score', got:\n%s", healed)
	}

	// Verify auto-injected mathematical and invariant documentation comments
	if !strings.Contains(healed, "// [AI-INVARIANT: Strict Zero-Ambient Verification Passed]") {
		t.Fatalf("missing AI-INVARIANT auto-comment in healed source")
	}

	if !strings.Contains(healed, "// [AI-PROOF: Topological Knot Validated]") {
		t.Fatalf("missing AI-PROOF auto-comment in healed source")
	}
}

func TestMicroNeuralHealer_FuzzyTypoTolerance(t *testing.T) {
	healer := ai.NewMicroNeuralHealer()

	fuzzyCode := `
req cap[Console.Out]
intetn "Тест нечеткого расстояния Левенштейна" {
    print("Успешно исправлено");
}
`
	healed, suggestions := healer.HealSource(fuzzyCode)
	if len(suggestions) == 0 {
		t.Fatalf("expected fuzzy healer to detect and fix typos")
	}

	if !strings.Contains(healed, "require cap[Console.Out]") {
		t.Fatalf("expected 'require', got:\n%s", healed)
	}

	if !strings.Contains(healed, "intent ") {
		t.Fatalf("expected 'intent', got:\n%s", healed)
	}
}
