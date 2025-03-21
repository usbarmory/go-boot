// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uefi

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
)

const (
	EFI_LOADED_IMAGE_PROTOCOL_GUID       = "5b1b31a1-9562-11d2-8e3f-00a0c969723b"
	EFI_SIMPLE_FILE_SYSTEM_PROTOCOL_GUID = "964e5b22-6459-11d2-8e39-00a0c969723b"

	EFI_LOADED_IMAGE_PROTOCOL_REVISION       = 0x00001000
	EFI_SIMPLE_FILE_SYSTEM_PROTOCOL_REVISION = 0x00010000
)

// FilePath represents an EFI File Path Media Device Path instance.
type FilePath struct {
	Type     uint8
	SubType  uint8
	Length   uint16
	PathName []byte
}

// Bytes converts the descriptor structure to byte array format.
func (d *FilePath) Bytes() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, d.Type)
	binary.Write(buf, binary.LittleEndian, d.SubType)
	binary.Write(buf, binary.LittleEndian, d.Length)
	binary.Write(buf, binary.LittleEndian, d.PathName)

	// Device Path End
	binary.Write(buf, binary.LittleEndian, []byte{
		0x7f, // Type    - End of Hardware Device Path
		0xff, // SubType - End Entire Device Path
		0x04, // Length
	})

	return buf.Bytes()
}

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
	status := callService(ptrval(&root.OpenVolume), 2,
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
	fs     *simpleFileSystem
	volume *File
	addr   uint64
}

// FilePath returns the EFI Device Path associated with the named file.
func (root *FS) FilePath(name string) (filePath *FilePath) {
	pathName := toUTF16(name)

	return &FilePath{
		Type:     4,
		SubType:  4,
		Length:   uint16(4 + len(pathName)),
		PathName: pathName,
	}
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

// Root returns an EFI Simple File System instance for the current EFI image
// root volume.
func (s *Services) Root() (root *FS, err error) {
	var addr uint64

	if addr, err = s.Boot.HandleProtocol(s.imageHandle, EFI_LOADED_IMAGE_PROTOCOL_GUID); err != nil {
		return
	}

	image := &loadedImage{}

	if err = decode(image, addr); err != nil {
		return
	}

	if image.Revision != EFI_LOADED_IMAGE_PROTOCOL_REVISION {
		return nil, errors.New("invalid protocol revision")
	}

	if addr, err = s.Boot.HandleProtocol(image.DeviceHandle, EFI_SIMPLE_FILE_SYSTEM_PROTOCOL_GUID); err != nil {
		return
	}

	root = &FS{
		image:  image,
		fs:     &simpleFileSystem{},
		volume: &File{},
		addr:   addr,
	}

	if err = decode(root.fs, addr); err != nil {
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
