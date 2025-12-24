// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/usbarmory/tamago/dma"
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

	// DMA region must be allocated before GHCB initialization
	x64.AllocateDMA(1 << 20) // FIXME: C-bit
}

func sevCmd(_ *shell.Interface, _ []string) (res string, err error) {
	var buf bytes.Buffer
	var key []byte

	var snp *uefi.SNPConfigurationTable
	var report *svm.AttestationReport

	if !x64.AMD64.Features().SNP {
		return "", errors.New("AMD SEV-SNP unavailable")
	}

	if snp, _ = x64.UEFI.GetSNPConfiguration(); err != nil {
		return "", errors.New("could find AMD SEV-SNP pages")
	}

	fmt.Fprintf(&buf, "Revision ...........: %d\n", snp.Version)
	fmt.Fprintf(&buf, "Secrets Page .......: %x (%d)\n", snp.SecretsPagePhysicalAddress, snp.SecretsPageSize)
	fmt.Fprintf(&buf, "  CPUID Page .......: %x (%d)\n", snp.CPUIDPagePhysicalAddress, snp.CPUIDPageSize)

	secrets := &svm.SNPSecrets{}

	if err = secrets.Init(uint(snp.SecretsPagePhysicalAddress), int(snp.SecretsPageSize)); err != nil {
		return "", errors.New("could not initialize AMD SEV-SNP secrets")
	}

	if key, err = secrets.VMPCK(0); err != nil {
		return
	}

	ghcb := &svm.GHCB{
		SharedMemory: dma.Default(),
	}

	if err = ghcb.Init(); err != nil {
		return
	}

	data := make([]byte, 64)
	rand.Read(data)

	if report, err = ghcb.GetAttestationReport(data, key, 0); err != nil {
		return
	}

	fmt.Fprintf(&buf, "Attestation Report:\n%+v\n", report)

	return buf.String(), err
}
