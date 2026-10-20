package fs_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"AxiomLang/stdlib/fs"
)

func TestHostFileSystem_CapabilityControlled(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "axiom_config.dat")

	writeCap := fs.IssueCapability("test_runner", targetFile, fs.RightWrite, 1*time.Minute)
	readCap := fs.IssueCapability("test_runner", targetFile, fs.RightRead, 1*time.Minute)

	payload := []byte("AXIOM_STORAGE_BLOCK_4096")

	// 1. Write file
	err := fs.WriteFile(writeCap, payload, 0644)
	if err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// 2. Read with writeCap must fail
	_, err = fs.ReadFile(writeCap)
	if err == nil {
		t.Fatalf("expected error reading with Write-only capability")
	}

	// 3. Read with readCap
	data, err := fs.ReadFile(readCap)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(data) != string(payload) {
		t.Fatalf("expected %q, got %q", payload, data)
	}

	// 4. Test Stream handle OpenFile
	f, err := fs.OpenFile(readCap, os.O_RDONLY, 0644)
	if err != nil {
		t.Fatalf("failed to open stream: %v", err)
	}
	defer f.Close()

	buf := make([]byte, len(payload))
	n, err := f.Read(buf)
	if err != nil || n != len(payload) {
		t.Fatalf("stream read failed: %v", err)
	}
}
