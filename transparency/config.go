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

	// BootPolicy represents the boot policy in JSON format
	// following the policy syntax supported by boot-transparency library.
	BootPolicy []byte

	// WitnessPolicy represents the witness policy following
	// the Sigsum plaintext witness policy format.
	WitnessPolicy []byte

	// ProofBundle represents the proof bundle in JSON format
	// following the proof bundle format supported by boot-transparency library.
	ProofBundle []byte

	// SubmitKey represents the log submitter public key in·OpenSSH·format.
	SubmitKey []byte

	// LogKey represents the log public key in OpenSSH format.
	LogKey []byte
}

// Load reads the transparency configuration files from disk.
// The entry argument allows per-bundle configurations.
func (c *Config) load(entryPath string) (err error) {
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
