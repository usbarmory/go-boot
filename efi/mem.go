// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package efi

import (
	"bytes"
	"encoding/binary"

	"github.com/u-root/u-root/pkg/boot/bzimage"
)

const (
	// EFI Boot Services offset for GetMemoryMap
	getMemoryMap = 0x38
	maxEntries   = 1000
)

// PageSize represents the EFI page size in bytes
const PageSize = 4096 //  4 KiB

// MemoryMap represents an EFI Memory Descriptor
type MemoryMap struct {
	Type          uint32
	_             uint32
	PhysicalStart uint64
	VirtualStart  uint64
	NumberOfPages uint64
	Attribute     uint64
	_             uint64
}

// MarshalBinary implements the [encoding.BinaryMarshaler] interface.
func (d *MemoryMap) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.LittleEndian, d)
	return buf.Bytes(), nil
}

// UnmarshalBinary implements the [encoding.BinaryUnmarshaler] interface.
func (d *MemoryMap) UnmarshalBinary(data []byte) (err error) {
	_, err = binary.Decode(data, binary.LittleEndian, d)
	return
}

// End returns the descriptor physical end address.
func (d *MemoryMap) PhysicalEnd() uint64 {
	return d.PhysicalStart + d.NumberOfPages * PageSize
}

// Size returns the descriptor size.
func (d *MemoryMap) Size() int {
	return int(d.NumberOfPages * PageSize)
}

// E820() converts an EFI Memory Map entry to an x86 E820 one suitable for use
// after exiting EFI Boot Services.
func (d *MemoryMap) E820() (bzimage.E820Entry, error) {
	e := bzimage.E820Entry{
		Addr: d.PhysicalStart,
		Size: d.NumberOfPages * PageSize,
	}

	// Unified Extensible Firmware Interface (UEFI) Specification
	// Version 2.10 - Table 7.10: Memory Type Usage after ExitBootServices()
	switch d.Type {
	case EfiLoaderCode, EfiLoaderData, EfiBootServicesCode, EfiBootServicesData, EfiConventionalMemory:
		e.MemType = bzimage.RAM
	case EfiPersistentMemory:
		e.MemType = 7; // E820_TYPE_PMEM
	case EfiACPIReclaimMemory:
		e.MemType = bzimage.ACPI
	case EfiACPIMemoryNVS:
		e.MemType = bzimage.NVS
	default:
		e.MemType = bzimage.Reserved
	}

	return e, nil
}

// GetMemoryMap calls EFI_BOOT_SERVICES.GetMemoryMap().
func (s *BootServices) GetMemoryMap() (m []*MemoryMap, mapKey uint64, err error) {
	d := &MemoryMap{}
	t, _ := d.MarshalBinary()

	buf := make([]byte, len(t)*maxEntries)
	size := uint64(len(buf))

	status := callService(
		s.base+getMemoryMap,
		ptrval(&size),
		ptrval(&buf[0]),
		ptrval(&mapKey),
		0,
	)

	for i := 0; i < int(size); i += len(t) {
		if err = d.UnmarshalBinary(buf[i:]); err != nil {
			break
		}

		m = append(m, d)
		d = &MemoryMap{}
	}

	err = parseStatus(status)

	return
}
