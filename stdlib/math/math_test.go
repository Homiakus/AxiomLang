package math_test

import (
	"testing"

	"AxiomLang/stdlib/math"
)

func TestMath_Basics(t *testing.T) {
	if math.Sqrt(16.0) != 4.0 {
		t.Fatalf("expected sqrt(16) = 4")
	}
	if math.Clamp(150, 0, 100) != 100 {
		t.Fatalf("expected clamp(150, 0, 100) = 100")
	}
	if math.Abs(-42.5) != 42.5 {
		t.Fatalf("expected abs(-42.5) = 42.5")
	}
}
