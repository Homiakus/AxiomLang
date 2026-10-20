//go:build amd64

#include "textflag.h"

// func clearPixelsAVX2(pixels []uint32, color uint32)
// pixels_base+0(FP): *uint32
// pixels_len+8(FP):  int
// pixels_cap+16(FP): int
// color+24(FP):      uint32
TEXT ·clearPixelsAVX2(SB), NOSPLIT, $0-32
	MOVQ pixels_base+0(FP), DI
	MOVQ pixels_len+8(FP), CX
	MOVL color+24(FP), AX

	TESTQ CX, CX
	JLE done

	// Broadcast 32-bit color to all 8 lanes of Y0 (256-bit)
	VMOVD AX, X0
	VPBROADCASTD X0, Y0

loop_avx:
	CMPQ CX, $8
	JL loop_scalar

	VMOVDQU Y0, (DI)
	ADDQ $32, DI
	SUBQ $8, CX
	JMP loop_avx

loop_scalar:
	TESTQ CX, CX
	JLE cleanup

	MOVL AX, (DI)
	ADDQ $4, DI
	DECQ CX
	JMP loop_scalar

cleanup:
	VZEROUPPER

done:
	RET
