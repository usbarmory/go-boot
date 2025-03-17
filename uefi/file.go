// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uefi

import (
	"errors"
	"fmt"
)

const (
	EFI_LOADED_IMAGE_PROTOCOL_GUID     = "5b1b31a1-9562-11d2-8e3f-00a0c969723b"
	EFI_LOADED_IMAGE_PROTOCOL_REVISION = 0x00001000

	EFI_SIMPLE_FILE_SYSTEM_PROTOCOL_GUID     = "964e5b22-6459-11d2-8e39-00a0c969723b"
	EFI_SIMPLE_FILE_SYSTEM_PROTOCOL_REVISION = 0x00010000

	EFI_FILE_INFO_ID            = "9576e92-6d3f-11d2-8e39-00a0c969723b"
	EFI_FILE_PROTOCOL_REVISION2 = 0x00020000
)

// LoadedImage represents an EFI Loaded Image Protocol instance.
type LoadedImage struct {
	Revision        uint32
	_               uint32
	ParentHandle    uint64
	SystemTable     uint64
	DeviceHandle    uint64
	FilePath        uint64
	_               uint64
	LoadOptionsSize uint32
	LoadOptions     uint64
	ImageBase       uint64
	ImageSize       uint64
	ImageCodeType   uint64
	ImageDataType   uint64
	Unload          uint64
}

// SimpleFileSystem represents an EFI Simple File System Protocol instance.
type SimpleFileSystem struct {
	Revision   uint64
	OpenVolume uint64
}

// Root represents an EFI image root volume.
type Root struct {
	FS      *SimpleFileSystem
	address uint64
}

// Open calls EFI_SIMPLE_FILE SYSTEM_PROTOCOL.OpenVolume().
func (root *Root) Open() (f *File, err error) {
	var addr uint64

	status := callService(
		ptrval(&root.FS.OpenVolume),
		root.address,
		ptrval(&addr),
		0,
		0,
	)

	if err = parseStatus(status); err != nil {
		return
	}

	f = &File{}

	if err = decode(f, addr); err != nil {
		return
	}

	if f.Revision != EFI_FILE_PROTOCOL_REVISION2 {
		return nil, fmt.Errorf("invalid protocol revision (%x)", f)
	}

	return
}

// FileInfo represents an EFI_FILE_INFO instance.
type FileInfo struct {
	Size uint64
	FileSize uint64
	PhysicalSize uint64
	CreateTime uint64
	LastAccessTime uint64
	ModificationTime uint64
	Attribute uint64
	Filename [512]byte
}

// File represents an EFI File Protocol instance.
type File struct {
	Revision    uint64
	Open        uint64
	Close       uint64
	Delete      uint64
	Read        uint64
	Write       uint64
	GetPosition uint64
	SetPosition uint64
	GetInfo     uint64
	SetInfo     uint64
	Flush       uint64
	OpenEx      uint64
	ReadEx      uint64
	WriteEx     uint64
	FlushEx     uint64
}

// Info calls EFI_FILE SYSTEM_PROTOCOL.GetInfo().
func (f *File) Info() (info *FileInfo, err error) {
	return nil, errors.New("TODO")

	guid := GUID(EFI_FILE_INFO_ID).Bytes()
	buf := make([]byte, 8*7+512)
	size := uint64(len(buf))

	status := callService(
		ptrval(&f.GetInfo),
		0, // FIXME
		ptrval(&guid),
		ptrval(&size),
		ptrval(&buf[0]),
	)

	if err = parseStatus(status); err != nil {
		return
	}

	return

}

// Root returns EFI Simple File System instance for the current EFI image
// root volume.
func (s *Services) Root() (root *Root, err error) {
	var addr uint64

	if addr, err = s.Boot.HandleProtocol(s.imageHandle, EFI_LOADED_IMAGE_PROTOCOL_GUID); err != nil {
		return
	}

	image := &LoadedImage{}

	if err = decode(image, addr); err != nil {
		return
	}

	if image.Revision != EFI_LOADED_IMAGE_PROTOCOL_REVISION {
		return nil, errors.New("invalid protocol revision")
	}

	if addr, err = s.Boot.HandleProtocol(image.DeviceHandle, EFI_SIMPLE_FILE_SYSTEM_PROTOCOL_GUID); err != nil {
		return
	}

	root = &Root{
		FS:      &SimpleFileSystem{},
		address: addr,
	}

	if err = decode(root.FS, addr); err != nil {
		return
	}

	if root.FS.Revision != EFI_SIMPLE_FILE_SYSTEM_PROTOCOL_REVISION {
		return nil, errors.New("invalid protocol revision")
	}

	return
}

// FileInfo locates a file at the argument path within the current EFI image
// root volume and calls EFI_FILE_PROTOCOL.GetInfo() on it.
func (s *Services) FileInfo(path string) (_ string, err error) {
	var root *Root
	var f *File
	var info *FileInfo

	if root, err = s.Root(); err != nil {
		return
	}

	if f, err = root.Open(); err != nil {
		return
	}

	if info, err = f.Info(); err != nil {
		return
	}

	return "", fmt.Errorf("TODO (%v)", info)
}

// TODO: use io.File ?
