// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uefi

// EFI Runtime Services offset for ResetSystem
const resetSystem = 0x68

// EFI_RESET_SYSTEM
const (
	EfiResetCold = iota
	EfiResetWarm
	EfiResetShutdown
	EfiResetPlatformSpecific
)

// ResetSystem calls EFI_RUNTIME_SERVICES.ResetSystem().
func (s *RuntimeServices) ResetSystem(resetType int) (err error) {
	status := callService(s.base+resetSystem,
		[]uint64{
			uint64(resetType),
			EFI_SUCCESS,
			0,
			0,
		},
	)

	return parseStatus(status)
}
