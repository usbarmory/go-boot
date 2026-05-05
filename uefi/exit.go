// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uefi

// EFI Boot Services offsets
const (
	exit             = 0xd8
	exitBootServices = 0xe8
)

// Exit calls EFI_BOOT_SERVICES.Exit().
func (s *BootServices) Exit(code int) (err error) {
	status := callService(s.base+exit,
		[]uint64{
			uint64(s.imageHandle),
			uint64(code),
			0,
			0,
		},
	)

	return parseStatus(status)
}

// ExitServices calls EFI_BOOT_SERVICES.ExitBootServices(), it is the caller
// responsability to avoid using any EFI Boot Service after this call is
// successful.
//
// Typically the following actions are required after exiting boot services:
//   - silencing active UEFI consoles by setting [Console.Out] to 0
//   - replacing runtime stdout by setting [x64.Stdout]
//   - trap CPU exceptions with [x64.AMD64.EnableExceptions]
//   - initializing APs as needed with [x64.AMD64.InitSMP]
func (s *BootServices) ExitBootServices() (memoryMap *MemoryMap, err error) {
	if memoryMap, err = s.GetMemoryMap(); err != nil {
		return
	}

	status := callService(s.base+exitBootServices,
		[]uint64{
			uint64(s.imageHandle),
			memoryMap.MapKey,
		},
	)

	return memoryMap, parseStatus(status)
}
