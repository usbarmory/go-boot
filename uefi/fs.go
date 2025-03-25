// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uefi

import (
	"errors"
	"fmt"
	"io/fs"
)

const (
	EFI_LOADED_IMAGE_PROTOCOL_GUID             = "5b1b31a1-9562-11d2-8e3f-00a0c969723b"
	EFI_LOADED_IMAGE_DEVICE_PATH_PROTOCOL_GUID = "09576e91-6d3f-11d2-8e39-00a0c969723b"
	EFI_SIMPLE_FILE_SYSTEM_PROTOCOL_GUID       = "964e5b22-6459-11d2-8e39-00a0c969723b"

	EFI_LOADED_IMAGE_PROTOCOL_REVISION       = 0x00001000
	EFI_SIMPLE_FILE_SYSTEM_PROTOCOL_REVISION = 0x00010000
)

// loadedImage represents an EFI Loaded Image Protocol instance.
type loadedImage struct {
	Revision        uint32
	_               uint32
	ParentHandle    uint64
	SystemTable     uint64
	DeviceHandle    uint64
	FilePath        uint64
	_               uint64
	LoadOptionsSize uint32
	_               uint32
	LoadOptions     uint64
	ImageBase       uint64
	ImageSize       uint64
	ImageCodeType   uint64
	ImageDataType   uint64
	Unload          uint64
}

// simpleFileSystem represents an EFI Simple File System Protocol instance.
type simpleFileSystem struct {
	Revision   uint64
	OpenVolume uint64
}

// openVolume calls EFI_SIMPLE_FILE SYSTEM_PROTOCOL.OpenVolume().
func (root *simpleFileSystem) openVolume(handle uint64) (f *fileProtocol, addr uint64, err error) {
	status := callService(ptrval(&root.OpenVolume),
		[]uint64{
			handle,
			ptrval(&addr),
		},
	)

	if err = parseStatus(status); err != nil {
		return
	}

	f = &fileProtocol{}

	if err = decode(f, addr); err != nil {
		return
	}

	if f.Revision != EFI_FILE_PROTOCOL_REVISION && f.Revision != EFI_FILE_PROTOCOL_REVISION2 {
		return nil, 0, fmt.Errorf("invalid protocol revision (%x)", f)
	}

	return
}

// FS implements the [fs.FS] interface for an EFI Simple File System.
type FS struct {
	image  *loadedImage
	device uint64
	addr   uint64

	fs     *simpleFileSystem
	volume *File
}

// Open opens the named file [File.Close] must be called to release any
// associated resources.
func (root *FS) Open(name string) (fs.File, error) {
	var err error

	f := &File{
		name: name,
	}

	if root.volume == nil || root.volume.file == nil || root.volume.addr == 0 {
		return nil, errors.New("invalid file system instance")
	}

	if f.file, f.addr, err = root.volume.file.open(root.volume.addr, name, EFI_FILE_MODE_READ); err != nil {
		return nil, err
	}

	return fs.File(f), nil
}

func (s *BootServices) loadImageHandle(imageHandle uint64) (image *loadedImage, err error) {
	var addr uint64

	if addr, err = s.HandleProtocol(imageHandle, EFI_LOADED_IMAGE_PROTOCOL_GUID); err != nil {
		return
	}

	image = &loadedImage{}

	if err = decode(image, addr); err != nil {
		return
	}

	if image.Revision != EFI_LOADED_IMAGE_PROTOCOL_REVISION {
		return nil, errors.New("invalid protocol revision")
	}

	return
}

// Root returns an EFI Simple File System instance for the current EFI image
// root volume.
func (s *Services) Root() (root *FS, err error) {
	root = &FS{
		fs:     &simpleFileSystem{},
		volume: &File{},
	}

	if root.image, err = s.Boot.loadImageHandle(s.imageHandle); err != nil {
		return
	}

	if root.device, err = s.Boot.HandleProtocol(root.image.DeviceHandle, EFI_LOADED_IMAGE_DEVICE_PATH_PROTOCOL_GUID); err != nil {
		return
	}

	if root.addr, err = s.Boot.HandleProtocol(root.image.DeviceHandle, EFI_SIMPLE_FILE_SYSTEM_PROTOCOL_GUID); err != nil {
		return
	}

	if err = decode(root.fs, root.addr); err != nil {
		return
	}

	if root.fs.Revision != EFI_SIMPLE_FILE_SYSTEM_PROTOCOL_REVISION {
		return nil, errors.New("invalid protocol revision")
	}

	if root.volume.file, root.volume.addr, err = root.fs.openVolume(root.addr); err != nil {
		return
	}

	return
}
