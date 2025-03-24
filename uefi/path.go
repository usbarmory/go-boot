// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uefi

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/usbarmory/tamago/dma"
)

const (
	bufferSize = (1 << 16)
	maxDepth   = 16
)

// DevicePath represents an EFI Generic Device Path Node structure.
type DevicePathNode struct {
	Type    uint8
	SubType uint8
	Length  uint16
}

// Bytes converts the descriptor structure to byte array format.
func (d *DevicePathNode) Bytes() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, d.Type)
	binary.Write(buf, binary.LittleEndian, d.SubType)
	binary.Write(buf, binary.LittleEndian, d.Length)

	return buf.Bytes()
}

// DevicePath represents an EFI Device Path Protocol node.
type DevicePath struct {
	DevicePathNode
	Data []byte
}

// While we could use UEFI functions to perform the same, we prefer to keep
// have control on this parsing tiven that UEFI firmware does not handle
// gracefully invalid pointers (e.g. DoS condition).
func (root *FS) devicePath() (devicePath []*DevicePath, desc []byte, err error) {
	addr := uint(root.device)
	off := uint(0)

	r, err := dma.NewRegion(uint(addr), bufferSize, false)

	if err != nil {
		return
	}

	defer r.Release(addr)
	_, buf := r.Reserve(bufferSize, 0)

	d := &DevicePath{}

	for i := 0; i <= maxDepth; i++ {
		if i == maxDepth {
			return nil, nil, errors.New("device path nodes limit exceeded")
		}

		node := &DevicePathNode{}

		if err = unmarshalBinary(buf[off:off+4], node); err != nil {
			return nil, nil, err
		}

		if node.Type == 0x7f && // End of Hardware Device Path
			node.SubType == 0xff { // End Entire Device Path
			break
		}

		if node.Length == 0 || node.Length > 0xff {
			return nil, nil, errors.New("invalid length")
		}

		off += 4

		d.Type = node.Type
		d.SubType = node.SubType
		d.Length = node.Length

		dataSize := uint(d.Length - 4)
		d.Data = make([]byte, dataSize)

		copy(d.Data, buf[off:off+dataSize])
		off += dataSize

		devicePath = append(devicePath, d)
		d = &DevicePath{}
	}

	desc = make([]byte, off)
	copy(desc, buf)

	return
}

// FilePath represents an EFI File Path Media Device Path instance.
type FilePath struct {
	DevicePathNode
	PathName []byte
}

// Bytes converts the descriptor structure to byte array format.
func (d *FilePath) Bytes() []byte {
	return append(d.DevicePathNode.Bytes(), d.PathName...)
}

// FilePath returns the full EFI Device Path associated with the named file.
func (root *FS) FilePath(name string) (devicePath []*DevicePath, filePath *FilePath, desc []byte, err error) {
	pathName := toUTF16(name)

	filePath = &FilePath{
		PathName: pathName,
	}

	filePath.Type = 0x04    // Media Device Path
	filePath.SubType = 0x04 // File Path
	filePath.Length = uint16(4 + len(pathName))

	if devicePath, desc, err = root.devicePath(); err != nil {
		return
	}

	devicePathEnd := &DevicePathNode{
		Type:    0x7f, // End of Hardware Device Path
		SubType: 0xff, // End Entire Device Path
		Length:  4,
	}

	desc = append(desc, filePath.Bytes()...)
	desc = append(desc, devicePathEnd.Bytes()...)

	return
}
