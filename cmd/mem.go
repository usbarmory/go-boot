// Copyright (c) WithSecure Corporation
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
		Syntax:  "<hex offset> <size>",
		Help:    "memory display (use with caution)",
		Fn:      memReadCmd,
	})

	shell.Add(shell.Cmd{
		Name:    "poke",
		Args:    2,
		Pattern: regexp.MustCompile(`^poke ([[:xdigit:]]+) ([[:xdigit:]]+)$`),
		Syntax:  "<hex offset> <hex value>",
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
	addr, err := strconv.ParseUint(arg[0], 16, 64)

	if err != nil {
		return "", fmt.Errorf("invalid address, %v", err)
	}

	size, err := strconv.ParseUint(arg[1], 10, 64)

	if err != nil {
		return "", fmt.Errorf("invalid size, %v", err)
	}

	if (addr%8) != 0 || (size%8) != 0 {
		return "", fmt.Errorf("only 64-bit aligned accesses are supported")
	}

	if size > maxBufferSize {
		return "", fmt.Errorf("size argument must be <= %d", maxBufferSize)
	}

	return hex.Dump(mem(uint(addr), int(size), nil)), nil
}

func memWriteCmd(_ *shell.Interface, arg []string) (res string, err error) {
	addr, err := strconv.ParseUint(arg[0], 16, 64)

	if err != nil {
		return "", fmt.Errorf("invalid address, %v", err)
	}

	val, err := strconv.ParseUint(arg[1], 16, 64)

	if err != nil {
		return "", fmt.Errorf("invalid data, %v", err)
	}

	buf := make([]byte, 8)
	binary.BigEndian.PutUint32(buf, uint32(val))

	mem(uint(addr), 8, buf)

	return
}
