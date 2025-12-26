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

var testBootPolicy = []byte(`[
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
]`)

var testWitnessPolicy = []byte(`log 4644af2abd40f4895a003bca350f9d5912ab301a49c77f13e5b6d905c20a5fe6 https://test.sigsum.org/barreleye

witness poc.sigsum.org/nisse 1c25f8a44c635457e2e391d1efbca7d4c2951a0aef06225a881e46b98962ac6c
witness rgdd.se/poc-witness  28c92a5a3a054d317c86fc2eeb6a7ab2054d6217100d0be67ded5b74323c5806

group  demo-quorum-rule any poc.sigsum.org/nisse rgdd.se/poc-witness
quorum demo-quorum-rule
`)

var testProofBundle = []byte(`{
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
					"file_name": "vmlinuz-6.14.0-36-generic",
					"file_hash": "4551848b4ab43cb4321c4d6ba98e1d215f950cee21bfd82c8c82ab64e34ec9a6",
					"version": "v6.14.0-36-generic",
					"tainted": false,
					"build_args": {
						"CONFIG_STACKPROTECTOR_STRONG": "y"
					}
				}
			},
			{
				"category": 2,
				"claims": {
					"file_name": "initrd.img-6.14.0-36-generic",
					"file_hash": "337630b74e55eae241f460faadf5a2f9a2157d6de2853d4106c35769e4acf538",
					"tainted": false
				}
			}
		],
		"signatures": [
			{
				"pub_key": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIP5rbNcIOcwqBHzLOhJEfdKFHa+pIs10idfTm8c+HDnK",
				"signature": "d29fe25fb7b0f8e977662bd50adac19ce0fdf3a51d8c2d8142d3fdd8a856200c7c4f1bf9ac8bb0bb011070869e585ebc8237d1e863e108b35788404de40bb40a"
			},
			{
				"pub_key": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIL0zV5fSWzzXa4R7Kpk6RAXkvWsJGpvkQ+9/xxpHC49J",
				"signature": "fac959414a3032f93490790193460e3ab014eaae674dd2fef1211edaf0630d47eac774065e6f9897b1a49b6e66c89ec7c3cc4387068dc2261bd4115be096aa0d"
			}
		]
	},
	"probe": {
		"origin": "https://test.sigsum.org/barreleye",
		"leaf_signature": "b380a0b0df0ae1710c1a51a708d7f45c0df840047a91c820dac04fe3f2c03b40b10b3f0cbcb189de410f83ad735ccf0d8146fe0b9c1408eda2b421d1daea8101",
		"log_public_key_hash": "4e89cc51651f0d95f3c6127c15e1a42e3ddf7046c5b17b752689c402e773bb4d",
		"submit_public_key_hash": "302928c2e0e01da52e3b161c54906de9b55ce250f0f47e80e022d04036e2765c"
	},
	"proof": "version=2\nlog=4e89cc51651f0d95f3c6127c15e1a42e3ddf7046c5b17b752689c402e773bb4d\nleaf=302928c2e0e01da52e3b161c54906de9b55ce250f0f47e80e022d04036e2765c b380a0b0df0ae1710c1a51a708d7f45c0df840047a91c820dac04fe3f2c03b40b10b3f0cbcb189de410f83ad735ccf0d8146fe0b9c1408eda2b421d1daea8101\n\nsize=17460\nroot_hash=4b07f20245d211b6fe058dd396baf2821d260d4180da47b561833ec3611cdfcd\nsignature=016235c8ce22d377028da78453b8b9d30e664c36f1892b436fca2a29aab182ff6251c14ae0c3d453fddc28fb3f5a73ea932517ba2657716a7d251ab2c5b77b0a\ncosignature=70b861a010f25030de6ff6a5267e0b951e70c04b20ba4a3ce41e7fba7b9b7dfc 1765901875 d713b5ed21fc156e942d735808586d91f27c494f61a59414ecc19048866d46a57a265e82eaa2c80acbcbb7cf90a57d9e11f1a5c22ecdd880264246d6dac7580e\ncosignature=26983895bdf491838dd4885848670562e7b728b6efa15fd9047b5b97a9a0618f 1765901875 ff48b36e74670e2e74910bca4b0bc1d9ab0c62f162c9ccd1ad86aca05f2b1a31e1920aa0aa029973866174780d3ee6f480c8337d9927feeb118a40202a07cf01\ncosignature=d960fcff859a34d677343e4789c6843e897c9ff195ea7140a6ef382566df3b65 1765901875 9019fccf24c91a5dc535c5f73b6b8bc85da8bfa97136704f9c13733808ba52b7213ac664a2cef4aa67e2512736a77853524cbb0ee1ace6d3451b47351475a20e\ncosignature=e4a6a1e4657d8d7a187cc0c20ed51055d88c72f340d29534939aee32d86b4021 1765901875 2ccd5af89a0d25d94679c86e7f782a868f1b61da6fb3365f59558b355be583d588acd11c57c1a33e2f5683c82cb4237eb62d6d8275d25a51014e42f06a2e110e\ncosignature=42351ad474b29c04187fd0c8c7670656386f323f02e9a4ef0a0055ec061ecac8 1765901875 f679eb4fb0bf40b516b30eb6e74c6d6293e8797be0cb8ab6335698f547e1aae1397f91b4ee60ceecf02e02daf202e65b26117a0c86676c956d72eaaa8e829a07\ncosignature=c1d2d6935c2fb43bef395792b1f3c1dfe4072d4c6cadd05e0cc90b28d7141ed3 1765901875 8faa05703639bc15bd5f783214393e50e57a3f0eb61d9f79de042b2428c613586261f357f31d57a81c8793fabb50d889f5b80c6e9c9da8a94b675e3485ccbd01\ncosignature=1c997261f16e6e81d13f420900a2542a4b6a049c2d996324ee5d82a90ca3360c 1765901875 580be37fda6053b5fe1442b174e97b4ea5d49404855b1a5f3de034ee16ba988c2d694b6bc56683b707f640507271f78b367691dcebb7ee53f6e2f797e3653101\ncosignature=49c4cd6124b7c572f3354d854d50b2a4b057a750f786cf03103c09de339c4ea3 1765901875 1be97c08da21bf8adfccfbf4b69166fb8529ec13123f84bebef3c5a582d7e5a5c0f579e96f813410c8ca1f966f15944c4e48272570cd670d62b9a6321fa4f500\ncosignature=86b5414ae57f45c2953a074640bb5bedebad023925d4dc91a31de1350b710089 1765901875 d4daf5a7be9482af90bcb665c6bc75dd05421b7fcb5f053b3bcb9a6994b9e516339e2736e03c831cd273947cda964898f92c4288fba8001bf84e699be0b7300e\n\nleaf_index=17459\nnode_hash=83deb0c1a1b1d1468883321dc1ce686c437142c23eba4cf842dd0520ac6d5025\nnode_hash=503d30009fc38c5179225883ada9a9cb1d27cfdd9b2d730218ee95ec31c83d7c\nnode_hash=5d1d426d73f0a178a3c6385e2cc8811e18c81edee8dee566c03d82c7f4404626\nnode_hash=4f1e71bd71a62ab4255e6eb2d739b1f28b21fb0cbda97434c76da71399031a74\nnode_hash=9ad205a7b91cfbdcbe22a13e77d877509f2eb90c4b19355247781ef32c8446bb\nnode_hash=297c6cc8145fd585a9e009057191d91c9ce013957ee48ae26982ac96d52b578b\n"
}`)

var testSubmitKey = []byte(`ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIMqym9S/tFn6B/Eri5hGJiEV8BpGumEPcm65uxC+FG6K sigsum key`)

var testLogKey = []byte(`ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEZEryq9QPSJWgA7yjUPnVkSqzAaScd/E+W22QXCCl/m`)

var testUefiRoot = fstest.MapFS{
	bootPolicy: {
		Data: testBootPolicy,
	},
	witnessPolicy: {
		Data: testWitnessPolicy,
	},
	submitKey: {
		Data: testSubmitKey,
	},
	logKey: {
		Data: testLogKey,
	},
	proofBundle: {
		Data: testProofBundle,
	},
}

var testBootPolicyUnauthorized = []byte(`[
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
]`)
