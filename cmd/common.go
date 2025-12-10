// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"time"

	"github.com/hako/durafmt"

	"github.com/usbarmory/go-boot/shell"
)

const testDiversifier = "\xde\xad\xbe\xef"
const LogPath = "/go-boot.log"

var Banner string

func init() {
	Banner = fmt.Sprintf("go-boot • %s/%s (%s) • UEFI x64",
		runtime.GOOS, runtime.GOARCH, runtime.Version())

	shell.Add(shell.Cmd{
		Name: "build",
		Help: "build information",
		Fn:   buildInfoCmd,
	})

	shell.Add(shell.Cmd{
		Name: "log",
		Help: "show runtime logs",
		Fn:   logCmd,
	})

	shell.Add(shell.Cmd{
		Name:    "exit,quit",
		Args:    1,
		Pattern: regexp.MustCompile(`^(exit|quit)$`),
		Help:    "exit application",
		Fn:      exitCmd,
	})

	shell.Add(shell.Cmd{
		Name: "stack",
		Help: "goroutine stack trace (current)",
		Fn:   stackCmd,
	})

	shell.Add(shell.Cmd{
		Name: "stackall",
		Help: "goroutine stack trace (all)",
		Fn:   stackallCmd,
	})

	shell.Add(shell.Cmd{
		Name:    "date",
		Args:    1,
		Pattern: regexp.MustCompile(`^date(.*)`),
		Syntax:  "(time in RFC339 format)?",
		Help:    "show/change runtime date and time",
		Fn:      dateCmd,
	})

	shell.Add(shell.Cmd{
		Name: "uptime",
		Help: "show system running time",
		Fn:   uptimeCmd,
	})
}

func buildInfoCmd(_ *shell.Interface, _ []string) (string, error) {
	res := new(bytes.Buffer)

	if bi, ok := debug.ReadBuildInfo(); ok {
		res.WriteString(bi.String())
	}

	return res.String(), nil
}

func logCmd(_ *shell.Interface, _ []string) (string, error) {
	res, err := os.ReadFile(LogPath)
	return string(res), err
}

func exitCmd(_ *shell.Interface, _ []string) (res string, err error) {
	return "", io.EOF
}

func stackCmd(_ *shell.Interface, _ []string) (string, error) {
	return string(debug.Stack()), nil
}

func stackallCmd(_ *shell.Interface, _ []string) (string, error) {
	buf := new(bytes.Buffer)
	pprof.Lookup("goroutine").WriteTo(buf, 1)

	return buf.String(), nil
}

func dateCmd(_ *shell.Interface, arg []string) (res string, err error) {
	if len(arg[0]) > 1 {
		t, err := time.Parse(time.RFC3339, arg[0][1:])

		if err != nil {
			return "", err
		}

		date(t.UnixNano())
	}

	return fmt.Sprintf("%s", time.Now().Format(time.RFC3339)), nil
}

func uptimeCmd(_ *shell.Interface, _ []string) (string, error) {
	ns := uptime()
	return fmt.Sprintf("%s\n", durafmt.Parse(time.Duration(ns)*time.Nanosecond)), nil
}
