# Copyright (c) WithSecure Corporation
#
# Use of this source code is governed by the license
# that can be found in the LICENSE file.

BUILD_TAGS = linkcpuinit,linkramsize,linkramstart,linkprintk
SHELL = /bin/bash
APP ?= go-boot
CONSOLE ?= com1

TEXT_START := 0x10010000 # ramStart (defined in mem.go under tamago/amd64 package) + 0x10000
GOFLAGS := -tags ${BUILD_TAGS} -trimpath -ldflags "-T $(TEXT_START) -R 0x1000 -X 'main.Console=${CONSOLE}'"
GOENV := GOOS=tamago GOARCH=amd64

OVMFCODE ?= OVMF_CODE.fd
OVMFVARS ?= OVMF_VARS.fd
LOG ?= qemu.log

QEMU ?= qemu-system-x86_64 \
        -enable-kvm -cpu host,invtsc=on -m 4G \
        -drive file=fat:rw:$(CURDIR) \
        -drive if=pflash,format=raw,readonly,file=$(OVMFCODE) \
        -drive if=pflash,format=raw,file=$(OVMFVARS) \
        -debugcon file:$(LOG) -global isa-debugcon.iobase=0x402 \
        -serial stdio -nographic -nodefaults

.PHONY: clean

#### primary targets ####

all: $(APP)

elf: $(APP)

efi: $(APP).efi

qemu: $(APP).efi
	@if [ "${QEMU}" == "" ]; then \
		echo 'qemu not available for this target'; \
		exit 1; \
	fi
	$(QEMU)

qemu-gdb: GOFLAGS := $(GOFLAGS:-w=)
qemu-gdb: GOFLAGS := $(GOFLAGS:-s=)
qemu-gdb: $(APP)
	$(QEMU) -S -s

#### utilities ####

check_tamago:
	@if [ "${TAMAGO}" == "" ] || [ ! -f "${TAMAGO}" ]; then \
		echo 'You need to set the TAMAGO variable to a compiled version of https://github.com/usbarmory/tamago-go'; \
		exit 1; \
	fi

clean:
	@rm -fr $(APP) $(APP).efi

#### dependencies ####

$(APP): check_tamago
	$(GOENV) $(TAMAGO) build $(GOFLAGS) -o ${APP}

$(APP).efi: $(APP)
	objcopy \
		--strip-debug \
		--target efi-app-x86_64 \
		--subsystem=efi-app \
		--image-base $(subst 1000,0000,$(TEXT_START)) \
		--stack=0x10000 \
		${APP} ${APP}.efi
	printf '\x26\x02' | dd of=${APP}.efi bs=1 seek=150 count=2 conv=notrunc,fsync # adjust Characteristics
