package fs_test

import (
	"testing"
	"time"

	"AxiomLang/stdlib/fs"
)

func TestGovernedFileSystem_ZeroAmbient(t *testing.T) {
	gfs := fs.NewGovernedFileSystem()

	// 1. Unprivileged write must fail
	err := gfs.Write(nil, []byte("unauthorized data"))
	if err == nil {
		t.Fatalf("expected error on unprivileged write, got nil")
	}

	// 2. Granted write capability
	writeCap := &fs.FileCapability{
		TokenID:   "cap_test_01",
		Path:      "/data/test.json",
		Rights:    fs.RightWrite,
		ExpiresAt: time.Now().Add(10 * time.Minute).Unix(),
		Issuer:    "KernelRoot",
	}

	payload := []byte(`{"key": "axiom_data"}`)
	err = gfs.Write(writeCap, payload)
	if err != nil {
		t.Fatalf("failed to write with valid capability: %v", err)
	}

	// 3. Read with Write-only capability must fail
	_, _, err = gfs.Read(writeCap)
	if err == nil {
		t.Fatalf("expected error reading with write-only capability, got nil")
	}

	// 4. Read with valid Read capability
	readCap := &fs.FileCapability{
		TokenID:   "cap_test_02",
		Path:      "/data/test.json",
		Rights:    fs.RightRead,
		ExpiresAt: time.Now().Add(10 * time.Minute).Unix(),
		Issuer:    "KernelRoot",
	}

	data, hash, err := gfs.Read(readCap)
	if err != nil {
		t.Fatalf("failed to read with valid read capability: %v", err)
	}

	if string(data) != string(payload) {
		t.Fatalf("data mismatch: got %s, want %s", string(data), string(payload))
	}

	if hash == "" {
		t.Fatalf("expected non-empty cryptographic hash")
	}
}
