Introduction
============

The [go-boot](https://github.com/usbarmory/go-boot) project is a
[TamaGo](https://github.com/usbarmory/tamago) unikernel implementing a UEFI
Shell and OS loader for AMD64 platforms, allowing UEFI API interaction and OS
loading.

The OS loading functionality currently supports parsing of boot loader entries
compliant with the Linux Userspace API (UAPI)
[specifications](https://uapi-group.org/specifications/specs/boot_loader_specification/)
such as ones supported by
[systemd-boot](https://www.freedesktop.org/wiki/Software/systemd/systemd-boot/).

Authors
=======

Andrea Barisani  
andrea@inversepath.com  

Operation
=========

```
Shell> go-boot.efi

initializing EFI services
initializing console (com1)

go-boot • tamago/amd64 (go1.24.1) • UEFI x64

build                                    # build information
cat             <path>                   # show file contents
clear                                    # clear screen
cpuid           <leaf> <subleaf>         # show CPU capabilities
date            (time in RFC339 format)? # show/change runtime date and time
exit,quit                                # exit application
halt,shutdown                            # shutdown system
info                                     # device information
linux,l,\r      (loader entry path)?     # boot Linux kernel image
log                                      # show runtime logs
memmap          (e820)?                  # show UEFI memory map
mode            <mode>                   # set screen mode
peek            <hex offset> <size>      # memory display (use with caution)
poke            <hex offset> <hex value> # memory write   (use with caution)
protocol        <registry format GUID>   # locate UEFI protocol
reset           (cold|warm)?             # reset system
stack                                    # goroutine stack trace (current)
stackall                                 # goroutine stack trace (all)
stat            <path>                   # show file information
uefi                                     # UEFI information
uptime                                   # show system running time

> uefi
Firmware Vendor ....: Lenovo
Firmware Revision ..: 0x1560
Runtime Services  ..: 0x90e2eb98
Boot Services ......: 0x6bd17690
Frame Buffer .......: 1920x1200 @ 0x4000000000
Configuration Tables: 0x8f426018
  ee4e5898-3914-4259-9d6e-dc7bd79403cf (0x8db6dc98)
  dcfa911d-26eb-469f-a220-38b7dc461220 (0x8b037018)
...

> alloc 90000000 4096
allocating memory range 0x90000000 - 0x90001000

> memmap
Type Start            End              Pages            Attributes
02   0000000090000000 0000000090000fff 0000000000000001 000000000000000f
...

> linux \loader\entries\arch.conf
loading boot loader entry \loader\entries\arch.conf
go-boot exiting EFI boot services and jumping to kernel
Linux version 6.13.6-arch1-1 (linux@archlinux) (gcc (GCC) 14.2.1 20250207, GNU ld (GNU Binutils) 2.44)
...
```

Package documentation
=====================

[![Go Reference](https://pkg.go.dev/badge/github.com/usbarmory/go-boot.svg)](https://pkg.go.dev/github.com/usbarmory/go-boot)

Hardware Compatibility List
===========================

The list of supported hardware is available in the
project wiki [HCL](https://github.com/usbarmory/go-boot/wiki#hardware-compatibility-list).

The list provides test `IMAGE_BASE` values to pass while _Compiling_.

Compiling
=========

Build the [TamaGo compiler](https://github.com/usbarmory/tamago-go)
(or use the [latest binary release](https://github.com/usbarmory/tamago-go/releases/latest)):

```
wget https://github.com/usbarmory/tamago-go/archive/refs/tags/latest.zip
unzip latest.zip
cd tamago-go-latest/src && ./all.bash
cd ../bin && export TAMAGO=`pwd`/go
```

The `CONSOLE` environment variable must be set to either `com1` or `text`
(default) to configure the output console to serial port or UEFI console.

The `IMAGE_BASE` environment variable must be set within a memory range (in
hex) available in the target UEFI environment for the unikernel allocation, the
[HCL](https://github.com/usbarmory/go-boot/wiki#hardware-compatibility-list) or
`memmap` command from an [UEFI Shell](https://github.com/pbatard/UEFI-Shell)
can provide such value, when empty a common default value is set.

Build the `go-boot.efi` application executable:

```
git clone https://github.com/usbarmory/go-boot && cd go-boot
make efi IMAGE_BASE=00100000 CONSOLE=text
```

Executing as UEFI application
=============================

The `go-boot.efi` application executable, built after _Compiling_, can be
loaded from an [UEFI Shell](https://github.com/pbatard/UEFI-Shell)
or boot manager, the following example shows an entry for
[systemd-boot](https://www.freedesktop.org/wiki/Software/systemd/systemd-boot/):

```
# /boot/loader/entries/go-boot.conf
title Go Boot
efi /EFI/Linux/go-boot.efi
```

UEFI boot manager entry
=======================

The following example shows creation of an EFI boot entry using
[efibootmgr](https://github.com/rhboot/efibootmgr):

```
efibootmgr -C -L "go-boot" -d $DISK -p $PART -l '\EFI\go-boot.efi'
```

Emulated hardware with QEMU
===========================

QEMU supported targets can be executed under emulation, using the
[Open Virtual Machine Firmware](https://github.com/tianocore/tianocore.github.io/wiki/OVMF)
as follows:

```
make qemu IMAGE_BASE=40000000 CONSOLE=com1 OVMFCODE=<path to OVMF_CODE.fd> OVMFVARS=<path to OVMF_VARS.fd>
```

The emulation run will provide an interactive console.

An emulated target can be [debugged with GDB](https://retrage.github.io/2019/12/05/debugging-ovmf-en.html/)
using `make qemu-gdb`, this will make qemu waiting for a GDB connection that
can be launched as follows:

```
gdb -ex "target remote 127.0.0.1:1234"
```

Breakpoints can be set in the usual way:

```
b CoreStartImage
continue
```

License
=======

go-boot | https://github.com/usbarmory/go-boot  
Copyright (c) WithSecure Corporation

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/go-boot/blob/main/LICENSE) file.
