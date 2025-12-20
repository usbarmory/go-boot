// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package transparency implements an interface to the
// boot-transparency library functions to ease boot bundle
// validation.
package transparency

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/usbarmory/go-boot/uefi/x64"
)

// Represents the status of the boot transparency functionality.
type Status int

// Represents boot transparency status codes.
const (
	// Boot transparency disabled.
	None Status = iota

	// Boot transparency enabled in offline mode.
	Offline

	// Boot transparency enabled in online mode.
	Online
)

// Resolve resolves Status codes into a human-readable strings.
func (s Status) Resolve() string {
	statusName := map[Status]string{
		None:    "none",
		Offline: "offline",
		Online:  "online",
	}

	return statusName[s]
}

// Boot transparency configuration root directory and filenames.
const (
	transparencyRoot = `/transparency`

	bootPolicy    = `policy.json`
	witnessPolicy = `trust_policy`
	proofBundle   = `proof-bundle.json`
	submitKey     = `submit-key.pub`
	logKey        = `log-key.pub`
)

// Config represents the configuration for the boot transparency functionality.
type Config struct {
	// Status represents the status of the boot transparency functionality.
	Status Status

	// ExternalLoader is set to true when the transparency pkg is used
	// externally to a boot loader context (e.g. installers or user-space tools).
	// In such cases, the configuration files are not loaded automatically
	// from the UEFI partition during artifact(s) validation.
	ExternalLoader bool

	// BootPolicy represents the boot policy in JSON format
	// following the policy syntax supported by boot-transparency library.
	BootPolicy []byte

	// WitnessPolicy represents the witness policy following
	// the Sigsum plaintext witness policy format.
	WitnessPolicy []byte

	// SubmitKey represents the log submitter public key in·OpenSSH·format.
	SubmitKey []byte

	// LogKey represents the log public key in OpenSSH format.
	LogKey []byte

	// ProofBundle represents the proof bundle in JSON format
	// following the proof bundle format supported by boot-transparency library.
	ProofBundle []byte
}

// Path returns a unique configuration path for a given set of
// artifacts (i.e. boot entry).
// Returns error if one of the artifacts does not include a valid
// SHA-256 hash.
func (c *Config) Path(b *BootEntry) (entryPath string, err error) {
	if len(*b) == 0 {
		return "", fmt.Errorf("cannot build configuration path, got an invalid boot entry pointer")
	}

	artifacts := *b

	// Sort the passed artifacts, by their Category, to ensure
	// consistency in the way the entry path is build.
	sort.Slice(artifacts, func(i, j int) bool {
		return artifacts[i].Category < artifacts[j].Category
	})

	entryPath = transparencyRoot
	for _, artifact := range artifacts {
		h, err := hex.DecodeString(artifact.Hash)

		if err != nil || len(h) != sha256.Size {
			return "", fmt.Errorf("cannot build configuration path, got an invalid artifact hash")
		}

		entryPath = path.Join(entryPath, artifact.Hash)
	}

	// Do not rewrite paths when the pkg is used externally to
	// the UEFI boot loader (i.e. installers or user-space tools).
	if c.ExternalLoader {
		entryPath = strings.ReplaceAll(entryPath, `/`, `\`)
	}

	return
}

// loadFromUefiPart reads the transparency configuration files from
// the UEFI partition. The entry argument allows per-bundle configurations.
func (c *Config) loadFromUefiPart(entryPath string) (err error) {
	root, err := x64.UEFI.Root()

	if err != nil {
		return fmt.Errorf("could not open root volume, %v", err)
	}

	assets := map[string]*[]byte{
		bootPolicy:    &c.BootPolicy,
		witnessPolicy: &c.WitnessPolicy,
		submitKey:     &c.SubmitKey,
		logKey:        &c.LogKey,
		proofBundle:   &c.ProofBundle,
	}

	for filename, dst := range assets {
		p := path.Join(entryPath, filename)
		p = strings.ReplaceAll(p, `/`, `\`)

		if *dst, err = fs.ReadFile(root, p); err != nil {
			return fmt.Errorf("cannot load configuration file: %v", filename)
		}
	}

	return
}
