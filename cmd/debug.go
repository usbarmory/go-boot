// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build net && debug

package cmd

import (
	"github.com/arl/statsviz"
	_ "net/http/pprof"
)

func init() {
	statsviz.RegisterDefault()
}
