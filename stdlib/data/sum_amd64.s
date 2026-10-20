//go:build amd64

#include "textflag.h"

// func sumInt64Asm(data []int64) int64
// data_base+0(FP): *int64
// data_len+8(FP):  int
// data_cap+16(FP): int
// ret+24(FP):      int64
TEXT ·sumInt64Asm(SB), NOSPLIT, $0-32
	MOVQ data_base+0(FP), SI
	MOVQ data_len+8(FP), CX
	XORQ AX, AX

	TESTQ CX, CX
	JLE done

	// If length < 8, handle scalar
	CMPQ CX, $8
	JL scalar_loop

	// Clear AVX registers (Y0, Y1)
	VPXOR Y0, Y0, Y0
	VPXOR Y1, Y1, Y1

avx_loop:
	CMPQ CX, $8
	JL avx_reduce

	VMOVDQU (SI), Y2
	VPADDQ Y2, Y0, Y0

	VMOVDQU 32(SI), Y3
	VPADDQ Y3, Y1, Y1

	ADDQ $64, SI
	SUBQ $8, CX
	JMP avx_loop

avx_reduce:
	// Sum Y0 and Y1 into Y0
	VPADDQ Y1, Y0, Y0

	// Extract high 128-bit lane into X1
	VEXTRACTI128 $1, Y0, X1
	// Add X1 and X0 (lower 128-bit)
	VPADDQ X1, X0, X0

	// Move lower 64-bit to AX
	VMOVQ X0, AX

	// Shift high 64-bit of X0 into DX and add
	VPEXTRQ $1, X0, DX
	ADDQ DX, AX

	// Clear AVX upper state
	VZEROUPPER

scalar_loop:
	TESTQ CX, CX
	JLE done

	ADDQ (SI), AX
	ADDQ $8, SI
	DECQ CX
	JMP scalar_loop

done:
	MOVQ AX, ret+24(FP)
	RET
