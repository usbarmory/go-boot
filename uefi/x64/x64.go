// Copyright (c) The go-boot authors. All Rights Reserved.
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
	"runtime/goos"
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

// set in x64.s
var (
	imageHandle uint64
	systemTable uint64
	conIn       uint64
	conOut      uint64
)

// Peripheral instances
var (
	// AMD64 core
	AMD64 = &amd64.CPU{
		// required before Init()
		TimerMultiplier: 1,
	}

	// Real-Time Clock
	RTC = &rtc.RTC{}

	// Serial port
	UART0 = &uart.UART{
		Index: 1,
		Base:  COM1,
		DTR:   true,
		RTS:   true,
	}

	// UEFI services
	UEFI = &uefi.Services{}
)

//go:linkname nanotime runtime/goos.Nanotime
func nanotime() int64 {
	return AMD64.GetTime()
}

// Init takes care of the lower level initialization triggered early in runtime
// setup.
//
//go:linkname Init runtime/goos.Hwinit1
func Init() {
	// initialize CPU
	AMD64.Init()

	// disable CPU idle time management
	goos.Idle = nil

	// initialize serial console
	UART0.Init()
}

func init() {
	if t, err := RTC.Now(); err == nil {
		AMD64.SetTime(t.UnixNano())
	}

	Console.ClearScreen()

	print("initializing EFI services\n")

	if err := UEFI.Init(imageHandle, systemTable); err != nil {
		fmt.Printf("could not initialize EFI services, %v\n", err)
	}

	// allocate runtime heap in UEFI memory
	allocateHeap()
}
