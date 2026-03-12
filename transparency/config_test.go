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
	"github.com/usbarmory/boot-transparency/policy"
	"github.com/usbarmory/boot-transparency/transparency"
)

func TestPath(t *testing.T) {
	c := Config{
		Status: Offline,
		Engine: transparency.Sigsum,
	}

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

	p, err := c.Path(&b)
	if err != nil {
		t.Fatal(err)
	}

	if p != testEntryPath {
		t.Fatal("got an invalid path.")
	}
}

func TestLoadFromRoot(t *testing.T) {
	c := Config{
		Status: Offline,
		Engine: transparency.Sigsum,
		Root:   testRoot,
	}

	if err := c.loadFromRoot(""); err != nil {
		t.Fatal(err)
	}
}

func TestLoadFromRootWithPathPrefix(t *testing.T) {
	c := Config{
		Status:     Offline,
		Engine:     transparency.Sigsum,
		Root:       testRootWithPrefix,
		PathPrefix: `transparency`,
	}

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

	entryPath, err := c.Path(&b)
	if err != nil {
		t.Fatal(err)
	}

	if err = c.loadFromRoot(entryPath); err != nil {
		t.Fatal(err)
	}
}

func TestLoadFromRootInCorrectPathPrefix(t *testing.T) {
	c := Config{
		Status:     Offline,
		Engine:     transparency.Sigsum,
		Root:       testRootWithPrefix,
		PathPrefix: `FOO`,
	}

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

	entryPath, err := c.Path(&b)
	if err != nil {
		t.Fatal(err)
	}

	if err = c.loadFromRoot(entryPath); err == nil {
		t.Fatal("missing error due to incorrect path prefix")
	}
}
