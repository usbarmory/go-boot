// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package shell implements a terminal console handler for user defined
// commands.
package shell

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/usbarmory/go-boot/uefi"
)

// DefaultPrompt represents the command prompt when none is set for the
// Interface instance.
var DefaultPrompt = "> "

// PaginationPrompt represents the prompt before displaying the next output
// page when [Interface.Pagination] is enabled.
var PaginationPrompt = "<press enter to continue>"

// Interface represents a terminal interface.
type Interface struct {
	// Prompt represents the command prompt
	Prompt string
	// Banner represents the welcome message
	Banner string

	// Log represents the interface log file
	Log *os.File

	// ReadWriter represents the terminal connection
	ReadWriter io.ReadWriter

	// Output represents the interface output
	Output io.Writer
	// Terminal represents the VT100 terminal output
	Terminal *term.Terminal
	// Console represents the UEFI Console
	Console *uefi.Console

	// Pagination enables console pagination to avoid frame buffer
	// scrolling.
	Pagination bool

	t   *term.Terminal
}

func (c *Interface) paginate(prompt bool) (err error) {
	var mode *uefi.OutputMode
	var rows uint64

	if mode, err = c.Console.GetMode(); err != nil {
		return
	}

	if _, rows, err = c.Console.QueryMode(uint64(mode.Mode)); err != nil {
		return
	}

	if mode.CursorRow < int32(rows) - 2 {
		return
	}

	if prompt {
		fmt.Fprintf(c.Output, PaginationPrompt)
		c.t.ReadLine()
	}

	c.Console.ClearScreen()

	return
}

func (c *Interface) handleLine(line string) (err error) {
	var match *Cmd
	var arg []string
	var res string

	for _, cmd := range cmds {
		if cmd.Pattern == nil {
			if cmd.Name == line {
				match = cmd
				break
			}
		} else if m := cmd.Pattern.FindStringSubmatch(line); len(m) > 0 && (len(m)-1 == cmd.Args) {
			match = cmd
			arg = m[1:]
			break
		}
	}

	if match == nil {
		return errors.New("unknown command, type `help`")
	}

	if res, err = match.Fn(c, arg); err != nil {
		return
	}

	if len(res) == 0 {
		return
	}

	if c.Console == nil || !c.Pagination {
		fmt.Fprintln(c.Output, res)
		return
	}

	if err = c.paginate(false); err != nil {
		fmt.Fprintln(c.Output, res)
		return
	}

	for line := range strings.Lines(res) {
		fmt.Fprintf(c.Output, "%s", line)
		c.paginate(true)
	}

	return
}

func (c *Interface) readLine() error {
	switch {
	case c.Terminal != nil:
		fmt.Fprint(c.Output, string(c.Terminal.Escape.Red)+c.Prompt+string(c.Terminal.Escape.Reset))
	case c.Console != nil:
		c.Console.SetAttribute(uefi.EFI_RED)
		fmt.Fprint(c.Output, c.Prompt)
		c.Console.SetAttribute(uefi.EFI_WHITE)
	default:
		fmt.Fprint(c.Output, c.Prompt)
	}

	s, err := c.t.ReadLine()

	if err == io.EOF {
		return err
	}

	if err != nil {
		log.Printf("readline error, %v", err)
		return nil
	}

	if err = c.handleLine(s); err != nil {
		if err == io.EOF {
			return err
		}

		fmt.Fprintf(c.Output, "command error, %v\n", err)
		return nil
	}

	return nil
}

// Exec executes an individual command.
func (c *Interface) Exec(cmd []byte) {
	if err := c.handleLine(string(cmd)); err != nil {
		fmt.Fprintf(c.Output, "command error (%s), %v\n", cmd, err)
	}
}

func (c *Interface) handle() {
	if len(c.Prompt) == 0 {
		c.Prompt = DefaultPrompt
	}

	if c.Terminal != nil {
		c.Output = c.Terminal
	} else {
		c.Output = c.ReadWriter
	}

	fmt.Fprintf(c.t, "\n%s\n\n", c.Banner)
	Help(c, nil)

	for {
		if err := c.readLine(); err != nil {
			return
		}
	}
}

// Start handles registered commands over the interface Terminal or ReadWriter,
// the argument specifies whether ReadWriter is VT100 compatible.
func (c *Interface) Start(vt100 bool) {
	Add(Cmd{
		Name: "help",
		Help: "this help",
		Fn:   Help,
	})

	switch {
	case c.Terminal != nil:
		c.t = c.Terminal
		c.handle()
	case c.ReadWriter != nil:
		c.t = term.NewTerminal(c.ReadWriter, "")

		if vt100 {
			c.Terminal = c.t
		}

		c.handle()
	}
}
