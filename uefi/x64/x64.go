// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package x64 provides hardware initialization, automatically on import, for
// the Unified Extensible Firmware Interface (UEFI) application environment
// under a single x86_64 core.
//
// This package is only meant to be used with `GOOS=tamago` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package x64

import (
	"fmt"
	_ "unsafe"

	"github.com/usbarmory/tamago/amd64"
	"github.com/usbarmory/tamago/soc/intel/rtc"
	"github.com/usbarmory/tamago/soc/intel/uart"

	"github.com/usbarmory/go-boot/uefi"
)

// Peripheral registers
const (
	// Keyboard controller port
	KBD_PORT = 0x64

	// Communication port
	COM1 = 0x3f8
)

// set in amd64.s
var (
	imageHandle uint64
	systemTable uint64
	conIn       uint64
	conOut      uint64
)

// Peripheral instances
var (
	// AMD64 core
	AMD64 = &amd64.CPU{}

	// Real-Time Clock
	RTC = &rtc.RTC{}

	// Serial port
	UART0 = &uart.UART{
		Index: 1,
		Base:  COM1,
	}

	// UEFI services
	UEFI = &uefi.Services{}
)

//go:linkname nanotime1 runtime.nanotime1
func nanotime1() int64 {
	return int64(float64(AMD64.TimerFn())*AMD64.TimerMultiplier) + AMD64.TimerOffset
}

// Init takes care of the lower level initialization triggered early in runtime
// setup.
//
//go:linkname Init runtime.hwinit
func Init() {
	// initialize CPU
	AMD64.Init()

	// initialize serial console
	UART0.Init()
}

func init() {
	// Real-Time Clock
	RTC = &rtc.RTC{}

	if t, err := RTC.Now(); err == nil {
		AMD64.SetTimer(t.UnixNano())
	}

	print("initializing EFI services\n")

	if err := UEFI.Init(imageHandle, systemTable); err != nil {
		fmt.Printf("could not initialize EFI services, %v\n", err)
	}

	// allocate runtime heap in UEFI memory
	allocateHeap()
}
