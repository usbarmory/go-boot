// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build amd64

package cmd

import (
	"bytes"
	"fmt"
	"runtime"

	"github.com/usbarmory/go-boot/efi"
	"github.com/usbarmory/go-boot/shell"
)

// This unikernel is reallocated based on build time variable IMAGE_BASE, for
// simplicity we do not use it to set runtime.ramStart and therefore we avoid
// using runtime.MemRegion() here.
func memRegion() (start uint64, end uint64) {
	textStart, _ := runtime.TextRegion()

	start = textStart - 0x10000
	end = start + efi.RamSize

	return
}

func mem(start uint, size int, w []byte) (b []byte) {
	return memCopy(start, size, w)
}

func infoCmd(_ *shell.Interface, _ []string) (string, error) {
	var res bytes.Buffer

	ramStart, ramEnd := memRegion()

	fmt.Fprintf(&res, "Runtime ......: %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(&res, "RAM ..........: %#08x-%#08x (%d MiB)\n", ramStart, ramEnd, (ramEnd-ramStart)/(1025*1024))
	fmt.Fprintf(&res, "CPU ..........: %s\n", efi.AMD64.Name())

	return res.String(), nil
}

func date(epoch int64) {
	efi.AMD64.SetTimer(epoch)
}

func uptime() (ns int64) {
	return int64(float64(efi.AMD64.TimerFn()) * efi.AMD64.TimerMultiplier)
}
