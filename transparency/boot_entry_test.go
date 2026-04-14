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
	_ "github.com/usbarmory/boot-transparency/engine/sigsum"
	_ "github.com/usbarmory/boot-transparency/engine/tessera"
	"github.com/usbarmory/boot-transparency/policy"
	"github.com/usbarmory/boot-transparency/transparency"
)

var testConfig = &Config{
	Status: Offline,
	Engine: transparency.Sigsum,

	BootPolicy:    []byte(testBootPolicy),
	WitnessPolicy: []byte(testWitnessPolicy),
	SubmitKey:     []byte(testSubmitKey),
	LogKey:        []byte(testLogKey),
	ProofBundle:   []byte(testProofBundle),
}

func TestOfflineValidate(t *testing.T) {
	testConfig.Status = Offline

	b := &policy.BootEntry{
		Artifacts: []policy.BootArtifact{
			policy.BootArtifact{
				Category: artifact.LinuxKernel,
				Data:     []byte(testKernel),
			},
			policy.BootArtifact{
				Category: artifact.Initrd,
				Data:     []byte(testInitrd),
			},
		},
	}

	if err := Validate(testConfig, b); err != nil {
		t.Fatal(err)
	}
}

func TestOnlineValidate(t *testing.T) {
	testConfig.Status = Online

	be := &policy.BootEntry{
		Artifacts: []policy.BootArtifact{
			policy.BootArtifact{
				Category: artifact.LinuxKernel,
				Data:     []byte(testKernel),
			},
			policy.BootArtifact{
				Category: artifact.Initrd,
				Data:     []byte(testInitrd),
			},
		},
	}

	if err := Validate(testConfig, be); err != nil {
		t.Fatal(err)
	}
}

func TestOfflineValidateInvalidBootEntry(t *testing.T) {
	testConfig.Status = Offline

	be := &policy.BootEntry{
		Artifacts: []policy.BootArtifact{
			policy.BootArtifact{
				Category: artifact.LinuxKernel,
				Data:     []byte(testKernel),
			},
			policy.BootArtifact{
				Category: artifact.Initrd,
				// missing Data
			},
		},
	}

	// Error expected: invalid boot entry error.
	if err := Validate(testConfig, be); err == nil || !errors.Is(err, policy.ErrInvalidBootEntry) {
		t.Fatal("missing invalid boot entry error")
	}
}

func TestOfflineValidateHashMismatch(t *testing.T) {
	testConfig.Status = Offline

	be := &policy.BootEntry{
		Artifacts: []policy.BootArtifact{
			policy.BootArtifact{
				Category: artifact.LinuxKernel,
				Data:     []byte(testIncorrectKernel),
			},
			policy.BootArtifact{
				Category: artifact.Initrd,
				Data:     []byte(testInitrd),
			},
		},
	}

	// Error expected: incorrect hash.
	if err := Validate(testConfig, be); err == nil || !errors.Is(err, policy.ErrValidate) {
		t.Fatal("missing incorrect hash error")
	}
}

func TestOfflineValidatePolicyNotMet(t *testing.T) {
	testConfig.Status = Offline
	testConfig.BootPolicy = []byte(testBootPolicyUnauthorized)

	be := &policy.BootEntry{
		Artifacts: []policy.BootArtifact{
			policy.BootArtifact{
				Category: artifact.LinuxKernel,
				Data:     []byte(testKernel),
			},
			policy.BootArtifact{
				Category: artifact.Initrd,
				Data:     []byte(testInitrd),
			},
		},
	}

	// Error expected: requirement not met.
	if err := Validate(testConfig, be); err == nil || !errors.Is(err, policy.ErrValidate) {
		t.Fatal("missing policy validation error")
	}
}
