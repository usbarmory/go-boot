// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uefi

import "errors"

// ErrTruncated is returned when a ReceiveData transfer exceeds the supplied buffer.
var ErrTruncated = errors.New("storage security: transfer size exceeds buffer")

// EFI_STORAGE_SECURITY_COMMAND_PROTOCOL_GUID identifies the Storage Security Command
// Protocol, the carrier for TCG Opal IF-SEND/IF-RECV.
var EFI_STORAGE_SECURITY_COMMAND_PROTOCOL_GUID = MustParseGUID("c88b0b6d-0dfc-49a7-9cb4-49074b4c3a78")

// EFI_STORAGE_SECURITY_COMMAND_PROTOCOL member offsets.
const (
	receiveData = 0x00
	sendData    = 0x08
)

// StorageSecurity is a located EFI_STORAGE_SECURITY_COMMAND_PROTOCOL instance.
type StorageSecurity struct {
	base uint64
	recv uint64
	send uint64
}

// GetStorageSecurity locates the first EFI_STORAGE_SECURITY_COMMAND_PROTOCOL instance.
// It uses LocateProtocol and yields no handle; use LocateStorageSecurityHandles with
// GetStorageSecurityByHandle when the handle's Block I/O MediaId is also needed.
func (s *BootServices) GetStorageSecurity() (ssc *StorageSecurity, err error) {
	base, err := s.LocateProtocol(EFI_STORAGE_SECURITY_COMMAND_PROTOCOL_GUID)
	if err != nil {
		return nil, err
	}
	return newStorageSecurity(base)
}

// LocateStorageSecurityHandles returns the handles exposing
// EFI_STORAGE_SECURITY_COMMAND_PROTOCOL.
func (s *BootServices) LocateStorageSecurityHandles() ([]uint64, error) {
	return s.LocateHandleBuffer(ByProtocol, EFI_STORAGE_SECURITY_COMMAND_PROTOCOL_GUID)
}

// GetStorageSecurityByHandle resolves EFI_STORAGE_SECURITY_COMMAND_PROTOCOL on a handle.
func (s *BootServices) GetStorageSecurityByHandle(handle uint64) (*StorageSecurity, error) {
	base, err := s.HandleProtocol(handle, EFI_STORAGE_SECURITY_COMMAND_PROTOCOL_GUID)
	if err != nil {
		return nil, err
	}
	return newStorageSecurity(base)
}

// newStorageSecurity decodes the member slots, failing closed on a NULL instance.
func newStorageSecurity(base uint64) (*StorageSecurity, error) {
	var fn struct {
		ReceiveData uint64
		SendData    uint64
	}
	if err := decode(&fn, base); err != nil {
		return nil, err
	}
	if fn.ReceiveData == 0 || fn.SendData == 0 {
		return nil, errors.New("storage security: NULL ReceiveData/SendData pointer")
	}
	return &StorageSecurity{base: base, recv: base + receiveData, send: base + sendData}, nil
}

// SendData calls EFI_STORAGE_SECURITY_COMMAND_PROTOCOL.SendData(). spSpecific carries
// the TCG ComID; timeout is in 100 ns units (0 = none).
func (p *StorageSecurity) SendData(mediaID uint32, timeout uint64, securityProtocol uint8, spSpecific uint16, payload []byte) error {
	var ptr uint64
	if len(payload) > 0 {
		ptr = ptrval(&payload[0])
	}
	status := callService(p.send, []uint64{
		p.base,
		uint64(mediaID),
		timeout,
		uint64(securityProtocol),
		uint64(spSpecific),
		uint64(len(payload)),
		ptr,
	})
	return parseStatus(status)
}

// ReceiveData calls EFI_STORAGE_SECURITY_COMMAND_PROTOCOL.ReceiveData(), returning the
// valid prefix (PayloadTransferSize) of the transfer.
func (p *StorageSecurity) ReceiveData(mediaID uint32, timeout uint64, securityProtocol uint8, spSpecific uint16, size int) ([]byte, error) {
	buf := make([]byte, size)
	var xfer uintptr
	var ptr uint64
	if size > 0 {
		ptr = ptrval(&buf[0])
	}
	status := callService(p.recv, []uint64{
		p.base,
		uint64(mediaID),
		timeout,
		uint64(securityProtocol),
		uint64(spSpecific),
		uint64(size),
		ptr,
		ptrval(&xfer),
	})
	if err := parseStatus(status); err != nil {
		return nil, err
	}
	if int(xfer) > size {
		return nil, ErrTruncated
	}
	return buf[:xfer], nil
}
