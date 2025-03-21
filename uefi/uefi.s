// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

// func callService(fn uint64, n int, args []uint64) (status uint64)
TEXT ·callService(SB),$0-48
	MOVQ	fn+0(FP), DI
	MOVQ	n+8(FP), R13
	MOVQ	args+16(FP), R12

	// len(args)
	CMPQ	R13, $0
	JE	ret

	// &args[0]
	CMPQ	R12, $0
	JE	ret

	MOVQ	SP, BX		// callee-saved

	// Unified Extensible Firmware Interface (UEFI) Specification
	// Version 2.10 - 2.3.4.2 Detailed Calling Conventions
	MOVQ	(R12), CX	// 1st argument
	SUBQ	$1, R13
	CMPQ	R13, $0
	JE	call

	ADDQ	$8, R12
	MOVQ	(R12), DX	// 2nd argument
	SUBQ	$1, R13
	CMPQ	R13, $0
	JE	call

	ADDQ	$8, R12
	MOVQ	(R12), R8	// 3rd argument
	SUBQ	$1, R13
	CMPQ	R13, $0
	JE	call

	ADDQ	$8, R12
	MOVQ	(R12), R9	// 4th argument
	SUBQ	$1, R13
	CMPQ	R13, $0
	JE	call

	// 5th arguments and above are passed on the stack
	MOVQ	R13, R14
	ANDQ	$1, R14
	CMPQ	R13, R14
	JNE	align
	PUSHQ	$0		// ensure 16-byte alignment
align:
	ANDQ	$~15, SP	// alignment for x86_64 ABI
	ADJSP	$32		// shadow stack
	MOVQ	R13, R14
push:
	ADDQ	$8, R12
	PUSHQ	(R12)
	SUBQ	$1, R13
	CMPQ	R13, $0
	JNE	push

	CALL	(DI)
pop:
	POPQ	CX
	SUBQ	$1, R14
	CMPQ	R14, $0
	JNE	pop
	JMP	done

dummy:
	// balance PUSH/POP Go assembler error for conditional alignment
	ADJSP	$-32
	POPQ	CX

call:
	ANDQ	$~15, SP	// alignment for x86_64 ABI
	ADJSP	$32		// shadow stack
	CALL	(DI)
	ADJSP	$-32

done:
	MOVQ	BX, SP
	MOVQ	AX, status+40(FP)
ret:
	RET
