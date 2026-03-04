// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"

	"github.com/usbarmory/go-boot/shell"
	"github.com/usbarmory/tamago/dma"
)

const maxBufferSize = 102400

func init() {
	shell.Add(shell.Cmd{
		Name:    "peek",
		Args:    2,
		Pattern: regexp.MustCompile(`^peek ([[:xdigit:]]+) (\d+)$`),
		Syntax:  "<hex addr> <size>",
		Help:    "memory display (use with caution)",
		Fn:      memReadCmd,
	})

	shell.Add(shell.Cmd{
		Name:    "poke",
		Args:    2,
		Pattern: regexp.MustCompile(`^poke ([[:xdigit:]]+) ([[:xdigit:]]+)$`),
		Syntax:  "<hex addr> <hex value>",
		Help:    "memory write   (use with caution)",
		Fn:      memWriteCmd,
	})
}

func memCopy(start uint, size int, w []byte) (b []byte) {
	mem, err := dma.NewRegion(start, size, true)

	if err != nil {
		panic("could not allocate memory copy DMA")
	}

	start, buf := mem.Reserve(size, 0)
	defer mem.Release(start)

	if len(w) > 0 {
		copy(buf, w)
	} else {
		b = make([]byte, size)
		copy(b, buf)
	}

	return
}

func memReadCmd(_ *shell.Interface, arg []string) (res string, err error) {
	addr, err := strconv.ParseUint(arg[0], 16, dma.DefaultAlignment*8)

	if err != nil {
		return "", fmt.Errorf("invalid address, %v", err)
	}

	size, err := strconv.ParseUint(arg[1], 10, 32)

	if err != nil {
		return "", fmt.Errorf("invalid size, %v", err)
	}

	if (addr%dma.DefaultAlignment) != 0 || (size%dma.DefaultAlignment) != 0 {
		return "", fmt.Errorf("only %d-bit aligned accesses are supported", dma.DefaultAlignment*8)
	}

	if size > maxBufferSize {
		return "", fmt.Errorf("size argument must be <= %d", maxBufferSize)
	}

	return hex.Dump(memCopy(uint(addr), int(size), nil)), nil
}

func memWriteCmd(_ *shell.Interface, arg []string) (res string, err error) {
	addr, err := strconv.ParseUint(arg[0], 16, dma.DefaultAlignment*8)

	if err != nil {
		return "", fmt.Errorf("invalid address, %v", err)
	}

	val, err := strconv.ParseUint(arg[1], 16, 32)

	if err != nil {
		return "", fmt.Errorf("invalid data, %v", err)
	}

	size := 4

	if (addr%dma.DefaultAlignment) != 0 || (size%dma.DefaultAlignment) != 0 {
		return "", fmt.Errorf("only %d-bit aligned accesses are supported", dma.DefaultAlignment*8)
	}

	buf := make([]byte, size)
	binary.BigEndian.PutUint32(buf, uint32(val))

	memCopy(uint(addr), size, buf)

	return
}
