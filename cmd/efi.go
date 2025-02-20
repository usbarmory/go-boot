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

	"github.com/usbarmory/go-boot/efi"
	"github.com/usbarmory/go-boot/shell"
)

var (
	systemTable  *efi.SystemTable
	bootServices *efi.BootServices
)

func init() {
	if systemTable, _ = efi.GetSystemTable(); systemTable != nil {
		bootServices, _ = systemTable.GetBootServices()
	}

	shell.Add(shell.Cmd{
		Name: "uefi",
		Help: "UEFI information",
		Fn:   uefiCmd,
	})

	shell.Add(shell.Cmd{
		Name: "memmap",
		Help: "EFI_BOOT_SERVICES.GetMemoryMap()",
		Fn:   memmapCmd,
	})

	shell.Add(shell.Cmd{
		Name:    "alloc",
		Args:    2,
		Pattern: regexp.MustCompile(`^alloc ([[:xdigit:]]+) (\d+)$`),
		Syntax:  "<hex offset> <size>",
		Help:    "EFI_BOOT_SERVICES.AllocatePages()",
		Fn:      allocCmd,
	})
}

func uefiCmd(_ []string) (res string, err error) {
	var buf bytes.Buffer

	if systemTable == nil {
		return "", errors.New("EFI System Table unavailable")
	}

	fmt.Fprintf(&buf, "Firmware Revision .: %x\n", systemTable.FirmwareRevision)
	fmt.Fprintf(&buf, "Runtime Services  .: %#x\n", systemTable.RuntimeServices)
	fmt.Fprintf(&buf, "Boot Services .....: %#x\n", systemTable.BootServices)
	fmt.Fprintf(&buf, "Table Entries .....: %d\n", systemTable.NumberOfTableEntries)

	return buf.String(), err
}

func memmapCmd(_ []string) (res string, err error) {
	var buf bytes.Buffer
	var mmap []*efi.MemoryMap

	if bootServices == nil {
		return "", errors.New("EFI Boot Services unavailable")
	}

	if mmap, _, err = bootServices.GetMemoryMap(); err != nil {
		return
	}

	fmt.Fprintf(&buf, "Type\tStart\t\t\tEnd\t\t\tPages\tAttributes\t\n")

	for _, desc := range mmap {
		fmt.Fprintf(&buf, "%02d\t%#016x\t%#016x\t%d\t%016x\n",
			desc.Type, desc.PhysicalStart, desc.PhysicalEnd()-1, desc.NumberOfPages, desc.Attribute)
	}

	return buf.String(), err
}

func allocCmd(arg []string) (res string, err error) {
	addr, err := strconv.ParseUint(arg[0], 16, 64)

	if err != nil {
		return "", fmt.Errorf("invalid address, %v", err)
	}

	size, err := strconv.ParseUint(arg[1], 10, 64)

	if err != nil {
		return "", fmt.Errorf("invalid size, %v", err)
	}

	if (addr%8) != 0 || (size%8) != 0 {
		return "", fmt.Errorf("only 64-bit aligned accesses are supported")
	}

	if bootServices == nil {
		return "", errors.New("EFI Boot Services unavailable")
	}

	err = bootServices.AllocatePages(
		efi.AllocateAddress,
		efi.EfiLoaderData,
		int(size),
		addr,
	)

	return "", err
}
