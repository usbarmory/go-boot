// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build amd64

package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"

	_ "github.com/usbarmory/go-boot/cmd"
	"github.com/usbarmory/go-boot/efi"
	"github.com/usbarmory/go-boot/shell"
)

// Decide whether to allocate output on console or serial port.
var useUART = false

func init() {
	print("go-boot initializing\n")

	log.SetFlags(0)

	logFile, _ := os.OpenFile("/go-boot.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
}

func main() {
	banner := fmt.Sprintf("%s/%s (%s) • UEFI",
		runtime.GOOS, runtime.GOARCH, runtime.Version())

	iface := &shell.Interface{
		Banner:   banner,
	}

	if useUART {
		iface.ReadWriter = efi.UART0
		iface.VT100 = true
		iface.Start()
	} else {
		iface.ReadWriter = efi.CONSOLE
		iface.Start()
	}

	runtime.Exit(0)
}
