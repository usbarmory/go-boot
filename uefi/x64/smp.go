// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package x64

import (
	"fmt"
	"unsafe"

	"github.com/usbarmory/tamago/amd64"
	"github.com/usbarmory/tamago/amd64/lapic"
)

const (
	// AP task address
	taskAddress = 0x6020

	// Intel Local Advanced Programmable Interrupt Controller
	LAPIC_SVR  = amd64.LAPIC_BASE + lapic.LAPIC_SVR
	SVR_ENABLE = lapic.SVR_ENABLE
)

// defined in smp.s
func apstart()

// task represents a CPU task
type task struct {
	sp uint64 // stack pointer
	gp uint64 // G
	pc uint64 // fn
}

// InitSMP enables Secure Multiprocessor (SMP) operation by initializing
// Application Processors.
//
// After initialization [runtime.NumCPU] or [runtime.GOMAXPROCS] can be used to
// verify SMP use by the runtime.
func InitSMP() (err error) {
	mp, err := UEFI.Boot.GetMultiProcessor()

	if err != nil {
		return fmt.Errorf("could not locate multiprocessor services, %v", err)
	}

	// trap CPU exceptions
	AMD64.EnableExceptions()

	fn := apstart
	pc := **(**uintptr)(unsafe.Pointer(&fn))

	// initialize APs
	if err := mp.StartupAllAPs(pc, true, 0); err != nil {
		return fmt.Errorf("could not start APs, %v", err)
	}

	// set runtime CPUs
	AMD64.GOMAXPROCS(amd64.NumCPU())

	return
}
