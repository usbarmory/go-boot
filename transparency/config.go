// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package transparency implements an interface to the
// boot-transparency library functions to ease boot bundle
// validation.
package transparency

import (
	"encoding/hex"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"

	"github.com/usbarmory/boot-transparency/policy"
	"github.com/usbarmory/boot-transparency/transparency"
)

// Status represents the status of the boot-transparency functionality.
type Status int

// Status codes.
const (
	// Boot-transparency disabled.
	None Status = iota

	// Boot-transparency enabled in offline mode.
	Offline

	// Boot-transparency enabled in online mode.
	Online
)

// String resolves Status codes into a human-readable strings.
func (s Status) String() string {
	statusName := map[Status]string{
		None:    "none",
		Offline: "offline",
		Online:  "online",
	}

	return statusName[s]
}

// Boot-transparency configuration root directory and filenames.
const (
	// DefaultPathPrefix represents the default prefix to the asset paths.
	DefaultPathPrefix = `transparency`

	bootPolicy    = `policy.json`
	witnessPolicy = `trust_policy`
	proofBundle   = `proof-bundle.json`
	submitKey     = `submit-key.pub`
	logKey        = `log-key.pub`
)

// Config represents the configuration for the boot-transparency functionality.
type Config struct {
	// Engine represents the transparency engine chosen among the ones
	// supported by boot-transparency library.
	Engine transparency.EngineCode

	// Status represents the status of the boot-transparency functionality.
	Status Status

	// Root represents the filesystem root directory where the transparency assets
	// are stored.
	Root fs.FS

	// PathPrefix represent the directory used to load the boot entry
	// assets from their correspondent unique paths (see [*Config Path()]).
	// If left not configured, by default the [DefaultPathPrefix] is pre-pended
	// to all the asset paths.
	PathPrefix string

	// BootPolicy represents the boot policy in JSON format
	// following the policy syntax supported by boot-transparency library.
	BootPolicy []byte

	// WitnessPolicy represents the witness policy following
	// the Sigsum plaintext witness policy format.
	WitnessPolicy []byte

	// SubmitKey represents the log submitter public keys.
	// The format should match the one(s) supported by the
	// chosen transparency engine.
	SubmitKey []byte

	// LogKey represents the log public keys.
	// The format should match the one(s) supported by the
	// chosen transparency engine.
	LogKey []byte

	// ProofBundle represents the proof bundle in JSON format
	// following the proof bundle format supported by boot-transparency library.
	ProofBundle []byte
}

// Path returns a unique configuration path for a given set of
// artifacts (i.e. boot entry).
// Returns error if one of the artifacts does not include a valid
// SHA-256 hash.
func (c *Config) Path(be *policy.BootEntry) (entryPath string, err error) {
	if len(be.Artifacts) == 0 {
		return "", fmt.Errorf("invalid boot entry")
	}

	// Sort the passed artifacts, by their Category, to ensure
	// consistency in the way the entry path is built.
	sort.Slice(be.Artifacts, func(i, j int) bool {
		return be.Artifacts[i].Category < be.Artifacts[j].Category
	})

	entryPath = DefaultPathPrefix
	if c.PathPrefix != "" {
		entryPath = c.PathPrefix
	}

	for _, artifact := range be.Artifacts {
		entryPath = filepath.Join(entryPath, hex.EncodeToString(artifact.Hash()))
	}

	return
}

// loadFromRoot reads the transparency configuration files from
// the disk root filesystem. The entry argument allows per-bundle configurations.
func (c *Config) loadFromRoot(entryPath string) (err error) {
	assets := map[string]*[]byte{
		bootPolicy:    &c.BootPolicy,
		witnessPolicy: &c.WitnessPolicy,
		submitKey:     &c.SubmitKey,
		logKey:        &c.LogKey,
		proofBundle:   &c.ProofBundle,
	}

	if c.Root == nil {
		return fmt.Errorf("cannot open root filesystem")
	}

	for filename, dst := range assets {
		p := filepath.Join(entryPath, filename)

		if *dst, err = fs.ReadFile(c.Root, p); err != nil {
			return fmt.Errorf("cannot load configuration file: %v", filename)
		}
	}

	return
}
