// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"strings"

	"github.com/usbarmory/go-boot/shell"
	"github.com/usbarmory/go-boot/transparency"
	"github.com/usbarmory/go-boot/uefi/x64"
)

var btConfig transparency.BtConfig

const (
	transparencyRoot  = `/transparency`
	bootPolicyFile    = `policy.json`
	witnessPolicyFile = `trust_policy`
	proofBundleFile   = `proof-bundle.json`
	submitKeyFile     = `submit-key.pub`
	logKeyFile        = `log-key.pub`
)

func init() {
	shell.Add(shell.Cmd{
		Name:    "bt",
		Args:    1,
		Pattern: regexp.MustCompile(`^(?:bt)( none| offline| online)?$`),
		Syntax:  "(none|offline|online)?",
		Help:    "show/set boot-transparency status",
		Fn:      btCmd,
	})
}

func btCmd(_ *shell.Interface, arg []string) (res string, err error) {
	if len(arg[0]) > 0 {
		switch strings.TrimSpace(arg[0]) {
		case "none":
			btConfig.Status = transparency.None
		case "offline":
			btConfig.Status = transparency.Offline
		case "online":
			btConfig.Status = transparency.Online
		}
	}

	switch btConfig.Status {
	case transparency.None:
		return fmt.Sprintf("boot-transparency is disabled\n"), nil
	case transparency.Offline, transparency.Online:
		return fmt.Sprintf("boot-transparency is enabled in %s mode\n", btConfig.Status.ToString()), nil
	}

	return
}

// btLoadConfig loads the boot-transparency configuration from files on disk,
// the entryPath argument allows per-bundle configurations.
func btLoadConfig(entry string) (err error) {
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

	if btConfig.BootPolicy, err = fs.ReadFile(root, bootPolicyPath); err != nil {
		return fmt.Errorf("cannot read boot policy, %v", err)
	}

	if btConfig.WitnessPolicy, err = fs.ReadFile(root, witnessPolicyPath); err != nil {
		return fmt.Errorf("cannot read witness policy, %v", err)
	}

	if btConfig.SubmitKey, err = fs.ReadFile(root, submitKeyPath); err != nil {
		return fmt.Errorf("cannot read log submitter key, %v", err)
	}

	if btConfig.LogKey, err = fs.ReadFile(root, logKeyPath); err != nil {
		return fmt.Errorf("cannot read log key, %v", err)
	}

	if btConfig.ProofBundle, err = fs.ReadFile(root, proofBundlePath); err != nil {
		return fmt.Errorf("cannot read proof bundle, %v", err)
	}

	return
}
