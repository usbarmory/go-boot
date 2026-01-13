// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package uapi implements Boot Loader Entries parsing
// following the specifications at:
//
//	https://uapi-group.org/specifications/specs/boot_loader_specification
package uapi

import (
	"fmt"
	"io/fs"
	"strings"
)

// Entry represents the contents loaded from a Type #1 Boot Loader Entry.
type Entry struct {
	// Title is the human-readable entry title.
	Title string
	// Linux is the kernel image to execute.
	Linux []byte
	// Initrd is the ramdisk cpio image, multiple entries are concatenated.
	Initrd []byte
	// Options is the kernel parameters.
	Options string

	parsed  string
	ignored string

	fsys fs.FS
}

func (e *Entry) loadKeyValue(v string) ([]byte, error) {
	return fs.ReadFile(e.fsys, v)
}

func (e *Entry) parseKey(line string) (err error) {
	kv := strings.SplitN(line, " ", 2)

	if len(kv) < 2 {
		return
	}

	k := kv[0]
	v := strings.Trim(kv[1], "\n\r")
	v = strings.TrimSpace(v)

	switch k {
	case "title":
		e.Title = v
	case "linux":
		if e.Linux, err = e.loadKeyValue(v); err != nil {
			return
		}
	case "initrd":
		var initrd []byte

		if initrd, err = e.loadKeyValue(v); err != nil {
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

// String returns the successfully parsed entry keys.
func (e *Entry) String() string {
	return e.parsed
}

// Ignored returns the entry keys ignored during parsing.
func (e *Entry) Ignored() string {
	return e.ignored
}

// LoadEntry parses Type #1 Boot Loader Specification Entries from the argument
// file and loads each key contents from the argument file system.
func LoadEntry(fsys fs.FS, path string) (e *Entry, err error) {
	e = &Entry{
		fsys: fsys,
	}

	entry, err := fs.ReadFile(fsys, path)

	if err != nil {
		return nil, fmt.Errorf("error reading entry file, %v", err)
	}

	for line := range strings.Lines(string(entry)) {
		if err = e.parseKey(line); err != nil {
			return nil, fmt.Errorf("error parsing entry line, %v line:%s", err, line)
		}
	}

	return
}
