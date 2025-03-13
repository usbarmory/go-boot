// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build amd64

package efi

import (
	"fmt"
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

// Services represents the UEFI service instances.
type Services struct {
	SystemTable     *SystemTable
	BootServices    *BootServices
	RuntimeServices *RuntimeServices
}

// Peripheral instances
var (
	// AMD64 core
	AMD64 = &amd64.CPU{}

	// EFI Console I/O
	CONSOLE = &Console{}

	// Real-Time Clock
	RTC = &rtc.RTC{}

	// Serial port
	UART0 = &uart.UART{
		Index: 1,
		Base:  COM1,
	}

	// UEFI services
	UEFI = &Services{}
)

// set in amd64.s
var (
	imageHandle uint64
	systemTable uint64
)

//go:linkname _unused runtime.ramStart
var _unused uint64 = 0x00100000 // overridden in amd64.s

//go:linkname RamSize runtime.ramSize
var RamSize uint64 = 0x20000000 // 512MB

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

	// cache UEFI services
	UEFI.SystemTable, _ = GetSystemTable()

	if UEFI.SystemTable != nil {
		UEFI.BootServices, _ = UEFI.SystemTable.GetBootServices()
		UEFI.RuntimeServices, _ = UEFI.SystemTable.GetRuntimeServices()
	}

	if UEFI.BootServices == nil {
		return
	}

	memoryMap, err := UEFI.BootServices.GetMemoryMap()

	if err != nil {
		fmt.Printf("WARNING: could not get memory map, %err\n", err)
		return
	}

	heapStart := uint64(0)
	ramStart, ramEnd := runtime.MemRegion()

	// locate runtime heap offset according to UEFI memory allocation
	for _, desc := range memoryMap.Descriptors {
		if desc.Type == EfiLoaderCode && desc.PhysicalStart == ramStart {
			heapStart = desc.PhysicalEnd()
			break
		}
	}

	if heapStart == 0 {
		fmt.Println("WARNING: could not find heap offset")
	}

	// reserve runtime heap in UEFI memory
	if err := UEFI.BootServices.AllocatePages(
		AllocateAddress,
		EfiLoaderData,
		int(ramEnd-heapStart),
		heapStart,
	); err != nil {
		fmt.Printf("WARNING: could not allocate heap at %x\n", heapStart)
	}
}
