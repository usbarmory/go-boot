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
	getVariable         = 0x48
	getNextVariableName = 0x50
)

// VariableAttributes represents the attributes of a UEFI variable.
// See: https://uefi.org/specs/UEFI/2.11/08_Services_Runtime_Services.html#getvariable
type VariableAttributes struct {
	NonVolatile              bool
	BootServiceAccess        bool
	RuntimeServiceAccess     bool
	HardwareErrorRecord      bool
	AuthWriteAccess          bool
	TimeBasedAuthWriteAccess bool
	AppendWrite              bool
	EnhancedAuthAccess       bool
}

// GetVariable calls EFI_RUNTIME_SERVICES.GetVariable().
// See: https://uefi.org/specs/UEFI/2.11/08_Services_Runtime_Services.html#getvariable
func (s *RuntimeServices) GetVariable(name string, guid GUID, withData bool) (attr VariableAttributes, dataSize uint64, data []byte, err error) {
	// Convert lastName to UTF-16 for UEFI
	nameUTF16 := toUTF16(name)

	var attributes uint32

	// The first call retrieves the attributes and size of data
	status := callService(s.base+getVariable,
		[]uint64{
			ptrval(&nameUTF16[0]),
			guid.ptrval(),
			ptrval(&attributes),
			ptrval(&dataSize),
			0,
		},
	)

	if status != EFI_SUCCESS && status&0xff != EFI_BUFFER_TOO_SMALL {
		err = parseStatus(status)
		return VariableAttributes{}, 0, nil, err
	}

	// Parse attributes
	attr.NonVolatile = attributes&0x1 != 0
	attr.BootServiceAccess = attributes&0x2 != 0
	attr.RuntimeServiceAccess = attributes&0x4 != 0
	attr.HardwareErrorRecord = attributes&0x8 != 0
	attr.AuthWriteAccess = attributes&0x10 != 0
	attr.TimeBasedAuthWriteAccess = attributes&0x20 != 0
	attr.AppendWrite = attributes&0x40 != 0
	attr.EnhancedAuthAccess = attributes&0x80 != 0

	if !withData {
		return attr, dataSize, nil, nil
	}

	// The second call retrieves the data
	data = make([]byte, dataSize)
	status = callService(s.base+getVariable,
		[]uint64{
			ptrval(&nameUTF16[0]),
			ptrval(&guid[0]),
			0,
			ptrval(&dataSize),
			ptrval(&data[0]),
		},
	)

	if err = parseStatus(status); err != nil {
		return attr, 0, nil, err
	}

	return attr, dataSize, data, nil
}

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
