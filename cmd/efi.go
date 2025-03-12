// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"unicode/utf16"

	"github.com/usbarmory/go-boot/efi"
	"github.com/usbarmory/go-boot/shell"
)

const maxVendorSize = 32

var (
	systemTable     *efi.SystemTable
	bootServices    *efi.BootServices
	runtimeServices *efi.RuntimeServices
)

func init() {
	efi.ForceLine = true
	efi.ReplaceTabs = 8

	print("initializing EFI services\n")

	systemTable = efi.UEFI.SystemTable
	bootServices = efi.UEFI.BootServices
	runtimeServices = efi.UEFI.RuntimeServices

	shell.Add(shell.Cmd{
		Name: "uefi",
		Help: "UEFI information",
		Fn:   uefiCmd,
	})

	shell.Add(shell.Cmd{
		Name:    "protocol",
		Args:    1,
		Pattern: regexp.MustCompile(`^protocol ([[:xdigit:]]{8}-[[:xdigit:]]{4}-[[:xdigit:]]{4}-[[:xdigit:]]{4}-[[:xdigit:]]{12})$`),
		Syntax:  "<registry format GUID>",
		Help:    "EFI_BOOT_SERVICES.LocateProtocol()",
		Fn:      locateCmd,
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
		Help:    "EFI_RUNTIME_SERVICES.ResetSystem()",
		Syntax:  "(cold|warm)?",
		Fn:      resetCmd,
	})

	shell.Add(shell.Cmd{
		Name:    "halt, shutdown",
		Args:    1,
		Pattern: regexp.MustCompile(`^(halt|shutdown)$`),
		Help:    "shutdown system",
		Fn:      shutdownCmd,
	})
}

func uefiCmd(_ *shell.Interface, _ []string) (res string, err error) {
	var buf bytes.Buffer
	var s []uint16

	b := mem(uint(systemTable.FirmwareVendor), maxVendorSize, nil)

	for i := 0; i < maxVendorSize; i += 2 {
		if b[i] == 0x00 && b[i+1] == 0 {
			break
		}

		s = append(s, binary.LittleEndian.Uint16(b[i:i+2]))
	}

	fmt.Fprintf(&buf, "Firmware Vendor ....: %s\n", string(utf16.Decode(s)))
	fmt.Fprintf(&buf, "Firmware Revision ..: %#x\n", systemTable.FirmwareRevision)
	fmt.Fprintf(&buf, "Runtime Services  ..: %#x\n", systemTable.RuntimeServices)
	fmt.Fprintf(&buf, "Boot Services ......: %#x\n", systemTable.BootServices)

	if s, err := screenInfo(); err == nil {
		fmt.Fprintf(&buf, "Frame Buffer .......: %dx%d @ %#x\n",
			s.LfbWidth, s.LfbHeight,
			uint64(s.ExtLfbBase)<<32|uint64(s.LfbBase))
	}

	fmt.Fprintf(&buf, "Configuration Tables: %#x\n", systemTable.ConfigurationTable)

	if c, err := systemTable.ConfigurationTables(); err == nil {
		for _, t := range c {
			fmt.Fprintf(&buf, "  %s (%#x)\n", t.RegistryFormat(), t.VendorTable)
		}
	}

	return buf.String(), err
}

func locateCmd(_ *shell.Interface, arg []string) (res string, err error) {
	if bootServices == nil {
		return "", errors.New("EFI Boot Services unavailable")
	}

	addr, err := bootServices.LocateProtocolString(arg[0])

	return fmt.Sprintf("%s: %#08x", arg[0], addr), err
}

func memmapCmd(_ *shell.Interface, _ []string) (res string, err error) {
	var buf bytes.Buffer
	var memoryMap *efi.MemoryMap

	if bootServices == nil {
		return "", errors.New("EFI Boot Services unavailable")
	}

	if memoryMap, err = bootServices.GetMemoryMap(); err != nil {
		return
	}

	fmt.Fprintf(&buf, "Type Start            End              Pages            Attributes\n")

	for _, desc := range memoryMap.Descriptors {
		fmt.Fprintf(&buf, "%02d   %016x %016x %016x %016x\n",
			desc.Type, desc.PhysicalStart, desc.PhysicalEnd()-1, desc.NumberOfPages, desc.Attribute)
	}

	return buf.String(), err
}

func allocCmd(_ *shell.Interface, arg []string) (res string, err error) {
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

	log.Printf("allocating memory range %#08x - %#08x", addr, addr+size)

	err = bootServices.AllocatePages(
		efi.AllocateAddress,
		efi.EfiLoaderData,
		int(size),
		addr,
	)

	return "", err
}

func resetCmd(_ *shell.Interface, arg []string) (_ string, err error) {
	var resetType int

	if runtimeServices == nil {
		return "", errors.New("EFI Runtime Services unavailable")
	}

	switch arg[0] {
	case "cold":
		resetType = efi.EfiResetCold
	case "warm", "":
		resetType = efi.EfiResetWarm
	case "shutdown":
		resetType = efi.EfiResetShutdown
	}

	log.Printf("performing system reset type %d", resetType)
	err = runtimeServices.ResetSystem(resetType)

	return
}

func shutdownCmd(_ *shell.Interface, _ []string) (_ string, err error) {
	return resetCmd(nil, []string{"shutdown"})
}
