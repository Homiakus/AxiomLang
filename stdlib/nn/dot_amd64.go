//go:build amd64

package nn

//go:noescape
func dotProductFloat32(a, b []float32) float32
