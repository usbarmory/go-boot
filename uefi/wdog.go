// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uefi

const (
	// EFI Boot Services offset for SetWatchdogTimer
	setWatchdogTimer = 0x100
	watchdogCode     = 0xba3e5e7a1
)

// SetWatchdogTimer calls EFI_BOOT_SERVICES.SetWatchdogTimer()
func (s *BootServices) SetWatchdogTimer(sec int) (err error) {
	status := callService(s.base+setWatchdogTimer,
		[]uint64{
			uint64(sec),
			watchdogCode,
			0,
			0,
		},
	)

	return parseStatus(status)
}
