// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uefi

// EFI Boot Services offsets
const (
	handleProtocol = 0x098
	locateProtocol = 0x140
)

// HandleProtocol calls EFI_BOOT_SERVICES.HandleProtocol().
func (s *BootServices) HandleProtocol(handle uint64, guid GUID) (addr uint64, err error) {
	status := callService(s.base+handleProtocol,
		[]uint64{
			handle,
			guid.ptrval(),
			ptrval(&addr),
		},
	)

	return addr, parseStatus(status)
}

// LocateProtocol calls EFI_BOOT_SERVICES.LocateProtocol().
func (s *BootServices) LocateProtocol(guid GUID) (addr uint64, err error) {
	status := callService(s.base+locateProtocol,
		[]uint64{
			guid.ptrval(),
			0,
			ptrval(&addr),
		},
	)

	return addr, parseStatus(status)
}
