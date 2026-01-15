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

	"github.com/usbarmory/boot-transparency/artifact"
	"github.com/usbarmory/boot-transparency/engine/sigsum"
	"github.com/usbarmory/boot-transparency/policy"
	"github.com/usbarmory/boot-transparency/transparency"
)

// Artifact represents a boot artifact.
type Artifact struct {
	// Category represents the artifact category as defined
	// in the boot-transparency library.
	Category uint

	// Hash represents the SHA-256 hash of the artifact.
	Hash string
}

// BootEntry represent a boot entry as a set of artifacts.
type BootEntry []Artifact

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

	proof := pb.(*sigsum.ProofBundle)

	claims, err := policy.ParseStatement(proof.Statement)
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

// Hash performs data hashing using SHA-256, which is
// the same algorithm used by boot-transparency library.
// Returns the computed hash as hex string.
func Hash(data *[]byte) (hexHash string) {
	h := sha256.New()
	h.Write(*data)
	hash := h.Sum(nil)

	return hex.EncodeToString(hash)
}

// Validate the matching between loaded artifact hashes and
// the ones included in the proof bundle.
// This step is vital to ensure the correspondence between the artifacts
// loaded in memory during the boot and the claims that will be validated
// by the boot-transparency policy function.
func (b BootEntry) validateProofHashes(s *policy.Statement) (err error) {
	for _, a := range b {
		if err = a.validateProofHash(s); err != nil {
			return err
		}
	}

	return
}

func (a Artifact) validateProofHash(s *policy.Statement) (err error) {
	var h artifact.Handler
	var found bool

	if err = a.hasValidHash(); err != nil {
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

func (a Artifact) hasValidHash() (err error) {
	h, err := hex.DecodeString(a.Hash)

	if err != nil || len(h) != sha256.Size {
		return fmt.Errorf("invalid artifact hash")
	}

	return
}
