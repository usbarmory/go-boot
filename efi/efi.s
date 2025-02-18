// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

// func callService(fn uintptr, a1 uint64, a2 uint64, a3 uint64, a4 uint64) uint64
TEXT ·callService(SB),$0-48
	MOVQ	fn+0(FP), DI
	MOVQ	a1+8(FP), AX
	MOVQ	a2+16(FP), BX
	MOVQ	a3+24(FP), CX
	MOVQ	a4+32(FP), DX

	PUSHQ	AX
	PUSHQ	BX
	PUSHQ	CX
	PUSHQ	DX

	CALL	(DI)
	MOVQ	AX, ret+40(FP)

	POPQ	DX
	POPQ	CX
	POPQ	BX
	POPQ	AX

	RET
