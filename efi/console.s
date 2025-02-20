// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

// func consoleInput(k *inputKey) (status uint64)
TEXT ·consoleInput(SB),$0-16
	MOVQ	·readKeyStroke(SB), DI

	MOVQ	·conIn(SB), CX
	MOVQ	inputKey+0(FP), DX

	CALL	DI
	MOVQ	AX, status+8(FP)

	RET

// func consoleOutput(c *byte) (status uint64)
TEXT ·consoleOutput(SB),$0-16
	MOVQ	·outputString(SB), DI

	MOVQ	·conOut(SB), CX
	MOVQ	c+0(FP), DX

	CALL	DI
	MOVQ	AX, status+8(FP)

	RET
