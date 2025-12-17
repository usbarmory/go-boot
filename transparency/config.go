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

// Boot transparency configuration filenames.
const (
	transparencyRoot = `/transparency`

	bootPolicyFile    = `policy.json`
	witnessPolicyFile = `trust_policy`
	proofBundleFile   = `proof-bundle.json`
	submitKeyFile     = `submit-key.pub`
	logKeyFile        = `log-key.pub`
)

// Load reads the transparency configuration files from disk.
// The entry argument allows per-bundle configurations.
func (c *Config) load(entryPath string) (err error) {
	root, err := x64.UEFI.Root()

	if err != nil {
		return fmt.Errorf("could not open root volume, %v", err)
	}

	bootPolicyPath := path.Join(entryPath, bootPolicyFile)
	bootPolicyPath = strings.ReplaceAll(bootPolicyPath, `/`, `\`)

	if c.BootPolicy, err = fs.ReadFile(root, bootPolicyPath); err != nil {
		return fmt.Errorf("cannot read boot policy, %v", err)
	}

	witnessPolicyPath := path.Join(entryPath, witnessPolicyFile)
	witnessPolicyPath = strings.ReplaceAll(witnessPolicyPath, `/`, `\`)

	if c.WitnessPolicy, err = fs.ReadFile(root, witnessPolicyPath); err != nil {
		return fmt.Errorf("cannot read witness policy, %v", err)
	}

	submitKeyPath := path.Join(entryPath, submitKeyFile)
	submitKeyPath = strings.ReplaceAll(submitKeyPath, `/`, `\`)

	if c.SubmitKey, err = fs.ReadFile(root, submitKeyPath); err != nil {
		return fmt.Errorf("cannot read log submitter key, %v", err)
	}

	logKeyPath := path.Join(entryPath, logKeyFile)
	logKeyPath = strings.ReplaceAll(logKeyPath, `/`, `\`)

	if c.LogKey, err = fs.ReadFile(root, logKeyPath); err != nil {
		return fmt.Errorf("cannot read log key, %v", err)
	}

	proofBundlePath := path.Join(entryPath, proofBundleFile)
	proofBundlePath = strings.ReplaceAll(proofBundlePath, `/`, `\`)

	if c.ProofBundle, err = fs.ReadFile(root, proofBundlePath); err != nil {
		return fmt.Errorf("cannot read proof bundle, %v", err)
	}

	return
}
