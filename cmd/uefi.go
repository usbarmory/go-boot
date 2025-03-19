// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io/fs"
	"log"
	"regexp"
	"strconv"
	"unicode/utf16"

	"github.com/usbarmory/go-boot/shell"
	"github.com/usbarmory/go-boot/uefi"
	"github.com/usbarmory/go-boot/uefi/x64"
)

const maxVendorSize = 32

func init() {
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
		Name:    "cat",
		Args:    1,
		Pattern: regexp.MustCompile(`^cat (.*)`),
		Syntax:  "<path>",
		Help:    "EFI_FILE_PROTOCOL.Read()",
		Fn:      catCmd,
	})

	shell.Add(shell.Cmd{
		Name:    "stat",
		Args:    1,
		Pattern: regexp.MustCompile(`^stat (.*)`),
		Syntax:  "<path>",
		Help:    "EFI_FILE_PROTOCOL.GetInfo()",
		Fn:      statCmd,
	})

	shell.Add(shell.Cmd{
		Name:    "memmap",
		Args:    1,
		Pattern: regexp.MustCompile(`^memmap( e820)?$`),
		Help:    "EFI_BOOT_SERVICES.GetMemoryMap()",
		Syntax:  "(e820)?",
		Fn:      memmapCmd,
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
		Pattern: regexp.MustCompile(`^reset(?: (cold|warm))?$`),
		Help:    "EFI_RUNTIME_SERVICES.ResetSystem()",
		Syntax:  "(cold|warm)?",
		Fn:      resetCmd,
	})

	shell.Add(shell.Cmd{
		Name:    "halt,shutdown",
		Args:    1,
		Pattern: regexp.MustCompile(`^(halt|shutdown)$`),
		Help:    "shutdown system",
		Fn:      shutdownCmd,
	})
}

func uefiCmd(_ *shell.Interface, _ []string) (res string, err error) {
	var buf bytes.Buffer
	var s []uint16

	t := x64.UEFI.SystemTable
	b := mem(uint(t.FirmwareVendor), maxVendorSize, nil)

	for i := 0; i < maxVendorSize; i += 2 {
		if b[i] == 0x00 && b[i+1] == 0 {
			break
		}

		s = append(s, binary.LittleEndian.Uint16(b[i:i+2]))
	}

	fmt.Fprintf(&buf, "Firmware Vendor ....: %s\n", string(utf16.Decode(s)))
	fmt.Fprintf(&buf, "Firmware Revision ..: %#x\n", t.FirmwareRevision)
	fmt.Fprintf(&buf, "Runtime Services  ..: %#x\n", t.RuntimeServices)
	fmt.Fprintf(&buf, "Boot Services ......: %#x\n", t.BootServices)

	if s, err := screenInfo(); err == nil {
		fmt.Fprintf(&buf, "Frame Buffer .......: %dx%d @ %#x\n",
			s.LfbWidth, s.LfbHeight,
			uint64(s.ExtLfbBase)<<32|uint64(s.LfbBase))
	}

	fmt.Fprintf(&buf, "Configuration Tables: %#x\n", t.ConfigurationTable)

	if c, err := t.ConfigurationTables(); err == nil {
		for _, t := range c {
			fmt.Fprintf(&buf, "  %s (%#x)\n", t.RegistryFormat(), t.VendorTable)
		}
	}

	return buf.String(), err
}

func locateCmd(_ *shell.Interface, arg []string) (res string, err error) {
	addr, err := x64.UEFI.Boot.LocateProtocol(uefi.GUID(arg[0]))
	return fmt.Sprintf("%s: %#08x", arg[0], addr), err
}

func catCmd(_ *shell.Interface, arg []string) (res string, err error) {
	root, err := x64.UEFI.Root()

	if err != nil {
		return "", fmt.Errorf("could not open root volume, %v", err)
	}

	f, err := fs.ReadFile(root, arg[0])

	if err != nil {
		return "", fmt.Errorf("could not read file, %v", err)
	}

	return string(f), nil
}

func statCmd(_ *shell.Interface, arg []string) (res string, err error) {
	root, err := x64.UEFI.Root()

	if err != nil {
		return "", fmt.Errorf("could not open root volume, %v", err)
	}

	f, err := root.Open(arg[0])

	if err != nil {
		return "", fmt.Errorf("could not open file, %v", err)
	}

	defer f.Close()

	stat, err := f.Stat()

	if err != nil {
		return
	}

	buf := make([]byte, stat.Size())

	if _, err = f.Read(buf); err != nil {
		return
	}

	return fmt.Sprintf("Size:%d ModTime:%s IsDir:%v Sys:%#x Sum256:%x",
		stat.Size(),
		stat.ModTime(),
		stat.IsDir(),
		stat.Sys(),
		sha256.Sum256(buf),
	), nil
}

func memmapCmd(_ *shell.Interface, arg []string) (res string, err error) {
	var buf bytes.Buffer
	var memoryMap *uefi.MemoryMap

	if memoryMap, err = x64.UEFI.Boot.GetMemoryMap(); err != nil {
		return
	}

	fmt.Fprintf(&buf, "Type Start            End              Pages            ")

	switch {
	case arg[0] == "":
		fmt.Fprintf(&buf, "Attributes\n")
		for _, desc := range memoryMap.Descriptors {
			fmt.Fprintf(&buf, "%02d   %016x %016x %016x %016x\n",
				desc.Type, desc.PhysicalStart, desc.PhysicalEnd()-1, desc.NumberOfPages, desc.Attribute)
		}
	case arg[0] == " e820":
		fmt.Fprintf(&buf, "\n")
		for _, desc := range memoryMap.E820() {
			fmt.Fprintf(&buf, "%02d   %016x %016x %016x\n",
				desc.MemType, desc.Addr, desc.Addr+desc.Size-1, desc.Size/4096)
		}
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

	log.Printf("allocating memory range %#08x - %#08x", addr, addr+size)

	err = x64.UEFI.Boot.AllocatePages(
		uefi.AllocateAddress,
		uefi.EfiLoaderData,
		int(size),
		addr,
	)

	return "", err
}

func resetCmd(_ *shell.Interface, arg []string) (_ string, err error) {
	var resetType int

	switch arg[0] {
	case "cold":
		resetType = uefi.EfiResetCold
	case "warm", "":
		resetType = uefi.EfiResetWarm
	case "shutdown":
		resetType = uefi.EfiResetShutdown
	}

	log.Printf("performing system reset type %d", resetType)
	err = x64.UEFI.Runtime.ResetSystem(resetType)

	return
}

func shutdownCmd(_ *shell.Interface, _ []string) (_ string, err error) {
	return resetCmd(nil, []string{"shutdown"})
}
