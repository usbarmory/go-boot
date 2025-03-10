// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

// func callService(fn uint64, a1, a2, a3, a4 uint64) (status uint64)
TEXT ·callService(SB),$0-48
	MOVQ	fn+0(FP), DI

	// Unified Extensible Firmware Interface (UEFI) Specification
	// Version 2.10 - 2.3.4.2 Detailed Calling Conventions
	MOVQ	a1+8(FP), CX
	MOVQ	a2+16(FP), DX
	MOVQ	a3+24(FP), R8
	MOVQ	a4+32(FP), R9

	MOVQ	SP, BX		// callee-saved
	ANDQ	$~15, SP	// alignment for x86_64 ABI
	ADJSP	$32
	CALL	(DI)
	ADJSP	$-32
	MOVQ	BX, SP

	MOVQ	AX, status+40(FP)
	RET
