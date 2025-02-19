// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package efi

import (
	"bytes"
	"encoding/binary"
	"unsafe"
)

// EFI Boot Services offset for GetMemoryMap
const getMemoryMap = 0x38
const maxEntries = 1000

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

// GetMemoryMap calls EFI_BOOT_SERVICES.GetMemoryMap().
func (s *BootServices) GetMemoryMap() (m []*MemoryMap, mapKey uint64, err error) {
	d := &MemoryMap{}
	t, _ := d.MarshalBinary()

	buf := make([]byte, len(t)*maxEntries)
	mmap := uint64(uintptr(unsafe.Pointer(&buf[0])))
	size := uint64(len(buf))

	status := callService(
		s.base+getMemoryMap,
		ptrval(&size),
		mmap,
		ptrval(&mapKey),
		0,
	)

	for i := 0; i < int(size); i += len(t) {
		d = &MemoryMap{}

		if err = d.UnmarshalBinary(buf[i:]); err != nil {
			break
		}

		m = append(m, d)
	}

	err = parseStatus(status)

	return
}
