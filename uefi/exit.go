// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uefi

// EFI Boot Services offsets
const (
	exit             = 0xd8
	exitBootServices = 0xe8
)

// Exit calls EFI_BOOT_SERVICES.Exit().
func (s *BootServices) Exit(code int) (err error) {
	status := callService(
		s.base+exit,
		uint64(s.imageHandle),
		uint64(code),
		0,
		0,
	)

	return parseStatus(status)
}

// ExitServices calls EFI_BOOT_SERVICES.ExitBootServices().
func (s *BootServices) ExitBootServices() (err error) {
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

	return parseStatus(status)
}
