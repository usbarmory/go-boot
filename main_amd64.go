// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

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

	cmd.Banner = fmt.Sprintf("%s/%s (%s) • UEFI",
		runtime.GOOS, runtime.GOARCH, runtime.Version())
}

func main() {
	logFile, _ := os.OpenFile("/runtime.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))

	console := &cmd.Interface{
		CPU: efi.AMD64,
		UART: efi.UART0,
		Log: logFile,
	}

	cmd.StartTerminal(console)

	runtime.Exit(0)
}
