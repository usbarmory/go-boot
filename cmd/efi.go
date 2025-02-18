// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"golang.org/x/term"

	"github.com/usbarmory/go-boot/efi"
)

var systemTable *efi.SystemTable

func init() {
	Add(Cmd{
		Name: "uefi",
		Help: "UEFI information",
		Fn:   uefiCmd,
	})

	Add(Cmd{
		Name:    "alloc",
		Args:    2,
		Pattern: regexp.MustCompile(`^alloc ([[:xdigit:]]+) (\d+)$`),
		Syntax:  "<hex offset> <size>",
		Help:    "allocate pages via UEFI boot services",
		Fn:      allocCmd,
	})

}

func uefiCmd(_ *Interface, term *term.Terminal, _ []string) (res string, err error) {
	var buf bytes.Buffer

	if systemTable, err = efi.GetSystemTable(); err != nil {
		return
	}

	fmt.Fprintf(&buf, "Firmware Revision .: %x\n", systemTable.FirmwareRevision)
	fmt.Fprintf(&buf, "Runtime Services  .: %#x\n", systemTable.RuntimeServices)
	fmt.Fprintf(&buf, "Boot Services .....: %#x\n", systemTable.BootServices)
	fmt.Fprintf(&buf, "Table Entries .....: %d\n", systemTable.NumberOfTableEntries)

	return buf.String(), err
}

func allocCmd(_ *Interface, _ *term.Terminal, arg []string) (res string, err error) {
	addr, err := strconv.ParseUint(arg[0], 16, 32)

	if err != nil {
		return "", fmt.Errorf("invalid address, %v", err)
	}

	size, err := strconv.ParseUint(arg[1], 10, 32)

	if err != nil {
		return "", fmt.Errorf("invalid size, %v", err)
	}

	if (addr%8) != 0 || (size%8) != 0 {
		return "", fmt.Errorf("only 64-bit aligned accesses are supported")
	}

	if systemTable == nil {
		return "", errors.New("run `uefi` first")
	}

	b, err := systemTable.GetBootServices()

	if err != nil {
		return "", err
	}

	err = b.AllocatePages(
		efi.AllocateAddress,
		efi.EfiBootServicesCode,
		int(size),
		addr,
	)

	return "", err
}
