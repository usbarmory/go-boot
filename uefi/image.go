// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uefi

// EFI Boot Services offsets
const (
	loadImage  = 0xc8
	startImage = 0xd0
)

// LoadImage calls EFI_BOOT_SERVICES.LoadImage().
func (s *BootServices) LoadImage(boot int, devicePath *FilePath) (imageHandle uint64, err error) {
	filePath := devicePath.Bytes()

	status := callService(s.base+loadImage, 6,
		[]uint64{
			uint64(boot),
			s.imageHandle,
			ptrval(&filePath[0]),
			0,
			0,
			ptrval(&imageHandle),
		},
	)

	return imageHandle, parseStatus(status)
}

// StartImage calls EFI_BOOT_SERVICES.StartImage().
func (s *BootServices) StartImage(imageHandle uint64) (err error) {
	status := callService(s.base+startImage, 3,
		[]uint64{
			imageHandle,
			0,
			0,
		},
	)

	return parseStatus(status)
}
