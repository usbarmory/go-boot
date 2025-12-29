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

var vmpck0 []byte

func init() {
	if !svm.Features(x64.AMD64).SEV.SEV {
		return
	}

	shell.Add(shell.Cmd{
		Name: "sev",
		Help: "AMD SEV-SNP information",
		Fn:   sevCmd,
	})

	shell.Add(shell.Cmd{
		Name: "sev-report",
		Help: "AMD SEV-SNP attestation report",
		Fn:   attestationCmd,
	})
}

func sevCmd(_ *shell.Interface, _ []string) (res string, err error) {
	var buf bytes.Buffer
	var snp *uefi.SNPConfigurationTable

	defer func() {
		res = buf.String()
		err = nil
	}()

	features := svm.Features(x64.AMD64)

	fmt.Fprintf(&buf, "SEV ................: %v\n", features.SEV.SEV)
	fmt.Fprintf(&buf, "SEV-ES .............: %v\n", features.SEV.ES)
	fmt.Fprintf(&buf, "SEV-SNP ............: %v\n", features.SEV.SNP)

	if !features.SEV.SNP {
		return
	}

	if snp, err = x64.UEFI.GetSNPConfiguration(); err != nil {
		fmt.Fprintf(&buf, " could not find AMD SEV-SNP pages, %v", err)
		return
	}

	fmt.Fprintf(&buf, "Revision ...........: %d\n", snp.Version)
	fmt.Fprintf(&buf, "Secrets Page .......: %#x (%d bytes)\n", snp.SecretsPagePhysicalAddress, snp.SecretsPageSize)
	fmt.Fprintf(&buf, "CPUID Page .........: %#x (%d bytes)\n", snp.CPUIDPagePhysicalAddress, snp.CPUIDPageSize)

	secrets := &svm.SNPSecrets{
		Address: uint(snp.SecretsPagePhysicalAddress),
		Size:    int(snp.SecretsPageSize),
	}

	if err = secrets.Init(); err != nil {
		fmt.Fprintf(&buf, " could not initialize AMD SEV-SNP secrets, %v", err)
		return
	}

	if vmpck0, err = secrets.VMPCK(0); err != nil {
		fmt.Fprintf(&buf, " could not get VMPCK0, %v", err)
		return
	}

	n := len(vmpck0)
	fmt.Fprintf(&buf, "VMPCK0 .............: %x -- %x (%d bytes)\n", vmpck0[0], vmpck0[n-1], n)

	return
}

func attestationCmd(_ *shell.Interface, _ []string) (res string, err error) {
	var ghcbAddr uint64
	var report *svm.AttestationReport

	if len(vmpck0) == 0 {
		return "", errors.New("AMD SEV-SNP secrsts unavailable, run `sev` first")
	}

	ghcb := &svm.GHCB{}

	// OVMF already allocates 2 shared pages, the first for GHCB and the
	// second for per-CPU variables, we re-use them four our purposes.

	if ghcbAddr = x64.AMD64.MSR(svm.MSR_AMD_GHCB); ghcbAddr == 0 {
		return "", errors.New("could not find GHCB address")
	}

	if ghcb.SharedMemory, err = dma.NewRegion(uint(ghcbAddr), 4096*2, false); err != nil {
		return "", fmt.Errorf("could not allocate GHCB, %v", err)
	}

	if err = ghcb.Init(false); err != nil {
		return "", fmt.Errorf("could not initialize GHCB, %v", err)
	}

	data := make([]byte, 64)
	rand.Read(data)

	if report, err = ghcb.GetAttestationReport(data, vmpck0, 0); err != nil {
		return "", fmt.Errorf("could not get report, %v", err)
	}

	return fmt.Sprintf("%+v", report), nil
}
