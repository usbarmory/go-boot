TamaGo - bare metal Go - UEFI x64 support
=========================================

go-boot | https://github.com/usbarmory/go-boot  

Copyright (c) WithSecure Corporation  

![TamaGo gopher](https://github.com/usbarmory/tamago/wiki/images/tamago.svg?sanitize=true)

Authors
=======

Andrea Barisani  
andrea@inversepath.com  

Introduction
============

TamaGo is a framework that enables compilation and execution of unencumbered Go
applications on bare metal AMD64/ARM/RISC-V processors.

The [uefi](https://github.com/usbarmory/go-boot/tree/main/uefi) and
[uefi/x64](https://github.com/usbarmory/go-boot/tree/main/uefi/x64)
packages provides support for unikernels running under the Unified Extensible
Firmware Interface [UEFI](https://uefi.org/) on an AMD64 core.

Documentation
=============

[![Go Reference](https://pkg.go.dev/badge/github.com/usbarmory/go-boot/uefi.svg)](https://pkg.go.dev/github.com/usbarmory/go-boot/uefi)

For more information about TamaGo see its
[repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki).

For usage of these packages in the context of an UEFI application see the
[go-boot](https://github.com/usbarmory/go-boot) unikernel project.

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/go-boot).

Compiling
=========

Go applications are simply required to import, the relevant board package to
ensure that hardware initialization and runtime support take place:

```golang
import (
	_ "github.com/usbarmory/go-boot/uefi/x64"
)
```

Build the [TamaGo compiler](https://github.com/usbarmory/tamago-go)
(or use the [latest binary release](https://github.com/usbarmory/tamago-go/releases/latest)):

```
wget https://github.com/usbarmory/tamago-go/archive/refs/tags/latest.zip
unzip latest.zip
cd tamago-go-latest/src && ./all.bash
cd ../bin && export TAMAGO=`pwd`/go
```

Go applications can be compiled as usual, using the compiler built in the
previous step, but with the addition of the following flags/variables:

```
GOOS=tamago GOARCH=amd64 ${TAMAGO} build -ldflags "-E cpuinit -T $(TEXT_START) -R 0x1000" main.go
```

The resulting ELF must be converted to a PE32+ executable for EFI for execution
under UEFI:

```
objcopy \
	--strip-debug \
	--target efi-app-x86_64 \
	--subsystem=efi-app \
	--image-base 0x$(IMAGE_BASE) \
	--stack=0x10000 \
	main main.efi
printf '\x26\x02' | dd of=${APP}.efi bs=1 seek=150 count=2 conv=notrunc,fsync # adjust Characteristics
```

An example application, targeting the UEFI environment,
is [go-boot](https://github.com/usbarmory/go-boot).

License
=======

go-boot | https://github.com/usbarmory/go-boot  
Copyright (c) WithSecure Corporation

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/go-boot/blob/main/LICENSE) file.

The TamaGo logo is adapted from the Go gopher designed by Renee French and
licensed under the Creative Commons 3.0 Attributions license. Go Gopher vector
illustration by Hugo Arganda.
