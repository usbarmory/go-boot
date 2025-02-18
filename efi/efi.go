// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package efi implements a driver for the Unified Extensible Firmware Interface (UEFI)
// interface following the specifications at:
//
//	https://uefi.org/specs/UEFI/2.10/
//
// This package is only meant to be used with `GOOS=tamago` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package efi

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/usbarmory/tamago/dma"
)

// EFI Table Header Signature
const signature = 0x5453595320494249 // TSYS IBI

// EFI Boot Service function prototype
func callService(fn uintptr, a1 uint64, a2 uint64, a3 uint64, a4 uint64) uint64

// BootServices represents an EFI Boot Services instance.
type BootServices struct {
	base uintptr
}

// TableHeader represents the data structure that precedes all of the standard
// EFI table types.
type TableHeader struct {
	Signature  uint64
	Revision   uint32
	HeaderSize uint32
	CRC32      uint32
	Reserved   uint32
}

// SystemTable represents the EFI System Table, containing pointers to the
// runtime and boot services tables.
type SystemTable struct {
	Header               TableHeader
	FirmwareVendor       uint64
	FirmwareRevision     uint32
	_                    uint32
	ConsoleInHandle      uint64
	ConIn                uint64
	ConsoleOutHandle     uint64
	ConOut               uint64
	StandardErrorHandle  uint64
	StdErr               uint64
	RuntimeServices      uint64
	BootServices         uint64
	NumberOfTableEntries uint64
	ConfigurationTable   uint64
}

// MarshalBinary implements the [encoding.BinaryMarshaler] interface.
func (d *SystemTable) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.LittleEndian, d)
	return buf.Bytes(), nil
}

// UnmarshalBinary implements the [encoding.BinaryUnmarshaler] interface.
func (d *SystemTable) UnmarshalBinary(data []byte) (err error) {
	_, err = binary.Decode(data, binary.LittleEndian, d)
	return
}

// GetSystemTable returns the EFI System Table if the runtime has been launched
// as an UEFI application.
func GetSystemTable() (t *SystemTable, err error) {
	t = &SystemTable{}

	if systemTable == 0 {
		return nil, errors.New("EFI System Table pointer is nil")
	}

	buf, _ := t.MarshalBinary()
	r, err := dma.NewRegion(uint(systemTable), len(buf), false)

	if err != nil {
		return
	}

	_, buf = r.Reserve(len(buf), 0)

	if err = t.UnmarshalBinary(buf); err != nil {
		return
	}

	if t.Header.Signature != signature {
		return nil, errors.New("EFI System Table pointer is invalid")
	}

	return
}

// GetBootServices returns an EFI Boot Services instance.
func (d *SystemTable) GetBootServices() (b *BootServices, err error) {
	if d.BootServices == 0 {
		return nil, errors.New("EFI Boot Servies pointer is nil")
	}

	b = &BootServices{
		base: uintptr(d.BootServices),
	}

	return
}
