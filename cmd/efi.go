// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	"golang.org/x/term"

	"github.com/usbarmory/go-boot/efi"
)

func init() {
	Add(Cmd{
		Name:    "uefi",
		Help:    "UEFI information",
		Fn:      uefiCmd,
	})
}

func uefiCmd(_ *Interface, term *term.Terminal, _ []string) (string, error) {
	fmt.Fprintf(term, "System table: %x\n", efi.SystemTable)
	return "", nil
}
