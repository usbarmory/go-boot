// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

TEXT cpuinit(SB),NOSPLIT|NOFRAME,$0
	// Unified Extensible Firmware Interface (UEFI) Specification
	// Version 2.10

	// 2.3.4.1 Handoff State
	MOVQ	CX, ·imageHandle(SB)
	MOVQ	DX, ·systemTable(SB)

	// 12.3 Simple Text Input Protocol
	MOVQ	48(DX), AX			// SystemTable_ConIn(DX)
	MOVQ	AX, ·conIn(SB)			// Console.In

	// 12.4 Simple Text Output Protocol
	MOVQ	64(DX), AX			// SystemTable_ConOut(DX)
	MOVQ	AX, ·conOut(SB)			// Console.Out

	// Enable SSE
	CALL	sse_enable(SB)

	// ramStart is relocated based on build time variable IMAGE_BASE.
	MOVQ	$runtime·text(SB), AX
	SUBQ	$(64*1024), AX
	MOVQ	AX, runtime·ramStart(SB)

	MOVQ	runtime·ramStart(SB), SP
	MOVQ	runtime·ramSize(SB), AX
	MOVQ	runtime·ramStackOffset(SB), BX
	ADDQ	AX, SP
	SUBQ	BX, SP

	JMP	runtime·rt0_amd64_tamago(SB)
