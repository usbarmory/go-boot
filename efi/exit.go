// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package efi

// EFI Boot Services offset for ExitBootServices
const exitBootServices = 0xe8

// Exit calls EFI_BOOT_SERVICES.ExitBootServices().
func (s *BootServices) Exit(imageHandle uint64, mapKey uint64) error {
	status := callService(
		s.base+exitBootServices,
		imageHandle,
		ptrval(&mapKey),
		0,
		0,
	)

	return parseStatus(status)
}
