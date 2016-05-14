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

	AuthMechanism string
	Login         string
	Password      string

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

	if err := c.ProcessGreeting(); err != nil {
		return err
	}

	if c.State == ClientStateNotAuthenticated {
		if err := c.Authenticate(); err != nil {
			return err
		}
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

func (c *Client) ProcessGreeting() error {
	resp, err := ReadResponse(c.Stream)
	if err != nil {
		return err
	}

	var caps []string
	hasCaps := false

	authenticated := false

	switch tresp := resp.(type) {
	case *ResponseOk:
		if tresp.Text.Code == "CAPABILITY" {
			hasCaps = true
			caps = tresp.Text.CodeData.([]string)
		}

	case *ResponsePreAuth:
		authenticated = true
		if tresp.Text.Code == "CAPABILITY" {
			hasCaps = true
			caps = tresp.Text.CodeData.([]string)
		}

	case *ResponseBye:
		return fmt.Errorf("server shutting down: %v", tresp.Text.Text)

	default:
		return fmt.Errorf("invalid greeting %s", resp.Name())
	}

	if hasCaps {
		if err := c.ProcessCaps(caps); err != nil {
			return err
		}
	}

	if authenticated {
		c.State = ClientStateAuthenticated
	} else {
		c.State = ClientStateNotAuthenticated
	}

	return nil
}

func (c *Client) Authenticate() error {
	mechanisms := []string{"CRAM-MD5", "PLAIN"}

	var mechanism string

	if c.AuthMechanism == "" {
		for _, mech := range mechanisms {
			if c.HasCap("AUTH=" + mech) {
				mechanism = mech
				break
			}
		}
		if mechanism == "" {
			return fmt.Errorf("no supported authentication " +
				"mechanism found")
		}
	} else {
		if !c.HasCap("AUTH=" + c.AuthMechanism) {
			return fmt.Errorf("unsupported authentication " +
				"mechanism")
		}
		mechanism = c.AuthMechanism
	}

	var cmd Command
	switch mechanism {
	case "CRAM-MD5":
		cmd = &CommandAuthenticateCramMD5{
			Login:    c.Login,
			Password: c.Password,
		}
	case "PLAIN":
		cmd = &CommandAuthenticatePlain{
			Login:    c.Login,
			Password: c.Password,
		}
	default:
		return fmt.Errorf("unknown authentication mechanism")
	}

	if _, _, err := c.SendCommand(cmd); err != nil {
		return err
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
			if err := cmd.Continue(c.Writer, tresp); err != nil {
				return nil, nil, err
			}

			if err := c.Writer.Flush(); err != nil {
				return nil, nil, err
			}

		case *ResponseStatus:
			statusResponse = tresp

			switch status := tresp.Response.(type) {
			case *ResponseOk:
				res = nil
				break loop
			case *ResponseNo:
				res = errors.New(status.Text.Text)
				break loop
			case *ResponseBad:
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

func (c *Client) HasCap(cap string) bool {
	_, found := c.Caps[cap]
	return found
}
