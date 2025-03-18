// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package uapi implements Boot Loader Entries parsing
// following the specifications at:
//
//	https://uapi-group.org/specifications/specs/boot_loader_specification/
package uapi

import (
	"io/fs"
	"strings"
)

// Entry represents the parsed contents of Type #1 Boot Loader Entry Keys.
type Entry struct {
	Title   string
	Linux   []byte
	Initrd  []byte
	Options string

	parsed  string
	ignored string
}

func (e *Entry) parseKey(fsys fs.FS, line string) (err error) {
	kv := strings.SplitN(line, " ", 2)

	if len(kv) < 2 {
		return
	}

	k := kv[0]
	v := strings.Trim(kv[1], "\n\r")

	switch k {
	case "title":
		e.Title = v
	case "linux":
		v = strings.ReplaceAll(v, `/`, `\`)

		if e.Linux, err = fs.ReadFile(fsys, v); err != nil {
			return
		}
	case "initrd":
		v = strings.ReplaceAll(v, `/`, `\`)

		var initrd []byte

		if initrd, err = fs.ReadFile(fsys, v); err != nil {
			return
		}

		e.Initrd = append(e.Initrd, initrd...)
	case "options":
		e.Options += v
	default:
		e.ignored += line
		return
	}

	e.parsed += line

	return
}

// String returns the lines successfully parsed.
func (e *Entry) String() string {
	return e.parsed
}

// Ignored returns the lines ignored during parsing.
func (e *Entry) Ignored() string {
	return e.ignored
}

// LoadEntry parses Type #1 Boot Loader Specification Entries from the argument
// file and loads each key contents from the argument file system.
func LoadEntry(fsys fs.FS, path string) (e *Entry, err error) {
	e = &Entry{}

	entry, err := fs.ReadFile(fsys, path)

	if err != nil {
		return
	}

	for line := range strings.Lines(string(entry)) {
		if err = e.parseKey(fsys, line); err != nil {
			return
		}
	}

	return
}
