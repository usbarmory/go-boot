// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package x64

import (
	"github.com/usbarmory/go-boot/uefi"
	_ "unsafe"
)

// Console represents the early UEFI services console for pre UEFI.Init()
// standard output.
var Console = &uefi.Console{
	ForceLine: true,
	In:        conIn,
	Out:       conOut,
}

// Stdout allows to override [Console] as standard output, useful when exiting
// UEFI runtime services.
var Stdout func(byte)

//go:linkname printk runtime/goos.Printk
func printk(c byte) {
	if Stdout != nil {
		Stdout(c)
	}

	if Console.Out == 0 {
		return
	}

	Console.Output([]byte{c})

	if c == 0x0a && Console.ForceLine { // LF
		Console.Output([]byte{0x0d}) // CR
	}
}
