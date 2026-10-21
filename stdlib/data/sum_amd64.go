//go:build amd64

package data

//go:noescape
func sumInt64Asm(data []int64) int64

//go:noescape
func minMaxFloat64Asm(data []float64) (float64, float64)
