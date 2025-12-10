// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"errors"
	"fmt"
	"log"
	"regexp"

	"github.com/u-root/u-root/pkg/boot/bzimage"

	"github.com/usbarmory/armory-boot/exec"
	"github.com/usbarmory/go-boot/shell"
	"github.com/usbarmory/go-boot/uapi"
	"github.com/usbarmory/go-boot/uefi"
	"github.com/usbarmory/go-boot/uefi/x64"
	"github.com/usbarmory/go-boot/transparency"
	"github.com/usbarmory/tamago/dma"
)

const (
	// avoid initial DMA region
	minLoadAddr = 0x01000000
	paramsSize  = 0x1000
	exitRetries = 3
)

// DefaultLinuxEntry represents the default path for the UAPI Type #1 Boot
// Loader Entry (`linux,l,\\r` command).
var DefaultLinuxEntry string

func init() {
	shell.Add(shell.Cmd{
		Name:    "linux,l",
		Args:    1,
		Pattern: regexp.MustCompile(`^(?:linux|l)(?: (\S+))?$`),
		Syntax:  "(loader entry path)?",
		Help:    "boot Linux kernel image",
		Fn:      linuxCmd,
	})

	if len(DefaultLinuxEntry) > 0 {
		shell.Add(shell.Cmd{
			Name:    "linux,l,\\r",
			Args:    1,
			Pattern: regexp.MustCompile(`^(?:linux|l|)(?: (\S+))?$`),
			Help:    fmt.Sprintf("`l %s`", DefaultLinuxEntry),
			Fn:      linuxCmd,
		})
	}
}

func reserveMemory(m *uefi.MemoryMap, image *exec.LinuxImage) (err error) {
	size := len(image.BzImage.KernelCode) + len(image.InitialRamDisk)

	// Convert UEFI Memory Map to E820 as it reflects availability after
	// exiting EFI Boot Services.
	image.Memory = m.E820()

	// find unallocated UEFI memory for kernel and ramdisk loading
	for _, entry := range image.Memory {
		if entry.MemType != bzimage.RAM ||
			int(entry.Size) < size {
			continue
		}

		// shift above minLoadAddr as required and recheck size
		if entry.Addr < minLoadAddr {
			off := minLoadAddr - entry.Addr
			entry.Addr += off
			entry.Size -= off

			if int(entry.Size) < size {
				continue
			}
		}

		// opportunistic size increase
		size = int(entry.Size)

		// reserve unallocated UEFI memory for our runtime DMA
		if image.Region, err = dma.NewRegion(uint(entry.Addr), size, false); err != nil {
			// skip our own runtime pages
			continue
		}

		log.Printf("reserving memory %#x - %#x", entry.Addr, entry.Addr+entry.Size)
		image.Region.Reserve(size, 0)

		break
	}

	if image.Region == nil {
		return errors.New("could not find memory for kernel loading")
	}

	// enforce required alignment on kernel and ramdisk offsets
	align := int(image.BzImage.Header.Kernelalignment)
	base := int(image.Region.Start())

	image.InitialRamDiskOffset = 0
	image.InitialRamDiskOffset += -(base + image.InitialRamDiskOffset) & (align - 1)

	image.KernelOffset = image.InitialRamDiskOffset + len(image.InitialRamDisk)
	image.KernelOffset += -(base + image.KernelOffset) & (align - 1)

	// place boot parameters at the far end
	image.CmdLineOffset = size - int(image.BzImage.Header.CmdLineSize)
	image.ParamsOffset = image.CmdLineOffset - paramsSize

	return
}

func efiInfo(memoryMap *uefi.MemoryMap) (efi *exec.EFI, err error) {
	return &exec.EFI{
		LoaderSignature:   exec.EFI64LoaderSignature,
		SystemTable:       uint32(x64.UEFI.Address()),
		SystemTableHigh:   uint32(x64.UEFI.Address() >> 32),
		MemoryMapHigh:     uint32(memoryMap.Address() >> 32),
		MemoryMapSize:     uint32(memoryMap.MapSize),
		MemoryMap:         uint32(memoryMap.Address()),
		MemoryDescSize:    uint32(memoryMap.DescriptorSize),
		MemoryDescVersion: 1, // Linux only accepts this value
	}, nil
}

