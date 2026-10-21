package time_test

import (
	"testing"

	"AxiomLang/stdlib/time"
)

func TestTime_Now(t *testing.T) {
	t0 := time.NowMs()
	if t0 <= 0 {
		t.Fatalf("expected positive timestamp")
	}
}
