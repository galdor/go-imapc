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
	"bytes"
	"fmt"
)

type Response interface {
	Name() string
	Read(*Stream) error
}

func ReadResponse(s *Stream) (Response, error) {
	prefix, err := s.Peek(1)
	if err != nil {
		return nil, err
	}

	var readFn func(*Stream) (Response, error)

	if prefix[0] == '+' {
		// Command continuation
		readFn = ReadResponseContinuation
	} else if prefix[0] == '*' {
		// Server data or untagged status response
		// The only untagged status response is BYE.
		const byePrefix = "* BYE "
		prefix, err := s.Peek(len(byePrefix))
		if err != nil {
			return nil, err
		}

		if bytes.Equal(prefix, []byte(byePrefix)) {
			// Untagged status response
			readFn = ReadResponseStatus
		} else {
			// Server data
			readFn = ReadResponseData
		}
	} else {
		// Tagged status response
		readFn = ReadResponseStatus
	}

	r, err := readFn(s)
	if err != nil {
		return nil, err
	}

	if err := r.Read(s); err != nil {
		return nil, err
	}

	return r, nil
}

// ---------------------------------------------------------------------------
//  Status responses
//  RFC 3501 7.1.
// ---------------------------------------------------------------------------
func ReadResponseStatus(s *Stream) (Response, error) {
	// Read the tag
	tag, err := s.ReadUntilByteAndSkip(' ')
	if err != nil {
		return nil, err
	}

	// Read the response name
	name, err := s.ReadUntilByteAndSkip(' ')
	if err != nil {
		return nil, err
	}

	r := &ResponseStatus{
		Tag:          string(tag),
		ResponseName: string(name),
	}

	switch string(name) {
	case "OK":
		r.Response = &ResponseOk{}
	case "NO":
		r.Response = &ResponseNo{}
	case "BAD":
		r.Response = &ResponseBad{}
	default:
		return nil, fmt.Errorf("unknown response %q", name)
	}

	return r, nil
}

type ResponseStatus struct {
	Tag          string
	ResponseName string
	Response     Response
}

func (r *ResponseStatus) Name() string { return r.ResponseName }

func (r *ResponseStatus) Read(s *Stream) error {
	return r.Response.Read(s)
}

// OK
type ResponseOk struct {
	Text *ResponseText
}

func (r *ResponseOk) Name() string { return "OK" }

func (r *ResponseOk) Read(s *Stream) error {
	r.Text = &ResponseText{}
	return r.Text.Read(s)
}

// NO
type ResponseNo struct {
	Text *ResponseText
}

func (r *ResponseNo) Name() string { return "NO" }

func (r *ResponseNo) Read(s *Stream) error {
	r.Text = &ResponseText{}
	return r.Text.Read(s)
}

// BAD
type ResponseBad struct {
	Text *ResponseText
}

func (r *ResponseBad) Name() string { return "BAD" }

func (r *ResponseBad) Read(s *Stream) error {
	r.Text = &ResponseText{}
	return r.Text.Read(s)
}

// ---------------------------------------------------------------------------
//  Server data responses
//  RFC 3501 7.2., 7.3., 7.4.
// ---------------------------------------------------------------------------
func ReadResponseData(s *Stream) (Response, error) {
	// Skip "* "
	if err := s.Skip(2); err != nil {
		return nil, err
	}

	// Read the response name
	name, err := s.ReadUntilByteAndSkip(' ')
	if err != nil {
		return nil, err
	}

	var r Response
	switch string(name) {
	case "OK":
		r = &ResponseOk{}
	default:
		return nil, fmt.Errorf("unknown response %q", name)
	}

	return r, nil
}

// PREAUTH
type ResponsePreAuth struct {
	Text *ResponseText
}

func (r *ResponsePreAuth) Name() string { return "PREAUTH" }

func (r *ResponsePreAuth) Read(s *Stream) error {
	r.Text = &ResponseText{}
	return r.Text.Read(s)
}

// BYE
type ResponseBye struct {
	Text *ResponseText
}

func (r *ResponseBye) Name() string { return "BYE" }

func (r *ResponseBye) Read(s *Stream) error {
	r.Text = &ResponseText{}
	return r.Text.Read(s)
}

// ---------------------------------------------------------------------------
//  Command continuation responses
//  RFC 3501 7.5.
// ---------------------------------------------------------------------------
func ReadResponseContinuation(s *Stream) (Response, error) {
	// Skip "+ "
	if err := s.Skip(2); err != nil {
		return nil, err
	}

	return &ResponseContinuation{}, nil
}

type ResponseContinuation struct {
	Text string
}

func (r *ResponseContinuation) Name() string { return "" }

func (r *ResponseContinuation) Read(s *Stream) error {
	text, err := s.ReadUntilAndSkip([]byte("\r\n"))
	if err != nil {
		return err
	}
	r.Text = string(text)

	return nil
}

// ---------------------------------------------------------------------------
//  Response text
//  RFC 3501 7.1.
// ---------------------------------------------------------------------------
type ResponseText struct {
	Text string

	// Optional
	Code       string
	CodeString string
	CodeData   interface{}
}

func (r *ResponseText) Read(s *Stream) error {
	found, err := s.SkipByte('[')
	if err != nil {
		return err
	} else if found {
		// Response text code
		codeBytes, err := s.ReadUntilAndSkip([]byte("] "))
		if err != nil {
			return err
		}
		r.CodeString = string(codeBytes)

		idx := bytes.IndexByte(codeBytes, ' ')
		if idx == -1 {
			r.Code = string(codeBytes)
		} else {
			r.Code = string(codeBytes[0:idx])
		}

		switch r.Code {
		case "BADCHARSET":
			// TODO
		case "PERMANENTFLAGS":
			// TODO
		case "UIDNEXT":
			// TODO
		case "UIDVALIDITY":
			// TODO
		case "UNSEEN":
			// TODO
		case "CAPABILITY":
			parts := bytes.Split(codeBytes, []byte{' '})
			caps := make([]string, len(parts))
			for i, cap := range parts {
				caps[i] = string(cap)
			}
			r.CodeData = caps
		}
	}

	// Response text
	text, err := s.ReadUntilAndSkip([]byte("\r\n"))
	if err != nil {
		return err
	}

	r.Text = string(text)
	fmt.Printf("TEXT  %#v\n", r)
	return nil
}
