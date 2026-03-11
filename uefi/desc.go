// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uefi

import (
	"bytes"
	"encoding/binary"
	"errors"
	"unicode/utf16"

	"github.com/usbarmory/tamago/dma"
)

const align = 8

func toUTF16(s string) (buf []byte) {
	for _, r := range utf16.Encode([]rune(s)) {
		buf = append(buf, byte(r&0xff))
		buf = append(buf, byte(r>>8))
	}

	return append([]byte(buf), []byte{0x00, 0x00}...)
}

func fromUTF16(buf []byte) string {
	if len(buf)%2 != 0 {
		// invalid UTF-16 byte slice
		return ""
	}

	codes := make([]uint16, len(buf)/2)

	for i := 0; i < len(codes); i++ {
		codes[i] = uint16(buf[i*2]) | (uint16(buf[i*2+1]) << 8)
	}

	for i, code := range codes {
		if code == null {
			codes = codes[:i]
			break
		}
	}

	return string(utf16.Decode(codes))
}

func marshalBinary(data any) (buf []byte, err error) {
	b := new(bytes.Buffer)
	err = binary.Write(b, binary.LittleEndian, data)
	return b.Bytes(), err
}

func unmarshalBinary(buf []byte, data any) (err error) {
	_, err = binary.Decode(buf, binary.LittleEndian, data)
	return
}

func encode(data any, addr uint64) (err error) {
	if addr == 0 {
		return errors.New("invalid address")
	}

	buf, _ := marshalBinary(data)
	n := len(buf)

	r, err := dma.NewRegion(uint(addr), n, false)

	if err != nil {
		return
	}

	ptr := r.Alloc(buf, 0)
	r.Free(ptr)

	return
}

func decode(data any, addr uint64) (err error) {
	if addr == 0 {
		return errors.New("invalid address")
	}

	t, _ := marshalBinary(data)
	n := len(t) + (len(t) % align)

	r, err := dma.NewRegion(uint(addr), n, false)

	if err != nil {
		return
	}

	ptr, buf := r.Reserve(len(t), 0)
	defer r.Release(ptr)

	return unmarshalBinary(buf, data)
}
