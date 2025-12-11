// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/usbarmory/go-boot/shell"
	"github.com/usbarmory/go-boot/transparency"
)

func init() {
	shell.Add(shell.Cmd{
		Name:    "bt",
		Args:    1,
		Pattern: regexp.MustCompile(`^(?:bt)( none| offline| online)?$`),
		Syntax:  "(none|offline|online)?",
		Help:    "show/set boot-transparency status",
		Fn:      btCmd,
	})
}

func btCmd(_ *shell.Interface, arg []string) (res string, err error) {
	if len(arg[0]) > 0 {
		transparency.Config.Status = strings.TrimSpace(arg[0])
	}

	switch transparency.Config.Status {
	case "none":
		transparency.CleanupConfig()

		return fmt.Sprintf("boot-transparency is disabled\n"), nil
	case "offline", "online":
		if err = transparency.LoadConfig(); err != nil {
			return "", err
		}

		return fmt.Sprintf("boot-transparency is enabled in %s mode\n", transparency.Config.Status), nil
	}

	return
}
