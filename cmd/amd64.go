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

	"golang.org/x/term"
)

func mem(start uint, size int, w []byte) (b []byte) {
	return memCopy(start, size, w)
}

func infoCmd(iface *Interface, _ *term.Terminal, _ []string) (string, error) {
	var res bytes.Buffer

	ramStart, ramEnd := runtime.MemRegion()

	fmt.Fprintf(&res, "Runtime ......: %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(&res, "RAM ..........: %#08x-%#08x (%d MiB)\n", ramStart, ramEnd, (ramEnd-ramStart)/(1024*1024))
	fmt.Fprintf(&res, "CPU ..........: %s\n", iface.CPU.Name())

	return res.String(), nil
}

func date(iface *Interface, epoch int64) {
	iface.CPU.SetTimer(epoch)
}

func uptime(iface *Interface) (ns int64) {
	return int64(float64(iface.CPU.TimerFn()) * iface.CPU.TimerMultiplier)
}

func rebootCmd(_ *Interface, _ *term.Terminal, _ []string) (_ string, err error) {
	runtime.Exit(0)
	return
}
