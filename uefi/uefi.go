// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package uefi implements a driver for the Unified Extensible Firmware
// Interface (UEFI) following the specifications at:
//
//	https://uefi.org/specs/UEFI/2.10/
//
// This package is only meant to be used with `GOOS=tamago` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package uefi

import (
	"errors"
	"unsafe"
)

// EFI Table Header Signature
const signature = 0x5453595320494249 // TSYS IBI

// defined in efi.s
func callService(fn uint64, n int, args []uint64) (status uint64)

// This function helps preparing callService arguments, allowing a single call
// for all EFI services.
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
	base        uint64
	imageHandle uint64
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

// Services represents the UEFI services instance.
type Services struct {
	// EFI System Table instance
	SystemTable *SystemTable

	// UEFI services
	Console *Console
	Boot    *BootServices
	Runtime *RuntimeServices

	imageHandle uint64
	systemTable uint64
}

// Init initializes an UEFI services instance using the argument pointers.
func (s *Services) Init(imageHandle uint64, systemTable uint64) (err error) {
	s.imageHandle = imageHandle
	s.systemTable = systemTable

	s.SystemTable = &SystemTable{}

	if err = decode(s.SystemTable, systemTable); err != nil {
		return
	}

	if s.SystemTable.Header.Signature != signature {
		return errors.New("EFI System Table pointer is invalid")
	}

	s.Console = &Console{
		ForceLine:   true,
		ReplaceTabs: 8,
		In:          s.SystemTable.ConIn,
		Out:         s.SystemTable.ConOut,
	}

	s.Boot = &BootServices{
		base:        s.SystemTable.BootServices,
		imageHandle: imageHandle,
	}

	s.Runtime = &RuntimeServices{
		base: s.SystemTable.RuntimeServices,
	}

	return
}

// Handle returns the UEFI image handle pointer.
func (s *Services) ImageHandle() uint64 {
	return s.imageHandle
}

// Address returns the EFI System Table pointer.
func (s *Services) Address() uint64 {
	return s.systemTable
}
