// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build amd64

package cmd

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/usbarmory/armory-boot/exec"
	"github.com/usbarmory/go-boot/efi"
	"github.com/usbarmory/go-boot/shell"
	"github.com/usbarmory/tamago/dma"
)

const (
	// require at least 256MB for kernel and ramdisk loading
	memorySize = 0x10000000
	paramsSize = 0x1000
)

// CommandLine represents the Linux kernel boot parameters
var CommandLine = "earlyprintk=ttyS0,115200,8n1,keep debug\x00"

//go:embed bzImage
var bzImage []byte

//go:embed initrd
var initrd []byte

func init() {
	shell.Add(shell.Cmd{
		Name:    "linux",
		Args:    1,
		Pattern: regexp.MustCompile(`^linux(.*)`),
		Syntax:  "(path)?",
		Help:    "boot Linux kernel bzImage",
		Fn:      linuxCmd,
	})
}

func reserveMemory(memdesc []*efi.MemoryDescriptor, image *exec.LinuxImage, size int) (err error) {
	// find unallocated UEFI memory for kernel and ramdisk loading
	for _, desc := range memdesc {
		if desc.Type != efi.EfiConventionalMemory ||
			desc.Size() < size {
			continue
		}

		// opportunistic size increase
		size = desc.Size()

		// reserve unallocated UEFI memory for our runtime DMA
		if image.Region, err = dma.NewRegion(uint(desc.PhysicalStart), size, false); err != nil {
			return
		}

		image.Region.Reserve(size, 0)
		break

	}

	if image.Region == nil {
		return errors.New("could not find memory for kernel loading")
	}

	// build E820 memory map
	for _, desc := range memdesc {
		e820, err := desc.E820()

		if err != nil {
			return err
		}

		image.Memory = append(image.Memory, e820)
	}

	// enforce required alignment on kernel and ramdisk offsets
	align := int(image.BzImage.Header.Kernelalignment)
	base := int(image.Region.Start())

	image.InitialRamDiskOffset = 0
	image.InitialRamDiskOffset += -(base + image.InitialRamDiskOffset) & (align - 1)

	image.KernelOffset = image.InitialRamDiskOffset + len(initrd)
	image.KernelOffset += -(base + image.KernelOffset) & (align - 1)

	// place boot parameters at the far end
	image.CmdLineOffset = size - int(image.BzImage.Header.CmdLineSize)
	image.ParamsOffset = image.CmdLineOffset - paramsSize

	return
}

func cleanup() {
	log.Printf("exiting EFI boot services")

	if err := bootServices.Exit(); err != nil {
		log.Printf("could not exit EFI boot services, %v\n", err)
	}

	bootServices = nil
}

func efiInfo(memoryMap *efi.MemoryMap) (efi *exec.EFI, err error) {
	return &exec.EFI{
		LoaderSignature:   exec.EFI64LoaderSignature,
		SystemTable:       uint32(systemTable.Address()),
		SystemTableHigh:   uint32(systemTable.Address() >> 32),
		MemoryMapHigh:     uint32(memoryMap.Address() >> 32),
		MemoryMapSize:     uint32(memoryMap.MapSize),
		MemoryMap:         uint32(memoryMap.Address()),
		MemoryDescSize:    uint32(memoryMap.DescriptorSize),
		MemoryDescVersion: 1, // Linux only accepts this value
	}, nil
}

func screenInfo() (screen *exec.Screen, err error) {
	var gop *efi.GraphicsOutput
	var mode *efi.ProtocolMode
	var info *efi.ModeInformation

	if gop, err = bootServices.GetGraphicsOutput(); err != nil {
		return
	}

	if mode, err = gop.GetMode(); err != nil {
		return
	}

	if info, err = mode.GetInfo(); err != nil {
		return
	}

	return &exec.Screen{
		OrigVideoIsVGA: 0x70,
		Lfbwidth:       uint16(info.HorizontalResolution),
		Lfbheight:      uint16(info.VerticalResolution),
		Lfbbase:        uint32(mode.FrameBufferBase),
		Lfbsize:        uint32(mode.FrameBufferSize),
		Lfblinelength:  uint16(info.HorizontalResolution * 4),
	}, nil
}

func linuxCmd(_ *shell.Interface, arg []string) (res string, err error) {
	path := strings.TrimSpace(arg[0])

	if len(path) != 0 {
		if bzImage, err = os.ReadFile(path); err != nil {
			return
		}
	} else if len(bzImage) == 0 {
		return "", errors.New("bzImage not embedded")
	}

	if bootServices == nil {
		return "", errors.New("EFI Boot Services unavailable")
	}

	memoryMap, err := bootServices.GetMemoryMap()

	if err != nil {
		return
	}

	image := &exec.LinuxImage{
		Kernel:         bzImage,
		InitialRamDisk: initrd,
		CmdLine:        CommandLine,
	}

	if err = image.Parse(); err != nil {
		return
	}

	// reserve runtime memory for kernel loading
	if err = reserveMemory(memoryMap.Descriptors, image, memorySize); err != nil {
		return
	}

	// release in case of error
	defer image.Region.Release(image.Region.Start())

	start := uint64(image.Region.Start())
	size := int(image.Region.Size())

	log.Printf("allocating memory pages %#08x - %#08x", start, int(start)+size)

	// reserve UEFI memory for kernel loading
	if err = bootServices.AllocatePages(
		efi.AllocateAddress,
		efi.EfiLoaderData,
		size,
		start,
	); err != nil {
		return
	}

	// free in case of error
	defer bootServices.FreePages(
		start,
		size,
	)

	// fill EFI information in boot parameters
	if image.EFI, err = efiInfo(memoryMap); err != nil {
		return
	}

	// fill screen_info
	if image.Screen, err = screenInfo(); err != nil {
		return
	}

	// load kernel in reserved memory
	if err = image.Load(); err != nil {
		return "", fmt.Errorf("could not load kernel, %v", err)
	}

	log.Printf("jumping to kernel entry %#08x", image.Entry())
	err = image.Boot(cleanup)

	return
}
