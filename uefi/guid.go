// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uefi

import (
	"uuid"
)

// GUID represents an EFI GUID (Globally Unique Identifier) as a 16-byte array
// with the native EFI byte order.
//
// Note: The registry string format (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx), as
// well as the RFC 9562 encoding used by [uuid.UUID], stores the first three
// fields as big-endian values. Internally, we keep the native EFI layout (as
// used in memory), i.e. 16 bytes where the first three fields are
// little-endian values.
type GUID [16]byte

// swap converts a 16-byte identifier between the RFC 9562 and the native EFI
// byte order by reversing its first three fields.
func swap(in [16]byte) (out [16]byte) {
	out = in

	out[0], out[1], out[2], out[3] = in[3], in[2], in[1], in[0]
	out[4], out[5] = in[5], in[4]
	out[6], out[7] = in[7], in[6]

	return
}

// FromUUID converts an RFC 9562 UUID to a native EFI GUID.
func FromUUID(u uuid.UUID) GUID {
	return GUID(swap(u))
}

// UUID converts a native EFI GUID to an RFC 9562 UUID.
func (g GUID) UUID() uuid.UUID {
	return uuid.UUID(swap(g))
}

// ParseGUID parses a GUID, in any of the string formats accepted by
// [uuid.Parse], into a native EFI GUID.
func ParseGUID(s string) (g GUID, err error) {
	var u uuid.UUID

	if u, err = uuid.Parse(s); err != nil {
		return GUID{}, err
	}

	return FromUUID(u), nil
}

// MustParseGUID is like ParseGUID but panics on error. It is intended for package
// level GUID declarations.
func MustParseGUID(s string) GUID {
	return FromUUID(uuid.MustParse(s))
}

// String returns the registry format string representation of the GUID.
// https://uefi.org/specs/UEFI/2.10/Apx_A_GUID_and_Time_Formats.html
func (g GUID) String() string {
	return g.UUID().String()
}
