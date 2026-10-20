//go:build !amd64

package gui

func clearPixelsAVX2(pixels []uint32, color uint32) {
	for i := range pixels {
		pixels[i] = color
	}
}

func blendRowAVX2(pixels []uint32, color uint32) {
	srcA := uint32(color >> 24)
	if srcA == 255 {
		for i := range pixels {
			pixels[i] = color
		}
		return
	}
	if srcA == 0 {
		return
	}

	srcR := (color >> 16) & 0xFF
	srcG := (color >> 8) & 0xFF
	srcB := color & 0xFF
	invA := 255 - srcA

	for i, dst := range pixels {
		dstR := (dst >> 16) & 0xFF
		dstG := (dst >> 8) & 0xFF
		dstB := dst & 0xFF

		outR := (srcR*srcA + dstR*invA) / 255
		outG := (srcG*srcA + dstG*invA) / 255
		outB := (srcB*srcA + dstB*invA) / 255

		pixels[i] = (255 << 24) | (outR << 16) | (outG << 8) | outB
	}
}
