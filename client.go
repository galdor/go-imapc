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
	"errors"
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
	Login      string
	Password   string

	Conn   net.Conn
	Stream *Stream
	Writer *BufferedWriter

	State ClientState

	Caps map[string]bool

	Tag int
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

	c.Stream = NewStream(c.Conn)
	c.Writer = NewBufferedWriter(c.Conn)

	c.State = ClientStateNotAuthenticated

	if err := c.Authenticate(); err != nil {
		return err
	}

	if err := c.FetchCaps(); err != nil {
		return err
	}

	// Main loop
loop:
	for {
		resp, err := ReadResponse(c.Stream)
		if err != nil {
			return err
		}

		switch tresp := resp.(type) {
		case *ResponseBye:
			fmt.Printf("BYE   %#v\n", tresp)
			break loop
		default:
			fmt.Printf("RESP  %#v\n", tresp)
			// TODO
		}
	}

	return nil
}

func (c *Client) Authenticate() error {
	// Read the greeting response
	resp, err := ReadResponse(c.Stream)
	if err != nil {
		return err
	}

	var caps []string
	hasCaps := false

	auth := false

	switch tresp := resp.(type) {
	case *ResponseOk:
		auth = true
		if tresp.Text.Code == "CAPABILITY" {
			hasCaps = true
			caps = tresp.Text.CodeData.([]string)
		}
	case *ResponsePreAuth:
		if tresp.Text.Code == "CAPABILITY" {
			hasCaps = true
			caps = tresp.Text.CodeData.([]string)
		}
	case *ResponseBye:
		// TODO Signal connection close and return
	default:
		return fmt.Errorf("invalid greeting %s", resp.Name())
	}

	// Check capabilities if they are provided
	if hasCaps {
		if err := c.ProcessCaps(caps); err != nil {
			return err
		}
	}

	// Authenticate if necessary
	if auth {
		// TODO select supported authentication mechanism

		cmd := &CommandAuthenticatePlain{
			Login:    c.Login,
			Password: c.Password,
		}

		if _, _, err := c.SendCommand(cmd); err != nil {
			return err
		}
	}

	c.State = ClientStateAuthenticated
	return nil
}

func (c *Client) SendCommand(cmd Command) ([]Response, *ResponseStatus, error) {
	// Send a new tag
	c.Tag++
	fmt.Fprintf(c.Writer, "c%07d ", c.Tag)

	// Send the command
	cmd.Write(c.Writer)

	if err := c.Writer.Flush(); err != nil {
		return nil, nil, err
	}

	// Read responses until we get either the status response or a bye
	// response
	var res error
	var dataResponses []Response
	var statusResponse *ResponseStatus

loop:
	for {
		resp, err := ReadResponse(c.Stream)
		if err != nil {
			return nil, nil, err
		}

		switch tresp := resp.(type) {
		case *ResponseContinuation:
			cmd.Continue(c.Writer, tresp)
			if err := c.Writer.Flush(); err != nil {
				return nil, nil, err
			}

		case *ResponseStatus:
			statusResponse = tresp

			switch status := tresp.Response.(type) {
			case *ResponseOk:
				fmt.Printf("OK    %#v\n", tresp)
				res = nil
				break loop
			case *ResponseNo:
				fmt.Printf("NO    %#v\n", tresp)
				res = errors.New(status.Text.Text)
				break loop
			case *ResponseBad:
				fmt.Printf("BAD   %#v\n", tresp)
				res = errors.New(status.Text.Text)
				break loop
			}

		case *ResponseBye:
			res = fmt.Errorf("server shutting down: %v",
				tresp.Text.Text)
			break loop

		default:
			dataResponses = append(dataResponses, resp)
		}
	}

	return dataResponses, statusResponse, res
}

func (c *Client) FetchCaps() error {
	cmd := &CommandCapability{}
	resps, _, err := c.SendCommand(cmd)
	if err != nil {
		return err
	}

	for _, resp := range resps {
		tresp, ok := resp.(*ResponseCapability)
		if ok {
			if err := c.ProcessCaps(tresp.Caps); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Client) ProcessCaps(caps []string) error {
	c.Caps = make(map[string]bool)

	for _, cap := range caps {
		c.Caps[cap] = true
	}

	if _, found := c.Caps["IMAP4rev1"]; !found {
		return fmt.Errorf("missing IMAP4rev1 capability")
	}

	return nil
}
