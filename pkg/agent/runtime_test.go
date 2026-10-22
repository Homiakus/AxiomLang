package agent

import (
	"strings"
	"testing"
	"time"
)

func TestAgentSandboxBudgetLimits(t *testing.T) {
	budget := ResourceBudget{
		MaxComputeCycles: 1000,
		MaxMemoryBytes:   1024 * 1024, // 1MB
		MaxWallTime:      500 * time.Millisecond,
		AllowedCaps:      []string{"Console.Out", "Net.Resolve"},
	}

	sandbox := NewAgentSandbox("agent_01", budget)

	// Test normal cycle consumption
	if err := sandbox.ConsumeCycles(500); err != nil {
		t.Fatalf("unexpected error consuming cycles: %v", err)
	}

	// Test exceeding compute cycles
	err := sandbox.ConsumeCycles(600)
	if err == nil || !strings.Contains(err.Error(), "EXCEEDED_COMPUTE_BUDGET") {
		t.Fatalf("expected EXCEEDED_COMPUTE_BUDGET error, got: %v", err)
	}

	// Subsequent calls must immediately fail
	if err := sandbox.ConsumeCycles(1); err == nil {
		t.Fatalf("expected execution to be halted on terminated agent")
	}
}

func TestAgentSandboxMemoryLimit(t *testing.T) {
	budget := ResourceBudget{
		MaxComputeCycles: 5000,
		MaxMemoryBytes:   512,
		AllowedCaps:      []string{"Storage.Read"},
	}

	sandbox := NewAgentSandbox("agent_mem", budget)

	if err := sandbox.AllocateMemory(256); err != nil {
		t.Fatalf("failed valid memory allocation: %v", err)
	}

	err := sandbox.AllocateMemory(300)
	if err == nil || !strings.Contains(err.Error(), "OUT_OF_MEMORY_BUDGET") {
		t.Fatalf("expected OUT_OF_MEMORY_BUDGET, got: %v", err)
	}
}

func TestAgentCapabilityIsolation(t *testing.T) {
	budget := ResourceBudget{
		AllowedCaps: []string{"Console.Out", "FileSystem.Read:/tmp/data"},
	}

	sandbox := NewAgentSandbox("agent_sec", budget)

	if !sandbox.CheckCapability("Console.Out") {
		t.Errorf("expected granted Console.Out")
	}
	if !sandbox.CheckCapability("FileSystem.Read:/tmp/data") {
		t.Errorf("expected granted FileSystem.Read:/tmp/data")
	}
	if sandbox.CheckCapability("Net.Connect") {
		t.Errorf("Net.Connect must NOT be granted by ambient authority")
	}
}
