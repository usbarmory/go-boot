
Introduction
============

> :warning: this is a Work in Progress

This [TamaGo](https://github.com/usbarmory/tamago) based unikernel acts as a
primary UEFI boot loader for AMD64 platforms, allowing boot of kernel images
(e.g. Linux).

Operation
=========

```
Shell> go.efi

tamago/amd64 (go1.24.0) • UEFI

build                                    # build information
date            (time in RFC339 format)? # show/change runtime date and time
dma             (free|used)?             # show allocation of default DMA region
exit, quit                               # close session
halt                                     # halt the machine
help                                     # this help
info                                     # device information
peek            <hex offset> <size>      # memory display (use with caution)
poke            <hex offset> <hex value> # memory write   (use with caution)
reboot                                   # reset device
stack                                    # goroutine stack trace (current)
stackall                                 # goroutine stack trace (all)
uptime                                   # show how long the system has been running

>
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
make efi
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
