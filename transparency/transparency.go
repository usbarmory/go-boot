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

	"github.com/usbarmory/boot-transparency/artifact"
	"github.com/usbarmory/boot-transparency/engine/sigsum"
	"github.com/usbarmory/boot-transparency/policy"
	"github.com/usbarmory/boot-transparency/transparency"
)

// BtStatus represents boot-transparency status codes
type BtStatus int

const (
    None BtStatus = iota
    Offline
    Online
)

// BtStatusName represents boot-transparency status names
var BtStatusName = map[BtStatus]string {
    None:    "none",
    Offline: "offline",
    Online:  "online",
}

// BtConfig represents boot-transparency configuration
type BtConfig struct {
	Status        BtStatus
	BootPolicy    []byte
	WitnessPolicy []byte
	ProofBundle   []byte
	SubmitKey     []byte
	LogKey        []byte
}

// BtArtifact represents boot-transparency requirements for a boot artifact.
// Requirements are expressed in JSON format, following the same key:value
// syntax supported by boot-transparency to define boot policy requirements.
// Category should be consistent with the artifact categories supported
// by boot-transparency.
type BtArtifact struct {
        Category     uint
        Requirements []byte
}

// Validate validates the transparency inclusion proof and the consistency
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

	err = te.SetKey([]string{string(c.LogKey)}, []string{string(c.SubmitKey)})
	if err != nil {
		return
	}

	wp, err := te.ParseWitnessPolicy(c.WitnessPolicy)
	if err != nil {
		return
	}

	err = te.SetWitnessPolicy(wp)
	if err != nil {
		return
	}

	pb, _, err := te.ParseProof(c.ProofBundle)
	if err != nil {
		return
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
		err = te.VerifyProof(freshBundle)
		if err != nil {
			return err
		}
	} else {
		// If network access is not available the inclusion proof verification
		// is performed using the proof included in the proof bundle.
		err = te.VerifyProof(pb)
		if err != nil {
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
                                h, err = artifact.GetHandler(a.Category)
                                if err != nil {
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

                                err = h.Validate(r, c)
                                if err != nil {
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
