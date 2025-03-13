// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package x64

import (
	"github.com/usbarmory/go-boot/uefi"
	_ "unsafe"
)

var earlyConsole = &uefi.Console{
	ForceLine: true,
	Out:       conOut,
}

//go:linkname printk runtime.printk
func printk(c byte) {
	earlyConsole.Output([]byte{c})

	if c == 0x0a && earlyConsole.ForceLine { // LF
		earlyConsole.Output([]byte{0x0d}) // CR
	}
}
