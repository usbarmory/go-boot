// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package x64

import (
	"fmt"
	"runtime"
	_ "unsafe"

	"github.com/usbarmory/tamago/dma"

	"github.com/usbarmory/go-boot/uefi"
)

//go:linkname _unused runtime.ramStart
var _unused uint64 = 0x00100000 // overridden in x64.s

//go:linkname RamSize runtime.ramSize
var RamSize uint64 = 0x2c000000 // 704MB

var dmaSize int

func allocateHeap() {
	memoryMap, err := UEFI.Boot.GetMemoryMap()

	if err != nil {
		fmt.Printf("WARNING: could not get memory map, %v\n", err)
		return
	}

	heapStart := uint64(0)
	ramStart, ramEnd := runtime.MemRegion()

	// locate runtime heap offset within UEFI memory allocation
	for _, desc := range memoryMap.Descriptors {
		if heapStart > 0 {
			// increase RamSize to cover entire page
			ramEnd = heapStart + uint64(desc.Size())
			RamSize = ramEnd - ramStart
			break
		}

		if desc.Type == uefi.EfiLoaderCode && desc.PhysicalStart == ramStart {
			heapStart = desc.PhysicalEnd()
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
		fmt.Printf("WARNING: could not allocate heap at %x, %v\n", heapStart, err)
	}
}

// AllocateDMA initializes the global memory region for DMA buffer allocation
// at the end of allocated Go runtime heap space.
//
// Once allocated `RamSize` is diminished accordingly, reducing available Go
// runtime memory. It is the caller responsibility to ensure that the DMA
// allocation takes place over an unused memory area.
func AllocateDMA(size int) (err error) {
	_, ramEnd := runtime.MemRegion()

	if size <= dmaSize {
		return
	}

	RamSize -= uint64(size)
	err = dma.Init(uint(ramEnd) - uint(size), size)

	if err != nil {
		RamSize += uint64(size)
	} else {
		dmaSize = size
	}

	return
}
