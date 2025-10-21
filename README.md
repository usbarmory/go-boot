Introduction
============

The [go-boot](https://github.com/usbarmory/go-boot) project is a
[TamaGo](https://github.com/usbarmory/tamago) unikernel implementing a UEFI
Shell and OS loader for AMD64 platforms, allowing UEFI API interaction and OS
loading.

The OS loading functionality supports launching of:
 * `.` EFI application images
 * `l` Linux kernels, with configuration parsed from Linux Userspace API (UAPI) [boot loader entries](https://uapi-group.org/specifications/specs/boot_loader_specification/)
 * `w` Windows UEFI boot manager

Authors
=======

Andrea Barisani  
andrea@inversepath.com  

Operation
=========

The default operation is to present an UEFI shell and its help, the ⏎ shortcut
(identically to `l` or `linux`) boots the default UAPI entry set at compile
time (see _Compiling_).

```
Shell> go-boot.efi

initializing EFI services
initializing console (com1)

go-boot • tamago/amd64 (go1.24.1) • UEFI x64

.               <path>                   # load and start EFI image
build                                    # build information
cat             <path>                   # show file contents
clear                                    # clear screen
cpuid           <leaf> <subleaf>         # show CPU capabilities
date            (time in RFC339 format)? # show/change runtime date and time
dns             <host>                   # resolve domain
exit,quit                                # exit application
halt,shutdown                            # shutdown system
info                                     # runtime information
linux,l         (loader entry path)?     # boot Linux kernel image
linux,l,\r                               # `l \loader\entries\arch.conf`
log                                      # show runtime logs
ls              (path)?                  # list directory contents
lspci                                    # list PCI devices
memmap          (e820)?                  # show UEFI memory map
mode            <mode>                   # set screen mode
net             <ip> <mac> <gw> (debug)? # start UEFI networking
peek            <hex offset> <size>      # memory display (use with caution)
poke            <hex offset> <hex value> # memory write   (use with caution)
protocol        <registry format GUID>   # locate UEFI protocol
reset           (cold|warm)?             # reset system
stack                                    # goroutine stack trace (current)
stackall                                 # goroutine stack trace (all)
stat            <path>                   # show file information
uefi                                     # UEFI information
uptime                                   # show system running time
windows,win,w                            # launch Windows UEFI boot manager

> uefi
UEFI Revision ......: 2.70
Firmware Vendor ....: Lenovo
Firmware Revision ..: 0x1560
Runtime Services  ..: 0x90e2eb98
Boot Services ......: 0x6bd17690
Frame Buffer .......: 1920x1200 @ 0x4000000000
Configuration Tables: 0x8f426018
  ee4e5898-3914-4259-9d6e-dc7bd79403cf (0x8db6dc98)
  dcfa911d-26eb-469f-a220-38b7dc461220 (0x8b037018)
...

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

The following environment variables configure the `go-boot.efi` executable
build:

* `IMAGE_BASE`: must be set (in hex) within a memory range
  available in the target UEFI environment for the unikernel allocation, the
  [HCL](https://github.com/usbarmory/go-boot/wiki#hardware-compatibility-list) or
  `memmap` command from an [UEFI Shell](https://github.com/pbatard/UEFI-Shell)
  can provide such value, when empty a common default value is set.

* `DEFAULT_EFI_ENTRY`: defines the `.` shortcut entry path
  for EFI image loading, it defaults to `\efi\boot\bootx64.efi`
  when unspecified.

* `DEFAULT_LINUX_ENTRY`: defines the `linux,l,\r` shortcut loader entry path
  for Linux kernel image booting, it defaults to `\loader\entries\arch.conf`
  when unspecified.

* `CONSOLE`: set to either `com1` or `text` (default) controls the output
  console to either serial port or UEFI console.

* `NET`: set to either `0` (default) or `1` controls enabling of UEFI
  networking support (see _UEFI networking_).

Build the `go-boot.efi` executable:

```
git clone https://github.com/usbarmory/go-boot && cd go-boot
make efi IMAGE_BASE=10000000 CONSOLE=text
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

UEFI networking
===============

With `NET=1` passed in the environment builds include UEFI networking support
through the [Simple Network Protocol](https://uefi.org/specs/UEFI/2.10_A/24_Network_Protocols_SNP_PXE_BIS.html)
(SNP) and [go-net](https://github.com/usbarmory/go-net).

On such builds the `net` and `dns` commands become available and `make qemu`
will require a tap0 interface.

The `net` command takes an IP address in CIDR notation, a fixed MAC address or
`:` to automatically generate a random MAC, and a gateway IP address as
arguments.

The optional `debug` strings can be passed as final argument to `net` to enable
Go [profiling server](https://pkg.go.dev/net/http/pprof) and an unauthenticated
SSH console exposing the UEFI shell.

```
> net 10.0.0.1/24 : 10.0.0.2 debug
starting debug servers:
        http://10.0.0.1:80/debug/pprof
        ssh://10.0.0.1:22
network initialized (10.0.0.1/24 da:e7:ac:e2:5e:05)

> dns golang.org
[142.251.209.17 2a00:1450:4002:410::2011]
```

Emulated hardware with QEMU
===========================

QEMU supported targets can be executed under emulation, using the
[Open Virtual Machine Firmware](https://github.com/tianocore/tianocore.github.io/wiki/OVMF)
as follows:

```
make qemu CONSOLE=com1 OVMFCODE=<path to OVMF_CODE.fd>
```

With `NET=1` tap0 should be configured as follows (Linux example):

```
ip tuntap add dev tap0 mode tap group <your user group>
ip addr add 10.0.0.2/24 dev tap0
ip link set tap0 up
```

An emulated target can be [debugged with GDB](https://retrage.github.io/2019/12/05/debugging-ovmf-en.html/)
using `make qemu-gdb`, this will make qemu waiting for a GDB connection that
can be launched as follows:

```
gdb -ex "target remote 127.0.0.1:1234"
```

Breakpoints can be set in the usual way:

```
b cpuinit
continue
```

License
=======

go-boot | https://github.com/usbarmory/go-boot  
Copyright (c) The go-boot authors. All Rights Reserved.

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/go-boot/blob/main/LICENSE) file.
