// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package transparency implements an interface to the
// boot-transparency library functions to ease boot bundle
// validation.
package transparency

import (
	"testing"

	"github.com/usbarmory/boot-transparency/artifact"
)

func TestPath(t *testing.T) {
	kernelHash := "70672136965536be27980489b0388d864c96c74efd73d21432d0bcf10f9269f3"
	initrdHash := "337630b74e55eae241f460faadf5a2f9a2157d6de2853d4106c35769e4acf538"
	expectedPath := "/transparency/70672136965536be27980489b0388d864c96c74efd73d21432d0bcf10f9269f3/337630b74e55eae241f460faadf5a2f9a2157d6de2853d4106c35769e4acf538"

	c := Config{
		Status: Offline,
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

	p, err := c.Path(&b)
	if err != nil {
		t.Fatal(err)
	}

	if p != expectedPath {
		t.Fatal("got an invalid path.")
	}
}

func TestPathInvalidHash(t *testing.T) {
	// Invalid sha-256 hash.
	kernelHash := "0672136965536be27980489b0388d864c96c74efd73d21432d0bcf10f9269f3"
	initrdHash := "337630b74e55eae241f460faadf5a2f9a2157d6de2853d4106c35769e4acf538"

	c := Config{
		Status: Offline,
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

	// Error expected due to the invalid hash in the test entry.
	if _, err := c.Path(&b); err == nil {
		t.Fatal(err)
	}
}

func TestLoadFromUefiRoot(t *testing.T) {
	c := Config{
		Status:   Offline,
		UefiRoot: testUefiRoot,
	}

	if err := c.loadFromUefiRoot(""); err != nil {
		t.Fatal(err)
	}
}
