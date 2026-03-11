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

	"github.com/usbarmory/boot-transparency/policy"
	"github.com/usbarmory/boot-transparency/transparency"
)

// Validate applies boot-transparency validation (e.g. inclusion proof,
// boot policy and claims consistency) for the argument [Config] representing
// the boot transparency configuration.
// Returns error if the boot artifacts are not passing the validation.
func Validate(c *Config, b *policy.BootEntry) (err error) {
	if c.Status == None {
		return
	}

	if len(*b) == 0 {
		return fmt.Errorf("invalid boot entry")
	}

	// Automatically load the configuration from the configured root filesystem.
	if c.Root != nil {
		entryPath, err := c.Path(b)
		if err != nil {
			return fmt.Errorf("cannot load boot-transparency configuration, %v", err)
		}

		if err = c.loadFromRoot(entryPath); err != nil {
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

	// Validate the matching between the logged claims and the policy requirements,
	// fot the given set of boot artifacts.
	if err = policy.Validate(requirements, claims, b); err != nil {
		// The boot bundle is NOT authorized.
		return
	}

	// boot-transparency validation passed, boot bundle is authorized.
	return
}
