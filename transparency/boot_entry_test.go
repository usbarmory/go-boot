// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package transparency implements an interface to the
// boot-transparency library functions to ease boot bundle
// validation.
package transparency

import (
	"encoding/hex"
	"errors"
	"testing"

	"github.com/usbarmory/boot-transparency/artifact"
	"github.com/usbarmory/boot-transparency/transparency"
	"github.com/usbarmory/boot-transparency/policy"
	_ "github.com/usbarmory/boot-transparency/engine/sigsum"
	_ "github.com/usbarmory/boot-transparency/engine/tessera"
)

var kernelHash []byte
var initrdHash []byte
var incorrectKernelHash []byte
var testConfig Config

func init() {
	kernelHash, _ = hex.DecodeString(testKernelHash)
	incorrectKernelHash, _ = hex.DecodeString(testIncorrectKernelHash)
	initrdHash, _ = hex.DecodeString(testInitrdHash)

	testConfig = Config{
		Status: Offline,
		Engine: transparency.Sigsum,

		BootPolicy:    []byte(testBootPolicy),
		WitnessPolicy: []byte(testWitnessPolicy),
		SubmitKey:     []byte(testSubmitKey),
		LogKey:        []byte(testLogKey),
		ProofBundle:   []byte(testProofBundle),
	}
}

func TestOfflineValidate(t *testing.T) {
	c := testConfig

	b := BootEntry{
		Artifact{
			Category: artifact.LinuxKernel,
			Hash:     kernelHash,
		},
		Artifact{
			Category: artifact.Initrd,
			Hash:     initrdHash,
		},
	}

	if err := b.Validate(&c); err != nil {
		t.Fatal(err)
	}
}

func TestOnlineValidate(t *testing.T) {
	c := testConfig
	c.Status = Online

	b := BootEntry{
		Artifact{
			Category: artifact.LinuxKernel,
			Hash:     kernelHash,
		},
		Artifact{
			Category: artifact.Initrd,
			Hash:     initrdHash,
		},
	}

	if err := b.Validate(&c); err != nil {
		t.Fatal(err)
	}
}

func TestOfflineValidateInvalidBootEntry(t *testing.T) {
	c := testConfig

	b := BootEntry{
		Artifact{
			Category: artifact.LinuxKernel,
			Hash:     kernelHash,
		},
		Artifact{
			Category: artifact.Initrd,
			// missing Hash
		},
	}

	// Error expected: missing required Hash.
	if err := b.Validate(&c); err == nil || !errors.Is(err, ErrHashInvalid) {
		t.Fatal("missing invalid hash error")
	}
}

func TestOfflineValidateHashMismatch(t *testing.T) {
	c := testConfig

	b := BootEntry{
		Artifact{
			Category: artifact.LinuxKernel,
			Hash:     incorrectKernelHash,
		},
		Artifact{
			Category: artifact.Initrd,
			Hash:     initrdHash,
		},
	}

	// Error expected: incorrect hash.
	if err := b.Validate(&c); err == nil || !errors.Is(err, ErrHashMismatch) {
		t.Fatal("missing incorrect hash error")
	}
}

func TestOfflineValidatePolicyNotMet(t *testing.T) {
	c := testConfig
	c.BootPolicy = []byte(testBootPolicyUnauthorized)

	b := BootEntry{
		Artifact{
			Category: artifact.LinuxKernel,
			Hash:     kernelHash,
		},
		Artifact{
			Category: artifact.Initrd,
			Hash:     initrdHash,
		},
	}

	// Error expected: requirement not met.
	if err := b.Validate(&c); err == nil || !errors.Is(err, policy.ErrValidate) {
		t.Fatal("missing policy validation error")
	}
}
