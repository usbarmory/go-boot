// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build transparency

package transparency

import (
	"fmt"
	"io/fs"

	// Maintained set of TLD roots for any potential TLS client request
	_ "golang.org/x/crypto/x509roots/fallback"

	"github.com/usbarmory/boot-transparency/engine/sigsum"
	"github.com/usbarmory/boot-transparency/policy"
	"github.com/usbarmory/boot-transparency/transparency"
)

const (
	// Boot-transparency paths
	bootPolicyPath    = "\\transparency\\policy.json"
	witnessPolicyPath = "\\transparency\\trust_policy"
	proofBundlePath   = "\\transparency\\proof-bundle.json"
	submitKeyPath     = "\\transparency\\submit-key.pub"
	logKeyPath        = "\\transparency\\log-key.pub"
)

func Check(fsys fs.FS, onLine bool) (err error) {
	bootPolicy, err := fs.ReadFile(fsys, bootPolicyPath)
	if err != nil {
		return fmt.Errorf("cannot read boot policy, %v", err)
	}

	witnessPolicy, err := fs.ReadFile(fsys, witnessPolicyPath)
	if err != nil {
		return fmt.Errorf("cannot read witness policy, %v", err)
	}

	submitKey, err := fs.ReadFile(fsys, submitKeyPath)
	if err != nil {
		return fmt.Errorf("cannot read log submitter key, %v", err)
	}

	logKey, err := fs.ReadFile(fsys, logKeyPath)
	if err != nil {
		return fmt.Errorf("cannot read log key, %v", err)
	}

	proofBundle, err := fs.ReadFile(fsys, proofBundlePath)
	if err != nil {
		return fmt.Errorf("cannot read proof bundle, %v", err)
	}

	// Select Sigsum as transparency engine.
	te, err := transparency.GetEngine(transparency.Sigsum)
	if err != nil {
		return fmt.Errorf("unable to configure the transparency engine, %w", err)
	}

	// Set public keys.
	err = te.SetKey([]string{string(logKey)}, []string{string(submitKey)})
	if err != nil {
		return err
	}

	// Parse witness policy.
	wp, err := te.ParseWitnessPolicy(witnessPolicy)
	if err != nil {
		return err
	}

	// Set witness policy.
	err = te.SetWitnessPolicy(wp)
	if err != nil {
		return err
	}

	// Parse the proof bundle, which is expected to contain
	// the logged statement and its inclusion proof, or the probe
	// data to request the inclusion proof when operating in
	// on-line mode.
	pb, _, err := te.ParseProof(proofBundle)
	if err != nil {
		return err
	}

	if onLine {
		// Probe the log to obtain a fresh inclusion proof.
		pr, err := te.GetProof(pb)
		if err != nil {
			return err
		}

		freshBundle := pb.(*sigsum.ProofBundle)
		freshBundle.Proof = string(pr)

		// Inclusion proof verification,
		// use the fresh inclusion proof obtained from the log, include
		// verification of the co-signing quorum as defined in the witness policy.
		err = te.VerifyProof(freshBundle)
		if err != nil {
			return err
		}
	} else {
		// Inclusion proof verification, including the co-signing quorum verification
		// as defined in the witness policy.
		err = te.VerifyProof(pb)
		if err != nil {
			return err
		}
	}

	// Parse the boot policy requirements.
	r, err := policy.ParseRequirements(bootPolicy)
	if err != nil {
		return err
	}

	// Convert to the proof bundle type expected by the selected engine.
	b := pb.(*sigsum.ProofBundle)

	// Parse the statement included in the proof bundle.
	c, err := policy.ParseStatement(b.Statement)
	if err != nil {
		return err
	}

	// Check if the logged claims are matching the policy requirements.
	if err = policy.Check(r, c); err != nil {
		// The boot bundle is NOT authorized for boot.
		return err
	}

	// All boot-transparency checks passed.
	return
}
