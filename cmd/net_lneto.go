// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build net && lneto

package cmd

import (
	"github.com/usbarmory/go-net"
)

func newStack() gnet.Stack {
	return gnet.NewLnetoStack(nil)
}
