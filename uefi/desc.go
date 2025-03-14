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

const align = 8

func marshalBinary(data any) (buf []byte, err error) {
	b := new(bytes.Buffer)
	err = binary.Write(b, binary.LittleEndian, data)
	return b.Bytes(), nil
}

func unmarshalBinary(buf []byte, data any) (err error) {
	_, err = binary.Decode(buf, binary.LittleEndian, data)
	return
}

func decode(data any, addr uint64) (err error) {
	if addr == 0 {
		return errors.New("invalid address")
	}

	t, _ := marshalBinary(data)
	n := len(t) + (len(t) % align)

	r, err := dma.NewRegion(uint(addr), n, true)

	if err != nil {
		return
	}

	ptr, buf := r.Reserve(len(t), 0)
	defer r.Release(ptr)

	return unmarshalBinary(buf, data)
}
