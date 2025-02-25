// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"

	"github.com/usbarmory/go-boot/efi"
	"github.com/usbarmory/go-boot/shell"
)

var (
	systemTable     *efi.SystemTable
	bootServices    *efi.BootServices
	runtimeServices *efi.RuntimeServices
)

func init() {
	shell.Add(shell.Cmd{
		Name: "init",
		Help: "init UEFI services",
		Fn:   initCmd,
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

	shell.Add(shell.Cmd{
		Name:    "reset",
		Args:    1,
		Pattern: regexp.MustCompile(`reset(?: (cold|warm))?$`),
		Help:    "reset system",
		Syntax:  "(cold|warm)?",
		Fn:      resetCmd,
	})

	shell.Add(shell.Cmd{
		Name: "shutdown",
		Help: "shutdown system",
		Fn:   shutdownCmd,
	})
}

func initCmd(_ []string) (res string, err error) {
	var buf bytes.Buffer

	if systemTable, _ = efi.GetSystemTable(); systemTable != nil {
		bootServices, _ = systemTable.GetBootServices()
		runtimeServices, _ = systemTable.GetRuntimeServices()
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
		return "", errors.New("EFI Boot Services unavailable (forgot `init`?)")
	}

	if mmap, _, err = bootServices.GetMemoryMap(); err != nil {
		return
	}

	fmt.Fprintf(&buf, "Type Start            End              Pages            Attributes\n")

	for _, desc := range mmap {
		fmt.Fprintf(&buf, "%02d   %016x %016x %016x %016x\n",
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
		return "", errors.New("EFI Boot Services unavailable (forgot `init`?)")
	}

	log.Printf("allocating memory range %#08x - %#08x", addr, addr+size)

	err = bootServices.AllocatePages(
		efi.AllocateAddress,
		efi.EfiLoaderData,
		int(size),
		addr,
	)

	return "", err
}

func resetCmd(arg []string) (_ string, err error) {
	var resetType int

	if runtimeServices == nil {
		return "", errors.New("EFI Runtime Services unavailable (forgot `init`?)")
	}

	switch arg[0] {
	case "", "cold":
		resetType = efi.EfiResetCold
	case "warm":
		resetType = efi.EfiResetWarm
	case "shutdown":
		resetType = efi.EfiResetShutdown
	}

	log.Printf("performing system reset type %d", resetType)

	err = runtimeServices.ResetSystem(resetType)

	return
}

func shutdownCmd(_ []string) (_ string, err error) {
	return resetCmd([]string{"shutdown"})
}
