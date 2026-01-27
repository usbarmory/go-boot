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
	"encoding/json"
	"errors"
	"fmt"

	"github.com/usbarmory/boot-transparency/artifact"
	"github.com/usbarmory/boot-transparency/policy"
	"github.com/usbarmory/boot-transparency/transparency"
)

// Artifact represents a boot artifact.
type Artifact struct {
	// Category represents the artifact category as defined
	// in the boot-transparency library.
	Category uint

	// Hash represents the SHA256 checksum of the artifact.
	Hash []byte
}

// BootEntry represents a boot entry as a set of artifacts.
type BootEntry []Artifact

// ErrHashMismatch represents an hash mismatch error.
var ErrHashMismatch = errors.New("hash mismatch")
// ErrHashInvalid represents an hash invalid error.
var ErrHashInvalid = errors.New("hash invalid")

// Validate applies boot-transparency validation (e.g. inclusion proof,
// boot policy and claims consistency) for the argument [Config] representing
// the boot artifacts.
// Returns error if the boot artifacts are not passing the validation.
func (b BootEntry) Validate(c *Config) (err error) {
	if c.Status == None {
		return
	}

	if len(b) == 0 {
		return fmt.Errorf("invalid boot entry")
	}

	// Automatically load the configuration from the UEFI partition
	// when the function is used within the UEFI boot loader.
	if c.UefiRoot != nil {
		entryPath, err := c.Path(b)
		if err != nil {
			return fmt.Errorf("cannot load boot-transparency configuration, %v", err)
		}

		if err = c.loadFromUefiRoot(entryPath); err != nil {
			return fmt.Errorf("cannot load boot-transparency configuration, %v", err)
		}
	}

	te, err := transparency.GetEngine(c.Engine)
	if err != nil {
		return fmt.Errorf("unable to get transparency engine, %v", err)
	}

	if err = te.SetKey(c.LogKey, c.SubmitKey); err != nil {
		return fmt.Errorf("unable to set log and submitter keys, %v", err)
	}

	if err = te.SetWitnessPolicy(c.WitnessPolicy); err != nil {
		return fmt.Errorf("unable to set witness policy, %v", err)
	}

	format, statement, proof, probe, _, err := transparency.ParseProofBundle(c.ProofBundle)
	if err != nil {
		return fmt.Errorf("unable to parse the proof bundle, %v", err)
	}

	if format != c.Engine {
		return fmt.Errorf("proof bundle format doesn't match the transparency engine.")
	}

	// If network access is available the inclusion proof verification
	// is performed using the proof fetched from the log.
	if c.Status == Online {
		proof, err = te.GetProof(statement, probe)
		if err != nil {
			return err
		}
	}

	if err = te.VerifyProof(statement, proof, nil); err != nil {
		return err
	}

	requirements, err := policy.ParseRequirements(c.BootPolicy)
	if err != nil {
		return
	}

	claims, err := policy.ParseStatement(statement)
	if err != nil {
		return
	}

	// Validate the matching between loaded artifact hashes and
	// the ones included in the proof bundle.
	if err = b.validateProofHashes(claims); err != nil {
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

func (b BootEntry) validateProofHashes(s *policy.Statement) (err error) {
	for _, a := range b {
		if err = a.validateProofHash(s); err != nil {
			return err
		}
	}

	return
}

// Validate the matching between loaded artifact hash and the one included
// in the proof bundle.
// This step is vital to ensure the correspondence between the artifacts
// loaded in memory during the boot and the claims that will be validated
// by the boot-transparency policy function.
func (a Artifact) validateProofHash(s *policy.Statement) (err error) {
	var h artifact.Handler
	var found bool

	if err = a.validHash(); err != nil {
		return
	}

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
		requirements, _ := json.Marshal(map[string]string{"file_hash": hex.EncodeToString(a.Hash)})

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
			return fmt.Errorf("%w for artifact category %d, hash %q", ErrHashMismatch, a.Category, hex.EncodeToString(a.Hash))
		}

		found = true
		break
	}

	if !found {
		return fmt.Errorf("one or more artifacts are not present in the proof bundle")
	}

	return
}

func (a Artifact) validHash() (err error) {
	if len(a.Hash) != artifact.HashSize {
		err = fmt.Errorf("%w for artifact category %d", ErrHashInvalid, a.Category)
	}

	return
}
