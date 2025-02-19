// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package efi

// EFI Boot Services offset
const allocatePages = 0x28

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
		uint64(allocateType),
		uint64(memoryType),
		uint64(size)/4096,
		&physicalAddress,
	)

	return parseStatus(status)
}
