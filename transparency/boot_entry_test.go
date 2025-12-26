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
	kernelHash := "4551848b4ab43cb4321c4d6ba98e1d215f950cee21bfd82c8c82ab64e34ec9a6"
	initrdHash := "337630b74e55eae241f460faadf5a2f9a2157d6de2853d4106c35769e4acf538"

	c := Config{
		Status: Offline,

		BootPolicy:    testBootPolicy,
		WitnessPolicy: testWitnessPolicy,
		SubmitKey:     testSubmitKey,
		LogKey:        testLogKey,
		ProofBundle:   testProofBundle,
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
	kernelHash := "4551848b4ab43cb4321c4d6ba98e1d215f950cee21bfd82c8c82ab64e34ec9a6"
	initrdHash := "337630b74e55eae241f460faadf5a2f9a2157d6de2853d4106c35769e4acf538"

	c := Config{
		Status: Online,

		BootPolicy:    testBootPolicy,
		WitnessPolicy: testWitnessPolicy,
		SubmitKey:     testSubmitKey,
		LogKey:        testLogKey,
		ProofBundle:   testProofBundle,
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
	kernelHash := "4551848b4ab43cb4321c4d6ba98e1d215f950cee21bfd82c8c82ab64e34ec9a6"

	c := Config{
		Status: Offline,

		BootPolicy:    testBootPolicy,
		WitnessPolicy: testWitnessPolicy,
		SubmitKey:     testSubmitKey,
		LogKey:        testLogKey,
		ProofBundle:   testProofBundle,
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
	kernelHash := "aabbccddeeffaabbccddeeffaabbccddeeffaabbccddeeffaabbccddeeffaabb"
	initrdHash := "337630b74e55eae241f460faadf5a2f9a2157d6de2853d4106c35769e4acf538"

	c := Config{
		Status: Offline,

		BootPolicy:    testBootPolicy,
		WitnessPolicy: testWitnessPolicy,
		SubmitKey:     testSubmitKey,
		LogKey:        testLogKey,
		ProofBundle:   testProofBundle,
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
	kernelHash := "4551848b4ab43cb4321c4d6ba98e1d215f950cee21bfd82c8c82ab64e34ec9a6"
	initrdHash := "337630b74e55eae241f460faadf5a2f9a2157d6de2853d4106c35769e4acf538"

	c := Config{
		Status: Offline,

		BootPolicy:    testBootPolicyUnauthorized,
		WitnessPolicy: testWitnessPolicy,
		SubmitKey:     testSubmitKey,
		LogKey:        testLogKey,
		ProofBundle:   testProofBundle,
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
