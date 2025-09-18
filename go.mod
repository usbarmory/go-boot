module github.com/usbarmory/go-boot

go 1.25.1

require (
	github.com/gliderlabs/ssh v0.1.2-0.20181113160402-cbabf5414432
	github.com/hako/durafmt v0.0.0-20210608085754-5c1018a4e16b
	github.com/u-root/u-root v0.15.0
	github.com/usbarmory/armory-boot v0.0.0-20250827125939-ac32f955c61c
	github.com/usbarmory/go-net v0.0.0-20250918102452-cc202caa88e2
	github.com/usbarmory/tamago v1.25.2-0.20250915195610-7b6e648e66bf
	golang.org/x/term v0.35.0
)

require (
	github.com/anmitsu/go-shlex v0.0.0-20200514113438-38f4b401e2be // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/klauspost/compress v1.17.4 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/therootcompany/xz v1.0.1 // indirect
	github.com/u-root/uio v0.0.0-20240224005618-d2acac8f3701 // indirect
	github.com/ulikunitz/xz v0.5.11 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/exp v0.0.0-20250305212735-054e65f0b394 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/time v0.7.0 // indirect
	gvisor.dev/gvisor v0.0.0-20250911055229-61a46406f068 // indirect
)

replace github.com/usbarmory/go-net => /mnt/git/public/go-net
