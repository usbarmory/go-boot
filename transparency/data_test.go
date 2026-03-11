// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package transparency implements an interface to the
// boot-transparency library functions to ease boot bundle
// validation.
package transparency

import (
	"testing/fstest"
)

const (
	testKernel = `test linux kernel`
	testInitrd = `test initrd`
	testIncorrectKernel = `test incorrect linux kernel`
	testEntryPath       = "transparency/5e6d8e01d75e3e0396d672b0e8c3e31f78532eef9fa2a3f464299ee7cc44a12e/b868d20383e979c588e7b16d24b9d3fcb9c1213c89135e6c656edf94cbf31542"

	testBootPolicy = `[
	{
		"artifacts": [
			{
				"category": 1,
				"requirements": {
					"min_version": "v6.14.0-29",
					"tainted": false,
					"build_args": {
						"CONFIG_STACKPROTECTOR_STRONG": "y"
					}
				}
			},
			{
				"category": 2,
				"requirements": {
					"tainted": false
				}
			}
		],
		"signatures": {
			"signers": [
				{
					"name": "dist signatory I",
					"pub_key": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIP5rbNcIOcwqBHzLOhJEfdKFHa+pIs10idfTm8c+HDnK"
				},
				{
					"name": "dist signatory II",
			 		"pub_key": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIL0zV5fSWzzXa4R7Kpk6RAXkvWsJGpvkQ+9/xxpHC49J"
				}
			],
			"quorum": 2
		}
	}
]`

	testWitnessPolicy = `log 4644af2abd40f4895a003bca350f9d5912ab301a49c77f13e5b6d905c20a5fe6 https://test.sigsum.org/barreleye

witness poc.sigsum.org/nisse 1c25f8a44c635457e2e391d1efbca7d4c2951a0aef06225a881e46b98962ac6c
witness rgdd.se/poc-witness  28c92a5a3a054d317c86fc2eeb6a7ab2054d6217100d0be67ded5b74323c5806

group  demo-quorum-rule any poc.sigsum.org/nisse rgdd.se/poc-witness
quorum demo-quorum-rule
`

	testProofBundle = `{
	"format": 1,
	"statement": {
		"header": {
			"description": "Linux bundle",
			"revision": "v1"
		},
		"artifacts": [
			{
				"category": 1,
				"claims": {
					"file_name": "test-vmlinuz-6.14.0-29-generic",
					"file_hash": "5e6d8e01d75e3e0396d672b0e8c3e31f78532eef9fa2a3f464299ee7cc44a12e",
					"version": "v6.14.0-29-generic",
					"tainted": false,
					"build_args": {
						"CONFIG_STACKPROTECTOR_STRONG": "y"
					}
				}
			},
			{
				"category": 2,
				"claims": {
					"file_name": "test-initrd.img-6.14.0-29-generic",
					"file_hash": "b868d20383e979c588e7b16d24b9d3fcb9c1213c89135e6c656edf94cbf31542",
					"version": "v6.14.0-29-generic",
					"tainted": false
				}
			}
		],
		"signatures": [
			{
				"pub_key": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIP5rbNcIOcwqBHzLOhJEfdKFHa+pIs10idfTm8c+HDnK",
				"signature": "8d984b482ab45de5a2f0171a338b6ce8e64d95a70d6ea14b9d2a5f772c21d339d4cd51091b8f4c93f6dc289ee32ad94d048c8badb4fc3cc0a3136bfb4886ba0f"
			},
			{
				"pub_key": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIL0zV5fSWzzXa4R7Kpk6RAXkvWsJGpvkQ+9/xxpHC49J",
				"signature": "99af82c7c559cf98902ac299f10a608353f2447729115667e924719d4ffb2b3cf80dc6715959afafaa5d255ae45e880351245cba6233ae093716136670b7d409"
			}
		]
	},
	"probe": {
		"origin": "https://test.sigsum.org/barreleye",
		"leaf_signature": "cf4e40c8bef8eb2742327ba8037bc531ca8929a58b38ba7ca4c48152eff3f62d8c9c71a720d4ea601bfe47a5c543409e2e0116958230a5e2e9356c123464920c",
		"log_public_key_hash": "4e89cc51651f0d95f3c6127c15e1a42e3ddf7046c5b17b752689c402e773bb4d",
		"submit_public_key_hash": "302928c2e0e01da52e3b161c54906de9b55ce250f0f47e80e022d04036e2765c"
	},
	"proof": "version=2\nlog=4e89cc51651f0d95f3c6127c15e1a42e3ddf7046c5b17b752689c402e773bb4d\nleaf=302928c2e0e01da52e3b161c54906de9b55ce250f0f47e80e022d04036e2765c cf4e40c8bef8eb2742327ba8037bc531ca8929a58b38ba7ca4c48152eff3f62d8c9c71a720d4ea601bfe47a5c543409e2e0116958230a5e2e9356c123464920c\n\nsize=162296\nroot_hash=2f63a7e021ae6bf254df038a89439c0867cb0b6845f235f2c52cb1f556ac9ba4\nsignature=5c4074ead7bd3b2f053790adb45ca776455f193e6370091a00159b711bfd7323b75c72cc667e8b27a8120ba4ea397ac964434a23363a6e35b24e81b1b925de0b\ncosignature=70b861a010f25030de6ff6a5267e0b951e70c04b20ba4a3ce41e7fba7b9b7dfc 1773223723 97ba8d7c37485a30507fef8cf1c67b5772ec3e5f9aee74c2a8412f56a1676b6a757e8b9f7483668a9320afb1ed8deaf2134ff2f5a439651fafc9e1096d9c6406\ncosignature=49c4cd6124b7c572f3354d854d50b2a4b057a750f786cf03103c09de339c4ea3 1773223723 44e29bed024f21be6d057abbbc6dbd2269351602cf5ead6e2fbcf24b04645047036779df90be5a3410a4446c3ede815a0972c9971e1bce1dff022d2f5bfca503\ncosignature=f308ac5bf00ef954f70fbe5e769258203cc792469154a1b9cf3003a6286a138d 1773223723 c60790b30b1a1cd5a8413e7568524dbd954d13ca93bfa884c2f836dc2fe7d8fbe72cfa6692223e080c8537b856d6ffb1d4cea3da875cd98fccd3dea4d613d00e\ncosignature=86b5414ae57f45c2953a074640bb5bedebad023925d4dc91a31de1350b710089 1773223723 9ae72e1972fe49aa0de13df3d7148f542ba42548b8146b4c731e71d753f491eff1cb033b8e7229fc10d70d3327a6a8226ab8b62251e6488eaefbc8b6e8a5c20c\ncosignature=c1d2d6935c2fb43bef395792b1f3c1dfe4072d4c6cadd05e0cc90b28d7141ed3 1773223723 b5758eed296cea8df0729d5f794ca173001991b588f81abe2264ff2111f04725a892760440fc1d5a39b95a114a1cb6e563a0afe7980cd72545a2c372425cd20b\ncosignature=26983895bdf491838dd4885848670562e7b728b6efa15fd9047b5b97a9a0618f 1773223723 7ea0e9bb590d7f8bfe5fc8ae532cb2da82bc9dfb55c393df1df33857f7b9b98b0ba153fd5f8627db44a905669648a3afc45c594381f9327dae17876ec194300a\ncosignature=e4a6a1e4657d8d7a187cc0c20ed51055d88c72f340d29534939aee32d86b4021 1773223723 5dad8a314a0fd11bc6f17b394294d07fa04c559ee6f0f19dce9136da31ecb787e1d192b78d76e1254cb4a71834cd433dd4ae7983107fbb1ff9d6c605109be809\ncosignature=d960fcff859a34d677343e4789c6843e897c9ff195ea7140a6ef382566df3b65 1773223723 8f5c9f0dc40c14e895e2d4ddcc2e21b9f79cb72f7aea1a186150ef2eb788c18943b8a69dffd77d81dd2418da612b01b8028c0b327d2f4213d58f87f6af7be40d\ncosignature=1c997261f16e6e81d13f420900a2542a4b6a049c2d996324ee5d82a90ca3360c 1773223723 cac4549af290fe0871a60f61ae78c5935783b99ed3230cd8cbaada7a344aafb9c154f69479411f8b3b5541b50ca52c2e4808770b94e5e1dcb18ddabf9871290a\ncosignature=42351ad474b29c04187fd0c8c7670656386f323f02e9a4ef0a0055ec061ecac8 1773223723 d8eecc0efe851467abd3eb4cb7c1092b5139b3261261167a1041edaa57ac30a866427e9d8f8dafbb3ede69a828f1d84c1b4ea2843f225f88d4fe915f9aaf1609\n\nleaf_index=162295\nnode_hash=c07ee3cc7acf42d84055f55e1ed33b87f28dde8da34831b3dea5885f6cbcbb43\nnode_hash=90dc785de30daf434287c6c01be8873feccb678f53e1230c1a72ce32ab44444d\nnode_hash=238cbbee2e28f5be71e876382bfbe104838e91cea920ea4b17dffd552b208ce4\nnode_hash=c509aad460c158ae33b1136e116d9140fb704917bd752828b24fcf865a2a420b\nnode_hash=444b347c61b994f387b67dab3d44476967ff7821d5087012783b1ab589d0dc63\nnode_hash=dacdfb38e8b1d5228164ebf5f124d5e967b96b209a5475b58bd58d566cece960\nnode_hash=8f53b17df95a109fcae2c9e7a630709c592ca9e6d481015f04b1a0da17ab6d4c\nnode_hash=4b2a225d2562949d50812ce7682fc23c53d630e686fdf012dda6414f08d481f6\nnode_hash=fec9d1376932767b7774a00fe6c72b9fef2d6156cf1195b11f1b252ad1a5b932\nnode_hash=cc6f233eb1e8e56292545c00f429318c19501a5609db82510dd128849025ce03\nnode_hash=a4909e6ea2bca271bee1338a87ea2617cf0354e7a894b6da017b7d1c08d7cdb7\nnode_hash=3d848e383e4b77edb95a496ad00f1e8764b8ddbfedd90f9eaeadae17449e2af6\nnode_hash=43bd28c79dec46786a85bdf0fe72eac3985a8fa172979cdbf7dc04d6c506d43d\n"
}`

	testSubmitKey = `ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIMqym9S/tFn6B/Eri5hGJiEV8BpGumEPcm65uxC+FG6K sigsum key`
	testLogKey    = `ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEZEryq9QPSJWgA7yjUPnVkSqzAaScd/E+W22QXCCl/m`

	testBootPolicyUnauthorized = `[
	{
		"artifacts": [
			{
				"category": 1,
				"requirements": {
					"build_args": {
						"I_WANT_CANDY": "y"
					}
				}
			},
			{
				"category": 2,
				"requirements": {
					"tainted": false
				}
			}
		],
		"signatures": {
			"signers": [
				{
					"name": "dist signatory I",
					"pub_key": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIP5rbNcIOcwqBHzLOhJEfdKFHa+pIs10idfTm8c+HDnK"
				},
				{
					"name": "dist signatory II",
			 		"pub_key": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIL0zV5fSWzzXa4R7Kpk6RAXkvWsJGpvkQ+9/xxpHC49J"
				}
			],
			"quorum": 2
		}
	}
]`
)

var testRoot = fstest.MapFS{
	bootPolicy: {
		Data: []byte(testBootPolicy),
	},
	witnessPolicy: {
		Data: []byte(testWitnessPolicy),
	},
	submitKey: {
		Data: []byte(testSubmitKey),
	},
	logKey: {
		Data: []byte(testLogKey),
	},
	proofBundle: {
		Data: []byte(testProofBundle),
	},
}
