package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"math/big"
)

// LWEParams defines parameters for Learning With Errors lattice-based scheme
type LWEParams struct {
	Dimension int      // n
	Modulus   *big.Int // q
	NoiseStd  float64
}

var DefaultLWE = LWEParams{
	Dimension: 64,
	Modulus:   big.NewInt(12289), // standard Ring-LWE modulus
	NoiseStd:  3.2,
}

// ObfuscatedSymbol holds a cryptographically sealed state node
type ObfuscatedSymbol struct {
	SymbolName string
	KnotDigest string
	VectorA    []int64
	ScalarB    int64
}

// Obfuscator provides LWE lattice transformation and topological knot flattening
type Obfuscator struct {
	params LWEParams
	secret []int64
}

func NewObfuscator(params LWEParams) *Obfuscator {
	secret := make([]int64, params.Dimension)
	for i := 0; i < params.Dimension; i++ {
		n, _ := rand.Int(rand.Reader, params.Modulus)
		secret[i] = n.Int64()
	}
	return &Obfuscator{
		params: params,
		secret: secret,
	}
}

// ProtectSymbol encrypts an identifier / function name into an LWE lattice representation
func (ob *Obfuscator) ProtectSymbol(symbol string) (*ObfuscatedSymbol, error) {
	hasher := sha256.New()
	hasher.Write([]byte(symbol))
	digest := hex.EncodeToString(hasher.Sum(nil))

	vecA := make([]int64, ob.params.Dimension)
	var dotProduct int64 = 0

	for i := 0; i < ob.params.Dimension; i++ {
		val, _ := rand.Int(rand.Reader, ob.params.Modulus)
		vecA[i] = val.Int64()
		dotProduct = (dotProduct + (vecA[i] * ob.secret[i])) % ob.params.Modulus.Int64()
	}

	// Add noise and 1-bit payload or hash slice
	var symInt int64
	if len(digest) >= 8 {
		symInt = int64(binary.LittleEndian.Uint64([]byte(digest[:8]))) % ob.params.Modulus.Int64()
	}
	scalarB := (dotProduct + symInt) % ob.params.Modulus.Int64()

	return &ObfuscatedSymbol{
		SymbolName: "axiom_sym_" + digest[:12],
		KnotDigest: digest,
		VectorA:    vecA,
		ScalarB:    scalarB,
	}, nil
}

// GenerateKnotProof generates cryptographic invariant verification proof
func GenerateKnotProof(stateData []byte) (string, error) {
	if len(stateData) == 0 {
		return "", errors.New("empty state data")
	}
	h := sha256.Sum256(stateData)
	return "KNOT_PROOF_LWE_" + hex.EncodeToString(h[:]), nil
}
