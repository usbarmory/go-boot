// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package transparency implements an interface to the
// boot-transparency library functions to ease boot bundle
// validation.
package transparency

import (
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
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

	// UefiRoot represents the UEFI filesystem to "automatically" load the
	// configuration files when running within the boot loader context.
	// If the transparency pkg is imported "externally" by user-space tools
	// this field is not set.
	UefiRoot fs.FS

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
		if err = artifact.hasValidHash(); err != nil {
			return "", fmt.Errorf("cannot build configuration path, %v", err)
		}

		entryPath = path.Join(entryPath, artifact.Hash)
	}

	// Rewrite paths only when the pkg is used in the context
	// of the UEFI boot loader.
	if c.UefiRoot != nil {
		entryPath = strings.ReplaceAll(entryPath, `/`, `\`)
	}

	return
}

// loadFromUefiRoot reads the transparency configuration files from
// the UEFI partition. The entry argument allows per-bundle configurations.
func (c *Config) loadFromUefiRoot(entryPath string) (err error) {
	assets := map[string]*[]byte{
		bootPolicy:    &c.BootPolicy,
		witnessPolicy: &c.WitnessPolicy,
		submitKey:     &c.SubmitKey,
		logKey:        &c.LogKey,
		proofBundle:   &c.ProofBundle,
	}

	if c.UefiRoot == nil {
		return fmt.Errorf("cannot open uefi root filesystem")
	}

	for filename, dst := range assets {
		p := path.Join(entryPath, filename)
		p = strings.ReplaceAll(p, `/`, `\`)

		if *dst, err = fs.ReadFile(c.UefiRoot, p); err != nil {
			return fmt.Errorf("cannot load configuration file: %v", filename)
		}
	}

	return
}
