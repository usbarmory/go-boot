// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"bytes"
	"fmt"
	"regexp"
	"runtime"
	"strconv"

	"github.com/usbarmory/tamago/soc/intel/pci"

	"github.com/usbarmory/go-boot/shell"
	"github.com/usbarmory/go-boot/uefi/x64"
)

func init() {
	shell.Add(shell.Cmd{
		Name: "info",
		Help: "device information",
		Fn:   infoCmd,
	})

	shell.Add(shell.Cmd{
		Name:    "cpuid",
		Args:    2,
		Pattern: regexp.MustCompile(`^cpuid\s+([[:xdigit:]]+) ([[:xdigit:]]+)$`),
		Syntax:  "<leaf> <subleaf>",
		Help:    "show CPU capabilities",
		Fn:      cpuidCmd,
	})

	shell.Add(shell.Cmd{
		Name:    "lspci",
		Help:    "list PCI devices",
		Fn:      lspciCmd,
	})
}

func mem(start uint, size int, w []byte) (b []byte) {
	return memCopy(start, size, w)
}

func infoCmd(_ *shell.Interface, _ []string) (string, error) {
	var res bytes.Buffer

	ramStart, ramEnd := runtime.MemRegion()
	textStart, textEnd := runtime.TextRegion()
	_, heapStart := runtime.DataRegion()

	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)

	fmt.Fprintf(&res, "Runtime ......: %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(&res, "RAM ..........: %#08x-%#08x (%d MiB)\n", ramStart, ramEnd, (ramEnd-ramStart)/(1024*1024))
	fmt.Fprintf(&res, "Text .........: %#08x-%#08x\n", textStart, textEnd)
	fmt.Fprintf(&res, "Heap .........: %#08x-%#08x Alloc:%d MiB Sys:%d MiB\n", heapStart, ramEnd, m.HeapAlloc/(1024*1024), m.HeapSys/(1024*1024))
	fmt.Fprintf(&res, "CPU ..........: %s\n", x64.AMD64.Name())
	fmt.Fprintf(&res, "Frequency ....: %v GHz\n", float32(x64.AMD64.Freq())/1e9)

	return res.String(), nil
}

func cpuidCmd(_ *shell.Interface, arg []string) (string, error) {
	var res bytes.Buffer

	leaf, err := strconv.ParseUint(arg[0], 16, 32)

	if err != nil {
		return "", fmt.Errorf("invalid leaf, %v", err)
	}

	subleaf, err := strconv.ParseUint(arg[1], 10, 32)

	if err != nil {
		return "", fmt.Errorf("invalid subleaf, %v", err)
	}

	eax, ebx, ecx, edx := x64.AMD64.CPUID(uint32(leaf), uint32(subleaf))

	fmt.Fprintf(&res, "EAX      EBX      ECX      EDX\n")
	fmt.Fprintf(&res, "%08x %08x %08x %08x\n", eax, ebx, ecx, edx)

	return res.String(), nil
}

func lspciCmd(_ *shell.Interface, arg []string) (string, error) {
	var res bytes.Buffer

	fmt.Fprintf(&res, "Bus Vendor Device Bar0\n")

	for i := 0; i < 256; i++ {
		for _, d := range pci.Devices(i) {
			fmt.Fprintf(&res, "%03d %04x   %04x   %#016x\n", i, d.Vendor, d.Device, d.BaseAddress(0))
		}
	}

	return res.String(), nil
}

func date(epoch int64) {
	x64.AMD64.SetTimer(epoch)
}

func uptime() (ns int64) {
	return int64(float64(x64.AMD64.TimerFn()) * x64.AMD64.TimerMultiplier)
}
