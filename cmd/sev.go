// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/usbarmory/tamago/kvm/svm"

	"github.com/usbarmory/go-boot/shell"
	"github.com/usbarmory/go-boot/uefi"
	"github.com/usbarmory/go-boot/uefi/x64"
)

func init() {
	shell.Add(shell.Cmd{
		Name: "sev",
		Help: "AMD SEV-SNP information",
		Fn:   sevCmd,
	})
}

func sevCmd(_ *shell.Interface, _ []string) (res string, err error) {
	var buf bytes.Buffer
	var snp *uefi.SNPConfigurationTable

	if !x64.AMD64.Features().SNP {
		return "", errors.New("AMD SEV-SNP unavailable")
	}

	if snp, _ = x64.UEFI.GetSNPConfiguration(); err != nil {
		return "", errors.New("could find AMD SEV-SNP pages")
	}

	fmt.Fprintf(&buf, "Version ............: %d\n", snp.Version)
	fmt.Fprintf(&buf, "Secrets Page .......: %x (%d)\n", snp.SecretsPagePhysicalAddress, snp.SecretsPageSize)
	fmt.Fprintf(&buf, "  CPUID Page .......: %x (%d)\n", snp.CPUIDPagePhysicalAddress, snp.CPUIDPageSize)

	secrets := &svm.SNPSecrets{}

	if err = secrets.Init(uint(snp.SecretsPagePhysicalAddress), int(snp.SecretsPageSize)); err != nil {
		return "", errors.New("could not initialize AMD SEV-SNP secrets")
	}

	for i := 0; i < 4; i++ {
		vmpck, err := secrets.VMPCK(i)

		if err != nil {
			return "", fmt.Errorf("could not read VMPCK%d, %v", i, err)
		}

		fmt.Fprintf(&buf, " VMPCK%d ...........: %x\n", i, vmpck)
	}

	ghcb := &svm.GHCB{}
	x64.AllocateDMA(svm.GHCBSize)

	ghcb.Init()
	ghcb.Yield()

	return buf.String(), err
}
