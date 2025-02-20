// Copyright (c) WithSecure Corporation
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
	MOVQ	SystemTable_ConIn(DX), AX
	MOVQ	AX, ·conIn(SB)
	MOVQ	SimpleTextInput_ReadKeyStroke(AX), AX
	MOVQ	AX, ·readKeyStroke(SB)

	// 12.4 Simple Text Output Protocol
	MOVQ	SystemTable_ConOut(DX), AX
	MOVQ	AX, ·conOut(SB)
	MOVQ	SimpleTextOutput_OutputString(AX), AX
	MOVQ	AX, ·outputString(SB)

	JMP	runtime·rt0_amd64_tamago(SB)
