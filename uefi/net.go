// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uefi

const (
	EFI_SIMPLE_NETWORK_PROTOCOL_GUID     = "a19832b9-ac25-11d3-9a2d-0090273fc14d"
	EFI_SIMPLE_NETWORK_PROTOCOL_REVISION = 0x00010000

	EFI_SIMPLE_NETWORK_TRANSMIT_INTERRUPT = 0x02
)

// EFI Simple Network Protocol offsets
const (
	start      = 0x08
	stop       = 0x10
	initialize = 0x18
	getStatus  = 0x58
	transmit   = 0x60
	receive    = 0x68
)

// SimpleNetwork represents an EFI Simple Network Protocol instance.
type SimpleNetwork struct {
	base uint64
}

// Start calls EFI_SIMPLE_NETWORK.Start()
func (sn *SimpleNetwork) Start() (err error) {
	status := callService(sn.base+start,
		[]uint64{
			sn.base,
		},
	)

	return parseStatus(status)
}

// Stop calls EFI_SIMPLE_NETWORK.Stop()
func (sn *SimpleNetwork) Stop() (err error) {
	status := callService(sn.base+stop,
		[]uint64{
			sn.base,
		},
	)

	return parseStatus(status)
}

// Initialize calls EFI_SIMPLE_NETWORK.Initialize()
func (sn *SimpleNetwork) Initialize() (err error) {
	status := callService(sn.base+initialize,
		[]uint64{
			sn.base,
			0,
			0,
		},
	)

	return parseStatus(status)
}

// GetStatus calls EFI_SIMPLE_NETWORK.GetStatus()
func (sn *SimpleNetwork) GetStatus() (interruptStatus uint32, txBuf uint64, err error) {
	status := callService(sn.base+getStatus,
		[]uint64{
			sn.base,
			ptrval(&interruptStatus),
			ptrval(&txBuf),
		},
	)

	err = parseStatus(status)

	return
}

// Transmit calls EFI_SIMPLE_NETWORK.Transmit(), the function waits for
// EFI_SIMPLE_NETWORK.GetStatus() to report a transmit interrupt before
// returning.
func (sn *SimpleNetwork) Transmit(buf []byte) (err error) {
	var interruptStatus uint32

	status := callService(sn.base+transmit,
		[]uint64{
			sn.base,
			0,
			uint64(len(buf)),
			ptrval(&buf[0]),
			0,
			0,
			0,
		},
	)

	if err = parseStatus(status); err != nil {
		return
	}

	for {
		if interruptStatus, _, err = sn.GetStatus(); err != nil {
			return
		}

		if interruptStatus & EFI_SIMPLE_NETWORK_TRANSMIT_INTERRUPT != 0 {
			break
		}
	}

	return
}

// Receive calls EFI_SIMPLE_NETWORK.Receive()
func (sn *SimpleNetwork) Receive(buf []byte) (n int, err error) {
	size := uint64(len(buf))

	status := callService(sn.base+receive,
		[]uint64{
			sn.base,
			0,
			ptrval(&size),
			ptrval(&buf[0]),
			0,
			0,
			0,
		},
	)

	if status&0xff == EFI_NOT_READY {
		return 0, nil
	}

	return int(size), parseStatus(status)
}

// GetNetwork locates and returns the EFI Simple Network Protocol instance.
func (s *BootServices) GetNetwork() (sn *SimpleNetwork, err error) {
	sn = &SimpleNetwork{}
	sn.base, err = s.LocateProtocol(EFI_SIMPLE_NETWORK_PROTOCOL_GUID)
	return
}
