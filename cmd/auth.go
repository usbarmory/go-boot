// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/usbarmory/go-boot/shell"
	"github.com/usbarmory/go-boot/transparency"
)

var btConfig transparency.Config

func init() {
	shell.Add(shell.Cmd{
		Name:    "bt",
		Args:    1,
		Pattern: regexp.MustCompile(`^(?:bt)( none| offline| online)?$`),
		Syntax:  "(none|offline|online)?",
		Help:    "show/change boot-transparency status",
		Fn:      btCmd,
	})
}

func btCmd(_ *shell.Interface, arg []string) (res string, err error) {
	if len(arg[0]) > 0 {
		switch strings.TrimSpace(arg[0]) {
		case "none":
			btConfig.Status = transparency.None
		case "offline":
			btConfig.Status = transparency.Offline
		case "online":
			if net.SocketFunc == nil {
				return "", errors.New("network unavailable")
			}
			btConfig.Status = transparency.Online
		}
	}

	switch btConfig.Status {
	case transparency.None:
		res = fmt.Sprintf("boot-transparency is disabled\n")
	case transparency.Offline, transparency.Online:
		res = fmt.Sprintf("boot-transparency is enabled in %s mode\n", btConfig.Status.Resolve())
	}

	return
}
