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

func init() {
	log.SetFlags(0)

	logFile, _ := os.OpenFile("/go-boot.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
}

func main() {
	banner := fmt.Sprintf("%s/%s (%s) • UEFI",
		runtime.GOOS, runtime.GOARCH, runtime.Version())

	console := &cmd.Interface{
		CPU:    efi.AMD64,
		UART:   efi.UART0,
		Banner: banner,
	}

	cmd.StartTerminal(console)

	runtime.Exit(0)
}
