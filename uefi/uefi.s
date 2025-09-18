// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

// func callFn(fn uint64, n int, args []uint64) (status uint64)
TEXT ·callFn(SB),NOSPLIT,$0-48
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

	// set stack to unused RAM boundary (see runtime.ramStackOffset)
	MOVQ	runtime·ramStart(SB), SP
	MOVQ	runtime·ramSize(SB), CX
	ADDQ	CX, SP

	ANDQ	$~15, SP	// alignment for x86_64 ABI

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

	// 5th arguments and above are pushed in reverse order on the stack

	// move to last element
	MOVQ	R13, R15
	ADDQ	$1, R15
	IMULQ	$8, R15
	ADDQ	R15, R12

	MOVQ	R13, R14
	ANDQ	$1, R14
	CMPQ	R13, R14
	JNE	aligned
	PUSHQ	$0		// ensure 16-byte alignment
aligned:
	MOVQ	R13, R14
push:
	SUBQ	$8, R12
	PUSHQ	(R12)
	SUBQ	$1, R13
	CMPQ	R13, $0
	JNE	push
	JMP	call

dummy:
	// avoid unbalanced PUSH/POP Go assembler error
	POPQ	CX
	POPQ	CX

call:
	ADJSP	$(4*8)		// shadow stack
	CALL	(DI)
	ADJSP	$-(4*8)

done:
	MOVQ	BX, SP
	MOVQ	AX, status+40(FP)
ret:
	RET
