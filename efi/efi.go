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
	"unsafe"

	"github.com/usbarmory/tamago/dma"
)

// EFI Table Header Signature
const signature = 0x5453595320494249 // TSYS IBI

// defined in efi.s
func callService(fn uint64, a1, a2, a3, a4 uint64) (status uint64)

// This function helps preparing callService arguments, allowing a single call
// for all EFI services with 4 or less arguments.
//
// Obtaining a pointer in this fashion is typically unsafe and tamago/dma
// package would be best to handle this. However, as arguments are prepared
// right before invoking Go assembly, it is considered safe as it is identical
// as having *uint64 as callService prototype.
func ptrval(ptr any) uint64 {
	var p unsafe.Pointer

	switch v := ptr.(type) {
	case *uint64:
		p = unsafe.Pointer(v)
	case *uint32:
		p = unsafe.Pointer(v)
	case *byte:
		p = unsafe.Pointer(v)
	case *InputKey:
		p = unsafe.Pointer(v)
	default:
		panic("internal error, invalid ptrval")
	}

	return uint64(uintptr(p))
}

// BootServices represents an EFI Boot Services instance.
type BootServices struct {
	base uint64
}

// RuntimeServices represents an EFI Runtime Services instance.
type RuntimeServices struct {
	base uint64
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

	addr, buf := r.Reserve(len(buf), 0)
	defer r.Release(addr)

	if err = t.UnmarshalBinary(buf); err != nil {
		return
	}

	if t.Header.Signature != signature {
		return nil, errors.New("EFI System Table pointer is invalid")
	}

	return
}

// Address returns the EFI System Table pointer.
func (d *SystemTable) Address() uint64 {
	return systemTable
}

// GetBootServices returns an EFI Boot Services instance.
func (d *SystemTable) GetBootServices() (*BootServices, error) {
	if d.BootServices == 0 {
		return nil, errors.New("EFI Boot Services pointer is nil")
	}

	return &BootServices{
		base: d.BootServices,
	}, nil
}

// GetRuntimeServices returns an EFI Runtime Services instance.
func (d *SystemTable) GetRuntimeServices() (*RuntimeServices, error) {
	if d.RuntimeServices == 0 {
		return nil, errors.New("EFI Runtime Services pointer is nil")
	}

	return &RuntimeServices{
		base: d.RuntimeServices,
	}, nil
}
