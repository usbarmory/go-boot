// Copyright (c) WithSecure Corporation
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
	"runtime/debug"
	"runtime/pprof"
	"sort"
	"strings"
	"time"

	"github.com/hako/durafmt"

	"github.com/usbarmory/go-boot/shell"
	"github.com/usbarmory/tamago/dma"
)

const testDiversifier = "\xde\xad\xbe\xef"
const LogPath = "/go-boot.log"

func init() {
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
		Name:    "dma",
		Args:    1,
		Pattern: regexp.MustCompile(`^dma(?: (free|used))?$`),
		Help:    "show default DMA region allocation",
		Syntax:  "(free|used)?",
		Fn:      dmaCmd,
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

func dmaCmd(_ *shell.Interface, arg []string) (string, error) {
	var res []string

	if dma.Default() == nil {
		return "no default DMA region is present", nil
	}

	dump := func(blocks map[uint]uint, tag string) string {
		var r []string
		var t uint

		for addr, n := range blocks {
			t += n
			r = append(r, fmt.Sprintf("%#08x-%#08x %10d", addr, addr+n, n))
		}

		sort.Strings(r)
		r = append(r, fmt.Sprintf("%21s %10d bytes %s", "", t, tag))

		return strings.Join(r, "\n")
	}

	if arg[0] == "" || arg[0] == "free" {
		if blocks := dma.Default().FreeBlocks(); len(blocks) > 0 {
			res = append(res, dump(blocks, "free"))
		}
	}

	if arg[0] == "" || arg[0] == "used" {
		if blocks := dma.Default().UsedBlocks(); len(blocks) > 0 {
			res = append(res, dump(blocks, "used"))
		}
	}

	return strings.Join(res, "\n"), nil
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
	return fmt.Sprintf("%s", durafmt.Parse(time.Duration(ns)*time.Nanosecond)), nil
}
