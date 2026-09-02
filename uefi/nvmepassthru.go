// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uefi

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// EFI_NVM_EXPRESS_PASS_THRU_PROTOCOL_GUID identifies the NVM Express Pass Thru
// Protocol, the carrier for TCG Opal IF-SEND/IF-RECV as the NVMe Security Send/Receive
// admin commands (unmediated by any firmware TCG stack).
var EFI_NVM_EXPRESS_PASS_THRU_PROTOCOL_GUID = MustParseGUID("52c78312-8edc-4233-98f2-1a1aa5e388a5")

// EFI_NVM_EXPRESS_PASS_THRU_PROTOCOL member offset (PassThru).
const nvmePassThru = 0x08

// NVMe Security Send/Receive admin opcodes.
const (
	nvmeSecuritySend    = 0x81
	nvmeSecurityReceive = 0x82
)

// nvmeAdminQueue is the command-packet QueueType for admin commands.
const nvmeAdminQueue = 0x00

// cdwFlagsCdw10Cdw11 marks Command Dwords 10 and 11 valid.
const cdwFlagsCdw10Cdw11 = 0x04 | 0x08

// NVMePassThru is a located EFI_NVM_EXPRESS_PASS_THRU_PROTOCOL instance.
type NVMePassThru struct {
	base     uint64
	passThru uint64
}

// LocateNVMePassThruHandles returns the handles exposing
// EFI_NVM_EXPRESS_PASS_THRU_PROTOCOL (one per NVMe controller).
func (s *BootServices) LocateNVMePassThruHandles() ([]uint64, error) {
	return s.LocateHandleBuffer(ByProtocol, EFI_NVM_EXPRESS_PASS_THRU_PROTOCOL_GUID)
}

// GetNVMePassThruByHandle resolves EFI_NVM_EXPRESS_PASS_THRU_PROTOCOL on a handle.
func (s *BootServices) GetNVMePassThruByHandle(handle uint64) (*NVMePassThru, error) {
	base, err := s.HandleProtocol(handle, EFI_NVM_EXPRESS_PASS_THRU_PROTOCOL_GUID)
	if err != nil {
		return nil, err
	}
	return newNVMePassThru(base)
}

// newNVMePassThru decodes the PassThru member slot, failing closed on a NULL instance.
func newNVMePassThru(base uint64) (*NVMePassThru, error) {
	var fn struct {
		Mode     uint64
		PassThru uint64
	}
	if err := decode(&fn, base); err != nil {
		return nil, err
	}
	if fn.PassThru == 0 {
		return nil, errors.New("nvme passthru: NULL PassThru pointer")
	}
	return &NVMePassThru{base: base, passThru: base + nvmePassThru}, nil
}

// nvmeCommand builds an EFI_NVM_EXPRESS_COMMAND for a Security Send/Receive: opcode in
// Cdw0, Cdw10 = (SECP<<24)|(SPSP<<8), Cdw11 = length. SPSP carries the ComID in NVMe's
// native byte order.
func nvmeCommand(opcode uint8, securityProtocol uint8, spSpecific uint16, length uint32) []byte {
	cmd := make([]byte, 44)
	binary.LittleEndian.PutUint32(cmd[0:], uint32(opcode))
	cmd[4] = cdwFlagsCdw10Cdw11
	binary.LittleEndian.PutUint32(cmd[20:], uint32(securityProtocol)<<24|uint32(spSpecific)<<8)
	binary.LittleEndian.PutUint32(cmd[24:], length)
	return cmd
}

// submit issues one admin-queue command packet. It fails closed on a pass-thru error
// and on a nonzero NVMe completion status.
func (p *NVMePassThru) submit(cmd, data, completion []byte) error {
	packet := make([]byte, 56)
	if len(data) > 0 {
		binary.LittleEndian.PutUint64(packet[8:], ptrval(&data[0]))
		binary.LittleEndian.PutUint32(packet[16:], uint32(len(data)))
	}
	packet[36] = nvmeAdminQueue
	binary.LittleEndian.PutUint64(packet[40:], ptrval(&cmd[0]))
	binary.LittleEndian.PutUint64(packet[48:], ptrval(&completion[0]))

	status := callService(p.passThru, []uint64{
		p.base,
		0,
		ptrval(&packet[0]),
		0,
	})
	if err := parseStatus(status); err != nil {
		return err
	}

	// EFI_NVM_EXPRESS_COMPLETION.DW3 holds the completion-queue Status Field in bits
	// 31:17 (SC and SCT); a nonzero value means the controller rejected the command.
	if sf := binary.LittleEndian.Uint32(completion[12:]) >> 17; sf != 0 {
		return fmt.Errorf("nvme command failed (status %#x)", sf)
	}
	return nil
}

// SecuritySend issues an NVMe Security Send (IF-SEND).
func (p *NVMePassThru) SecuritySend(securityProtocol uint8, comID uint16, payload []byte) error {
	cmd := nvmeCommand(nvmeSecuritySend, securityProtocol, comID, uint32(len(payload)))
	completion := make([]byte, 16)
	return p.submit(cmd, payload, completion)
}

// SecurityReceive issues an NVMe Security Receive (IF-RECV) and returns size bytes.
func (p *NVMePassThru) SecurityReceive(securityProtocol uint8, comID uint16, size int) ([]byte, error) {
	cmd := nvmeCommand(nvmeSecurityReceive, securityProtocol, comID, uint32(size))
	completion := make([]byte, 16)
	buf := make([]byte, size)
	if err := p.submit(cmd, buf, completion); err != nil {
		return nil, err
	}
	return buf, nil
}

// NVMe Identify Controller (opcode 0x06, CNS=0x01); the Serial Number is 20 bytes at
// offset 4 of the 4096-byte result.
const (
	nvmeIdentify    = 0x06
	cnsIdentifyCtrl = 0x01
	cdwFlagsCdw10   = 0x04
	nvmeIdentifyLen = 4096
	nvmeSerialOff   = 4
	nvmeSerialLen   = 20
)

// IdentifyController issues NVMe Identify Controller and returns the raw 4096-byte data.
func (p *NVMePassThru) IdentifyController() ([]byte, error) {
	cmd := make([]byte, 44)
	binary.LittleEndian.PutUint32(cmd[0:], nvmeIdentify)
	cmd[4] = cdwFlagsCdw10
	binary.LittleEndian.PutUint32(cmd[20:], cnsIdentifyCtrl)
	data := make([]byte, nvmeIdentifyLen)
	completion := make([]byte, 16)
	if err := p.submit(cmd, data, completion); err != nil {
		return nil, err
	}
	return data, nil
}

// SerialNumber returns the controller's 20-byte Serial Number, as-is (space-padded
// ASCII, not trimmed) so a caller can reproduce the drive's exact bytes.
func (p *NVMePassThru) SerialNumber() ([]byte, error) {
	id, err := p.IdentifyController()
	if err != nil {
		return nil, err
	}
	sn := make([]byte, nvmeSerialLen)
	copy(sn, id[nvmeSerialOff:nvmeSerialOff+nvmeSerialLen])
	return sn, nil
}
