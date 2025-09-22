// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uefi

import (
	"errors"
	"time"
)

const (
	EFI_SIMPLE_NETWORK_PROTOCOL_GUID     = "a19832b9-ac25-11d3-9a2d-0090273fc14d"
	EFI_SIMPLE_NETWORK_PROTOCOL_REVISION = 0x00010000

	EFI_SIMPLE_NETWORK_TRANSMIT_INTERRUPT = 0x02

	EFI_SIMPLE_NETWORK_RECEIVE_UNICAST               = 0x01
	EFI_SIMPLE_NETWORK_RECEIVE_MULTICAST             = 0x02
	EFI_SIMPLE_NETWORK_RECEIVE_BROADCAST             = 0x04
	EFI_SIMPLE_NETWORK_RECEIVE_PROMISCUOUS           = 0x08
	EFI_SIMPLE_NETWORK_RECEIVE_PROMISCUOUS_MULTICAST = 0x10
)

// EFI Simple Network Protocol offsets
const (
	start          = 0x08
	stop           = 0x10
	initialize     = 0x18
	shutdown       = 0x28
	receiveFilters = 0x30
	stationAddress = 0x38
	getStatus      = 0x58
	transmit       = 0x60
	receive        = 0x68
)

// TransmitTimeout represents the timeout for [SimpleNetwork.Transmit].
var TransmitTimeout = 10 * time.Millisecond

// SimpleNetwork represents an EFI Simple Network Protocol instance.
type SimpleNetwork struct {
	base uint64
}

// Start calls EFI_SIMPLE_NETWORK.Start().
func (sn *SimpleNetwork) Start() (err error) {
	status := callService(sn.base+start,
		[]uint64{
			sn.base,
		},
	)

	if status&0xff == EFI_ALREADY_STARTED {
		return nil
	}

	return parseStatus(status)
}

// Stop calls EFI_SIMPLE_NETWORK.Stop().
func (sn *SimpleNetwork) Stop() (err error) {
	status := callService(sn.base+stop,
		[]uint64{
			sn.base,
		},
	)

	return parseStatus(status)
}

// Initialize calls EFI_SIMPLE_NETWORK.Initialize().
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

// Shutdown calls EFI_SIMPLE_NETWORK.Shutdown().
func (sn *SimpleNetwork) Shutdown() (err error) {
	status := callService(sn.base+shutdown,
		[]uint64{
			sn.base,
		},
	)

	return parseStatus(status)
}

// ReceiveFilters calls EFI_SIMPLE_NETWORK.ReceiveFilters().
func (sn *SimpleNetwork) ReceiveFilters(enableMask uint32, disableMask uint32) (err error) {
	status := callService(sn.base+receiveFilters,
		[]uint64{
			sn.base,
			uint64(enableMask),
			0,
			0,
			0,
			0,
		},
	)

	return parseStatus(status)
}

// StationAddress calls EFI_SIMPLE_NETWORK.StationAddress().
func (sn *SimpleNetwork) StationAddress(reset bool, mac []byte) (err error) {
	var r uint64
	var m uint64

	if reset {
		r = 1
	}

	if n := len(mac); n > 0 && n <= 32 {
		mac = append(mac, make([]byte, 32-n)...)
		m = ptrval(&mac[0])
	}

	status := callService(sn.base+receiveFilters,
		[]uint64{
			sn.base,
			r,
			m,
		},
	)

	return parseStatus(status)
}

// GetStatus calls EFI_SIMPLE_NETWORK.GetStatus().
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
// EFI_SIMPLE_NETWORK.GetStatus() to report a transmit interrupt within
// [TransmitTimeout] before returning.
func (sn *SimpleNetwork) Transmit(buf []byte) (err error) {
	var interruptStatus uint32

	start := time.Now()

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

		if interruptStatus&EFI_SIMPLE_NETWORK_TRANSMIT_INTERRUPT != 0 {
			break
		}

		if time.Since(start) > TransmitTimeout {
			return errors.New("timeout")
		}
	}

	return
}

// Receive calls EFI_SIMPLE_NETWORK.Receive().
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
