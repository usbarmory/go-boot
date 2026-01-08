// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package transparency implements an interface to the
// boot-transparency library functions to ease boot bundle
// validation.
package transparency

import (
	"regexp"
	"testing"

	"github.com/usbarmory/boot-transparency/artifact"
)

func TestOfflineValidate(t *testing.T) {
	c := Config{
		Status: Offline,

		BootPolicy:    []byte(testBootPolicy),
		WitnessPolicy: []byte(testWitnessPolicy),
		SubmitKey:     []byte(testSubmitKey),
		LogKey:        []byte(testLogKey),
		ProofBundle:   []byte(testProofBundle),
	}

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
	c := Config{
		Status: Online,

		BootPolicy:    []byte(testBootPolicy),
		WitnessPolicy: []byte(testWitnessPolicy),
		SubmitKey:     []byte(testSubmitKey),
		LogKey:        []byte(testLogKey),
		ProofBundle:   []byte(testProofBundle),
	}

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
	c := Config{
		Status: Offline,

		BootPolicy:    []byte(testBootPolicy),
		WitnessPolicy: []byte(testWitnessPolicy),
		SubmitKey:     []byte(testSubmitKey),
		LogKey:        []byte(testLogKey),
		ProofBundle:   []byte(testProofBundle),
	}

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

	err := b.Validate(&c)

	// Error expected: missing required Hash.
	if err == nil {
		t.Fatal(err)
	}

	if !regexp.MustCompile(`invalid artifact hash`).MatchString(err.Error()) {
		t.Fatal(err)
	}
}

func TestOfflineValidateHashMismatch(t *testing.T) {
	c := Config{
		Status: Offline,

		BootPolicy:    []byte(testBootPolicy),
		WitnessPolicy: []byte(testWitnessPolicy),
		SubmitKey:     []byte(testSubmitKey),
		LogKey:        []byte(testLogKey),
		ProofBundle:   []byte(testProofBundle),
	}

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

	err := b.Validate(&c)

	// Error expected: incorrect hash.
	if err == nil {
		t.Fatal(err)
	}

	if !regexp.MustCompile(`file hash mismatch`).MatchString(err.Error()) {
		t.Fatal(err)
	}
}

func TestOfflineValidatePolicyNotMet(t *testing.T) {
	c := Config{
		Status: Offline,

		BootPolicy:    []byte(testBootPolicyUnauthorized),
		WitnessPolicy: []byte(testWitnessPolicy),
		SubmitKey:     []byte(testSubmitKey),
		LogKey:        []byte(testLogKey),
		ProofBundle:   []byte(testProofBundle),
	}

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
	err := b.Validate(&c)

	if err == nil {
		t.Fatal(err)
	}

	if !regexp.MustCompile(`build args requirement .+ not met`).MatchString(err.Error()) {
		t.Fatal(err)
	}
}
