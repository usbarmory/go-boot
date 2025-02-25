
Introduction
============

> :warning: this is a Work in Progress not meant for production use

This [TamaGo](https://github.com/usbarmory/tamago) based unikernel acts as a
primary UEFI boot loader for AMD64 platforms, allowing boot of kernel images
(e.g. Linux) and UEFI API interaction.

Operation
=========

```
Shell> go-boot.efi

go-boot initializing (console=com1)

tamago/amd64 (go1.24.0) • UEFI

alloc           <hex offset> <size>      # EFI_BOOT_SERVICES.AllocatePages()
build                                    # build information
date            (time in RFC339 format)? # show/change runtime date and time
dma             (free|used)?             # show allocation of default DMA region
exit, quit                               # close session
halt                                     # halt the machine
info                                     # device information
init                                     # init UEFI services
linux           (path)?                  # boot Linux kernel bzImage
log                                      # show runtime log
memmap                                   # EFI_BOOT_SERVICES.GetMemoryMap()
peek            <hex offset> <size>      # memory display (use with caution)
poke            <hex offset> <hex value> # memory write   (use with caution)
reset           (cold|warm)?             # reset system
shutdown                                 # shutdown system
stack                                    # goroutine stack trace (current)
stackall                                 # goroutine stack trace (all)
uptime                                   # show how long the system has been running

> init
Firmware Revision .: 10000
Runtime Services  .: 0x79ecb98
Boot Services .....: 0x7ea5720
Table Entries .....: 10

> alloc 90000000 4096
allocating memory range 0x90000000 - 0x90001000

> memmap
Type Start            End              Pages            Attributes
02   0000000090000000 0000000090000fff 0000000000000001 000000000000000f

> linux
allocating memory range 0x80000000 - 0x90000000
loading kernel@80000000
starting kernel@81000000
exiting EFI boot services
Linux version 5.10.233 (root@tamago) (gcc (GCC) 14.2.1 20250128, GNU ld (GNU Binutils) 2.43.1)
...
```

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

The `CONSOLE` environment variable must be set to either `com1` or `text` to
configure the output console to serial port or UEFI console.

Build the `go-boot.efi` application executable:

```
git clone https://github.com/usbarmory/go-boot && cd go-boot
make efi CONSOLE=com1
```

Debugging
=========

QEMU supported targets can be executed under emulation, using the
[Open Virtual Machine Firmware](https://github.com/tianocore/tianocore.github.io/wiki/OVMF)
as follows:

```
make qemu CONSOLE=com1 OVMFCODE=<path to OVMF_CODE.fd> OVMFVARS=<path to OVMF_VARS.fd>
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

Authors
=======

Andrea Barisani  
andrea@inversepath.com  

License
=======

go-boot | https://github.com/usbarmory/go-boot  
Copyright (c) WithSecure Corporation

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/go-boot/blob/master/LICENSE) file.
