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
	"regexp"

	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/kvm/sev"

	"github.com/usbarmory/go-boot/shell"
	"github.com/usbarmory/go-boot/uefi"
	"github.com/usbarmory/go-boot/uefi/x64"
)

var (
	features sev.SVMFeatures
	ghcb     *sev.GHCB
	secrets  *sev.SecretsPage
)

func init() {
	if !sev.Features(x64.AMD64).SEV.SEV {
		return
	}

	shell.Add(shell.Cmd{
		Name: "sev",
		Help: "AMD SEV-SNP information",
		Fn:   sevCmd,
	})

	shell.Add(shell.Cmd{
		Name:    "sev-report",
		Args:    1,
		Pattern: regexp.MustCompile(`^sev-report(?: (raw))?$`),
		Syntax:  "(raw)?",
		Help:    "AMD SEV-SNP attestation report",
		Fn:      attestationCmd,
	})

	shell.Add(shell.Cmd{
		Name: "sev-kdf",
		Help: "AMD SEV-SNP key derivation",
		Fn:   kdfCmd,
	})

	shell.Add(shell.Cmd{
		Name: "sev-tsc",
		Help: "AMD SEV-SNP TSC information",
		Fn:   tscCmd,
	})
}

func sevCmd(_ *shell.Interface, _ []string) (res string, err error) {
	var buf bytes.Buffer
	var snp *uefi.SNPConfigurationTable

	defer func() {
		res = buf.String()
		err = nil
	}()

	features = sev.Features(x64.AMD64)

	fmt.Fprintf(&buf, "SEV ................: %v\n", features.SEV.SEV)
	fmt.Fprintf(&buf, "SEV-ES .............: %v\n", features.SEV.ES)
	fmt.Fprintf(&buf, "SEV-SNP ............: %v\n", features.SEV.SNP)
	fmt.Fprintf(&buf, "Encrypted bit ......: %d\n", features.EncryptedBit)

	if !features.SEV.SNP {
		return
	}

	if snp, err = x64.UEFI.GetSNPConfiguration(); err != nil {
		fmt.Fprintf(&buf, " could not find AMD SEV-SNP pages, %v", err)
		return
	}

	fmt.Fprintf(&buf, "SNP Version ........: %d\n\n", snp.Version)
	fmt.Fprintf(&buf, "Secrets Page .......: %#x (%d bytes)\n", snp.SecretsPagePhysicalAddress, snp.SecretsPageSize)

	secrets = &sev.SecretsPage{}

	if err = secrets.Init(uint(snp.SecretsPagePhysicalAddress), int(snp.SecretsPageSize)); err != nil {
		fmt.Fprintf(&buf, " could not initialize AMD SEV-SNP secrets, %v", err)
		return
	}

	fmt.Fprintf(&buf, "Secrets Version ....: %d\n", secrets.Version)
	fmt.Fprintf(&buf, "TSC Factor .........: %#x\n", secrets.TSCFactor)
	fmt.Fprintf(&buf, "Launch Mitigations .: %#x\n", secrets.LaunchMitVector)
	fmt.Fprintf(&buf, "VMPCK0 .............: %#02x -- %#02x\n", secrets.VMPCK0[0], secrets.VMPCK0[31])
	fmt.Fprintf(&buf, "VMPCK1 .............: %#02x -- %#02x\n", secrets.VMPCK1[0], secrets.VMPCK1[31])
	fmt.Fprintf(&buf, "VMPCK2 .............: %#02x -- %#02x\n", secrets.VMPCK2[0], secrets.VMPCK2[31])
	fmt.Fprintf(&buf, "VMPCK3 .............: %#02x -- %#02x\n", secrets.VMPCK3[0], secrets.VMPCK3[31])

	return
}

func initSharedDMA(dmaSize int) (err error) {
	// align to 2MB page
	dmaStart := int(x64.RamSize) - dmaSize
	dmaSize += dmaStart % (2 << 20)

	if err = x64.AllocateDMA(dmaSize); err != nil {
		return
	}

	start := uint64(dma.Default().Start())
	end := uint64(dma.Default().End())

	// disable encryption for DMA region
	return sev.SetEncryptedBit(x64.AMD64, start, end, features.EncryptedBit, false)
}

