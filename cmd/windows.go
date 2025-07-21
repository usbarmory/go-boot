// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"regexp"

	"github.com/usbarmory/go-boot/shell"
)

const WindowsBootManager = `\EFI\Microsoft\Boot\bootmgfw.efi`

func init() {
	shell.Add(shell.Cmd{
		Name:    "windows,win,w",
		Pattern: regexp.MustCompile(`^(?:windows|win|w)$`),
		Help:    "launch Windows UEFI boot manager",
		Fn:      winCmd,
	})
}

func winCmd(_ *shell.Interface, arg []string) (res string, err error) {
	return imageCmd(nil, []string{WindowsBootManager})
}
