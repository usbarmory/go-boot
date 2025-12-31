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

	"github.com/usbarmory/tamago/amd64"
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
	fmt.Fprintf(&buf, "Encryption bit .....: %d\n", features.EncryptionBit)

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
	var buf bytes.Buffer
	var ghcbAddr uint64
	var report *svm.AttestationReport

	if len(vmpck0) == 0 {
		return "", errors.New("AMD SEV-SNP secrsts unavailable, run `sev` first")
	}

	if ghcbAddr = x64.AMD64.MSR(svm.MSR_AMD_GHCB); ghcbAddr == 0 {
		return "", errors.New("could not find GHCB address")
	}

	// OVMF allocates 2*ncpu contiguous pages, a first shared page
	// for GHCB and a second private one for vCPU variables.
	//
	// We are running on CPU0, so we obtain the first GHCB page and use the
	// unencrypted GHCB page allocated for CPU1 as request/response buffer,
	// sparing us from MMU re-configuration.
	if amd64.NumCPU() < 2 {
		return "", errors.New("cannot hijack unencrypted pages on single-core")
	}

	ghcbGPA := uint(ghcbAddr)
	reqGPA := ghcbGPA + uefi.PageSize*2
	ghcb := &svm.GHCB{}

	if ghcb.GHCBPage, err = dma.NewRegion(ghcbGPA, uefi.PageSize, false); err != nil {
		return "", fmt.Errorf("could not allocate GHCB layout page, %v", err)
	}

	if ghcb.RequestPage, err = dma.NewRegion(reqGPA, uefi.PageSize, false); err != nil {
		return "", fmt.Errorf("could not allocate GHCB request page, %v", err)
	}

	if err = ghcb.Init(false); err != nil {
		return "", fmt.Errorf("could not initialize GHCB, %v", err)
	}

	data := make([]byte, 64)
	rand.Read(data)

	if report, err = ghcb.GetAttestationReport(data, vmpck0, 0); err != nil {
		return "", fmt.Errorf("could not get report, %v", err)
	}

	fmt.Fprintf(&buf, "Version ......: %x\n", report.Version)
	fmt.Fprintf(&buf, "VMPL .........: %x\n", report.VMPL)
	fmt.Fprintf(&buf, "SignatureAlgo : %x\n", report.SignatureAlgo)
	fmt.Fprintf(&buf, "CurrentTCB ...: %x\n", report.CurrentTCB)
	fmt.Fprintf(&buf, "ReportedTCB ..: %x\n", report.ReportedTCB)
	fmt.Fprintf(&buf, "CommittedTCB .: %x\n", report.CommittedTCB)
	fmt.Fprintf(&buf, "Measurement ..: %x\n", report.Measurement)
	fmt.Fprintf(&buf, "SignatureR ...: %x\n", report.Signature[1:1+48])
	fmt.Fprintf(&buf, "SignatureS ...: %x\n", report.Signature[73:73+48])

	return buf.String(), nil
}
