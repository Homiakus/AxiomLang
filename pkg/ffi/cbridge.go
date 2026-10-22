package ffi

import (
	"fmt"
	"strings"
)

// FFIType represents supported types across the C/Rust boundary.
type FFIType string

const (
	TypeVoid    FFIType = "void"
	TypeInt32   FFIType = "i32"
	TypeInt64   FFIType = "i64"
	TypeFloat32 FFIType = "f32"
	TypeFloat64 FFIType = "f64"
	TypePointer FFIType = "ptr"
	TypeCString FFIType = "cstring"
)

// ForeignFunction defines the symbol contract with external libraries.
type ForeignFunction struct {
	SymbolName string
	ReturnType FFIType
	ParamTypes []FFIType
	Library    string
}

// CBridge manages zero-copy bindings and invocation descriptors for C/Rust shared libraries (.so / .dll / .dylib).
type CBridge struct {
	LoadedLibraries map[string]bool
	Functions       map[string]ForeignFunction
}

// NewCBridge initializes a new Foreign Function Interface bridge.
func NewCBridge() *CBridge {
	return &CBridge{
		LoadedLibraries: make(map[string]bool),
		Functions:       make(map[string]ForeignFunction),
	}
}

// RegisterForeignFunction registers an external C/Rust symbol into AxiomLang's safe runtime table.
func (b *CBridge) RegisterForeignFunction(lib string, symbol string, ret FFIType, params ...FFIType) error {
	if strings.TrimSpace(symbol) == "" {
		return fmt.Errorf("empty symbol name")
	}
	b.LoadedLibraries[lib] = true
	b.Functions[symbol] = ForeignFunction{
		SymbolName: symbol,
		ReturnType: ret,
		ParamTypes: params,
		Library:    lib,
	}
	return nil
}

// EmitLLVMDeclare emits the LLVM IR external declaration for zero-overhead linkage.
func (b *CBridge) EmitLLVMDeclare(symbol string) (string, error) {
	fn, ok := b.Functions[symbol]
	if !ok {
		return "", fmt.Errorf("symbol '%s' not registered in FFI bridge", symbol)
	}

	llvmRet := mapToLLVMType(fn.ReturnType)
	var llvmParams []string
	for _, p := range fn.ParamTypes {
		llvmParams = append(llvmParams, mapToLLVMType(p))
	}

	return fmt.Sprintf("declare %s @%s(%s)", llvmRet, fn.SymbolName, strings.Join(llvmParams, ", ")), nil
}

func mapToLLVMType(t FFIType) string {
	switch t {
	case TypeVoid:
		return "void"
	case TypeInt32:
		return "i32"
	case TypeInt64:
		return "i64"
	case TypeFloat32:
		return "float"
	case TypeFloat64:
		return "double"
	case TypePointer, TypeCString:
		return "i8*"
	default:
		return "i8*"
	}
}
