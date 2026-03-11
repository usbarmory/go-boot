// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uefi

var EFI_GRAPHICS_OUTPUT_PROTOCOL_GUID = MustParseGUID("9042a9de-23dc-4a38-96fb-7aded080516a")

// EFI Graphics Output Protocol offsets
const (
	blt = 0x10
)

type BltOperation int

// EFI_GRAPHICS_OUTPUT_BLT_OPERATION
const (
	EfiBltVideoFill = iota
	EfiBltVideoToBltBuffer
	EfiBltBufferToVideo
	EfiBltVideoToVideo
	EfiGraphicsOutputBltOperationMax
)

// ModeInformation represents an EFI Graphics Output Mode Information instance.
type ModeInformation struct {
	Version              uint32
	HorizontalResolution uint32
	VerticalResolution   uint32
	PixelFormat          uint32
	RedMask              uint32
	GreenMask            uint32
	BlueMask             uint32
	ReservedMask         uint32
	PixelsPerScanLine    uint32
}

// ProtocolMode represents an EFI Graphics Output Protocol Mode instance.
type ProtocolMode struct {
	MaxMode         uint32
	Mode            uint32
	Info            uint64
	SizeOfInfo      uint64
	FrameBufferBase uint64
	FrameBufferSize uint64
}

// GetInfo returns the EFI Graphics Output Mode information instance.
func (d *ProtocolMode) GetInfo() (m *ModeInformation, err error) {
	m = &ModeInformation{}
	err = decode(m, d.Info)
	return
}

// GraphicsOutput represents an EFI Graphics Output Protocol instance.
type GraphicsOutput struct {
	base uint64
	mode uint64
}

// GetMode returns the EFI Graphics Output Mode instance.
func (gop *GraphicsOutput) GetMode() (pm *ProtocolMode, err error) {
	pm = &ProtocolMode{}
	err = decode(pm, gop.mode)
	return
}

// Blt calls EFI_GRAPHICS_OUTPUT_PROTCOL.Blt().
func (gop *GraphicsOutput) Blt(buf []byte, op BltOperation, srcX, srcY, dstX, dstY, width, height, delta uint64) (err error) {
	if gop.base == 0 {
		return nil
	}

	status := callService(gop.base+blt,
		[]uint64{
			gop.base,
			ptrval(&buf[0]),
			uint64(op),
			srcX,
			srcY,
			dstX,
			dstY,
			width,
			height,
			delta,
		},
	)

	return parseStatus(status)
}

// GetGraphicsOutput locates and returns the EFI Graphics Output Protocol
// instance.
func (s *BootServices) GetGraphicsOutput() (gop *GraphicsOutput, err error) {
	gop = &GraphicsOutput{}

	var data struct {
		QueryMode uint64
		SetMode   uint64
		Blt       uint64
		Mode      uint64
	}

	if gop.base, err = s.LocateProtocol(EFI_GRAPHICS_OUTPUT_PROTOCOL_GUID); err != nil {
		return
	}

	if err = decode(&data, gop.base); err != nil {
		return
	}

	gop.mode = data.Mode

	return
}
