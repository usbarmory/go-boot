// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

TEXT cpuinit(SB),NOSPLIT|NOFRAME,$0
	// Unified Extensible Firmware Interface (UEFI) Specification
	// Version 2.10 - 2.3.4.1 Handoff State
	MOVQ	CX, ·imageHandle(SB)
	MOVQ	DX, ·systemTable(SB)
	JMP	runtime·rt0_amd64_tamago(SB)
