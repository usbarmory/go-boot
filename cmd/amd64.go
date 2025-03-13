// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build amd64

package cmd

import (
	"bytes"
	"fmt"
	"regexp"
	"runtime"
	"strconv"

	"github.com/usbarmory/go-boot/efi"
	"github.com/usbarmory/go-boot/shell"
)

func init() {
	shell.Add(shell.Cmd{
		Name:    "cpuid",
		Args:    2,
		Pattern: regexp.MustCompile(`^cpuid\s+([[:xdigit:]]+) ([[:xdigit:]]+)$`),
		Syntax:  "<leaf> <subleaf>",
		Help:    "display CPU capabilities",
		Fn:      cpuidCmd,
	})
}

func mem(start uint, size int, w []byte) (b []byte) {
	return memCopy(start, size, w)
}

func infoCmd(_ *shell.Interface, _ []string) (string, error) {
	var res bytes.Buffer

	ramStart, ramEnd := runtime.MemRegion()

	fmt.Fprintf(&res, "Runtime ......: %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(&res, "RAM ..........: %#08x-%#08x (%d MiB)\n", ramStart, ramEnd, (ramEnd-ramStart)/(1025*1024))
	fmt.Fprintf(&res, "CPU ..........: %s\n", efi.AMD64.Name())
	fmt.Fprintf(&res, "Frequency ....: %v GHz\n", float32(efi.AMD64.Freq())/1e9)

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

	eax, ebx, ecx, edx := efi.AMD64.CPUID(uint32(leaf), uint32(subleaf))

	fmt.Fprintf(&res, "EAX      EBX      ECX      EDX\n")
	fmt.Fprintf(&res, "%08x %08x %08x %08x\n", eax, ebx, ecx, edx)

	return res.String(), nil

}

func date(epoch int64) {
	efi.AMD64.SetTimer(epoch)
}

func uptime() (ns int64) {
	return int64(float64(efi.AMD64.TimerFn()) * efi.AMD64.TimerMultiplier)
}
