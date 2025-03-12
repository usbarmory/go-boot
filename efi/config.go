// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package efi

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/usbarmory/tamago/dma"
)

// Configuration represents an EFI Configuration Table.
type ConfigurationTable struct {
	GUID        [16]byte
	VendorTable uint64
}

// MarshalBinary implements the [encoding.BinaryMarshaler] interface.
func (d *ConfigurationTable) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.LittleEndian, d)
	return buf.Bytes(), nil
}

// UnmarshalBinary implements the [encoding.BinaryUnmarshaler] interface.
func (d *ConfigurationTable) UnmarshalBinary(data []byte) (err error) {
	_, err = binary.Decode(data, binary.LittleEndian, d)
	return
}

// RegistryFormat returns the table EF GUID in registry format.
func (d *ConfigurationTable) RegistryFormat() string {
	// https://uefi.org/specs/UEFI/2.10/Apx_A_GUID_and_Time_Formats.html
	return fmt.Sprintf("%08x-%04x-%04x-%x-%x",
		binary.LittleEndian.Uint32(d.GUID[0:4]),
		binary.LittleEndian.Uint16(d.GUID[4:6]),
		binary.LittleEndian.Uint16(d.GUID[6:8]),
		d.GUID[8:10],
		d.GUID[10:])
}

// ConfigurationTables returns the EFI Configuration Tables.
func (d *SystemTable) ConfigurationTables() (c []*ConfigurationTable, err error) {
	t := &ConfigurationTable{}

	if d.NumberOfTableEntries == 0 || d.ConfigurationTable == 0 {
		return nil, errors.New("EFI Configuration Table is invalid")
	}

	buf, _ := t.MarshalBinary()
	entrySize := len(buf)
	tableSize := entrySize * int(d.NumberOfTableEntries)

	r, err := dma.NewRegion(uint(d.ConfigurationTable), tableSize, false)

	if err != nil {
		return
	}

	addr, buf := r.Reserve(tableSize, 0)
	defer r.Release(addr)

	for i := 0; i < tableSize; i += entrySize {
		if err = t.UnmarshalBinary(buf[i : i+entrySize]); err != nil {
			return
		}

		c = append(c, t)
		t = &ConfigurationTable{}
	}

	return
}
