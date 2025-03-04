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

	"github.com/u-root/u-root/pkg/boot/bzimage"

	"github.com/usbarmory/armory-boot/exec"
	"github.com/usbarmory/go-boot/efi"
	"github.com/usbarmory/go-boot/shell"
	"github.com/usbarmory/tamago/dma"
)

// TODO: calculate from exec.LinuxImage.Region()
const (
	memoryStart = 0x80000000
	memorySize  = 0x10000000
)

// CommandLine represents the Linux kernel boot parameters
var CommandLine = "console=ttyS0,115200,8n1\x00"

// remove trailing space below to embed
//
// go:embed bzImage
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

func buildMemoryMap() (m []bzimage.E820Entry, err error) {
	var mmap []*efi.MemoryMap

	if mmap, _, err = bootServices.GetMemoryMap(); err != nil {
		return
	}

	for _, desc := range mmap {
		e, err := desc.E820()

		if err != nil {
			return nil, err
		}

		m = append(m, e)
	}

	return
}

func findMemory(m []bzimage.E820Entry, start int, size int) (mem *dma.Region, err error) {
	for _, e := range m {
		if e.MemType != bzimage.RAM || e.Size < uint64(size) {
			continue
		}

		if uint64(start) < e.Addr || uint64(start) >= e.Addr+e.Size {
			continue
		}

		if mem, err = dma.NewRegion(uint(start), size, false); err != nil {
			return
		}

		log.Printf("allocating memory range %#08x - %#08x", start, start+size)
		mem.Reserve(size, 0)

		break
	}

	if mem == nil {
		err = errors.New("could not find memory for kernel loading")
	}

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
	var mem *dma.Region
	var mmap []bzimage.E820Entry

	path := strings.TrimSpace(arg[0])

	if len(path) != 0 {
		if bzImage, err = os.ReadFile(path); err != nil {
			return
		}
	}

	if bootServices == nil {
		return "", errors.New("EFI Boot Services unavailable")
	}

	// build E820 memory map

	if mmap, err = buildMemoryMap(); err != nil {
		return
	}

	// find and reserve memory for kernel loading

	if mem, err = findMemory(mmap, memoryStart, memorySize); err != nil {
		return
	}

	// free reserved memory in case of error
	defer mem.Release(mem.Start())

	if err = bootServices.AllocatePages(
		efi.AllocateAddress,
		efi.EfiLoaderData,
		int(mem.Size()),
		uint64(mem.Start()),
	); err != nil {
		return
	}

	// free allocated pages in case of error
	defer bootServices.FreePages(
		uint64(mem.Start()),
		int(mem.Size()),
	)

	image := &exec.LinuxImage{
		Memory:  mmap,
		Region:  mem,
		Kernel:  bzImage,
		CmdLine: CommandLine,
	}

	// load kernel

	log.Printf("loading kernel@%0.8x", mem.Start())

	if err = image.Load(); err != nil {
		return "", fmt.Errorf("could not load kernel, %v", err)
	}

	// boot kernel

	log.Printf("starting kernel@%0.8x", image.Entry())

	// does not return on success
	return "", image.Boot(cleanup)
}
