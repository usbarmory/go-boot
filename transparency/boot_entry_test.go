// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package transparency implements an interface to the
// boot-transparency library functions to ease boot bundle
// validation.
package transparency

import (
	"errors"
	"testing"

	"github.com/usbarmory/boot-transparency/artifact"
	"github.com/usbarmory/boot-transparency/transparency"
	"github.com/usbarmory/boot-transparency/policy"
	_ "github.com/usbarmory/boot-transparency/engine/sigsum"
	_ "github.com/usbarmory/boot-transparency/engine/tessera"
)

var testConfig Config

func init() {
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

	b := policy.BootEntry{
		policy.BootArtifact{
			Category: artifact.LinuxKernel,
			Data:     []byte(testKernel),
		},
		policy.BootArtifact{
			Category: artifact.Initrd,
			Data:     []byte(testInitrd),
		},
	}

	if err := Validate(&c, &b); err != nil {
		t.Fatal(err)
	}
}

func TestOnlineValidate(t *testing.T) {
	c := testConfig
	c.Status = Online

	b := policy.BootEntry{
		policy.BootArtifact{
			Category: artifact.LinuxKernel,
			Data:     []byte(testKernel),
		},
		policy.BootArtifact{
			Category: artifact.Initrd,
			Data:     []byte(testInitrd),
		},
	}

	if err := Validate(&c, &b); err != nil {
		t.Fatal(err)
	}
}

func TestOfflineValidateInvalidBootEntry(t *testing.T) {
	c := testConfig

	b := policy.BootEntry{
		policy.BootArtifact{
			Category: artifact.LinuxKernel,
			Data:     []byte(testKernel),
		},
		policy.BootArtifact{
			Category: artifact.Initrd,
			// missing Data
		},
	}

	// Error expected: invalid boot entry error.
	if err := Validate(&c, &b); err == nil || !errors.Is(err, policy.ErrInvalidBootEntry) {
		t.Fatal("missing invalid boot entry error")
	}
}

func TestOfflineValidateHashMismatch(t *testing.T) {
	c := testConfig

	b := policy.BootEntry{
		policy.BootArtifact{
			Category: artifact.LinuxKernel,
			Data:     []byte(testIncorrectKernel),
		},
		policy.BootArtifact{
			Category: artifact.Initrd,
			Data:     []byte(testInitrd),
		},
	}

	// Error expected: incorrect hash.
	if err := Validate(&c, &b); err == nil || !errors.Is(err, policy.ErrValidate) {
		t.Fatal("missing incorrect hash error")
	}
}

func TestOfflineValidatePolicyNotMet(t *testing.T) {
	c := testConfig
	c.BootPolicy = []byte(testBootPolicyUnauthorized)

	b := policy.BootEntry{
		policy.BootArtifact{
			Category: artifact.LinuxKernel,
			Data:     []byte(testKernel),
		},
		policy.BootArtifact{
			Category: artifact.Initrd,
			Data:     []byte(testInitrd),
		},
	}

	// Error expected: requirement not met.
	if err := Validate(&c, &b); err == nil || !errors.Is(err, policy.ErrValidate) {
		t.Fatal("missing policy validation error")
	}
}
