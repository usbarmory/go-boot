# Copyright (c) WithSecure Corporation
#
# Use of this source code is governed by the license
# that can be found in the LICENSE file.

BUILD_TAGS = linkramsize,linkramstart,linkprintk

SHELL = /bin/bash

APP ?= go-boot
# FIXME
TEXT_START := 0x05c61b00 # ramStart (defined in mem.go under tamago/amd64 package) + 0x10000
TAMAGOFLAGS := -tags ${BUILD_TAGS} -trimpath -ldflags "-T $(TEXT_START) -R 0x1000"
GOENV := GOOS=tamago GOARCH=amd64

.PHONY: clean

#### primary targets ####

all: $(APP)

elf: $(APP)

efi: $(APP).efi

#### utilities ####

check_tamago:
	@if [ "${TAMAGO}" == "" ] || [ ! -f "${TAMAGO}" ]; then \
		echo 'You need to set the TAMAGO variable to a compiled version of https://github.com/usbarmory/tamago-go'; \
		exit 1; \
	fi

clean:
	@rm -fr $(APP)

#### dependencies ####

$(APP): check_tamago
	$(GOENV) $(TAMAGO) build $(TAMAGOFLAGS) -o ${APP}

$(APP).efi: $(APP)
	objcopy \
		--strip-debug \
		--image-base 0x0500e000 \
		--target efi-app-x86_64 \
		--subsystem=efi-app \
		--stack=0x10000 \
		${APP} ${APP}.efi
	printf '\x26\x02' | dd of=${APP}.efi bs=1 seek=150 count=2 conv=notrunc,fsync # ajust Characteristics
