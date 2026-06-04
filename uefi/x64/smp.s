// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

TEXT ·apstart(SB),NOSPLIT|NOFRAME,$0
	// apply BSP IDT
	MOVQ	$github·com∕usbarmory∕tamago∕amd64·idtptr(SB), AX
	LIDT	(AX)

	// use taskAddress as counting semaphore for SMP enabling
	MOVQ	$(const_taskAddress), BX
	MOVL	$1, AX
	LOCK
	XADDL	AX, 0(BX)
wait:
	// wait NMI from CPU.Task
	CLI
	HLT

	MOVQ	$(const_taskAddress), AX
	MOVQ	task_pc(AX), R12
	CMPQ	R12, $0
	JE	wait

	MOVQ	task_sp(AX), SP
	MOVQ	task_gp(AX), g

	// clear task
	MOVQ	$0, task_sp(AX)
	MOVQ	$0, task_gp(AX)
	MOVQ	$0, task_pc(AX)

	MOVQ	g, DI
	CALL	runtime·settls(SB)
	MOVQ	g, (TLS)

	// enable LAPIC
	MOVL	$(const_LAPIC_SVR), AX
	MOVL	$(1<<const_SVR_ENABLE), (AX)	// set SVR_ENABLE

	// call task target
	STI
	CALL	R12

	// go back to idle state in case we return
	JMP wait
