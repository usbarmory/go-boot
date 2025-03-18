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
	"github.com/usbarmory/go-boot/shell"
	"github.com/usbarmory/go-boot/uefi"
	"github.com/usbarmory/go-boot/uefi/x64"
)

// Build time variable
var Console string

func init() {
	fmt.Printf("initializing console (%s)\n", Console)

	log.SetFlags(0)

	logFile, _ := os.OpenFile(cmd.LogPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
}

func main() {
	banner := fmt.Sprintf("go-boot • %s/%s (%s) • UEFI x64",
		runtime.GOOS, runtime.GOARCH, runtime.Version())

	iface := &shell.Interface{
		Banner: banner,
	}

	// disable UEFI watchdog
	x64.UEFI.Boot.SetWatchdogTimer(0)

	switch Console {
	case "COM1", "com1", "":
		iface.ReadWriter = x64.UART0
		iface.Start(true)
	case "TEXT", "text":
		iface.ReadWriter = x64.UEFI.Console
		iface.Start(false)
	}

	log.Print("exit")

	if err := x64.UEFI.Boot.Exit(0); err != nil {
		log.Printf("halting due to exit error, %v", err)
		x64.UEFI.Runtime.ResetSystem(uefi.EfiResetShutdown)
	}
}
