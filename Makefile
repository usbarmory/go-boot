# Copyright (c) The go-boot authors. All Rights Reserved.
#
# Use of this source code is governed by the license
# that can be found in the LICENSE file.

NET ?= 0
BUILD_TAGS = linkcpuinit,linkramsize,linkramstart,linkprintk
SHELL = /bin/bash
APP ?= go-boot
CONSOLE ?= text
DEFAULT_EFI_ENTRY = \efi\boot\bootx64.efi
DEFAULT_LINUX_ENTRY = \loader\entries\arch.conf

ifeq ($(NET),1)
    BUILD_TAGS := $(BUILD_TAGS),net
endif

IMAGE_BASE := 10000000
TEXT_START := $(shell echo $$((16#$(IMAGE_BASE) + 16#10000)))
LDFLAGS := -s -w -E cpuinit -T $(TEXT_START) -R 0x1000 -X 'main.Console=${CONSOLE}'
LDFLAGS += -X 'github.com/usbarmory/go-boot/cmd.DefaultEFIEntry=${DEFAULT_EFI_ENTRY}'
LDFLAGS += -X 'github.com/usbarmory/go-boot/cmd.DefaultLinuxEntry=${DEFAULT_LINUX_ENTRY}'
GOFLAGS := -tags ${BUILD_TAGS} -trimpath -ldflags "${LDFLAGS}"
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

ifeq ($(NET),1)
        QEMU := $(QEMU) -device virtio-net-pci,netdev=net0 -netdev tap,id=net0,ifname=tap0,script=no,downscript=no
endif

.PHONY: clean

#### primary targets ####

all: $(APP).efi

elf: $(APP)

efi: $(APP).efi

qemu: $(APP).efi
	mkdir -p $(CURDIR)/qemu-disk/efi/boot && cp $(CURDIR)/$(APP).efi $(CURDIR)/qemu-disk/efi/boot/bootx64.efi
	$(QEMU)

qemu-gdb: GOFLAGS := $(GOFLAGS:-w=)
qemu-gdb: GOFLAGS := $(GOFLAGS:-s=)
qemu-gdb: $(APP).efi
	mkdir -p $(CURDIR)/qemu-disk/efi/boot && cp $(CURDIR)/$(APP).efi $(CURDIR)/qemu-disk/efi/boot/bootx64.efi
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
