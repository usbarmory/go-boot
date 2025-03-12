// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package efi

const EFI_GRAPHICS_OUTPUT_PROTOCOL_GUID = "9042a9de-23dc-4a38-96fb-7aded080516a"

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

// GetInfo() returns the EFI Graphics Output Mode information instance.
func (d *ProtocolMode) GetInfo() (mi *ModeInformation, err error) {
	mi = &ModeInformation{}
	err = decode(mi, d.Info)
	return
}

// GraphicsOutput represents an EFI Graphics Output Protocol instance.
type GraphicsOutput struct {
	QueryMode uint64
	SetMode   uint64
	Blt       uint64
	Mode      uint64
}

// GetMode returns the EFI Graphics Output Mode instance.
func (gop *GraphicsOutput) GetMode() (pm *ProtocolMode, err error) {
	pm = &ProtocolMode{}
	err = decode(pm, gop.Mode)
	return
}

// GetGraphicsOutput locates and returns the EFI Graphics Output Protocol
// instance.
func (s *BootServices) GetGraphicsOutput() (gop *GraphicsOutput, err error) {
	gop = &GraphicsOutput{}

	base, err := s.LocateProtocolString(EFI_GRAPHICS_OUTPUT_PROTOCOL_GUID)

	if err != nil {
		return
	}

	err = decode(gop, base)

	return
}
