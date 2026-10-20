//go:build amd64

#include "textflag.h"

// func dotProductFloat32(a, b []float32) float32
// a_base+0(FP):  *float32
// a_len+8(FP):   int
// a_cap+16(FP):  int
// b_base+24(FP): *float32
// b_len+32(FP):  int
// b_cap+40(FP):  int
// ret+48(FP):    float32
TEXT ·dotProductFloat32(SB), NOSPLIT, $0-52
	MOVQ a_base+0(FP), SI
	MOVQ a_len+8(FP), CX
	MOVQ b_base+24(FP), DI
	MOVQ b_len+32(FP), DX

	// CX = min(len(a), len(b))
	CMPQ DX, CX
	JGE have_min
	MOVQ DX, CX
have_min:
	VXORPS X0, X0, X0

	TESTQ CX, CX
	JLE done

	CMPQ CX, $16
	JL scalar_loop

	// Initialize Y0 and Y1 accumulators
	VXORPS Y0, Y0, Y0
	VXORPS Y1, Y1, Y1

fma_loop:
	CMPQ CX, $16
	JL fma_reduce

	VMOVUPS (SI), Y2
	VMOVUPS (DI), Y3
	VFMADD231PS Y2, Y3, Y0

	VMOVUPS 32(SI), Y4
	VMOVUPS 32(DI), Y5
	VFMADD231PS Y4, Y5, Y1

	ADDQ $64, SI
	ADDQ $64, DI
	SUBQ $16, CX
	JMP fma_loop

fma_reduce:
	VADDPS Y1, Y0, Y0

	// Extract high 128-bit lane from Y0 into X1
	VEXTRACTF128 $1, Y0, X1
	VADDPS X1, X0, X0

	// Horizontal add across 4 floats in X0
	VSHUFPS $0x4E, X0, X0, X1
	VADDPS X1, X0, X0
	VSHUFPS $0xB1, X0, X0, X1
	VADDPS X1, X0, X0

	VZEROUPPER

scalar_loop:
	TESTQ CX, CX
	JLE done

	VMOVSS (SI), X1
	VMOVSS (DI), X2
	VMULSS X1, X2, X1
	VADDSS X1, X0, X0

	ADDQ $4, SI
	ADDQ $4, DI
	DECQ CX
	JMP scalar_loop

done:
	VMOVSS X0, ret+48(FP)
	RET
