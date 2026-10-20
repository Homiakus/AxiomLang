//go:build amd64

#include "textflag.h"

// func findNonWhitespaceAVX2(s string, offset int) int
// s_base+0(FP):   *byte
// s_len+8(FP):    int
// offset+16(FP):  int
// ret+24(FP):     int
TEXT ·findNonWhitespaceAVX2(SB), NOSPLIT, $0-32
	MOVQ s_base+0(FP), SI
	MOVQ s_len+8(FP), DX
	MOVQ offset+16(FP), AX

	// Bounds check: if offset >= len, return len
	CMPQ AX, DX
	JGE return_len

	// SI = base + offset
	ADDQ AX, SI

loop_scalar:
	CMPQ AX, DX
	JGE return_len

	MOVBLZX (SI), BX
	// Check space (32), tab (9), CR (13), LF (10)
	CMPB BX, $32
	JE next_byte
	CMPB BX, $9
	JE next_byte
	CMPB BX, $10
	JE next_byte
	CMPB BX, $13
	JE next_byte

	// Found non-whitespace
	MOVQ AX, ret+24(FP)
	RET

next_byte:
	INCQ SI
	INCQ AX
	JMP loop_scalar

return_len:
	MOVQ DX, ret+24(FP)
	RET


// func fastCRC32String(s string) uint32
// s_base+0(FP): *byte
// s_len+8(FP):  int
// ret+16(FP):   uint32
TEXT ·fastCRC32String(SB), NOSPLIT, $0-20
	MOVQ s_base+0(FP), SI
	MOVQ s_len+8(FP), CX
	MOVL $0xFFFFFFFF, AX // Initial CRC

loop_crc:
	TESTQ CX, CX
	JLE done_crc

	MOVBLZX (SI), BX
	CRC32B BX, AX
	INCQ SI
	DECQ CX
	JMP loop_crc

done_crc:
	NOTL AX
	MOVL AX, ret+16(FP)
	RET
