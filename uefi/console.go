// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uefi

import (
	"bytes"
	"encoding/binary"
	"io"
	"time"
	"unicode/utf16"
)

// EFI ConOut offsets
const (
	outputString = 0x08
	setMode      = 0x20
	setAttribute = 0x28
	clearScreen  = 0x30
)

// EFI ConIn offsets
const (
	readKeyStroke = 0x08
)

// EFI text attributes
const (
	EFI_BLACK        = 0x00
	EFI_BLUE         = 0x01
	EFI_GREEN        = 0x02
	EFI_CYAN         = 0x03
	EFI_RED          = 0x04
	EFI_MAGENTA      = 0x05
	EFI_BROWN        = 0x06
	EFI_LIGHTGRAY    = 0x07
	EFI_BRIGHT       = 0x08
	EFI_DARKGRAY     = 0x08
	EFI_LIGHTBLUE    = 0x09
	EFI_LIGHTGREEN   = 0x0a
	EFI_LIGHTCYAN    = 0x0b
	EFI_LIGHTRED     = 0x0c
	EFI_LIGHTMAGENTA = 0x0d
	EFI_YELLOW       = 0x0e
	EFI_WHITE        = 0x0f
)

// ASCII control characters
const (
	null  = 0x00
	bs    = 0x08
	tab   = 0x09
	lf    = 0x0a
	cr    = 0x0d
	space = 0x20
)

// Control Sequence Introducer n D - CUB - Cursor Back
var cub = []byte{0x1b, 0x5b, 0x44, 0x20, 0x1b, 0x5b, 0x44}

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
func (c *Console) ClearScreen() error {
	if c.Out == 0 {
		return nil
	}

	status := callService(c.Out+clearScreen,
		[]uint64{
			c.Out,
		},
	)

	return parseStatus(status)
}

// SetMode calls EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL.SetMode().
func (c *Console) SetMode(mode uint64) error {
	if c.Out == 0 {
		return nil
	}

	status := callService(c.Out+setMode,
		[]uint64{
			c.Out,
			mode,
		},
	)

	return parseStatus(status)
}

// SetAttribute calls EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL.SetAttribute().
func (c *Console) SetAttribute(attr uint64) error {
	if c.Out == 0 {
		return nil
	}

	status := callService(c.Out+setAttribute,
		[]uint64{
			c.Out,
			attr,
		},
	)

	return parseStatus(status)
}

// Input calls EFI_SIMPLE_TEXT_INPUT_PROTOCOL.ReadKeyStroke().
func (c *Console) Input(k *InputKey) (status uint64) {
	if c.In == 0 {
		return
	}

	return callService(c.In+readKeyStroke,
		[]uint64{
			c.In,
			ptrval(k),
		},
	)
}

// Output calls EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL.OutputString().
func (c *Console) Output(p []byte) (status uint64) {
	if p[len(p)-1] != null {
		p = append(p, null)
	}

	if c.Out == 0 {
		return
	}

	return callService(c.Out+outputString,
		[]uint64{
			c.Out,
			ptrval(&p[0]),
		},
	)
}

// Read available data to buffer from console.
func (c *Console) Read(p []byte) (n int, err error) {
	k := &InputKey{}

	for n = 0; n < len(p); n += 2 {
		status := c.Input(k)

		switch {
		case status&0xff == EFI_NOT_READY:
			// Compatibility note:
			//
			// shell.(*Interface).readLine now starves the
			// scheduler, however this package currently has no
			// need for background goroutines.
			//
			// In case this becomes undesirable here add:
			//  runtime.Gosched()
			//
			// For now we just take an atomic nap as that eases a
			// benign HeapAlloc increase due to GC starvation.
			time.Sleep(1 * time.Millisecond)
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

	if len(p) == len(cub) && bytes.Equal(cub, p) {
		p = []byte{bs, 0x00}
	}

	// we receive an UTF-8 string and can output UTF-16
	b := utf16.Encode([]rune(string(p)))

	for _, r := range b {
		if r == tab && c.ReplaceTabs > 0 {
			for i := 0; i < c.ReplaceTabs; i++ {
				s = append(s, []byte{space, 0x00}...)
			}
			continue
		}

		s = append(s, byte(r&0xff))
		s = append(s, byte(r>>8))

		if r == lf && c.ForceLine {
			s = append(s, []byte{cr, 0x00}...)
		}
	}

	if status := c.Output(s); status != EFI_SUCCESS {
		return n, parseStatus(status)
	}

	return
}
