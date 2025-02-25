// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package efi

import (
	"io"
	"unicode/utf16"
	_ "unsafe"
)

const (
	// EFI ConOut offset for OutputString
	outputString = 0x08
	// EFI ConIn offset for ReadKeyStroke
	readKeyStroke = 0x08
)

// set in amd64.s
var (
	conIn  uint64
	conOut uint64
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
}

func consoleInput(k *InputKey) (status uint64) {
	return callService(
		conIn+readKeyStroke,
		conIn,
		ptrval(k),
		0,
		0,
	)
}

func consoleOutput(p []byte) (status uint64) {
	if p[len(p)-1] != 0x00 {
		p = append(p, 0x00)
	}

	return callService(
		conOut+outputString,
		conOut,
		ptrval(&p[0]),
		0,
		0,
	)
}

//go:linkname printk runtime.printk
func printk(c byte) {
	consoleOutput([]byte{c})

	// LF moves cursor to the next line
	if c == 0x0a {
		// CR moves cursor to left marging of the current line
		consoleOutput([]byte{0x0d})
	}
}

// Read available data to buffer from console.
func (c *Console) Read(p []byte) (n int, err error) {
	k := &InputKey{}

	for n = 0; n < len(p); n += 2 {
		status := consoleInput(k)

		switch {
		case status == EFI_SUCCESS:
			copy(p[n:], k.UnicodeChar[:])
		case status&0xff == EFI_NOT_READY:
			return
		default:
			return n, parseStatus(status)
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
		s = append(s, byte(r&0xff))
		s = append(s, byte(r>>8))
	}

	if status := consoleOutput(s); status != EFI_SUCCESS {
		return n, parseStatus(status)
	}

	return
}
