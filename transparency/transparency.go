// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package transparency

import (
	"fmt"
	"io/fs"

	"github.com/usbarmory/go-boot/uefi/x64"
	"github.com/usbarmory/boot-transparency/engine/sigsum"
	"github.com/usbarmory/boot-transparency/policy"
	"github.com/usbarmory/boot-transparency/transparency"
)

// Boot transparency configuration.
var Config struct {
	Status string

	bootPolicy    []byte
	witnessPolicy []byte
	proofBundle   []byte
	submitKey     []byte
	logKey        []byte
}

const (
	bootPolicyPath    = `\transparency\policy.json`
	witnessPolicyPath = `\transparency\trust_policy`
	proofBundlePath   = `\transparency\proof-bundle.json`
	submitKeyPath     = `\transparency\submit-key.pub`
	logKeyPath        = `\transparency\log-key.pub`
)

func init() {
	Config.Status = "none"
}

// Load boot-transparency configuration from files on disk.
func LoadConfig() (err error) {
	root, err := x64.UEFI.Root()

	if err != nil {
		return fmt.Errorf("could not open root volume, %v", err)
	}

	Config.bootPolicy, err = fs.ReadFile(root, bootPolicyPath)
	if err != nil {
		return fmt.Errorf("cannot read boot policy, %v", err)
	}

	Config.witnessPolicy, err = fs.ReadFile(root, witnessPolicyPath)
	if err != nil {
		return fmt.Errorf("cannot read witness policy, %v", err)
	}

	Config.submitKey, err = fs.ReadFile(root, submitKeyPath)
	if err != nil {
		return fmt.Errorf("cannot read log submitter key, %v", err)
	}

	Config.logKey, err = fs.ReadFile(root, logKeyPath)
	if err != nil {
		return fmt.Errorf("cannot read log key, %v", err)
	}

	Config.proofBundle, err = fs.ReadFile(root, proofBundlePath)
	if err != nil {
		return fmt.Errorf("cannot read proof bundle, %v", err)
	}

	return
}

// Cleanup boot-transparency configuration.
func CleanupConfig() {
	Config.bootPolicy = nil
	Config.witnessPolicy = nil
	Config.submitKey = nil
	Config.logKey = nil
	Config.proofBundle = nil
}

// Validate the transparency inclusion proof and consistency
// between the boot policy and the logged claims.
func Validate() (err error) {
	te, err := transparency.GetEngine(transparency.Sigsum)
	if err != nil {
		return fmt.Errorf("unable to configure the transparency engine, %w", err)
	}

	err = te.SetKey([]string{string(Config.logKey)}, []string{string(Config.submitKey)})
	if err != nil {
		return
	}

	wp, err := te.ParseWitnessPolicy(Config.witnessPolicy)
	if err != nil {
		return
	}

	err = te.SetWitnessPolicy(wp)
	if err != nil {
		return
	}

	// Parse the proof bundle, which is expected to contain
	// the logged statement and its inclusion proof, or the probe
	// data to request the inclusion proof when operating with
	// network access enabled.
	pb, _, err := te.ParseProof(Config.proofBundle)
	if err != nil {
		return
	}

	if Config.Status == "online" {
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

	r, err := policy.ParseRequirements(Config.bootPolicy)
	if err != nil {
		return
	}

	b := pb.(*sigsum.ProofBundle)

	c, err := policy.ParseStatement(b.Statement)
	if err != nil {
		return
	}

	// Check if the logged claims are matching the policy requirements.
	if err = policy.Check(r, c); err != nil {
		// The boot bundle is NOT authorized for boot.
		return
	}

	// All boot-transparency checks passed.
	return
}
