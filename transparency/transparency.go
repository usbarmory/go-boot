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
	"strings"

	"github.com/usbarmory/boot-transparency/artifact"
	"github.com/usbarmory/boot-transparency/engine/sigsum"
	"github.com/usbarmory/boot-transparency/policy"
	"github.com/usbarmory/boot-transparency/transparency"
	"github.com/usbarmory/go-boot/uefi/x64"
)

// Represents the status of the transparency functionality.
type Status int

// Represents transparency status codes.
const (
	// Transparency disabled.
	None Status = iota

	// Transparency enabled in offline mode.
	Offline

	// Transparency enabled in online mode.
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

// Config represents the configuration for the transparency functionality.
type Config struct {
	// Status represents the status of the transparency functionality.
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
	// Category should be consistent with the artifact categories supported
	// by boot-transparency.
	Category uint

	// SHA-256 hash of the artifact
	Hash string
}

// Transparency configuration filenames
const (
	transparencyRoot  = `/transparency`
	bootPolicyFile    = `policy.json`
	witnessPolicyFile = `trust_policy`
	proofBundleFile   = `proof-bundle.json`
	submitKeyFile     = `submit-key.pub`
	logKeyFile        = `log-key.pub`
)

// Hash performs data hashing via the same algoritm
// used by boot-transparency library (i.e. SHA-256).
// Returns the computed hash as hex string.
func Hash(data *[]byte) (hexHash string) {
	h := sha256.New()
	h.Write(*data)
	hash := h.Sum(nil)

	return hex.EncodeToString(hash)
}

// Validate the transparency inclusion proof and the consistency
// between the boot policy and the logged artifact claims.
// The function takes as input the pointers to the transparency
// configuration and to the artifacts requirements which contain the
// file hashes for the loaded boot artifacts.
// Returns error if the boot artifacts are not passing the validation.
func Validate(c *Config, a *[]Artifact) (err error) {
	if c.Status == None {
		return
	}

	if a == nil && len (*a) != 0 {
		return fmt.Errorf("got an invalid pointer to boot artifacts")
	}

	// Get the tranparency configuration path for the given boot entry (i.e. set of artifacts)
	entryPath := c.EntryPath(a)

	// Load the configuration from disk
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

	if c.Status == Online {
		// Probe the log to obtain a fresh inclusion proof.
		pr, err := te.GetProof(pb)
		if err != nil {
			return err
		}

		freshBundle := pb.(*sigsum.ProofBundle)
		freshBundle.Proof = string(pr)

		// If network access is available the inclusion proof verification
		// is performed using the proof fetched from the log.
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

	if err = validateProofHashes(claims, a); err != nil {
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

// EntryPath returns the unique configuration loading path
// for a given boot entry (i.e. set of artifacts)
func (c *Config) EntryPath(a *[]Artifact) (entryPath string) {
	entryPath = transparencyRoot

	for _, artifact := range *a {
		entryPath = path.Join(entryPath, artifact.Hash)
	}

	return strings.ReplaceAll(entryPath, `/`, `\`)
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



// Validate the matching between the file hashes of the boot artifacts
// and the ones claimed in the proof bundle.
// This step is vital to ensure the correspondency between the artifacts
// loaded in memory during the boot and the claims that will be validated
// by the boot-transparency policy function.
func validateProofHashes(s *policy.Statement, artifacts *[]Artifact) (err error) {
	for _, a := range *artifacts {
		if err = validateClaimedHash(s, a); err != nil {
			return err
		}
	}

	return
}

func validateClaimedHash(s *policy.Statement, a Artifact) (err error) {
	var h artifact.Handler
	var found = false

	for _, claimedArtifact := range s.Artifacts {
		if a.Category == claimedArtifact.Category {
			if h, err = artifact.GetHandler(a.Category); err != nil {
				return
			}

			// boot-transparency expect to parse requirements in JSON format
			requirements, _ := json.Marshal(map[string]string{"file_hash": a.Hash})
			r, err := h.ParseRequirements([]byte(requirements))
			if err != nil {
				return err
			}

			c, err := h.ParseClaims([]byte(claimedArtifact.Claims))
			if err != nil {
				return err
			}

			if err = h.Validate(r, c); err != nil {
				return fmt.Errorf("loaded boot artifacts do not correspond to the proof bundle ones, file hash mistmatch")
			}

			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("loaded boot artifacts do not correspond to the proof bundle ones, one or more artifacts are not present in the proof bundle")
	}

	return
}
