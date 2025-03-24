// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uefi

// EFI Boot Service offsets
const (
	allocatePages = 0x28
	freePages     = 0x30
)

// EFI_ALLOCATE_TYPE
const (
	AllocateAnyPages = iota
	AllocateMaxAddress
	AllocateAddress
	MaxAllocateType
)

// EFI_MEMORY_TYPE
const (
	EfiReservedMemoryType = iota
	EfiLoaderCode
	EfiLoaderData
	EfiBootServicesCode
	EfiBootServicesData
	EfiRuntimeServicesCode
	EfiRuntimeServicesData
	EfiConventionalMemory
	EfiUnusableMemory
	EfiACPIReclaimMemory
	EfiACPIMemoryNVS
	EfiMemoryMappedIO
	EfiMemoryMappedIOPortSpace
	EfiPalCode
	EfiPersistentMemory
	EfiUnacceptedMemoryType
	EfiMaxMemoryType
)

// AllocatePages calls EFI_BOOT_SERVICES.AllocatePages().
func (s *BootServices) AllocatePages(allocateType int, memoryType int, size int, physicalAddress uint64) error {
	status := callService(s.base+allocatePages,
		[]uint64{
			uint64(allocateType),
			uint64(memoryType),
			uint64(size) / PageSize,
			ptrval(&physicalAddress),
		},
	)

	return parseStatus(status)
}

// FreePages calls EFI_BOOT_SERVICES.FreePages().
func (s *BootServices) FreePages(physicalAddress uint64, size int) error {
	status := callService(s.base+freePages,
		[]uint64{
			physicalAddress,
			uint64(size) / PageSize,
		},
	)

	return parseStatus(status)
}
