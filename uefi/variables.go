// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uefi

var (
	EFI_GLOBAL_VARIABLE_GUID = MustParseGUID("8BE4DF61-93CA-11D2-AA0D-00E098032B8C")
)

// EFI Runtime Services offset for Variable Services
// See: https://uefi.org/specs/UEFI/2.11/08_Services_Runtime_Services.html#variable-services
const (
	getNextVariableName = 0x50
)

// GetNextVariableName calls EFI_RUNTIME_SERVICES.GetNextVariableName().
// See: https://uefi.org/specs/UEFI/2.11/08_Services_Runtime_Services.html#getnextvariablename
func (s *RuntimeServices) GetNextVariableName(name *string, guid *GUID) (err error) {
	// Convert name to UTF-16 for UEFI
	lastNameUTF16 := toUTF16(*name)

	// Calculate buffer size: need space for variable name (UTF-16) + null terminator
	// UEFI spec suggests 1024 bytes minimum, but we need more for longer names
	initialSize := uint64(1024)
	requiredSize := uint64(len(lastNameUTF16))
	if requiredSize > initialSize {
		initialSize = requiredSize
	}

	// Create a buffer that can hold UTF-16 encoded variable names
	nameBuf := make([]byte, initialSize)
	copy(nameBuf, lastNameUTF16)

	status := callService(s.base+getNextVariableName,
		[]uint64{
			ptrval(&initialSize),
			ptrval(&nameBuf[0]),
			ptrval(&guid[0]),
		},
	)

	err = parseStatus(status)
	if err != nil {
		if status&0xff == EFI_NOT_FOUND {
			err = ErrEfiNotFound
		} else {
			return err
		}
	}

	*name = fromUTF16(nameBuf)

	return err
}
