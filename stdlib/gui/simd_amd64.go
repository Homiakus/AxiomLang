//go:build amd64

package gui

// clearPixelsAVX2 fills pixels with color using AVX2 256-bit SIMD instructions.
// Defined in clear_amd64.s
//
//go:noescape
func clearPixelsAVX2(pixels []uint32, color uint32)

// blendRowAVX2 blends a row of pixels with a constant color using AVX2 SIMD.
// Defined in blend_amd64.s
//
//go:noescape
func blendRowAVX2(pixels []uint32, color uint32)
