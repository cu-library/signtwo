// Copyright 2015 Carleton University Library All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package ldap

import (
	"crypto/tls"
	"fmt"
	l "github.com/cu-library/signtwo/loglevel"
	"github.com/mavricknz/ldap"
	"net"
	"time"
)

var conn *ldap.LDAPConnection

func Connect(server string, port int, username string, password string, loglevel l.LogLevel) error {

	// The next block of code resolves the LDAP servers to one server.
	ips, err := net.LookupIP(server)
	if err != nil {
		return fmt.Errorf("Unable to get IP address of %v: %v", server, err)
	}
	hostnames, err := net.LookupAddr(ips[0].String())
	if err != nil {
		return fmt.Errorf("Unable to get hostname of %v: %v", server, err)
	}
	hostname := hostnames[0]

	tlsConfig := &tls.Config{ServerName: hostname}
	conn = ldap.NewLDAPSSLConnection(hostname, uint16(port), tlsConfig)

	debug := false
	if loglevel == l.DebugMessage || loglevel == l.TraceMessage {
		debug = true
	}
	conn.Debug = debug
	conn.NetworkConnectTimeout = time.Second * 10

	err = conn.Connect()
	if err != nil {
		return fmt.Errorf("Unable to connect to LDAP server at %v before 10 seconds: %v", hostname, err)
	}

	err = conn.Bind(username, password)
	if err != nil {
		return fmt.Errorf("Unable to bind to LDAP server at %v using credentials: %v", hostname, err)
	}

	return nil
}

func Close() {
	conn.Close()
}
