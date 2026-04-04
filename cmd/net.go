//go:build net

// Copyright (c) The go-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"regexp"
	"strings"

	"github.com/gliderlabs/ssh"
	gnet "github.com/usbarmory/go-net"

	"github.com/usbarmory/go-boot/shell"
	"github.com/usbarmory/go-boot/uefi"
	"github.com/usbarmory/go-boot/uefi/x64"

	// maintained set of TLS roots for any potential TLS client requests
	_ "golang.org/x/crypto/x509roots/fallback"
)

// Resolver represents the default name server
var Resolver = "8.8.8.8:53"

const receiveMask = uefi.EFI_SIMPLE_NETWORK_RECEIVE_UNICAST |
	uefi.EFI_SIMPLE_NETWORK_RECEIVE_BROADCAST |
	uefi.EFI_SIMPLE_NETWORK_RECEIVE_PROMISCUOUS

func init() {
	shell.Add(shell.Cmd{
		Name:    "net",
		Args:    4,
		Pattern: regexp.MustCompile(`^net (\S+) (\S+) (\S+)( debug)?$`),
		Syntax:  "<ip> <mac> <gw> (debug)?",
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

func newDefaultStack() gnet.Stack {
	return gnet.NewGVisorStack(1)
}

func netCmd(_ *shell.Interface, arg []string) (res string, err error) {
	nic, err := x64.UEFI.Boot.GetNetwork()

	if err != nil {
		return "", fmt.Errorf("could not locate network protocol, %v", err)
	}

	// clean up from previous initializations
	nic.Shutdown()
	nic.Stop()
	nic.Start()

	if err = nic.Initialize(); err != nil {
		return "", fmt.Errorf("could not initialize interface, %v", err)
	}

	if err = nic.ReceiveFilters(receiveMask, 0); err != nil {
		return "", fmt.Errorf("could not set receive filters, %v", err)
	}

	iface := gnet.Interface{}

	if arg[1] == ":" {
		arg[1] = ""
	}

	if err := iface.Init(nic, newDefaultStack(), arg[0], arg[1], arg[2]); err != nil {
		return "", fmt.Errorf("could not initialize networking, %v", err)
	}
	stack := iface.NetworkingStack()
	mac, err := stack.HardwareAddress()
	if err != nil {
		return "", fmt.Errorf("network stack failed on MAC get: %v", err)
	}
	if err = nic.StationAddress(false, mac); err != nil {
		fmt.Printf("could not set permanent station address, %v\n", err)
	}

	err = stack.EnableICMP()
	if err != nil {
		fmt.Printf("could not enable ICMP, %v\n", err)
	}
	go iface.StartRx()

	// hook interface into Go runtime
	net.SocketFunc = stack.Socket

	if len(arg[3]) > 0 {
		ip, _, _ := strings.Cut(arg[0], `/`)

		fmt.Printf("starting debug servers:\n")
		fmt.Printf("\thttp://%s:80/debug/pprof\n", ip)
		fmt.Printf("\tssh://%s:22\n", ip)

		ssh.Handle(func(s ssh.Session) {
			c := &shell.Interface{
				Banner:     Banner,
				ReadWriter: s,
			}

			log.SetOutput(io.MultiWriter(os.Stdout, s))
			defer log.SetOutput(os.Stdout)

			c.Start(true)
		})

		go ssh.ListenAndServe(":22", nil)
		go http.ListenAndServe(":80", nil)
	}

	return fmt.Sprintf("network initialized (%s %s)\n", arg[0], iface.NIC.MAC), nil
}

func dnsCmd(_ *shell.Interface, arg []string) (res string, err error) {
	cname, err := net.LookupHost(arg[0])

	if err != nil {
		return "", fmt.Errorf("query error: %v", err)
	}

	return fmt.Sprintf("%+v\n", cname), nil
}
