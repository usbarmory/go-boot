// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uefi

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"regexp"
)

var guidPattern = regexp.MustCompile(`^([[:xdigit:]]{8})-([[:xdigit:]]{4})-([[:xdigit:]]{4})-([[:xdigit:]]{4})-([[:xdigit:]]{12})$`)

// GUID represents an EFI GUID (Globally Unique Identifier) as a 16-byte array
// with the native EFI byte order.
//
// Note: The registry string format (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)
// reorders the first three fields as little-endian. Internally, we keep the
// native EFI layout (as used in memory), i.e. 16 bytes where the first three
// fields are little-endian values.
type GUID [16]byte

// ParseGUID parses a GUID in registry string format into a native EFI GUID
// byte slice (len 16). On parse error it returns nil and an error.
func ParseGUID(s string) (out GUID, err error) {
	var off int
	var buf []byte

	m := guidPattern.FindStringSubmatch(s)

	if len(m) != 6 {
		return GUID{}, fmt.Errorf("invalid GUID format: %q", s)
	}

	m = m[1:]

	for i, b := range m {
		if buf, err = hex.DecodeString(b); err != nil {
			return GUID{}, err
		}

		switch i {
		case 0:
			out[off+0] = buf[3]
			out[off+1] = buf[2]
			out[off+2] = buf[1]
			out[off+3] = buf[0]
			off += 4
		case 1, 2:
			out[off+0] = buf[1]
			out[off+1] = buf[0]
			off += 2
		default:
			copy(out[off:], buf)
			off += len(buf)
		}
	}

	return out, nil
}

// MustParseGUID is like ParseGUID but panics on error. It is intended for package
// level GUID declarations.
func MustParseGUID(s string) (g GUID) {
	var err error

	if g, err = ParseGUID(s); err != nil {
		panic(err)
	}

	return
}

// String returns the registry format string representation of the GUID.
// https://uefi.org/specs/UEFI/2.10/Apx_A_GUID_and_Time_Formats.html
func (g GUID) String() string {
	// First three fields are little-endian 32/16/16
	return fmt.Sprintf("%08x-%04x-%04x-%x-%x",
		binary.LittleEndian.Uint32(g[0:4]),
		binary.LittleEndian.Uint16(g[4:6]),
		binary.LittleEndian.Uint16(g[6:8]),
		g[8:10],
		g[10:])
}
