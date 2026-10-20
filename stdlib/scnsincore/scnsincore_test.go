package scnsincore_test

import (
	"testing"

	"AxiomLang/stdlib/scnsincore"
)

func TestSCNSCore_PulseAndInvariants(t *testing.T) {
	core := scnsincore.NewCore("core_alpha_0")

	state, err := core.EvaluateInvariants()
	if err != nil || state != scnsincore.StateNominal {
		t.Fatalf("unexpected state: %v, err: %v", state, err)
	}

	inputs := []float64{0.5, -1.2, 3.0, 0.0}
	outputs, proof, err := core.ComputePulse(inputs)
	if err != nil {
		t.Fatalf("ComputePulse failed: %v", err)
	}
	if len(outputs) != len(inputs) {
		t.Fatalf("expected output length %d, got %d", len(inputs), len(outputs))
	}
	if len(proof) != 64 {
		t.Fatalf("invalid proof hash length: %d", len(proof))
	}

	// Verify activation range [ -1.0, 1.0 ]
	for _, out := range outputs {
		if out < -1.0 || out > 1.0 {
			t.Fatalf("output out of bounded activation range: %f", out)
		}
	}
}

func TestSCNSCore_Isolation(t *testing.T) {
	core := scnsincore.NewCore("core_faulty")
	core.Invariants.FaultCount = 15 // Trigger critical failure

	state, err := core.EvaluateInvariants()
	if err == nil || state != scnsincore.StateCritical {
		t.Fatalf("expected StateCritical, got %v", state)
	}

	_, _, err = core.ComputePulse([]float64{1.0})
	if err == nil {
		t.Fatalf("expected pulse to be blocked for critical core")
	}
}
