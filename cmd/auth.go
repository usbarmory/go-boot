// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"net"
	"regexp"
	"strings"

	"github.com/usbarmory/go-boot/shell"
	"github.com/usbarmory/go-boot/transparency"
	"github.com/usbarmory/go-boot/uapi"

	"github.com/usbarmory/boot-transparency/artifact"
	_ "github.com/usbarmory/boot-transparency/engine/sigsum"
	_ "github.com/usbarmory/boot-transparency/engine/tessera"
	bt_transparency "github.com/usbarmory/boot-transparency/transparency"
)

var btConfig transparency.Config

func init() {
	shell.Add(shell.Cmd{
		Name:    "bt",
		Args:    2,
		Pattern: regexp.MustCompile(`^(?:bt)( none| offline| online)?( sigsum| tessera)?$`),
		Syntax:  "(none|offline|online)? (sigsum|tessera)?",
		Help:    "show/change boot-transparency configuration",
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

		if btConfig.Status != transparency.None && len(arg[1]) == 0 {
			return "", errors.New("invalid transparency engine")
		}

		if len(arg[1]) > 0 {
			e := bt_transparency.EngineCodeFromString[strings.TrimSpace(arg[1])]
			if _, err = bt_transparency.GetEngine(e); err != nil {
				return "", fmt.Errorf("unable to configure the transparency engine, %v", err)
			}
			btConfig.Engine = e
		}
	}

	switch btConfig.Status {
	case transparency.None:
		res = fmt.Sprintf("boot-transparency is disabled\n")
	case transparency.Offline, transparency.Online:
		res = fmt.Sprintf("boot-transparency is enabled in %s mode, %s engine selected\n", btConfig.Status, btConfig.Engine)
	}

	return
}

func btValidateLinux(entry *uapi.Entry, root fs.FS) (err error) {
	if entry == nil || len(entry.Linux) == 0 {
		return errors.New("invalid kernel entry")
	}

	btConfig.UefiRoot = root

	btEntry := transparency.BootEntry{
		transparency.Artifact{
			Category: artifact.LinuxKernel,
			Hash:     artifact.Sum(entry.Linux),
		},
		transparency.Artifact{
			Category: artifact.Initrd,
			Hash:     artifact.Sum(entry.Initrd),
		},
	}

	return btEntry.Validate(&btConfig)
}
