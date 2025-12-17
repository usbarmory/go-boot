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
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/usbarmory/boot-transparency/artifact"
	"github.com/usbarmory/boot-transparency/engine/sigsum"
	"github.com/usbarmory/boot-transparency/policy"
	"github.com/usbarmory/boot-transparency/transparency"
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

// Artifact represents a boot artifact.
type Artifact struct {
	// Category must be consistent with the artifact categories
	// supported by boot-transparency library.
	Category uint

	// SHA-256 hash of the artifact.
	Hash string
}

// BootEntry represent a boot antry as a set of artifacts.
type BootEntry []Artifact

// Boot transparency configuration filenames.
const (
	transparencyRoot  = `/transparency`
	bootPolicyFile    = `policy.json`
	witnessPolicyFile = `trust_policy`
	proofBundleFile   = `proof-bundle.json`
	submitKeyFile     = `submit-key.pub`
	logKeyFile        = `log-key.pub`
)

// Hash performs data hashing using SHA-256, which is
// the same algorithm used by boot-transparency library.
// Returns the computed hash as hex string.
func Hash(data *[]byte) (hexHash string) {
	h := sha256.New()
	h.Write(*data)
	hash := h.Sum(nil)

	return hex.EncodeToString(hash)
}

// ConfigPath returns a unique configuration loading path
// for a given set of artifacts (i.e. boot entry).
// Returns error if one of the artifacts does not include
// a valid SHA-256 hash.
func (b *BootEntry) ConfigPath() (entryPath string, err error) {
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

	return strings.ReplaceAll(entryPath, `/`, `\`), nil
}

// Validate the transparency inclusion proof and, the consistency
// between the boot policy and the logged claims for the boot artifacts.
// Takes as input the pointers to the transparency configuration and,
// to the information on the loaded boot artifacts.
// Returns error if the boot artifacts are not passing the validation.
func Validate(entry *BootEntry, c *Config) (err error) {
	if c.Status == None {
		return
	}

	if entry == nil || len(*entry) == 0 {
		return fmt.Errorf("got an invalid boot entry pointer")
	}

	// Get the boot transparency configuration path for a given
	// boot entry.
	entryPath, err := entry.ConfigPath()
	if err != nil {
		return
	}

	// Loads the configuration from disk.
	if err = c.load(entryPath); err != nil {
		return fmt.Errorf("cannot load boot transparency configuration, %v", err)
	}

	te, err := transparency.GetEngine(transparency.Sigsum)
	if err != nil {
		return fmt.Errorf("unable to configure the transparency engine, %w", err)
	}

	if err = te.SetKey([]string{string(c.LogKey)}, []string{string(c.SubmitKey)}); err != nil {
		return fmt.Errorf("unable to set log and submitter keys, %v", err)
	}

	wp, err := te.ParseWitnessPolicy(c.WitnessPolicy)
	if err != nil {
		return fmt.Errorf("unable to parse witness policy, %v", err)
	}

	if err = te.SetWitnessPolicy(wp); err != nil {
		return fmt.Errorf("unable to set witness policy, %v", err)
	}

	pb, _, err := te.ParseProof(c.ProofBundle)
	if err != nil {
		return fmt.Errorf("unable to parse proof bundle, %v", err)
	}

	// If network access is available the inclusion proof verification
	// is performed using the proof fetched from the log.
	if c.Status == Online {
		pr, err := te.GetProof(pb)
		if err != nil {
			return err
		}

		freshBundle := pb.(*sigsum.ProofBundle)
		freshBundle.Proof = string(pr)

		if err = te.VerifyProof(freshBundle); err != nil {
			return err
		}
	} else {
		// If network access is not available the inclusion proof verification
		// is performed using the proof included in the proof bundle.
		if err = te.VerifyProof(pb); err != nil {
			return err
		}
	}

	requirements, err := policy.ParseRequirements(c.BootPolicy)
	if err != nil {
		return
	}

	b := pb.(*sigsum.ProofBundle)

	claims, err := policy.ParseStatement(b.Statement)
	if err != nil {
		return
	}

	// Validate the matching between loaded artifact hashes and
	// the ones included in the proof bundle.
	if err = validateProofHashes(claims, entry); err != nil {
		return
	}

	// Validate the matching between the logged claims and the policy requirements.
	if err = policy.Validate(requirements, claims); err != nil {
		// The boot bundle is NOT authorized.
		return
	}

	// boot-transparency validation passed, boot bundle is authorized.
	return
}

// Load reads the transparency configuration files from disk.
// The entry argument allows per-bundle configurations.
func (c *Config) load(entryPath string) (err error) {
	bootPolicyPath := path.Join(entryPath, bootPolicyFile)
	bootPolicyPath = strings.ReplaceAll(bootPolicyPath, `/`, `\`)

	witnessPolicyPath := path.Join(entryPath, witnessPolicyFile)
	witnessPolicyPath = strings.ReplaceAll(witnessPolicyPath, `/`, `\`)

	submitKeyPath := path.Join(entryPath, submitKeyFile)
	submitKeyPath = strings.ReplaceAll(submitKeyPath, `/`, `\`)

	logKeyPath := path.Join(entryPath, logKeyFile)
	logKeyPath = strings.ReplaceAll(logKeyPath, `/`, `\`)

	proofBundlePath := path.Join(entryPath, proofBundleFile)
	proofBundlePath = strings.ReplaceAll(proofBundlePath, `/`, `\`)

	root, err := x64.UEFI.Root()
	if err != nil {
		return fmt.Errorf("could not open root volume, %v", err)
	}

	if c.BootPolicy, err = fs.ReadFile(root, bootPolicyPath); err != nil {
		return fmt.Errorf("cannot read boot policy, %v", err)
	}

	if c.WitnessPolicy, err = fs.ReadFile(root, witnessPolicyPath); err != nil {
		return fmt.Errorf("cannot read witness policy, %v", err)
	}

	if c.SubmitKey, err = fs.ReadFile(root, submitKeyPath); err != nil {
		return fmt.Errorf("cannot read log submitter key, %v", err)
	}

	if c.LogKey, err = fs.ReadFile(root, logKeyPath); err != nil {
		return fmt.Errorf("cannot read log key, %v", err)
	}

	if c.ProofBundle, err = fs.ReadFile(root, proofBundlePath); err != nil {
		return fmt.Errorf("cannot read proof bundle, %v", err)
	}

	return
}

// Validate the matching between loaded artifact hashes and
// the ones included in the proof bundle.
// This step is vital to ensure the correspondence between the artifacts
// loaded in memory during the boot and the claims that will be validated
// by the boot-transparency policy function.
func validateProofHashes(s *policy.Statement, b *BootEntry) (err error) {
	for _, a := range *b {
		if err = validateClaimedHash(s, a); err != nil {
			return err
		}
	}

	return
}

func validateClaimedHash(s *policy.Statement, a Artifact) (err error) {
	var h artifact.Handler
	var found bool

	for _, claimedArtifact := range s.Artifacts {
		// The claims are referring to a different artifact
		// category, try with next block of claims in the statement.
		if a.Category != claimedArtifact.Category {
			continue
		}

		if h, err = artifact.GetHandler(a.Category); err != nil {
			return
		}

		// boot-transparency expect to parse requirements in JSON format.
		requirements, _ := json.Marshal(map[string]string{"file_hash": a.Hash})

		r, err := h.ParseRequirements([]byte(requirements))
		if err != nil {
			return err
		}

		c, err := h.ParseClaims([]byte(claimedArtifact.Claims))
		if err != nil {
			return err
		}

		// The validation logic is safe in the sense that error is returned
		// if a file hash requested by the boot loader is not present in the
		// statement for a given artifact category.
		if err = h.Validate(r, c); err != nil {
			return fmt.Errorf("loaded boot artifacts do not correspond to the proof bundle ones, file hash mismatch")
		}

		found = true
		break
	}

	if !found {
		return fmt.Errorf("loaded boot artifacts do not correspond to the proof bundle ones, one or more artifacts are not present in the proof bundle")
	}

	return
}
