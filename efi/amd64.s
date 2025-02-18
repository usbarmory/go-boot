// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

TEXT cpuinit(SB),NOSPLIT|NOFRAME,$0
	MOVQ	DX, ·systemTable(SB)
	JMP	runtime·rt0_amd64_tamago(SB)
