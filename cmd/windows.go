// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"regexp"

	"github.com/usbarmory/go-boot/shell"
)

const windowsPath = `\EFI\Microsoft\Boot\bootmgfw.efi`

func init() {
	shell.Add(shell.Cmd{
		Name:    "windows,win,w",
		Pattern: regexp.MustCompile(`^(?:windows|win|w)$`),
		Help:    "start Windows UEFI bootloader",
		Fn:      winCmd,
	})
}

func winCmd(_ *shell.Interface, arg []string) (res string, err error) {
	return startCmd(nil, []string{windowsPath})
}
