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

// CommandLine represents the Linux kernel boot parameters
var CommandLine = "console=ttyS0,115200,8n1\x00"

// remove trailing space below to embed
//
//go:embed bzImage
var bzImage []byte

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

func reserveMemory(image *exec.LinuxImage) (err error) {
	mmap, _, err := bootServices.GetMemoryMap()

	if err != nil {
		return
	}

	start, end, err := image.Parse()

	if err != nil {
		return
	}

	log.Printf("loading kernel sections %#08x - %#08x", start, end)

	size := int(end - start)

	// ensure kernel sections can be allocated outside EFI
	for _, e := range mmap {
		if e.Type != efi.EfiConventionalMemory ||
			size > e.Size() ||
			start < e.PhysicalStart ||
			end > e.PhysicalEnd() {
			continue
		}

		log.Printf("found UEFI memory range %#08x - %#08x", e.PhysicalStart, e.PhysicalEnd())

		if image.Region, err = dma.NewRegion(uint(start), size, false); err != nil {
			return
		}

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

	image.Region.Reserve(size, 0)

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
		Kernel:  bzImage,
		CmdLine: CommandLine,
	}

	// reserve runtime memory for kernel loading
	if err = reserveMemory(image); err != nil {
		return
	}

	// release in case of error
	defer image.Region.Release(image.Region.Start())

	start := uint64(image.Region.Start())
	size := int(image.Region.Size())

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
