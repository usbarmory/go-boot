// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uefi

// EFI Boot Services offset for ExitBootServices
const exitBootServices = 0xe8

// Exit calls EFI_BOOT_SERVICES.ExitBootServices().
func (s *BootServices) Exit() (err error) {
	memoryMap, err := s.GetMemoryMap()

	if err != nil {
		return
	}

	status := callService(
		s.base+exitBootServices,
		uint64(s.imageHandle),
		memoryMap.MapKey,
		0,
		0,
	)

	if err = parseStatus(status); err != nil {
		return
	}

	return
}
