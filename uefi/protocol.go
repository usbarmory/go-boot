// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uefi

// EFI Boot Services offsets
const (
	handleProtocol     = 0x098
	locateHandleBuffer = 0x138
	locateProtocol     = 0x140
)

// EFI_LOCATE_SEARCH_TYPE values
const (
	AllHandles = iota
	ByRegisterNotify
	ByProtocol
)

// HandleProtocol calls EFI_BOOT_SERVICES.HandleProtocol().
func (s *BootServices) HandleProtocol(handle uint64, guid GUID) (addr uint64, err error) {
	status := callService(s.base+handleProtocol,
		[]uint64{
			handle,
			ptrval(&guid[0]),
			ptrval(&addr),
		},
	)

	return addr, parseStatus(status)
}

// LocateProtocol calls EFI_BOOT_SERVICES.LocateProtocol().
func (s *BootServices) LocateProtocol(guid GUID) (addr uint64, err error) {
	status := callService(s.base+locateProtocol,
		[]uint64{
			ptrval(&guid[0]),
			0,
			ptrval(&addr),
		},
	)

	return addr, parseStatus(status)
}

// LocateHandleBuffer calls EFI_BOOT_SERVICES.LocateHandleBuffer(), returning the
// handles matching searchType (for ByProtocol, guid selects the protocol).
func (s *BootServices) LocateHandleBuffer(searchType uint64, guid GUID) (handles []uint64, err error) {
	var noHandles uint64
	var buffer uint64

	status := callService(s.base+locateHandleBuffer,
		[]uint64{
			searchType,
			ptrval(&guid[0]),
			0,
			ptrval(&noHandles),
			ptrval(&buffer),
		},
	)

	if err = parseStatus(status); err != nil {
		return nil, err
	}

	if noHandles == 0 || buffer == 0 {
		return nil, nil
	}

	handles = make([]uint64, noHandles)

	if err = decode(handles, buffer); err != nil {
		return nil, err
	}

	if err = s.FreePool(buffer); err != nil {
		return nil, err
	}

	return handles, nil
}
