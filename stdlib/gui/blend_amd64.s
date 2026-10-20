//go:build amd64

#include "textflag.h"

// func blendRowAVX2(pixels []uint32, color uint32)
// pixels_base+0(FP): *uint32
// pixels_len+8(FP):  int
// pixels_cap+16(FP): int
// color+24(FP):      uint32
TEXT ·blendRowAVX2(SB), NOSPLIT, $0-32
	MOVQ pixels_base+0(FP), DI
	MOVQ pixels_len+8(FP), CX
	MOVL color+24(FP), AX

	TESTQ CX, CX
	JLE done

	// Extract src alpha
	SHRL $24, AX
	ANDL $0xFF, AX

	// If alpha == 0, nothing to do
	TESTL AX, AX
	JZ done

	// If alpha == 255, it's an opaque fill -> use clearPixelsAVX2 logic
	CMPL AX, $255
	JNE blend_setup

	// Broadcast full color to Y0
	MOVL color+24(FP), AX
	VMOVD AX, X0
	VPBROADCASTD X0, Y0

opaque_loop_avx:
	CMPQ CX, $8
	JL opaque_loop_scalar
	VMOVDQU Y0, (DI)
	ADDQ $32, DI
	SUBQ $8, CX
	JMP opaque_loop_avx

opaque_loop_scalar:
	TESTQ CX, CX
	JLE opaque_done
	MOVL AX, (DI)
	ADDQ $4, DI
	DECQ CX
	JMP opaque_loop_scalar

opaque_done:
	VZEROUPPER
	RET

blend_setup:
	// Reload color into AX
	MOVL color+24(FP), AX

	// srcR in BX
	MOVL AX, BX
	SHRL $16, BX
	ANDL $0xFF, BX

	// srcG in DX
	MOVL AX, DX
	SHRL $8, DX
	ANDL $0xFF, DX

	// srcB in R8
	MOVL AX, R8
	ANDL $0xFF, R8

	// srcA in R9
	MOVL AX, R9
	SHRL $24, R9
	MOVL $255, R10
	SUBL R9, R10   // invA in R10

	// Pre-multiply srcR*srcA, srcG*srcA, srcB*srcA
	MOVL BX, R11
	IMULL R9, R11  // srcR*srcA in R11

	MOVL DX, R12
	IMULL R9, R12  // srcG*srcA in R12

	MOVL R8, R13
	IMULL R9, R13  // srcB*srcA in R13

blend_pixel_loop:
	TESTQ CX, CX
	JLE done

	MOVL (DI), AX // dst pixel

	// dstR = (dst >> 16) & 0xFF
	MOVL AX, DX
	SHRL $16, DX
	ANDL $0xFF, DX
	IMULL R10, DX // dstR * invA
	ADDL R11, DX  // + srcR*srcA
	IMULL $32897, DX
	SHRL $23, DX

	// dstG = (dst >> 8) & 0xFF
	MOVL AX, BX
	SHRL $8, BX
	ANDL $0xFF, BX
	IMULL R10, BX // dstG * invA
	ADDL R12, BX  // + srcG*srcA
	IMULL $32897, BX
	SHRL $23, BX

	// dstB = dst & 0xFF
	ANDL $0xFF, AX
	IMULL R10, AX // dstB * invA
	ADDL R13, AX  // + srcB*srcA
	IMULL $32897, AX
	SHRL $23, AX

	// Pack out: (255 << 24) | (outR << 16) | (outG << 8) | outB
	SHLL $16, DX
	SHLL $8, BX
	ORL DX, AX
	ORL BX, AX
	ORL $0xFF000000, AX

	MOVL AX, (DI)
	ADDQ $4, DI
	DECQ CX
	JMP blend_pixel_loop

done:
	RET
