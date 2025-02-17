// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"sort"
	"strings"
	"time"

	"golang.org/x/term"

	"github.com/usbarmory/tamago/dma"

	"github.com/hako/durafmt"
)

const testDiversifier = "\xde\xad\xbe\xef"

func init() {
	Add(Cmd{
		Name: "help",
		Help: "this help",
		Fn:   helpCmd,
	})

	Add(Cmd{
		Name: "build",
		Help: "build information",
		Fn:   buildInfoCmd,
	})

	Add(Cmd{
		Name:    "exit, quit",
		Args:    1,
		Pattern: regexp.MustCompile(`^(exit|quit)$`),
		Help:    "close session",
		Fn:      exitCmd,
	})

	Add(Cmd{
		Name: "halt",
		Help: "halt the machine",
		Fn:   haltCmd,
	})

	Add(Cmd{
		Name: "stack",
		Help: "goroutine stack trace (current)",
		Fn:   stackCmd,
	})

	Add(Cmd{
		Name: "stackall",
		Help: "goroutine stack trace (all)",
		Fn:   stackallCmd,
	})

	Add(Cmd{
		Name:    "dma",
		Args:    1,
		Pattern: regexp.MustCompile(`^dma(?:(?: )(free|used))?$`),
		Help:    "show allocation of default DMA region",
		Syntax:  "(free|used)?",
		Fn:      dmaCmd,
	})

	Add(Cmd{
		Name:    "date",
		Args:    1,
		Pattern: regexp.MustCompile(`^date(.*)`),
		Syntax:  "(time in RFC339 format)?",
		Help:    "show/change runtime date and time",
		Fn:      dateCmd,
	})

	Add(Cmd{
		Name: "uptime",
		Help: "show how long the system has been running",
		Fn:   uptimeCmd,
	})

	// The following commands are board specific, therefore their Fn
	// pointers are defined elsewhere in the respective target files.

	Add(Cmd{
		Name: "info",
		Help: "device information",
		Fn:   infoCmd,
	})

	Add(Cmd{
		Name: "reboot",
		Help: "reset device",
		Fn:   rebootCmd,
	})
}

func helpCmd(_ *Interface, term *term.Terminal, _ []string) (string, error) {
	return Help(term), nil
}

func buildInfoCmd(_ *Interface, term *term.Terminal, _ []string) (string, error) {
	if bi, ok := debug.ReadBuildInfo(); ok {
		fmt.Fprintf(term, bi.String())
	}

	return "", nil
}

func exitCmd(_ *Interface, term *term.Terminal, _ []string) (string, error) {
	fmt.Fprintf(term, "Goodbye from %s/%s\n", runtime.GOOS, runtime.GOARCH)
	return "logout", io.EOF
}

func haltCmd(_ *Interface, term *term.Terminal, _ []string) (string, error) {
	fmt.Fprintf(term, "Goodbye from %s/%s\n", runtime.GOOS, runtime.GOARCH)
	go runtime.Exit(0)
	return "halted", io.EOF
}

func stackCmd(_ *Interface, _ *term.Terminal, _ []string) (string, error) {
	return string(debug.Stack()), nil
}

func stackallCmd(_ *Interface, _ *term.Terminal, _ []string) (string, error) {
	buf := new(bytes.Buffer)
	pprof.Lookup("goroutine").WriteTo(buf, 1)

	return buf.String(), nil
}

func dmaCmd(_ *Interface, term *term.Terminal, arg []string) (string, error) {
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

func dateCmd(iface *Interface, _ *term.Terminal, arg []string) (res string, err error) {
	if len(arg[0]) > 1 {
		t, err := time.Parse(time.RFC3339, arg[0][1:])

		if err != nil {
			return "", err
		}

		date(iface, t.UnixNano())
	}

	return fmt.Sprintf("%s", time.Now().Format(time.RFC3339)), nil
}

func uptimeCmd(iface *Interface, term *term.Terminal, _ []string) (string, error) {
	ns := uptime(iface)
	return fmt.Sprintf("%s", durafmt.Parse(time.Duration(ns)*time.Nanosecond)), nil
}
