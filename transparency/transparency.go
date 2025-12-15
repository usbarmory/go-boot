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

	"github.com/usbarmory/boot-transparency/artifact"
	"github.com/usbarmory/boot-transparency/engine/sigsum"
	"github.com/usbarmory/boot-transparency/policy"
	"github.com/usbarmory/boot-transparency/transparency"
	"github.com/usbarmory/go-boot/uefi/x64"
)

// Represents boot-transparency status.
type BtStatus int

// Represents boot-transparency status codes.
const (
	// Transparency disabled.
	None BtStatus = iota

	// Transparency enabled in offline mode.
	Offline

	// Transparency enabled in online mode.
	Online
)

// ToString resolves BtStatus codes into a human-readable strings.
func (s BtStatus) ToString() string {
	var btStatusName = map[BtStatus]string{
		None:    "none",
		Offline: "offline",
		Online:  "online",
	}

	return btStatusName[s]
}

// BtConfig represents boot-transparency configuration.
type BtConfig struct {
	// Status represents boot-transparency status.
	Status BtStatus

	// BootPolicy represents the boot policy in JSON format
	// following the boot-transparency policy syntax.
	BootPolicy []byte

	// WitnessPolicy represents the witness policy following
	// the Sigsum plaintext witness policy format.
	WitnessPolicy []byte

	// ProofBundle represents the proof bundle in JSON format
	// following the boot-transparency proof bundle syntax.
	ProofBundle []byte

	// SubmitKey represents the log submitter public key.
	SubmitKey []byte

	// LogKey represents the log public key.
	LogKey []byte
}

// BtArtifact represents boot-transparency requirements for a boot artifact.
type BtArtifact struct {
	// Category should be consistent with the artifact categories supported
	// by boot-transparency.
	Category uint

	// Requirements are expressed in JSON format, following the same key:value
	// syntax supported by boot-transparency to define boot policy requirements.
	Requirements []byte
}

// boot-transparency configuration filenames
const (
	transparencyRoot  = `/transparency`
	bootPolicyFile    = `policy.json`
	witnessPolicyFile = `trust_policy`
	proofBundleFile   = `proof-bundle.json`
	submitKeyFile     = `submit-key.pub`
	logKeyFile        = `log-key.pub`
)

// Load reads the boot-transparency configuration files from disk.
// The entry argument allows per-bundle configurations.
func (c *BtConfig) Load(entry string) (err error) {
	entryPath := path.Join(transparencyRoot, entry)

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

// Validate the transparency inclusion proof and the consistency
// between the boot policy and the logged artifact claims.
// The function takes as input the pointers to the boot-transparency
// configuration and to the artifacts requirements which contain the
// file hashes for the loaded boot artifacts.
// Returns error if the boot artifacts are not passing the validation.
func Validate(c *BtConfig, a *[]BtArtifact) (err error) {
	if c.Status == None {
		return
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

	if err = validateArtifacts(claims, a); err != nil {
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

// Validate the matching between the boot artifacts and the ones included
// in the proof bundle.
// This step is vital to ensure the correspondency between the artifacts
// loaded in memory during the boot and the claims that will be validated
// by the boot-transparency policy function.
func validateArtifacts(s *policy.Statement, btArtifacts *[]BtArtifact) (err error) {
	var h artifact.Handler

	for _, btArtifact := range *btArtifacts {
		found := false
		for _, a := range s.Artifacts {
			if btArtifact.Category == a.Category {
				if h, err = artifact.GetHandler(a.Category); err != nil {
					return
				}

				r, err := h.ParseRequirements([]byte(btArtifact.Requirements))
				if err != nil {
					return err
				}

				c, err := h.ParseClaims([]byte(a.Claims))
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
	}

	return
}
