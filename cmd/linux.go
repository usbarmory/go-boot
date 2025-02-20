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
	"golang.org/x/term"

	"github.com/usbarmory/armory-boot/exec"
	"github.com/usbarmory/go-boot/efi"
	"github.com/usbarmory/tamago/dma"
)

const (
	memoryStart = 0x80000000
	memorySize  = 0x10000000

	commandLine = "console=ttyS0,115200,8n1 mem=4G\x00"
)

// remove trailing space below to embed
//
// go:embed bzImage
var bzImage []byte

var memoryMap = []bzimage.E820Entry{
	// should always be usable (?)
	bzimage.E820Entry{
		Addr:    uint64(0x00000000),
		Size:    uint64(0x0009f000),
		MemType: bzimage.RAM,
	},
	// amd64.ramStart, microvm.ramSize
	bzimage.E820Entry{
		Addr:    efi.RamStart,
		Size:    efi.RamSize,
		MemType: bzimage.RAM,
	},
	bzimage.E820Entry{
		Addr:    memoryStart,
		Size:    memorySize,
		MemType: bzimage.RAM,
	},
}

func init() {
	Add(Cmd{
		Name:    "linux",
		Args:    1,
		Pattern: regexp.MustCompile(`^linux(.*)`),
		Syntax:  "(path)?",
		Help:    "boot Linux kernel bzImage",
		Fn:      linuxCmd,
	})
}

func linuxCmd(_ *Interface, term *term.Terminal, arg []string) (res string, err error) {
	var mem *dma.Region

	path := strings.TrimSpace(arg[0])

	if len(path) != 0 {
		if bzImage, err = os.ReadFile(path); err != nil {
			return
		}
	}

	if bootServices == nil {
		return "", errors.New("EFI Boot Services unavailable")
	}

	if err = bootServices.AllocatePages(
		efi.AllocateAddress,
		efi.EfiLoaderData,
		memorySize,
		memoryStart,
	); err != nil {
		return
	}

	log.Printf("allocated memory range %#08x - %#08x", memoryStart, memoryStart+memorySize)
	log.Printf("exiting EFI boot services")

	if err = bootServices.Exit(); err != nil {
		return "", fmt.Errorf("could not exit EFI boot services, %v\n", err)
	}

	if mem, err = dma.NewRegion(memoryStart, memorySize, false); err != nil {
		return
	}

	addr, _ := mem.Reserve(memorySize, 0)
	defer mem.Release(addr)

	image := &exec.LinuxImage{
		Memory:  memoryMap,
		Region:  mem,
		Kernel:  bzImage,
		CmdLine: commandLine,
	}

	if err = image.Load(); err != nil {
		return "", fmt.Errorf("could not load kernel, %v", err)
	}

	return "", image.Boot(nil)
}
