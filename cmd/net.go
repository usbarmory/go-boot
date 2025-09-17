// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"net"
	"regexp"

	"github.com/usbarmory/go-net"

	"github.com/usbarmory/go-boot/shell"
	"github.com/usbarmory/go-boot/uefi/x64"
)

// Resolver represents the default name server
var Resolver = "8.8.8.8:53"

func init() {
	shell.Add(shell.Cmd{
		Name:    "net",
		Args:    2,
		Pattern: regexp.MustCompile(`^net (\S+) (\S+)$`),
		Syntax:  "<ip> <gateway>",
		Help:    "start UEFI networking",
		Fn:      netCmd,
	})

	shell.Add(shell.Cmd{
		Name:    "dns",
		Args:    1,
		Pattern: regexp.MustCompile(`^dns (.*)`),
		Syntax:  "<host>",
		Help:    "resolve domain",
		Fn:      dnsCmd,
	})

	net.SetDefaultNS([]string{Resolver})
}

func netCmd(_ *shell.Interface, arg []string) (res string, err error) {
	nic, err := x64.UEFI.Boot.GetNetwork()

	if err != nil {
		return "", fmt.Errorf("could not locate network protocol, %v", err)
	}

	if err = nic.Start(); err != nil {
		return "", fmt.Errorf("could not start interface, %v", err)
	}

	if err = nic.Initialize(); err != nil {
		return "", fmt.Errorf("could not initialize interface, %v", err)
	}

	iface := gnet.Interface{}

	if err := iface.Init(nic, arg[0], "", arg[1]); err != nil {
		return "", fmt.Errorf("could not initialize networking, %v", err)
	}

	iface.EnableICMP()
	go iface.NIC.Start()

	// hook interface into Go runtime
	net.SocketFunc = iface.Socket

	return "network initialized", nil
}

func dnsCmd(_ *shell.Interface, arg []string) (res string, err error) {
	cname, err := net.LookupHost(arg[0])

	if err != nil {
		return "", fmt.Errorf("query error: %v", err)
	}

	return fmt.Sprintf("%+v", cname), nil
}
