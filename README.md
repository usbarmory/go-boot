
Introduction
============

> :warning: this is a Work in Progress

This [TamaGo](https://github.com/usbarmory/tamago) based unikernel acts as a
primary UEFI boot loader for AMD64 platforms, allowing boot of kernel images
(e.g. Linux).

Operation
=========

```
Shell> go-boot.efi

tamago/amd64 (go1.24.0) • UEFI

alloc           <hex offset> <size>      # EFI_BOOT_SERVICES.AllocatePages()
build                                    # build information
date            (time in RFC339 format)? # show/change runtime date and time
dma             (free|used)?             # show allocation of default DMA region
exit, quit                               # close session
halt                                     # halt the machine
help                                     # this help
info                                     # device information
linux           (path)?                  # boot Linux kernel bzImage
peek            <hex offset> <size>      # memory display (use with caution)
poke            <hex offset> <hex value> # memory write   (use with caution)
reboot                                   # reset device
stack                                    # goroutine stack trace (current)
stackall                                 # goroutine stack trace (all)
uefi                                     # UEFI information
uptime                                   # show how long the system has been running

> uefi
Firmware Revision .: 10000
Runtime Services  .: 0x79ecb98
Boot Services .....: 0x7ea5720
Table Entries .....: 10
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

Build the `go-boot.efi` application executable:

```
git clone https://github.com/usbarmory/go-boot && cd go-boot
# WiP: TEXT_START must be adjusted to reflect the actual EntryPoint from qemu logs
make efi TEXT_START=0x05c61b00
```

Debugging
=========

QEMU supported targets can be executed under emulation, using the
[Open Virtual Machine Firmware](https://github.com/tianocore/tianocore.github.io/wiki/OVMF)
as follows:

```
make qemu OVMFCODE=<path to OVMF_CODE.fd> OVMFVARS=<path to OVMF_VARS.fd>
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
