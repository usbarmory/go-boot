module github.com/usbarmory/efi-boot

go 1.24.0

require (
	github.com/hako/durafmt v0.0.0-20210608085754-5c1018a4e16b
	github.com/usbarmory/tamago v0.0.0-20250212123402-5facf762488d
	golang.org/x/term v0.29.0
)

require golang.org/x/sys v0.30.0 // indirect

replace github.com/usbarmory/tamago => /mnt/git/public/tamago
