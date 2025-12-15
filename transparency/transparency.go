// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package transparency

import (
	"fmt"

	"github.com/usbarmory/boot-transparency/artifact"
	"github.com/usbarmory/boot-transparency/engine/sigsum"
	"github.com/usbarmory/boot-transparency/policy"
	"github.com/usbarmory/boot-transparency/transparency"
)

// Defines boot-transparency status codes
type BtStatus int

const (
    None BtStatus = iota
    Offline
    Online
)

// Defines boot-transparency status names
var BtStatusName = map[BtStatus]string {
    None:    "none",
    Offline: "offline",
    Online:  "online",
}

// Defines boot-transparency configuration
type BtConfig struct {
	Status        BtStatus
	BootPolicy    []byte
	WitnessPolicy []byte
	ProofBundle   []byte
	SubmitKey     []byte
	LogKey        []byte
}

// Defines boot-transparency requirements for a given boot artifact
type BtArtifact struct {
        Category     uint
        Requirements []byte
}

// Validate the inclusion proof and the consistency between the
// boot policy and the logged claims.
// The function takes as input the pointers to the boot-transparency
// configuration and file hashes of the boot artifacts.
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

	// Parse the proof bundle, which is expected to contain
	// the logged statement and its inclusion proof.
	// The probe data is optionally required to request the inclusion
	// proof only when operating in online mode.
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

		// Inclusion proof verification with network access:
		// use the inclusion proof fetched from the log.
		err = te.VerifyProof(freshBundle)
		if err != nil {
			return err
		}
	} else {
		// Inclusion proof verification without network access:
		// use the inclusion proof included in the proof bundle.
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

// Ensure the matching between the boot artifacts and the ones included into a given proof bundle.
// This step is vital to ensure the correspondency between the artifacts actually
// loaded during the boot and the claims that will be validated by the  boot-transparency
// policy function.
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
