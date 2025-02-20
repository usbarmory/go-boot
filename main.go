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

	"github.com/usbarmory/go-boot/cmd"
	"github.com/usbarmory/go-boot/efi"
)

// Decide whether to allocate output on console or serial port.
var useUART = true

func init() {
	print("go-boot initializing\n")

	log.SetFlags(0)

	logFile, _ := os.OpenFile("/go-boot.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
}

func main() {
	banner := fmt.Sprintf("%s/%s (%s) • UEFI",
		runtime.GOOS, runtime.GOARCH, runtime.Version())

	iface := &cmd.Interface{
		CPU:      efi.AMD64,
		Banner:   banner,
	}

	if useUART {
		iface.Terminal = efi.UART0
		cmd.StartTerminal(iface)
	} else {
		iface.Terminal = efi.CONSOLE
		cmd.StartConsole(iface)
	}

	runtime.Exit(0)
}
