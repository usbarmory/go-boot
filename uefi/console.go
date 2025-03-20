// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uefi

import (
	"encoding/binary"
	"io"
	"unicode/utf16"
)

// EFI ConOut offsets
const (
	outputString = 0x08
	clearScreen  = 0x30
)

// EFI ConIn offsets
const (
	readKeyStroke = 0x08
)

// InputKey represents an EFI Input Key descriptor.
type InputKey struct {
	ScanCode    uint16
	UnicodeChar [2]byte
}

// Console implements the [io.ReadWriter] interface over EFI Simple Text
// Input/Output protocol.
type Console struct {
	io.ReadWriter

	// ForceLine controls whether line feeds (LF) should be supplemented
	// with a carriage return (CR).
	ForceLine bool

	// ReplaceTabs controls whether Console I/O output should have Tab
	// characters replaced with a number of spaces.
	ReplaceTabs int

	// In should be set to the EFI SystemTable ConIn address.
	In uint64
	// Out should be set to the EFI SystemTable ConOut address.
	Out uint64
}

// ClearScreen calls EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL.ClearScreen().
func (c *Console) ClearScreen() (status uint64) {
	if c.Out == 0 {
		return
	}

	return callService(
		c.Out+clearScreen,
		c.Out,
		0,
		0,
		0,
	)
}

// Input calls EFI_SIMPLE_TEXT_INPUT_PROTOCOL.ReadKeyStroke().
func (c *Console) Input(k *InputKey) (status uint64) {
	if c.In == 0 {
		return
	}

	return callService(
		c.In+readKeyStroke,
		c.In,
		ptrval(k),
		0,
		0,
	)
}

// Output calls EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL.OutputString().
func (c *Console) Output(p []byte) (status uint64) {
	if p[len(p)-1] != 0x00 {
		p = append(p, 0x00)
	}

	if c.Out == 0 {
		return
	}

	return callService(
		c.Out+outputString,
		c.Out,
		ptrval(&p[0]),
		0,
		0,
	)
}

// Read available data to buffer from console.
func (c *Console) Read(p []byte) (n int, err error) {
	k := &InputKey{}

	for n = 0; n < len(p); n += 2 {
		status := c.Input(k)

		switch {
		case status&0xff == EFI_NOT_READY:
			return
		case status != EFI_SUCCESS:
			return n, parseStatus(status)
		case k.ScanCode > 0:
			binary.LittleEndian.PutUint16(p[n:], k.ScanCode)
		default:
			copy(p[n:], k.UnicodeChar[:])
		}
	}

	return
}

// Write data from buffer to console.
func (c *Console) Write(p []byte) (n int, err error) {
	var s []byte

	if len(p) == 0 {
		return
	}

	b := utf16.Encode([]rune(string(p)))

	// We receive an UTF-8 string but we can output only UTF-16 ones.

	for _, r := range b {
		if r == 0x09 && c.ReplaceTabs > 0 { // Tab
			for i := 0; i < c.ReplaceTabs; i++ {
				s = append(s, []byte{0x20, 0x00}...) // Space
			}
			continue
		}

		s = append(s, byte(r&0xff))
		s = append(s, byte(r>>8))

		if r == 0x0a && c.ForceLine { // LF
			s = append(s, []byte{0x0d, 0x00}...) // CR
		}
	}

	if status := c.Output(s); status != EFI_SUCCESS {
		return n, parseStatus(status)
	}

	return
}
