// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uefi

import (
	"errors"

	"github.com/usbarmory/tamago/dma"
)

// Configuration represents an EFI Configuration Table.
type ConfigurationTable struct {
	GUID        GUID
	VendorTable uint64
}

// ConfigurationTables returns the EFI Configuration Tables.
func (d *SystemTable) ConfigurationTables() (c []*ConfigurationTable, err error) {
	t := &ConfigurationTable{}

	if d.NumberOfTableEntries == 0 || d.ConfigurationTable == 0 {
		return nil, errors.New("EFI Configuration Table is invalid")
	}

	buf, _ := marshalBinary(t)
	entrySize := len(buf)
	tableSize := entrySize * int(d.NumberOfTableEntries)

	r, err := dma.NewRegion(uint(d.ConfigurationTable), tableSize, false)

	if err != nil {
		return
	}

	addr, buf := r.Reserve(tableSize, 0)
	defer r.Release(addr)

	for i := 0; i < tableSize; i += entrySize {
		if err = unmarshalBinary(buf[i:i+entrySize], t); err != nil {
			return
		}

		c = append(c, t)
		t = &ConfigurationTable{}
	}

	return
}

// LocateConfiguration locates an EFI Configuration Table.
func (d *SystemTable) LocateConfiguration(guid GUID) (t *ConfigurationTable, err error) {
	var c []*ConfigurationTable

	if c, err = d.ConfigurationTables(); err != nil {
		return
	}

	for _, t := range c {
		if t.GUID == guid {
			return t, nil
		}
	}

	return nil, errors.New("could not find configuration table")
}
