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
	"github.com/usbarmory/go-boot/shell"
)

// Build time variable
var Console string

func init() {
	fmt.Printf("go-boot initializing (console=%s)\n", Console)

	log.SetFlags(0)

	logFile, _ := os.OpenFile(cmd.LogPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
}

func main() {
	banner := fmt.Sprintf("%s/%s (%s) • UEFI",
		runtime.GOOS, runtime.GOARCH, runtime.Version())

	iface := &shell.Interface{
		Banner:   banner,
	}

	switch Console {
	case "COM1", "com1", "":
		iface.ReadWriter = efi.UART0
		iface.Start(true)
	case "TEXT", "text":
		efi.ForceLine = true
		efi.ReplaceTabs = 8

		iface.ReadWriter = efi.CONSOLE
		iface.Start(false)
	}

	log.Print("halting")

	runtime.Exit(0)
}
