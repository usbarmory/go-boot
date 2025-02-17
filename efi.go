// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build amd64

package main

import (
	"runtime"
	_ "unsafe"

	"github.com/usbarmory/tamago/amd64"
	"github.com/usbarmory/tamago/soc/intel/rtc"
	"github.com/usbarmory/tamago/soc/intel/uart"
)

// Peripheral registers
const (
	// Keyboard controller port
	KBD_PORT = 0x64

	// Communication port
	COM1 = 0x3f8
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
)

//go:linkname ramStart runtime.ramStart
var ramStart uint64 = 0x10000000

//go:linkname ramSize runtime.ramSize
var ramSize uint64 = 0x40000000

//go:linkname nanotime1 runtime.nanotime1
func nanotime1() int64 {
	return int64(float64(AMD64.TimerFn())*AMD64.TimerMultiplier) + AMD64.TimerOffset
}

//go:linkname printk runtime.printk
func printk(c byte) {
	UART0.Tx(c)
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

	runtime.Exit = func(_ int32) {
		AMD64.Reset()
	}
}

func init() {
	// Real-Time Clock
	RTC = &rtc.RTC{}

	if t, err := RTC.Now(); err == nil {
		AMD64.SetTimer(t.UnixNano())
	}
}
