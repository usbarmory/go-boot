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

	"golang.org/x/term"

	"github.com/usbarmory/go-boot/uefi"
)

// DefaultPrompt represents the command prompt when none is set for the
// Interface instance.
var DefaultPrompt = "> "

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

	fmt.Fprintln(c.Output, res)

	return
}

func (c *Interface) readLine(t *term.Terminal) error {
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

	s, err := t.ReadLine()

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

func (c *Interface) handle(t *term.Terminal) {
	if len(c.Prompt) == 0 {
		c.Prompt = DefaultPrompt
	}

	if c.Terminal != nil {
		c.Output = c.Terminal
	} else {
		c.Output = c.ReadWriter
	}

	fmt.Fprintf(t, "\n%s\n\n", c.Banner)
	Help(c, nil)

	for {
		if err := c.readLine(t); err != nil {
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
		c.handle(c.Terminal)
	case c.ReadWriter != nil:
		t := term.NewTerminal(c.ReadWriter, "")

		if vt100 {
			c.Terminal = t
		}

		c.handle(t)
	}
}
