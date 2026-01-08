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

	p, err := c.Path(b)
	if err != nil {
		t.Fatal(err)
	}

	if p != entryPath {
		t.Fatal("got an invalid path.")
	}
}

func TestPathInvalidHash(t *testing.T) {
	c := Config{
		Status: Offline,
	}

	b := BootEntry{
		Artifact{
			Category: artifact.LinuxKernel,
			Hash:     invalidKernelHash,
		},
		Artifact{
			Category: artifact.Initrd,
			Hash:     initrdHash,
		},
	}

	// Error expected due to the invalid hash in the test entry.
	if _, err := c.Path(b); err == nil {
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
