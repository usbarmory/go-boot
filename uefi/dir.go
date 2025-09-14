// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uefi

import (
	"errors"
	"io"
	"io/fs"
)

// DirEntry implements the [fs.DirEntry] interface for the EFI File Protocol.
type DirEntry struct {
	fi *FileInfo
}

// Name returns the name of the  (or subdirectory) described by the entry.
func (d DirEntry) Name() string {
	return d.fi.name
}

// IsDir reports whether the entry describes a directory.
func (d DirEntry) IsDir() bool {
	return d.fi.IsDir()
}

// Type returns the file mode bits.
func (d DirEntry) Type() fs.FileMode {
	return d.fi.Mode()
}

// Info returns the FileInfo for the file or subdirectory described by the entry.
func (d DirEntry) Info() (fs.FileInfo, error) {
	return fs.FileInfo(d.fi), nil
}

// ReadDir reads the contents of the directory and returns
// a slice of up to n DirEntry values in directory order.
// Subsequent calls on the same file will yield further DirEntry values.
func (f *File) ReadDir(n int) (entries []fs.DirEntry, err error) {
	if fi, err := f.Stat(); err != nil || !fi.IsDir() {
		return nil, errors.New("not a directory")
	}

	if n < 0 {
		n = MaxDirEntries - f.n
	}

	for i := f.n; i < n; i++ {
		buf := make([]byte, fileInfoSize+MaxFileName*2)

		if _, err = f.Read(buf); err == io.EOF {
			return entries, nil
		}

		if err != nil {
			return nil, err
		}

		entry := DirEntry{
			fi: &FileInfo{
				info: &fileInfo{},
			},
		}

		if entry.fi.name, err = entry.fi.info.decode(buf); err != nil {
			return
		}

		entries = append(entries, fs.DirEntry(entry))
	}

	f.n += len(entries)

	return
}
