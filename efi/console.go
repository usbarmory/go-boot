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

var (
	// ForceLine controls whether line feeds (LF) should be supplemented
	// with a carriage return (CR).
	ForceLine bool

	// ReplaceTabs controls whether Console I/O output should have Tab
	// characters replaced with a number of spaces.
	ReplaceTabs int = 0
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

	// TODO: check if bootServices are present just to be sure they haven't
	// been exited

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

	// TODO: implement BufferedStdoutLog or similar

	if c == 0x0a && ForceLine { // LF
		consoleOutput([]byte{0x0d}) // CR
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
		if r == 0x09 && ReplaceTabs > 0 { // Tab
			for i := 0; i < ReplaceTabs; i++ {
				s = append(s, []byte{0x20, 0x00}...) // Space
			}
			continue
		}

		s = append(s, byte(r&0xff))
		s = append(s, byte(r>>8))

		if r == 0x0a && ForceLine { // LF
			s = append(s, []byte{0x0d, 0x00}...) // CR
		}
	}

	if status := consoleOutput(s); status != EFI_SUCCESS {
		return n, parseStatus(status)
	}

	return
}