func screenInfo() (screen *exec.Screen, err error) {
	var gop *uefi.GraphicsOutput
	var mode *uefi.ProtocolMode
	var info *uefi.ModeInformation

	if gop, err = x64.UEFI.Boot.GetGraphicsOutput(); err != nil {
		return
	}

	if mode, err = gop.GetMode(); err != nil {
		return
	}

	if info, err = mode.GetInfo(); err != nil {
		return
	}

	// values for efib selection
	screen = &exec.Screen{
		OrigVideoIsVGA: exec.VideoTypeEFI,
		LfbWidth:       uint16(info.HorizontalResolution),
		LfbHeight:      uint16(info.VerticalResolution),
		LfbBase:        uint32(mode.FrameBufferBase),
		LfbSize:        uint32(mode.FrameBufferSize),
		LfbLineLength:  uint16(info.HorizontalResolution * 4),
		ExtLfbBase:     uint32(mode.FrameBufferBase >> 32),
	}

	if screen.ExtLfbBase > 0 {
		screen.Capabilities = exec.Video64BitBase
	}

	return
}

func boot(image *exec.LinuxImage) (err error) {
	var memoryMap *uefi.MemoryMap

	log.Print("go-boot exiting EFI boot services and jumping to kernel")

	// fill screen_info
	if image.Screen, err = screenInfo(); err != nil {
		log.Printf("could not detect screen information, %v\n", err)
	}

	for i := 0; i < exitRetries; i++ {
		// own all available memory
		if memoryMap, err = x64.UEFI.Boot.ExitBootServices(); err != nil {
			log.Print("go-boot exiting EFI boot services (retrying)")
			continue
		}
		break
	}

	if err != nil {
		return fmt.Errorf("could not exit EFI boot services, %v\n", err)
	}

	// silence EFI Simple Text console
	x64.Console.Out = 0
	x64.UEFI.Console.Out = 0

	// parse kernel image
	if err = image.Parse(); err != nil {
		return
	}

	// reserve runtime memory for kernel loading
	if err = reserveMemory(memoryMap, image); err != nil {
		return
	}

	// release in case of error
	defer image.Region.Release(image.Region.Start())

	// fill EFI information in boot parameters
	if image.EFI, err = efiInfo(memoryMap); err != nil {
		return
	}

	// load kernel in reserved memory
	if err = image.Load(); err != nil {
		return fmt.Errorf("could not load kernel, %v", err)
	}

	log.Printf("booting kernel@%#x", image.Entry())
	return image.Boot(nil)
}

func linuxCmd(_ *shell.Interface, arg []string) (res string, err error) {
	var entry *uapi.Entry

	path := arg[0]

	if len(path) == 0 {
		path = DefaultLinuxEntry
	}

	if x64.UEFI.Boot == nil {
		return "", errors.New("EFI Boot Services unavailable")
	}

	root, err := x64.UEFI.Root()

	if err != nil {
		return "", fmt.Errorf("could not open root volume, %v", err)
	}

	if transparency.Config.Status != "none" {
		if err = transparency.Validate(); err != nil {
			return "", fmt.Errorf("boot-transparency validation failed\n%v", err)
		}
	}

	log.Printf("loading boot loader entry %s", path)

	if entry, err = uapi.LoadEntry(root, path); err != nil {
		return "", fmt.Errorf("error loading entry, %v", err)
	}

	if len(entry.Linux) == 0 {
		return "", errors.New("empty kernel entry")
	}

	image := &exec.LinuxImage{
		Kernel:         entry.Linux,
		InitialRamDisk: entry.Initrd,
		CmdLine:        entry.Options,
	}

	return "", boot(image)
}
