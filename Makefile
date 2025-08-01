# Copyright (c) WithSecure Corporation
#
# Use of this source code is governed by the license
# that can be found in the LICENSE file.

BUILD_TAGS = linkcpuinit,linkramsize,linkramstart,linkprintk
SHELL = /bin/bash
APP ?= go-boot
CONSOLE ?= text
DEFAULT_LINUX_ENTRY = \loader\entries\arch.conf

IMAGE_BASE := 10000000
TEXT_START := $(shell echo $$((16#$(IMAGE_BASE) + 16#10000)))
GOFLAGS := -tags ${BUILD_TAGS} -trimpath -ldflags "-s -w -E cpuinit -T $(TEXT_START) -R 0x1000 -X 'main.Console=${CONSOLE}' -X 'github.com/usbarmory/go-boot/cmd.DefaultLinuxEntry=${DEFAULT_LINUX_ENTRY}'"
GOENV := GOOS=tamago GOARCH=amd64

OVMFCODE ?= OVMF_CODE.fd
OVMFVARS ?= OVMF_VARS.fd
LOG ?= qemu.log

QEMU ?= qemu-system-x86_64 \
        -enable-kvm -cpu host,invtsc=on -m 8G \
        -drive file=fat:rw:$(CURDIR)/qemu-disk \
        -drive if=pflash,format=raw,readonly,file=$(OVMFCODE) \
        -drive if=pflash,format=raw,file=$(OVMFVARS) \
        -global isa-debugcon.iobase=0x402 \
        -serial stdio -vga virtio \
        # -debugcon file:$(LOG)

.PHONY: clean

#### primary targets ####

all: $(APP).efi

elf: $(APP)

efi: $(APP).efi

qemu: $(APP).efi
	@if [ "${QEMU}" == "" ]; then \
		echo 'qemu not available for this target'; \
		exit 1; \
	fi
	mkdir -p $(CURDIR)/qemu-disk/efi/boot && cp $(CURDIR)/$(APP).efi $(CURDIR)/qemu-disk/efi/boot/bootx64.efi
	$(QEMU)

qemu-gdb: GOFLAGS := $(GOFLAGS:-w=)
qemu-gdb: GOFLAGS := $(GOFLAGS:-s=)
qemu-gdb: $(APP).efi
	$(QEMU) -S -s

#### utilities ####

check_tamago:
	@if [ "${TAMAGO}" == "" ] || [ ! -f "${TAMAGO}" ]; then \
		echo 'You need to set the TAMAGO variable to a compiled version of https://github.com/usbarmory/tamago-go'; \
		exit 1; \
	fi

clean:
	@rm -fr $(APP) $(APP).efi $(CURDIR)/qemu-disk

#### dependencies ####

$(APP): check_tamago
	$(GOENV) $(TAMAGO) build $(GOFLAGS) -o ${APP}

$(APP).efi: $(APP)
	objcopy \
		--strip-debug \
		--target efi-app-x86_64 \
		--subsystem=efi-app \
		--image-base 0x$(IMAGE_BASE) \
		--stack=0x10000 \
		${APP} ${APP}.efi
	printf '\x26\x02' | dd of=${APP}.efi bs=1 seek=150 count=2 conv=notrunc,fsync # adjust Characteristics
