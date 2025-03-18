// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uefi

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"time"
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

	EFI_FILE_MODE_READ   = 0x0000000000000001
	EFI_FILE_MODE_WRITE  = 0x0000000000000002
	EFI_FILE_MODE_CREATE = 0x8000000000000000

	EFI_FILE_DIRECTORY = 0x0000000000000010
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
	status := callService(
		ptrval(&root.OpenVolume),
		handle,
		ptrval(&addr),
		0,
		0,
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

// efiTime represents an EFI_TIME instance.
type efiTime struct {
	Year       uint16
	Month      uint8
	Day        uint8
	Hour       uint8
	Minute     uint8
	Second     uint8
	_          uint8
	Nanosecond uint8
	TimeZone   int16
	Daylight   uint8
	_          uint32
}

// fileInfo represents an EFI_FILE_INFO instance.
type fileInfo struct {
	Size             uint64
	FileSize         uint64
	PhysicalSize     uint64
	CreateTime       efiTime
	LastAccessTime   efiTime
	ModificationTime efiTime
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

// close calls EFI_FILE_PROTOCOL.Close().
func (f *fileProtocol) close(handle uint64) (err error) {
	status := callService(
		ptrval(&f.Close),
		handle,
		0,
		0,
		0,
	)

	return parseStatus(status)
}

// read calls EFI_FILE_PROTOCOL.Read().
func (f *fileProtocol) read(handle uint64, buf []byte) (n int, err error) {
	size := uint64(len(buf))

	status := callService(
		ptrval(&f.Read),
		handle,
		ptrval(&size),
		ptrval(&buf[0]),
		0,
	)

	if status == EFI_DEVICE_ERROR || size == 0 {
		return 0, io.EOF
	}

	return int(size), parseStatus(status)
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
	addr uint64
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

// Mode returns the file mode bits.
func (fi *FileInfo) Mode() fs.FileMode {
	if fi.IsDir() {
		return fs.ModeDir
	}

	return 0
}

// ModTime returns the file modification time.
func (fi *FileInfo) ModTime() time.Time {
	m := fi.info.ModificationTime
	tz := time.FixedZone("tz", int(m.TimeZone))

	return time.Date(
		int(m.Year),
		time.Month(m.Month),
		int(m.Day),
		int(m.Hour),
		int(m.Minute),
		int(m.Second),
		int(m.Nanosecond),
		tz,
	)
}

// IsDir reports whether fi describes a directory.
func (fi *FileInfo) IsDir() bool {
	return (fi.info.Attribute & EFI_FILE_DIRECTORY) > 0
}

// Sys returns the underlying data source pointer.
func (fi *FileInfo) Sys() any {
	return fi.addr
}

// Stat returns a FileInfo describing the named file from the file system.
func (f *File) Stat() (fs.FileInfo, error) {
	var err error

	fi := &FileInfo{
		name: f.name,
		addr: f.addr,
	}

	if f.addr == 0 {
		return nil, errors.New("invalid file instance")
	}

	infoType := GUID(EFI_FILE_INFO_ID).Bytes()

	if fi.info, err = f.file.getInfo(f.addr, infoType); err != nil {
		return nil, err
	}

	return fs.FileInfo(fi), nil
}

// Read reads up to len(b) bytes from the File and stores them in b. It returns
// the number of bytes read and any error encountered. At end of file, Read
// returns 0, io.EOF.
func (f *File) Read(b []byte) (n int, err error) {
	if f.addr == 0 {
		return 0, errors.New("invalid file instance")
	}

	return f.file.read(f.addr, b)
}

// Close closes the File, rendering it unusable for I/O.
func (f *File) Close() (err error) {
	if f.addr == 0 {
		return errors.New("invalid file instance")
	}

	return f.file.close(f.addr)
}

// FS implements the [fs.FS] interface for an EFI Simple File System.
type FS struct {
	fs     *simpleFileSystem
	volume *File
	addr   uint64
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
