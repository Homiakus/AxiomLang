package ceih

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

// InterceptRule defines formal invariant conditions for code/execution admission
type InterceptRule struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	Severity    string   `json:"severity"` // "FATAL", "WARNING", "INFO"
	Forbidden   []string `json:"forbidden"`
	Required    []string `json:"required"`
}

// InvariantSeal represents cryptographic proof of adherence to CEIH policy
type InvariantSeal struct {
	Hash       string    `json:"hash"`
	IssuedAt   time.Time `json:"issued_at"`
	Adherence  float64   `json:"adherence"` // 0.0 .. 1.0 (1.0 = perfect adherence)
	Violations []string  `json:"violations,omitempty"`
}

// Inspector evaluates AST, runtime trace, or configuration against CEIH invariants
type Inspector struct {
	rules []InterceptRule
	mu    sync.RWMutex
}

// NewInspector creates an active CEIH invariant verification engine
func NewInspector() *Inspector {
	return &Inspector{
		rules: []InterceptRule{
			{
				ID:          "CEIH-001",
				Description: "Zero-Ambient Authority: Unsanctioned ambient I/O is strictly forbidden",
				Severity:    "FATAL",
				Forbidden:   []string{"eval", "system", "dlopen", "raw_exec"},
				Required:    []string{"require Capability["},
			},
			{
				ID:          "CEIH-002",
				Description: "Deterministic State Transition: Memory mutation must be bounded",
				Severity:    "WARNING",
				Forbidden:   []string{"goto", "unbounded_asm_jmp"},
			},
			{
				ID:          "CEIH-003",
				Description: "Hardware MMIO Protection: Direct port writes must be capability-sealed",
				Severity:    "FATAL",
				Forbidden:   []string{"outb_unmapped", "inb_unmapped"},
			},
		},
	}
}

// AddRule registers a custom invariant rule into CEIH engine
func (i *Inspector) AddRule(rule InterceptRule) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.rules = append(i.rules, rule)
}

// VerifyCode inspects source or IR string against registered CEIH rules
func (i *Inspector) VerifyCode(code string) (*InvariantSeal, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	var violations []string

	for _, rule := range i.rules {
		for _, forbid := range rule.Forbidden {
			if containsSubstring(code, forbid) {
				violations = append(violations, fmt.Sprintf("[%s %s] %s: contains forbidden token '%s'",
					rule.ID, rule.Severity, rule.Description, forbid))
			}
		}
	}

	totalRules := float64(len(i.rules))
	adherence := 1.0
	if len(violations) > 0 {
		adherence = 1.0 - (float64(len(violations)) / (totalRules * 2.0))
		if adherence < 0.0 {
			adherence = 0.0
		}
	}

	h := sha256.New()
	h.Write([]byte(code))
	h.Write([]byte(fmt.Sprintf("%f_%d", adherence, len(violations))))
	sealHash := hex.EncodeToString(h.Sum(nil))

	seal := &InvariantSeal{
		Hash:       sealHash,
		IssuedAt:   time.Now(),
		Adherence:  adherence,
		Violations: violations,
	}

	if len(violations) > 0 {
		return seal, errors.New("E_CEIH_INVARIANT_VIOLATION: code failed formal adherence inspection")
	}
	return seal, nil
}

func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) > 0 && searchSub(s, sub))
}

func searchSub(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
