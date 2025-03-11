// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package efi

import (
	"encoding/hex"
	"errors"
	"regexp"
)

// EFI Boot Services offset for LocateProtocol
const locateProtocol = 0x140

var guidPattern = regexp.MustCompile(`^([[:xdigit:]]{8})-([[:xdigit:]]{4})-([[:xdigit:]]{4})-([[:xdigit:]]{4})-([[:xdigit:]]{12})$`)

// LocateProtocol calls EFI_BOOT_SERVICES.LocateProtocol().
func (s *BootServices) LocateProtocol(guid []byte) (addr uint64, err error) {
	if len(guid) != 16 {
		return 0, errors.New("invalid argument")
	}

	status := callService(
		s.base+locateProtocol,
		ptrval(&guid[0]),
		0,
		ptrval(&addr),
		0,
	)

	return addr, parseStatus(status)
}

// LocateProtocolString calls EFI_BOOT_SERVICES.LocateProtocol()
func (s *BootServices) LocateProtocolString(g string) (addr uint64, err error) {
	var buf []byte
	var guid []byte

	m := guidPattern.FindStringSubmatch(g)

	if len(m) != 6 {
		return 0, errors.New("invalid argument")
	}

	m = m[1:]

	for i, b := range m {
		if buf, err = hex.DecodeString(b); err != nil {
			return
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

	return s.LocateProtocol(guid)
}