func initGHCB() (err error) {
	var ghcbAddr uint64

	if ghcb != nil {
		return
	}

	if secrets == nil || secrets.Version == 0 {
		return errors.New("AMD SEV-SNP secrets unavailable, run `sev` first")
	}

	// OVMF allocates 2*ncpu contiguous pages, a first shared page for GHCB
	// (which we re-use) and a second private one for vCPU variables.
	if ghcbAddr = x64.AMD64.MSR(sev.MSR_AMD_GHCB); ghcbAddr == 0 {
		return errors.New("could not find GHCB address")
	}

	ghcbGPA := uint(ghcbAddr)
	ghcb = &sev.GHCB{}

	if ghcb.GHCBPage, err = dma.NewRegion(ghcbGPA, uefi.PageSize, false); err != nil {
		return fmt.Errorf("could not allocate GHCB layout page, %v", err)
	}

	if err = ghcb.Init(false); err != nil {
		return fmt.Errorf("could not initialize GHCB, %v", err)
	}

	if err = initSharedDMA(1 << 20); err != nil {
		return fmt.Errorf("could not allocate shared DMA region, %v", err)
	}

	ghcb.Region = dma.Default()

	return
}

func attestationCmd(_ *shell.Interface, arg []string) (res string, err error) {
	var buf bytes.Buffer
	var report *sev.AttestationReport

	if err = initGHCB(); err != nil {
		return "", fmt.Errorf("could not initialize GHCB, %v", err)
	}

	data := make([]byte, 64)
	rand.Read(data)

	if report, err = ghcb.GetAttestationReport(data, secrets.VMPCK0[:], 0); err != nil {
		return "", fmt.Errorf("could not get report, %v", err)
	}

	if len(arg[0]) > 0 {
		return fmt.Sprintf("%x", report.Bytes()), nil
	}

	fmt.Fprintf(&buf, "Version ............: %x\n", report.Version)
	fmt.Fprintf(&buf, "VMPL ...............: %x\n", report.VMPL)
	fmt.Fprintf(&buf, "SignatureAlgo ......: %x\n", report.SignatureAlgo)
	fmt.Fprintf(&buf, "CurrentTCB .........: %x\n", report.CurrentTCB)
	fmt.Fprintf(&buf, "Measurement ........: %x\n", report.Measurement)
	fmt.Fprintf(&buf, "ReportedTCB ........: %x\n", report.ReportedTCB)
	fmt.Fprintf(&buf, "CommittedTCB .......: %x\n", report.CommittedTCB)
	fmt.Fprintf(&buf, "Launch  Mitigations : %#x\n", report.LaunchMitVector)
	fmt.Fprintf(&buf, "Current Mitigations : %#x\n", report.CurrentMitVector)
	fmt.Fprintf(&buf, "SignatureR .........: %x\n", report.Signature[0:48])
	fmt.Fprintf(&buf, "SignatureS .........: %x\n", report.Signature[72:72+48])

	return buf.String(), nil
}

func kdfCmd(_ *shell.Interface, arg []string) (res string, err error) {
	var key []byte

	if err = initGHCB(); err != nil {
		return "", fmt.Errorf("could not initialize GHCB, %v", err)
	}

	// Perform key derivation from the physical CPU keys, bound to the
	// Guest Policy and VM identity.
	req := &sev.KeyRequest{
		KeySelect:        sev.RootKeySelVCEK,
		GuestFieldSelect: sev.GuestSVN | sev.Measurement | sev.GuestPolicy,
	}

	if key, err = ghcb.DeriveKey(req, secrets.VMPCK0[:], 0); err != nil {
		return "", fmt.Errorf("could not derive key, %v", err)
	}

	return fmt.Sprintf("%x\n", key), nil
}

func tscCmd(_ *shell.Interface, arg []string) (res string, err error) {
	var buf bytes.Buffer
	var tsc *sev.TSCInfo

	if err = initGHCB(); err != nil {
		return "", fmt.Errorf("could not initialize GHCB, %v", err)
	}

	if tsc, err = ghcb.TSCInfo(secrets.VMPCK0[:], 0); err != nil {
		return "", fmt.Errorf("could not request TSC, %v", err)
	}

	fmt.Fprintf(&buf, "Guest TSC Scale ....: %d\n", tsc.GuestTSCScale)
	fmt.Fprintf(&buf, "Guest TSC Offset ...: %d\n", tsc.GuestTSCOffset)
	fmt.Fprintf(&buf, "TSC Factor .........: %d\n", tsc.TSCFactor)

	return buf.String(), nil
}
