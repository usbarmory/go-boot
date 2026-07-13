// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uefi

var EFI_MP_SERVICES_PROTOCOL_GUID = MustParseGUID("3fdda605-a76e-4f46-ad29-12f4531b3d08")

// EFI Multi Processor Services Protocol offsets
const (
	getNumberOfProcessors = 0x00
	startupAllAPs         = 0x10
)

// MultiProcessor represents an EFI Multi Processor Services instance.
type MultiProcessor struct {
	base uint64
}

// GetNumberOfProcessors calls EFI_MP_SERVICES_PROTOCOL.GetNumberOfProcessors().
func (mp *MultiProcessor) GetNumberOfProcessors() (num uint64, enabled uint64, err error) {
	status := callService(mp.base+getNumberOfProcessors,
		[]uint64{
			mp.base,
			ptrval(&num),
			ptrval(&enabled),
		},
	)

	err = parseStatus(status)

	return
}

// StartupAllAPs calls EFI_MP_SERVICES_PROTOCOL.StartupAllAPs().
func (mp *MultiProcessor) StartupAllAPs(pc uintptr, singleThread bool, timeout uint64) (err error) {
	st := uint64(0)

	if singleThread {
		st = 1
	}

	status := callService(mp.base+startupAllAPs,
		[]uint64{
			mp.base,
			uint64(pc),
			0,
			st,
			0,
			timeout,
			0,
		},
	)

	return parseStatus(status)
}

// GetMultiProcessor locates and returns the EFI Multi Processor Services
// instance.
func (s *BootServices) GetMultiProcessor() (mp *MultiProcessor, err error) {
	mp = &MultiProcessor{}
	mp.base, err = s.LocateProtocol(EFI_MP_SERVICES_PROTOCOL_GUID)
	return
}
