package toml_test

import (
	"testing"

	"AxiomLang/stdlib/toml"
)

func TestTOML_ParseAndGet(t *testing.T) {
	raw := `
# Axiom Node Configuration
title = "Axiom Autonomous Kernel"
version = 2
debug = true
rate = 1.618

[network]
host = "127.0.0.1"
port = 9000
enabled = true

[modbus]
slave_id = 1
baud_rate = 115200
`

	doc, err := toml.Parse(raw)
	if err != nil {
		t.Fatalf("failed to parse TOML: %v", err)
	}

	titleVal := doc.Get("title")
	if titleVal == nil || titleVal.String() != "Axiom Autonomous Kernel" {
		t.Fatalf("unexpected title: %v", titleVal)
	}

	verVal := doc.Get("version")
	ver, err := verVal.Int()
	if err != nil || ver != 2 {
		t.Fatalf("unexpected version: %v, err: %v", ver, err)
	}

	dbgVal := doc.Get("debug")
	dbg, err := dbgVal.Bool()
	if err != nil || !dbg {
		t.Fatalf("unexpected debug: %v", dbg)
	}

	portVal := doc.Get("network.port")
	port, err := portVal.Int()
	if err != nil || port != 9000 {
		t.Fatalf("unexpected network port: %v", port)
	}

	encoded := doc.Encode()
	if len(encoded) == 0 {
		t.Fatalf("encoded TOML is empty")
	}
}
