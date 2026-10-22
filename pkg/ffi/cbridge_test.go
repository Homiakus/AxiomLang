package ffi

import (
	"strings"
	"testing"
)

func TestCBridgeRegistrationAndLLVMEmit(t *testing.T) {
	bridge := NewCBridge()

	err := bridge.RegisterForeignFunction("libc.so.6", "puts", TypeInt32, TypeCString)
	if err != nil {
		t.Fatalf("failed to register puts: %v", err)
	}

	err = bridge.RegisterForeignFunction("libm.so", "sin", TypeFloat64, TypeFloat64)
	if err != nil {
		t.Fatalf("failed to register sin: %v", err)
	}

	putsDecl, err := bridge.EmitLLVMDeclare("puts")
	if err != nil {
		t.Fatalf("failed to emit puts LLVM declare: %v", err)
	}
	expectedPuts := "declare i32 @puts(i8*)"
	if putsDecl != expectedPuts {
		t.Errorf("expected %q, got %q", expectedPuts, putsDecl)
	}

	sinDecl, err := bridge.EmitLLVMDeclare("sin")
	if err != nil {
		t.Fatalf("failed to emit sin LLVM declare: %v", err)
	}
	expectedSin := "declare double @sin(double)"
	if sinDecl != expectedSin {
		t.Errorf("expected %q, got %q", expectedSin, sinDecl)
	}

	_, err = bridge.EmitLLVMDeclare("non_existent")
	if err == nil || !strings.Contains(err.Error(), "not registered") {
		t.Errorf("expected error for non existent function")
	}
}
