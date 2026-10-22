package agent

import (
	"fmt"
	"sync"
	"time"
)

// ResourceBudget defines execution constraints assigned to an autonomous agent sandbox.
type ResourceBudget struct {
	MaxComputeCycles int64         // Maximum instructions / cycles permitted
	MaxMemoryBytes   int64         // Maximum RAM consumption in bytes
	MaxWallTime      time.Duration // Wall-clock timeout threshold
	AllowedCaps      []string      // Whitelisted capabilities (Zero-Ambient Authority)
}

// AgentExecutionState tracks live metrics against the assigned budget.
type AgentExecutionState struct {
	AgentID       string
	Budget        ResourceBudget
	CyclesUsed    int64
	MemoryUsed    int64
	StartTime     time.Time
	Terminated    bool
	TerminationReason string
	mu            sync.Mutex
}

// NewAgentSandbox creates an isolated execution container with strict resource limits.
func NewAgentSandbox(agentID string, budget ResourceBudget) *AgentExecutionState {
	return &AgentExecutionState{
		AgentID:   agentID,
		Budget:    budget,
		StartTime: time.Now(),
	}
}

// ConsumeCycles increments cycle usage and aborts if budget is exceeded.
func (s *AgentExecutionState) ConsumeCycles(count int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Terminated {
		return fmt.Errorf("agent %s terminated: %s", s.AgentID, s.TerminationReason)
	}

	s.CyclesUsed += count
	if s.Budget.MaxComputeCycles > 0 && s.CyclesUsed > s.Budget.MaxComputeCycles {
		s.Terminated = true
		s.TerminationReason = fmt.Sprintf("EXCEEDED_COMPUTE_BUDGET: %d > %d", s.CyclesUsed, s.Budget.MaxComputeCycles)
		return fmt.Errorf("agent %s: %s", s.AgentID, s.TerminationReason)
	}

	if s.Budget.MaxWallTime > 0 && time.Since(s.StartTime) > s.Budget.MaxWallTime {
		s.Terminated = true
		s.TerminationReason = "TIMEOUT_EXCEEDED"
		return fmt.Errorf("agent %s: timeout reached", s.AgentID)
	}

	return nil
}

// AllocateMemory simulates heap allocation within the agent sandbox.
func (s *AgentExecutionState) AllocateMemory(bytes int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Terminated {
		return fmt.Errorf("agent %s terminated: %s", s.AgentID, s.TerminationReason)
	}

	s.MemoryUsed += bytes
	if s.Budget.MaxMemoryBytes > 0 && s.MemoryUsed > s.Budget.MaxMemoryBytes {
		s.Terminated = true
		s.TerminationReason = fmt.Sprintf("OUT_OF_MEMORY_BUDGET: %d bytes > %d bytes", s.MemoryUsed, s.Budget.MaxMemoryBytes)
		return fmt.Errorf("agent %s: %s", s.AgentID, s.TerminationReason)
	}

	return nil
}

// CheckCapability asserts that the agent holds the necessary capability.
func (s *AgentExecutionState) CheckCapability(capName string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, c := range s.Budget.AllowedCaps {
		if c == capName || c == "*" {
			return true
		}
	}
	return false
}
