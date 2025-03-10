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
const PageSize = 4096 // 4 KiB

// MemoryDescriptor represents an EFI Memory Descriptor
type MemoryDescriptor struct {
	Type          uint32
	_             uint32
	PhysicalStart uint64
	VirtualStart  uint64
	NumberOfPages uint64
	Attribute     uint64
	_             uint64
}

// MarshalBinary implements the [encoding.BinaryMarshaler] interface.
func (d *MemoryDescriptor) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.LittleEndian, d)
	return buf.Bytes(), nil
}

// UnmarshalBinary implements the [encoding.BinaryUnmarshaler] interface.
func (d *MemoryDescriptor) UnmarshalBinary(data []byte) (err error) {
	_, err = binary.Decode(data, binary.LittleEndian, d)
	return
}

// End returns the descriptor physical end address.
func (d *MemoryDescriptor) PhysicalEnd() uint64 {
	return d.PhysicalStart + d.NumberOfPages*PageSize
}

// Size returns the descriptor size.
func (d *MemoryDescriptor) Size() int {
	return int(d.NumberOfPages * PageSize)
}

// E820() converts an EFI Memory Map entry to an x86 E820 one suitable for use
// after exiting EFI Boot Services.
func (d *MemoryDescriptor) E820() (bzimage.E820Entry, error) {
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
		e.MemType = 7 // E820_TYPE_PMEM
	case EfiACPIReclaimMemory:
		e.MemType = bzimage.ACPI
	case EfiACPIMemoryNVS:
		e.MemType = bzimage.NVS
	default:
		e.MemType = bzimage.Reserved
	}

	return e, nil
}

// MemoryMap represents an EFI Memory Map
type MemoryMap struct {
	MapSize           uint64
	Descriptors       []*MemoryDescriptor
	MapKey            uint64
	DescriptorSize    uint64
	DescriptorVersion uint32

	buf []byte
}

// Address returns the EFI Memory Map pointer.
func (m *MemoryMap) Address() uint64 {
	return ptrval(&m.buf[0])
}

// GetMemoryMap calls EFI_BOOT_SERVICES.GetMemoryMap().
func (s *BootServices) GetMemoryMap() (m *MemoryMap, err error) {
	d := &MemoryDescriptor{}
	t, _ := d.MarshalBinary()
	n := len(t)

	m = &MemoryMap{
		MapSize:        uint64(n * maxEntries),
		DescriptorSize: uint64(n),
		buf:            make([]byte, n*maxEntries),
	}

	status := callService(
		s.base+getMemoryMap,
		ptrval(&m.MapSize),
		ptrval(&m.buf[0]),
		ptrval(&m.MapKey),
		ptrval(&m.DescriptorVersion),
	)

	for i := 0; i < int(m.MapSize); i += n {
		if err = d.UnmarshalBinary(m.buf[i:]); err != nil {
			break
		}

		m.Descriptors = append(m.Descriptors, d)
		d = &MemoryDescriptor{}
	}

	err = parseStatus(status)

	return
}
