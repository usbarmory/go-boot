// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build net && net-lneto

package cmd

import gnet "github.com/usbarmory/go-net"

func newDefaultStack() gnet.Stack {
	return gnet.NewLnetoStack()
}
