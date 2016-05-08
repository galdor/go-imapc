//
// Copyright (c) 2016 Nicolas Martyanoff <khaelin@gmail.com>
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

package imapc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
)

// ---------------------------------------------------------------------------
//  Client state
//  RFC 3501 3.
// ---------------------------------------------------------------------------
type ClientState string

const (
	ClientStateNotAuthenticated ClientState = "not-authenticated"
	ClientStateAuthenticated    ClientState = "authenticated"
	ClientStateSelected         ClientState = "selected"
	ClientStateLogout           ClientState = "logout"
)

// ---------------------------------------------------------------------------
//  Client
// ---------------------------------------------------------------------------
type Client struct {
	Host       string
	Port       int
	TLS        bool
	CACertPath string
	CertPath   string
	KeyPath    string

	Conn   net.Conn
	Stream *Stream

	State ClientState
}

func NewClient() *Client {
	return &Client{
		Host: "localhost",
		Port: 143,
	}
}

func (c *Client) Connect() error {
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)

	var err error
	if c.TLS {
		var caCerts *x509.CertPool
		if c.CACertPath != "" {
			caCertData, err := ioutil.ReadFile(c.CACertPath)
			if err != nil {
				return err
			}

			caCerts = x509.NewCertPool()
			if caCerts.AppendCertsFromPEM(caCertData) == false {
				return fmt.Errorf("cannot read ca certificate")
			}
		}

		var cert tls.Certificate
		if c.CertPath != "" && c.KeyPath != "" {
			cert, err = tls.LoadX509KeyPair(c.CertPath, c.KeyPath)
			if err != nil {
				return err
			}
		}

		cfg := tls.Config{
			RootCAs:      caCerts,
			Certificates: []tls.Certificate{cert},
		}

		c.Conn, err = tls.Dial("tcp", addr, &cfg)
	} else {
		c.Conn, err = net.Dial("tcp", addr)
	}
	if err != nil {
		return err
	}

	c.State = ClientStateNotAuthenticated

	c.Stream = NewStream(c.Conn)
	for {
		resp, err := ReadResponse(c.Stream)
		if err != nil {
			return err
		}

		fmt.Printf("RESP  %#v\n", resp)
	}

	return nil
}
