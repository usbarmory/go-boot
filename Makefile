# Copyright (c) WithSecure Corporation
#
# Use of this source code is governed by the license
# that can be found in the LICENSE file.

BUILD_USER ?= $(shell whoami)
BUILD_HOST ?= $(shell hostname)
BUILD_DATE ?= $(shell /bin/date -u "+%Y-%m-%d %H:%M:%S")
BUILD_TAGS = linkramsize,linkramstart,linkprintk
BUILD = ${BUILD_USER}@${BUILD_HOST} on ${BUILD_DATE}
REV = $(shell git rev-parse --short HEAD 2> /dev/null)

SHELL = /bin/bash

APP ?= efi-boot
TEXT_START := 0x10010000 # ramStart (defined in mem.go under tamago/amd64 package) + 0x10000
TAMAGOFLAGS := -tags ${BUILD_TAGS} -trimpath -ldflags "-T $(TEXT_START) -R 0x1000 -X 'main.Build=${BUILD}' -X 'main.Revision=${REV}' -X 'main.Boot=${BOOT}' -X 'main.Start=${START}' -X 'main.PublicKeyStr=${PUBLIC_KEY}'"
GOENV := GOOS=tamago GOARCH=amd64

.PHONY: clean

#### primary targets ####

all: $(APP)

elf: $(APP)

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
	objcopy \
		--strip-debug \
		--image-base 0x1000f000 \
		--target efi-app-x86_64 \
		--subsystem=efi-app \
		${APP}
	printf '\x26\x02' | dd of=$(APP) bs=1 seek=150 count=2 conv=notrunc,fsync # ajust Characteristics
