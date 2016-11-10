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
	ClientStateDisconnected     ClientState = "disconnected"
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

	stopChan chan int
	cmdChan  chan Command
	respChan chan *CommandResponse
}

func NewClient() *Client {
	return &Client{
		Host: "localhost",
		Port: 143,

		State: ClientStateDisconnected,
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

	if err := c.processGreeting(); err != nil {
		return err
	}

	c.stopChan = make(chan int)
	c.cmdChan = make(chan Command)
	c.respChan = make(chan *CommandResponse)

	go c.main()

	if c.State == ClientStateNotAuthenticated {
		if err := c.authenticate(); err != nil {
			c.stopChan <- 1
			return err
		}
	}

	if err := c.fetchCaps(); err != nil {
		c.stopChan <- 1
		return err
	}

	return nil
}

func (c *Client) main() {
loop:
	for {
		select {
		case <-c.stopChan:
			break loop

		case cmd := <-c.cmdChan:
			c.respChan <- c.processCommand(cmd)
		}
	}

	close(c.stopChan)
	close(c.cmdChan)
	close(c.respChan)
}

func (c *Client) processCommand(cmd Command) *CommandResponse {
	cmdResp := &CommandResponse{}

	sendLiteral := func(l Literal) error {
		data := []byte(l)

		fmt.Fprintf(c.Writer, "{%d}\r\n", len(data))
		if err := c.Writer.Flush(); err != nil {
			return err
		}

		var respErr error = nil

	loop:
		for {
			resp, err := ReadResponse(c.Stream)
			if err != nil {
				return err
			}

			switch tresp := resp.(type) {
			case *ResponseContinuation:
				c.Writer.Append(data)
				break loop

			case *ResponseStatus:
				cmdResp.Status = tresp

				switch tresp.Response.(type) {
				case *ResponseOk:
					respErr = errors.New("ok response " +
						" received while sending " +
						"literal")
				}

				break loop

			case *ResponseBye:
				respErr = fmt.Errorf("server shutting down: %v",
					tresp.Text.Text)
				break loop

			default:
				cmdResp.Data = append(cmdResp.Data, resp)
			}
		}

		return respErr
	}

	// Send a new tag
	c.Tag++
	fmt.Fprintf(c.Writer, "c%07d ", c.Tag)

	// Send the command
	args := cmd.Args()
	for i, arg := range args {
		if i > 0 {
			c.Writer.AppendString(" ")
		}

		switch targ := arg.(type) {
		case []byte:
			c.Writer.Append(targ)

		case string:
			c.Writer.AppendString(targ)

		case Literal:
			if err := sendLiteral(targ); err != nil {
				cmdResp.Error = err
				return cmdResp
			}

			// Do not send the rest of the command if the server
			// sent a NO or BAD status response.
			if cmdResp.Status != nil {
				switch cmdResp.Status.Response.(type) {
				case *ResponseNo:
					return cmdResp
				case *ResponseBad:
					return cmdResp
				}
			}

		default:
			panic("invalid command argument")
		}
	}

	c.Writer.AppendString("\r\n")

	if err := c.Writer.Flush(); err != nil {
		cmdResp.Error = err
		return cmdResp
	}

	// Read responses until we get either the status response or a bye
	// response

loop:
	for {
		resp, err := ReadResponse(c.Stream)
		if err != nil {
			cmdResp.Error = err
			return cmdResp
		}

		switch tresp := resp.(type) {
		case *ResponseContinuation:
			if err := cmd.Continue(c.Writer, tresp); err != nil {
				cmdResp.Error = err
				return cmdResp
			}

			if err := c.Writer.Flush(); err != nil {
				cmdResp.Error = err
				return cmdResp
			}

		case *ResponseStatus:
			cmdResp.Status = tresp
			break loop

		case *ResponseBye:
			cmdResp.Error = fmt.Errorf("server shutting down: %v",
				tresp.Text.Text)
			break loop

		default:
			cmdResp.Data = append(cmdResp.Data, resp)
		}
	}

	return cmdResp
}

func (c *Client) processGreeting() error {
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
		return fmt.Errorf("invalid greeting %#v", resp)
	}

	if hasCaps {
		if err := c.processCaps(caps); err != nil {
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

func (c *Client) authenticate() error {
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

func (c *Client) fetchCaps() error {
	cmd := &CommandCapability{}
	resps, _, err := c.SendCommand(cmd)
	if err != nil {
		return err
	}

	for _, resp := range resps {
		tresp, ok := resp.(*ResponseCapability)
		if ok {
			if err := c.processCaps(tresp.Caps); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Client) processCaps(caps []string) error {
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

func (c *Client) SendCommand(cmd Command) ([]Response, *ResponseStatus, error) {
	if c.State == ClientStateDisconnected {
		return nil, nil, errors.New("connection down")
	}

	c.cmdChan <- cmd
	resp := <-c.respChan

	var err error = nil

	if resp.Error != nil {
		err = resp.Error
	} else if resp.Status != nil {
		switch status := resp.Status.Response.(type) {
		case *ResponseNo:
			err = errors.New(status.Text.Text)
		case *ResponseBad:
			err = errors.New(status.Text.Text)
		}
	}

	return resp.Data, resp.Status, err
}

func (c *Client) SendCommandWithResponseSet(cmd Command, rs ResponseSet) error {
	resps, status, err := c.SendCommand(cmd)
	if err != nil {
		return err
	}

	if err := rs.Init(resps, status); err != nil {
		return err
	}

	return nil
}

func (c *Client) SendCommandList(ref, pattern string) (*ResponseSetList, error) {
	cmd := &CommandList{
		Ref:     ref,
		Pattern: pattern,
	}

	rs := &ResponseSetList{}

	if err := c.SendCommandWithResponseSet(cmd, rs); err != nil {
		return nil, err
	}

	return rs, nil
}

func (c *Client) SendCommandLSub(ref, pattern string) (*ResponseSetLSub, error) {
	cmd := &CommandLSub{
		Ref:     ref,
		Pattern: pattern,
	}

	rs := &ResponseSetLSub{}

	if err := c.SendCommandWithResponseSet(cmd, rs); err != nil {
		return nil, err
	}

	return rs, nil
}

func (c *Client) SendCommandSubscribe(mailboxName string) (*ResponseSetSubscribe, error) {
	cmd := &CommandSubscribe{
		MailboxName: mailboxName,
	}

	rs := &ResponseSetSubscribe{}

	if err := c.SendCommandWithResponseSet(cmd, rs); err != nil {
		return nil, err
	}

	return rs, nil
}

func (c *Client) SendCommandUnsubscribe(mailboxName string) (*ResponseSetUnsubscribe, error) {
	cmd := &CommandUnsubscribe{
		MailboxName: mailboxName,
	}

	rs := &ResponseSetUnsubscribe{}

	if err := c.SendCommandWithResponseSet(cmd, rs); err != nil {
		return nil, err
	}

	return rs, nil
}

func (c *Client) SendCommandExamine(mailboxName string) (*ResponseSetExamine, error) {
	cmd := &CommandExamine{
		MailboxName: mailboxName,
	}

	rs := &ResponseSetExamine{}

	if err := c.SendCommandWithResponseSet(cmd, rs); err != nil {
		return nil, err
	}

	return rs, nil
}
