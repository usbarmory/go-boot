// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"bytes"
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
	var res bytes.Buffer

	t, err := efi.GetSystemTable()

	if err != nil {
		return "", err
	}

	fmt.Fprintf(term, "Firmware Revision .: %x\n", t.FirmwareRevision)
	fmt.Fprintf(term, "Runtime Services  .: %#x\n", t.RuntimeServices)
	fmt.Fprintf(term, "Boot Services .....: %#x\n", t.BootServices)
	fmt.Fprintf(term, "Table Entries .....: %d\n", t.NumberOfTableEntries)

	return res.String(), nil
}
