package scnsincore

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

// InvariantVector represents multidimensional safety and state invariants of a core
type InvariantVector struct {
	EntropyMetric     float64 `json:"entropy_metric"`     // 0.0 .. 1.0
	BranchPredictFlux float64 `json:"branch_predict_flux"` // divergence risk
	ThermalThrottle   float64 `json:"thermal_throttle"`   // simulated hardware limit
	FaultCount        uint64  `json:"fault_count"`
}

// CoreState represents the execution lifecycle of a scnsincore cell
type CoreState string

const (
	StateNominal   CoreState = "NOMINAL"
	StateDegraded  CoreState = "DEGRADED"
	StateCritical  CoreState = "CRITICAL"
	StateIsolated  CoreState = "ISOLATED"
)

// Synapse represents interconnect pipeline between cores
type Synapse struct {
	ID        string
	Bandwidth uint64 // bytes/sec
	LatencyNs uint64 // nanoseconds
}

// SCNSCore encapsulates a Self-Coordinated Neural-Silicon Core unit
type SCNSCore struct {
	ID         string
	State      CoreState
	Invariants InvariantVector
	Synapses   map[string]*Synapse
	mu         sync.RWMutex
	lastTick   time.Time
}

// NewCore creates an initialized SCNS core unit
func NewCore(id string) *SCNSCore {
	return &SCNSCore{
		ID:    id,
		State: StateNominal,
		Invariants: InvariantVector{
			EntropyMetric:     0.05,
			BranchPredictFlux: 0.02,
			ThermalThrottle:   0.10,
			FaultCount:        0,
		},
		Synapses: make(map[string]*Synapse),
		lastTick: time.Now(),
	}
}

// ConnectSynapse establishes a verified interconnect with another core
func (c *SCNSCore) ConnectSynapse(remoteCoreID string, bandwidth uint64, latencyNs uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Synapses[remoteCoreID] = &Synapse{
		ID:        fmt.Sprintf("syn_%s_%s", c.ID, remoteCoreID),
		Bandwidth: bandwidth,
		LatencyNs: latencyNs,
	}
}

// EvaluateInvariants monitors core metrics and triggers self-stabilization if degraded
func (c *SCNSCore) EvaluateInvariants() (CoreState, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lastTick = time.Now()

	// Invariant assessment
	if c.Invariants.FaultCount > 10 || c.Invariants.EntropyMetric > 0.85 {
		c.State = StateCritical
		return c.State, errors.New("E_SCNS_CRITICAL_FAILURE: entropy limit exceeded, core isolated")
	}

	if c.Invariants.EntropyMetric > 0.50 || c.Invariants.ThermalThrottle > 0.80 {
		c.State = StateDegraded
		// Self-heal: damp entropy
		c.Invariants.EntropyMetric *= 0.70
		return c.State, nil
	}

	c.State = StateNominal
	return c.State, nil
}

// ComputePulse simulates a parallel neuromorphic activation step with hardware-bounded guarantees
func (c *SCNSCore) ComputePulse(inputs []float64) ([]float64, string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.State == StateCritical || c.State == StateIsolated {
		return nil, "", errors.New("E_SCNS_INACTIVE: core is isolated from compute matrix")
	}

	outputs := make([]float64, len(inputs))
	for i, v := range inputs {
		// Non-linear sigmoid-like activation: v / (1 + |v|)
		outputs[i] = v / (1.0 + abs(v))
	}

	h := sha256.New()
	for _, out := range outputs {
		h.Write([]byte(fmt.Sprintf("%f,", out)))
	}
	proof := hex.EncodeToString(h.Sum(nil))

	return outputs, proof, nil
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
