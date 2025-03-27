// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package x64

import (
	"fmt"
	"runtime"
	_ "unsafe"

	"github.com/usbarmory/go-boot/uefi"
)

//go:linkname _unused runtime.ramStart
var _unused uint64 = 0x00100000 // overridden in x64.s

//go:linkname RamSize runtime.ramSize
var RamSize uint64 = 0x2c000000 // 704MB

func allocateHeap() {
	memoryMap, err := UEFI.Boot.GetMemoryMap()

	if err != nil {
		fmt.Printf("WARNING: could not get memory map, %err\n", err)
		return
	}

	heapStart := uint64(0)
	ramStart, ramEnd := runtime.MemRegion()

	// locate runtime heap offset within UEFI memory allocation
	for _, desc := range memoryMap.Descriptors {
		if desc.Type == uefi.EfiLoaderCode && desc.PhysicalStart == ramStart {
			heapStart = desc.PhysicalEnd()
			break
		}
	}

	if heapStart == 0 {
		fmt.Println("WARNING: could not find heap offset")
	}

	if err := UEFI.Boot.AllocatePages(
		uefi.AllocateAddress,
		uefi.EfiLoaderData,
		int(ramEnd-heapStart),
		heapStart,
	); err != nil {
		fmt.Printf("WARNING: could not allocate heap at %x\n", heapStart)
	}
}
