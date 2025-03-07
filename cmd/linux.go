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
var CommandLine = "earlyprintk=ttyS0,115200,8n1 console=ttyS0,115200,8n1 debug\x00"

// remove trailing space below to embed
//
//go:embed bzImage
var bzImage []byte

// remove trailing space below to embed
//
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

func reserveMemory(image *exec.LinuxImage, size int) (err error) {
	mmap, _, err := bootServices.GetMemoryMap()

	if err != nil {
		return
	}

	// find unallocated UEFI memory for kernel and ramdisk loading
	for _, e := range mmap {
		if e.Type != efi.EfiConventionalMemory ||
			e.Size() < size {
			continue
		}

		// opportunistic size increase
		size = e.Size()

		// reserve unallocated UEFI memory for our runtime DMA
		if image.Region, err = dma.NewRegion(uint(e.PhysicalStart), size, false); err != nil {
			return
		}

		image.Region.Reserve(size, 0)
		break

	}

	if image.Region == nil {
		return errors.New("could not find memory for kernel loading")
	}

	// build E820 memory map
	for _, desc := range mmap {
		e, err := desc.E820()

		if err != nil {
			return err
		}

		image.Memory = append(image.Memory, e)
	}

	// enforce required kernel alignment on kernel and ramdisk offsets
	align := int(image.BzImage.Header.Kernelalignment)
	base := int(image.Region.Start())

	image.InitialRamDiskOffset = 0
	image.InitialRamDiskOffset += -(base + image.InitialRamDiskOffset) & (align - 1)

	image.KernelOffset = image.InitialRamDiskOffset + len(initrd)
	image.KernelOffset += -(base + image.KernelOffset) & (align - 1)

	// place boot parameters at the far end
	image.CmdLineOffset = size - int(image.BzImage.Header.CmdLineSize)
	image.ParamsOffset = size - image.CmdLineOffset - paramsSize

	return
}

func cleanup() {
	log.Printf("exiting EFI boot services")

	if err := bootServices.Exit(); err != nil {
		log.Printf("could not exit EFI boot services, %v\n", err)
	}

	bootServices = nil
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

	image := &exec.LinuxImage{
		Kernel:         bzImage,
		InitialRamDisk: initrd,
		CmdLine:        CommandLine,
	}

	if err = image.Parse(); err != nil {
		return
	}

	// reserve runtime memory for kernel loading
	if err = reserveMemory(image, memorySize); err != nil {
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

	// load kernel in reserved memory
	if err = image.Load(); err != nil {
		return "", fmt.Errorf("could not load kernel, %v", err)
	}

	log.Printf("jumping to kernel entry %#08x", image.Entry())
	err = image.Boot(cleanup)

	return
}
