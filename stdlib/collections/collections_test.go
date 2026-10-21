package collections_test

import (
	"testing"

	"AxiomLang/stdlib/collections"
)

func TestCollections_MapAndResult(t *testing.T) {
	m := collections.NewMap()
	m.Set("temperature", 24.5)
	val, ok := m.Get("temperature")
	if !ok || val.(float64) != 24.5 {
		t.Fatalf("map lookup failed")
	}

	res := collections.Ok(100)
	if !res.IsOk() || res.UnwrapOr(0) != 100 {
		t.Fatalf("result unwrap failed")
	}

	opt := collections.Some("active")
	if !opt.IsSome() {
		t.Fatalf("option isSome failed")
	}
}
