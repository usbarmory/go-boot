// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package efi

import (
	"encoding"
	"errors"

	"github.com/usbarmory/tamago/dma"
)

const align = 8

type BinaryMarshal interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

// TODO: expand this to remaining dma.NewRegion uses
func decode(data BinaryMarshal, addr uint64) (err error) {
	if addr == 0 {
		return errors.New("invalid address")
	}

	t, _ := data.MarshalBinary()
	n := len(t) + (len(t) % align)

	r, err := dma.NewRegion(uint(addr), n, false)

	if err != nil {
		return
	}

	ptr, buf := r.Reserve(len(t), 0)
	defer r.Release(ptr)

	return data.UnmarshalBinary(buf)
}
