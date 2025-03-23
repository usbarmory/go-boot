// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package shell

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"text/tabwriter"

	"github.com/usbarmory/go-boot/uefi"
)

// CmdFn represents a command handler.
type CmdFn func(c *Interface, arg []string) (res string, err error)

// Cmd represents a shell command.
type Cmd struct {
	// Name is the command name.
	Name string

	// Args defines the number of command arguments, meant to be in the
	// Pattern capturing brackets.
	Args int

	// Pattern defines the command syntax and arguments.
	Pattern *regexp.Regexp

	// Syntax defines the Help() command syntax field.
	Syntax string

	// Help defines the Help() command description field.
	Help string

	// Fn defines the command handler.
	Fn CmdFn
}

var cmds = make(map[string]*Cmd)

// Add registers a terminal interface command.
func Add(cmd Cmd) {
	cmds[cmd.Name] = &cmd
}

// Help returns a formatted string with instructions for all registered
// commands.
func (c *Interface) Help(_ *Interface, _ []string) (_ string, _ error) {
	var help bytes.Buffer
	var names []string

	t := tabwriter.NewWriter(&help, 16, 8, 0, ' ', tabwriter.TabIndent)

	for name, _ := range cmds {
		names = append(names, name)
	}

	sort.Strings(names)

	for _, name := range names {
		_, _ = fmt.Fprintf(t, "%s\t%s\t # %s\n", cmds[name].Name, cmds[name].Syntax, cmds[name].Help)
	}

	_ = t.Flush()
	res := help.String()

	switch {
	case c.Terminal != nil:
		res = string(c.Terminal.Escape.Cyan) + res + string(c.Terminal.Escape.Reset)
		fmt.Fprintln(c.Output, res)
	case c.Console != nil:
		c.Console.SetAttribute(uefi.EFI_LIGHTCYAN)
		fmt.Fprintln(c.Output, res)
		c.Console.SetAttribute(uefi.EFI_WHITE)
	}

	return
}
