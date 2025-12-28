// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uefi

import (
	"errors"
)

const snpSignature = 0x45444d41

// AMD SEV-ES Guest-Hypervisor Communication Block Standardization
var EFI_SEV_SNP_CC_BLOB_GUID = MustParseGUID("067b1f5f-cf26-44c5-8554-93d777912d42")

// SNPConfigurationTable represents an EFI SNP Confidential Computing Blob
// Configuration Table.
type SNPConfigurationTable struct {
	Header                     uint32
	Version                    uint16
	_                          uint16
	SecretsPagePhysicalAddress uint64
	SecretsPageSize            uint32
	_                          uint32
	CPUIDPagePhysicalAddress   uint64
	CPUIDPageSize              uint32
	_                          uint32
}

// GetSNPConfiguration returns the EFI SNP Confidential Computing Blob
// Configuration Table.
func (s *Services) GetSNPConfiguration() (snp *SNPConfigurationTable, err error) {
	var t *ConfigurationTable

	if s.SystemTable == nil {
		return nil, errors.New("EFI System Table is invalid")
	}

	if t, err = s.SystemTable.LocateConfiguration(EFI_SEV_SNP_CC_BLOB_GUID); err != nil {
		return
	}

	snp = &SNPConfigurationTable{}

	if err = decode(&snp, t.VendorTable); err != nil {
		return nil, err
	}

	if snp.Header != snpSignature || snp.Version < 2 {
		return snp, errors.New("EFI SNP Configuration Table is invalid")
	}

	return
}
