// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uefi

import (
	"errors"
	"fmt"
	"unicode/utf16"
)

const (
	EFI_LOADED_IMAGE_PROTOCOL_GUID       = "5b1b31a1-9562-11d2-8e3f-00a0c969723b"
	EFI_SIMPLE_FILE_SYSTEM_PROTOCOL_GUID = "964e5b22-6459-11d2-8e39-00a0c969723b"
	EFI_FILE_INFO_ID                     = "09576e92-6d3f-11d2-8e39-00a0c969723b"

	EFI_LOADED_IMAGE_PROTOCOL_REVISION       = 0x00001000
	EFI_SIMPLE_FILE_SYSTEM_PROTOCOL_REVISION = 0x00010000
	EFI_FILE_PROTOCOL_REVISION               = 0x00010000
	EFI_FILE_PROTOCOL_REVISION2              = 0x00020000

	EFI_FILE_MODE_CREATE = iota
	EFI_FILE_MODE_READ
	EFI_FILE_MODE_WRITE
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

// simpleFileSystem represents an EFI Simple File System Protocol instance.
type simpleFileSystem struct {
	Revision   uint64
	OpenVolume uint64
}

// openVolume calls EFI_SIMPLE_FILE SYSTEM_PROTOCOL.OpenVolume().
func (fs *simpleFileSystem) openVolume(handle uint64) (f *fileProtocol, addr uint64, err error) {
	status := callService(
		ptrval(&fs.OpenVolume),
		handle,
		ptrval(&addr),
		0,
		0,
	)

	if err = parseStatus(status); err != nil {
		return
	}

	f = &fileProtocol{}

	if err = decode(f, handle); err != nil {
		return
	}

	if f.Revision != EFI_FILE_PROTOCOL_REVISION && f.Revision != EFI_FILE_PROTOCOL_REVISION2 {
		return nil, 0, fmt.Errorf("invalid protocol revision (%x)", f)
	}

	return
}

// fileProtocol represents an EFI File Protocol instance.
type fileProtocol struct {
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
}

// fileInfo represents an EFI_FILE_INFO instance.
type fileInfo struct {
	Size             uint64
	FileSize         uint64
	PhysicalSize     uint64
	CreateTime       uint64
	LastAccessTime   uint64
	ModificationTime uint64
	Attribute        uint64
}

// open calls EFI_FILE_PROTOCOL.Open().
func (f *fileProtocol) open(handle uint64, name string, mode uint64) (o *fileProtocol, addr uint64, err error) {
	var fileName []byte

	for _, r := range utf16.Encode([]rune(name)) {
		fileName = append(fileName, byte(r&0xff))
		fileName = append(fileName, byte(r>>8))
	}

	fileName = append([]byte(fileName), []byte{0x00, 0x00}...)

	status := callService(
		ptrval(&f.Open),
		handle,
		ptrval(&addr),
		ptrval(&fileName[0]),
		mode,
	)

	if err = parseStatus(status); err != nil {
		return
	}

	o = &fileProtocol{}

	if err = decode(o, addr); err != nil {
		return
	}

	if o.Revision != EFI_FILE_PROTOCOL_REVISION && o.Revision != EFI_FILE_PROTOCOL_REVISION2 {
		return nil, 0, fmt.Errorf("invalid protocol revision (%x)", f)
	}

	return
}

// getInfo calls EFI_FILE SYSTEM_PROTOCOL.GetInfo().
func (f *fileProtocol) getInfo(handle uint64, guid []byte) (info *fileInfo, err error) {
	buf := make([]byte, 8*7+512)
	size := uint64(len(buf))

	status := callService(
		ptrval(&f.GetInfo),
		handle,
		ptrval(&guid[0]),
		ptrval(&size),
		ptrval(&buf[0]),
	)

	if err = parseStatus(status); err != nil {
		return
	}

	info = &fileInfo{}
	err = unmarshalBinary(buf[0:size], info)

	return
}

// File implements the [fs.File] interface for the EFI File Protocol.
type File struct {
	file *fileProtocol
	addr uint64
	name string
}

// FileInfo implements the [fs.FileInfo] interface for the EFI File Protocol.
type FileInfo struct {
	info *fileInfo
	name string
}

// Name returns the name of the file as presented to Open.
func (fi *FileInfo) Name() string {
	return fi.name
}

// Size returns the file length in bytes.
func (fi *FileInfo) Size() int64 {
	return int64(fi.info.FileSize)
}

// TODO
// 	Mode() FileMode     // file mode bits
//	ModTime() time.Time // modification time
//	IsDir() bool        // abbreviation for Mode().IsDir()
//	Sys() any           // underlying data source (can return nil)

// Stat returns a FileInfo describing the named file from the file system.
func (f *File) Stat() (fi *FileInfo, err error) {
	fi = &FileInfo{
		name: f.name,
	}

	if f.addr == 0 {
		return nil, errors.New("invalid file instance")
	}

	infoType := GUID(EFI_FILE_INFO_ID).Bytes()

	fi.info, err = f.file.getInfo(f.addr, infoType)

	return
}

// FS implements the [fs.FS] interface for an EFI Simple File System.
type FS struct {
	fs     *simpleFileSystem
	volume *File
	addr   uint64
}

// Open opens the named file [File.Close] must be called to release any
// associated resources.
func (fs *FS) Open(name string) (f *File, err error) {
	f = &File{
		name: name,
	}

	if fs.volume == nil || fs.volume.file == nil || fs.volume.addr == 0 {
		return nil, errors.New("invalid file system instance")
	}

	f.file, f.addr, err = fs.volume.file.open(fs.volume.addr, name, EFI_FILE_MODE_READ)

	return
}

// Root returns the EFI Simple File System instance for the current EFI image
// root volume.
func (s *Services) Root() (root *FS, err error) {
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

	root = &FS{
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
