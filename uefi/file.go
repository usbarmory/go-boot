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
	EFI_FILE_INFO_ID = "09576e92-6d3f-11d2-8e39-00a0c969723b"

	EFI_FILE_PROTOCOL_REVISION  = 0x00010000
	EFI_FILE_PROTOCOL_REVISION2 = 0x00020000

	EFI_FILE_MODE_READ   = 0x0000000000000001
	EFI_FILE_MODE_WRITE  = 0x0000000000000002
	EFI_FILE_MODE_CREATE = 0x8000000000000000

	EFI_FILE_DIRECTORY = 0x0000000000000010
)

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

func toUTF16(s string) (buf []byte) {
	for _, r := range utf16.Encode([]rune(s)) {
		buf = append(buf, byte(r&0xff))
		buf = append(buf, byte(r>>8))
	}

	return append([]byte(buf), []byte{0x00, 0x00}...)
}

// open calls EFI_FILE_PROTOCOL.Open().
func (f *fileProtocol) open(handle uint64, name string, mode uint64) (o *fileProtocol, addr uint64, err error) {
	fileName := toUTF16(name)

	status := callService(ptrval(&f.Open), 5,
		[]uint64{
			handle,
			ptrval(&addr),
			ptrval(&fileName[0]),
			mode,
			0,
		},
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
	status := callService(ptrval(&f.Close), 1,
		[]uint64{
			handle,
		},
	)

	return parseStatus(status)
}

// read calls EFI_FILE_PROTOCOL.Read().
func (f *fileProtocol) read(handle uint64, buf []byte) (n int, err error) {
	size := uint64(len(buf))

	status := callService(ptrval(&f.Read), 3,
		[]uint64{
			handle,
			ptrval(&size),
			ptrval(&buf[0]),
		},
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

	status := callService(ptrval(&f.GetInfo), 4,
		[]uint64{
			handle,
			ptrval(&guid[0]),
			ptrval(&size),
			ptrval(&buf[0]),
		},
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
