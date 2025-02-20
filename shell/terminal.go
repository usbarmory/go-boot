// Copyright (c) WithSecure Corporation
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
)

func init() {
}

// Interface represents a terminal interface.
type Interface struct {
	// Banner represents the welcome message
	Banner string

	// Log represents the interface log file
	Log *os.File

	// ReadWriter represents the terminal connection
	ReadWriter io.ReadWriter

	VT100 bool
}

func (iface *Interface) handleLine(line string, w io.Writer) (err error) {
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

	if res, err = match.Fn(arg); err != nil {
		return
	}

	fmt.Fprintln(w, res)

	return
}

func (iface *Interface) readLine(t *term.Terminal, w io.Writer) (error) {
	s, err := t.ReadLine()

	if err == io.EOF {
		return err
	}

	if err != nil {
		log.Printf("readline error, %v", err)
		return nil
	}

	if err = iface.handleLine(s, w); err != nil {
		if err == io.EOF {
			return err
		}

		fmt.Fprintf(w, "command error, %v\n", err)
		return nil
	}

	return nil
}

// Start handles registered commands over the interface ReadWriter.
func (iface *Interface) Start() {
	var w io.Writer

	t := term.NewTerminal(iface.ReadWriter, "")
	w = iface.ReadWriter

	if iface.VT100 {
		t.SetPrompt(string(t.Escape.Red) + "> " + string(t.Escape.Reset))
		w = t
	}

	help, _  := iface.Help(nil)

	fmt.Fprintf(t, "\n%s\n\n", iface.Banner)
	fmt.Fprintf(t, "%s\n", help)

	Add(Cmd{
		Name: "help",
		Help: "this help",
		Fn:   iface.Help,
	})

	for {
		if err := iface.readLine(t, w); err != nil {
			return
		}
	}
}
