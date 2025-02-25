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
)

func mem(start uint, size int, w []byte) (b []byte) {
	return memCopy(start, size, w)
}

func infoCmd(_ []string) (string, error) {
	var res bytes.Buffer

	ramStart, ramEnd := runtime.MemRegion()

	fmt.Fprintf(&res, "Runtime ......: %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(&res, "RAM ..........: %#08x-%#08x (%d MiB)\n", ramStart, ramEnd, (ramEnd-ramStart)/(1024*1024))
	fmt.Fprintf(&res, "CPU ..........: %s\n", efi.AMD64.Name())

	return res.String(), nil
}

func date(epoch int64) {
	efi.AMD64.SetTimer(epoch)
}

func uptime() (ns int64) {
	return int64(float64(efi.AMD64.TimerFn()) * efi.AMD64.TimerMultiplier)
}
