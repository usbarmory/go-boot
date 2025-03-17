// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uefi

import (
	"encoding/hex"
	"regexp"
)

var guidPattern = regexp.MustCompile(`^([[:xdigit:]]{8})-([[:xdigit:]]{4})-([[:xdigit:]]{4})-([[:xdigit:]]{4})-([[:xdigit:]]{12})$`)

// GUID represents an EFI GUID (Globally Unique Identifier).
type GUID string

// Bytes returns the GUID as byte slice.
func (g GUID) Bytes() (guid []byte) {
	var buf []byte
	var err error

	m := guidPattern.FindStringSubmatch(string(g))

	if len(m) != 6 {
		return make([]byte, 16)
	}

	m = m[1:]

	for i, b := range m {
		if buf, err = hex.DecodeString(b); err != nil {
			return make([]byte, 16)
		}

		switch i {
		case 0:
			guid = append(guid, buf[3])
			guid = append(guid, buf[2])
			guid = append(guid, buf[1])
			guid = append(guid, buf[0])
		case 1, 2:
			guid = append(guid, buf[1])
			guid = append(guid, buf[0])
		default:
			guid = append(guid, buf...)
		}
	}

	return
}

func (g GUID) ptrval() uint64 {
	buf := g.Bytes()
	return ptrval(&buf[0])
}
