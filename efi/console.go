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

// set in amd64.s
var (
	conIn         uintptr
	readKeyStroke uintptr

	conOut       uintptr
	outputString uintptr
)

// SimpleTextInput represents an EFI Simple Text Input Protocol descriptor.
type SimpleTextInput struct {
	Reset         uint64
	ReadKeyStroke uint64
	WaitForKey    uint64
}

// InputKey represents an EFI Input Key descriptor.
type InputKey struct {
	ScanCode    uint16
	UnicodeChar [2]byte
}

// SimpleTextOutput represents an EFI Simple Text Output Protocol descriptor.
type SimpleTextOutput struct {
	Reset             uint64
	OutputString      uint64
	TestString        uint64
	QueryMode         uint64
	SetMode           uint64
	SetAttribute      uint64
	ClearScreen       uint64
	SetCursorPosition uint64
	EnableCursor      uint64
	Mode              uint64
}

// Console implements the [io.ReadWriter] interface over EFI Simple Text
// Input/Output protocol.
type Console struct {
	io.ReadWriter
}

// defined in console.s
func consoleInput(k *InputKey) (status uint64)
func consoleOutput(c *byte) (status uint64)

//go:linkname printk runtime.printk
func printk(c byte) {
	consoleOutput(&c)

	// LF moves the cursor to the next line
	if c == 0x0a {
		// CR move cursor to left margin of the current line
		c = 0x0d
		consoleOutput(&c)
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
	b := utf16.Encode([]rune(string(p)))
	buf := make([]byte, 2)

	// We receive an UTF-8 string but we can output only UTF-16 ones.

	for _, r := range b {
		buf[1] = byte(r >> 8)
		buf[0] = byte(r & 0xff)

		if status := consoleOutput(&buf[0]); status != EFI_SUCCESS {
			return n, parseStatus(status)
		}
	}

	return
}
