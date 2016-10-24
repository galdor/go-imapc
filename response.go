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
	"strconv"
)

type Response interface {
	Name() string
	fmt.GoStringer
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

func (r *ResponseStatus) GoString() string {
	return fmt.Sprintf("#<response-status %#v>", r.Response)
}

func (r *ResponseStatus) Read(s *Stream) error {
	return r.Response.Read(s)
}

// OK
type ResponseOk struct {
	Text *ResponseText
}

func (r *ResponseOk) Name() string { return "OK" }

func (r *ResponseOk) GoString() string {
	return fmt.Sprintf("#<response-ok %#v>", r.Text)
}

func (r *ResponseOk) Read(s *Stream) error {
	r.Text = &ResponseText{}
	return r.Text.Read(s)
}

// NO
type ResponseNo struct {
	Text *ResponseText
}

func (r *ResponseNo) Name() string { return "NO" }

func (r *ResponseNo) GoString() string {
	return fmt.Sprintf("#<response-no %#v>", r.Text)
}

func (r *ResponseNo) Read(s *Stream) error {
	r.Text = &ResponseText{}
	return r.Text.Read(s)
}

// BAD
type ResponseBad struct {
	Text *ResponseText
}

func (r *ResponseBad) Name() string { return "BAD" }

func (r *ResponseBad) GoString() string {
	return fmt.Sprintf("#<response-bad %#v>", r.Text)
}

func (r *ResponseBad) Read(s *Stream) error {
	r.Text = &ResponseText{}
	return r.Text.Read(s)
}

// ---------------------------------------------------------------------------
//  Server data responses
// ---------------------------------------------------------------------------
func ReadResponseData(s *Stream) (Response, error) {
	// Skip "* "
	if err := s.Skip(2); err != nil {
		return nil, err
	}

	// Read either the response tag or a number
	data, err := s.ReadUntilByteAndSkip(' ')
	if err != nil {
		return nil, err
	}

	var r Response

	if ByteStringAll(data, IsDigitChar) {
		n, err := strconv.ParseUint(string(data), 10, 32)
		if err != nil {
			return nil, err
		}
		count := uint32(n)

		rest, err := s.ReadUntil([]byte("\r\n"))
		if err != nil {
			return nil, err
		}
		tag := string(rest)

		switch tag {
		case "EXISTS":
			r = &ResponseExists{Count: count}
		case "RECENT":
			r = &ResponseRecent{Count: count}
		default:
			return nil, fmt.Errorf("unknown response %q", tag)
		}
	} else {
		tag := string(data)

		switch tag {
		case "OK":
			r = &ResponseOk{}
		case "PREAUTH":
			r = &ResponsePreAuth{}
		case "BYE":
			r = &ResponseBye{}
		case "CAPABILITY":
			r = &ResponseCapability{}
		case "LIST":
			r = &ResponseList{}
		case "FLAGS":
			r = &ResponseFlags{}
		default:
			return nil, fmt.Errorf("unknown response %q", tag)
		}
	}

	return r, nil
}

// PREAUTH
type ResponsePreAuth struct {
	Text *ResponseText
}

func (r *ResponsePreAuth) Name() string { return "PREAUTH" }

func (r *ResponsePreAuth) GoString() string {
	return fmt.Sprintf("#<response-pre-auth %#v>", r.Text)
}

func (r *ResponsePreAuth) Read(s *Stream) error {
	r.Text = &ResponseText{}
	return r.Text.Read(s)
}

// BYE
type ResponseBye struct {
	Text *ResponseText
}

func (r *ResponseBye) Name() string { return "BYE" }

func (r *ResponseBye) GoString() string {
	return fmt.Sprintf("#<response-bye %#v>", r.Text)
}

func (r *ResponseBye) Read(s *Stream) error {
	r.Text = &ResponseText{}
	return r.Text.Read(s)
}

// CAPABILITY
type ResponseCapability struct {
	Caps []string
}

func (r *ResponseCapability) Name() string { return "CAPABILITY" }

func (r *ResponseCapability) GoString() string {
	return fmt.Sprintf("#<response-capability %v>", r.Caps)
}

func (r *ResponseCapability) Read(s *Stream) error {
	data, err := s.ReadUntilAndSkip([]byte("\r\n"))
	if err != nil {
		return err
	}

	parts := bytes.Split(data, []byte{' '})

	r.Caps = make([]string, len(parts))
	for i, cap := range parts {
		r.Caps[i] = string(cap)
	}

	return nil
}

// LIST
type ResponseList struct {
	MailboxFlags       []string
	HierarchyDelimiter rune
	MailboxName        string
}

func (r *ResponseList) Name() string { return "LIST" }

func (r *ResponseList) GoString() string {
	return fmt.Sprintf("#<response-list %q %v>",
		r.MailboxName, r.MailboxFlags)
}

func (r *ResponseList) Read(s *Stream) error {
	// Mailbox flags
	flags, err := s.ReadIMAPFlagList()
	if err != nil {
		return err
	}
	r.MailboxFlags = flags

	if found, err := s.SkipByte(' '); err != nil {
		return err
	} else if !found {
		return fmt.Errorf("missing space after mailbox flags")
	}

	// Hierarchy delimiter
	if found, err := s.SkipByte('"'); err != nil {
		return err
	} else if !found {
		return fmt.Errorf("missing first '\"' for hierarchy delimiter")
	}

	if found, err := s.SkipBytes([]byte("NIL")); err != nil {
		return err
	} else if !found {
		c, _, err := s.ReadIMAPQuotedChar()
		if err != nil {
			return err
		}

		r.HierarchyDelimiter = rune(c)
	}

	if found, err := s.SkipByte('"'); err != nil {
		return err
	} else if !found {
		return fmt.Errorf("missing last '\"' for hierarchy delimiter")
	}

	if found, err := s.SkipByte(' '); err != nil {
		return err
	} else if !found {
		return fmt.Errorf("missing space after hierarchy delimiter")
	}

	// Name
	encodedName, err := s.ReadIMAPAstring()
	if err != nil {
		return err
	}
	name, err := ModifiedUTF7Decode(encodedName)
	if err != nil {
		return fmt.Errorf("invalid mailbox name: %v", err)
	}
	r.MailboxName = string(name)

	// End
	if ok, err := s.SkipBytes([]byte("\r\n")); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("invalid character after mailbox name")
	}

	return nil
}

// FLAGS
type ResponseFlags struct {
	Flags []string
}

func (r *ResponseFlags) Name() string { return "FLAGS" }

func (r *ResponseFlags) GoString() string {
	return fmt.Sprintf("#<response-flags %v>", r.Flags)
}

func (r *ResponseFlags) Read(s *Stream) error {
	// Flags
	flags, err := s.ReadIMAPFlagList()
	if err != nil {
		return err
	}

	r.Flags = flags

	// End
	if ok, err := s.SkipBytes([]byte("\r\n")); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("invalid character after flag list")
	}

	return nil
}

// EXISTS
type ResponseExists struct {
	Count uint32
}

func (r *ResponseExists) Name() string { return "EXISTS" }

func (r *ResponseExists) GoString() string {
	return fmt.Sprintf("#<response-exists %d>", r.Count)
}

func (r *ResponseExists) Read(s *Stream) error {
	// End
	if ok, err := s.SkipBytes([]byte("\r\n")); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("invalid character after EXISTS")
	}

	return nil
}

// RECENT
type ResponseRecent struct {
	Count uint32
}

func (r *ResponseRecent) Name() string { return "RECENT" }

func (r *ResponseRecent) GoString() string {
	return fmt.Sprintf("#<response-recent %d>", r.Count)
}

func (r *ResponseRecent) Read(s *Stream) error {
	// End
	if ok, err := s.SkipBytes([]byte("\r\n")); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("invalid character after RECENT")
	}

	return nil
}

// ---------------------------------------------------------------------------
//  Command continuation responses
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

func (r *ResponseContinuation) GoString() string {
	return fmt.Sprintf("#<response-continuation %q>", r.Text)
}

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

		var codeData []byte

		idx := bytes.IndexByte(codeBytes, ' ')
		if idx == -1 {
			r.Code = string(codeBytes)
			codeData = []byte{}
		} else {
			r.Code = string(codeBytes[0:idx])
			codeData = codeBytes[idx+1:]
		}

		switch r.Code {
		case "CAPABILITY":
			parts := bytes.Split(codeData, []byte{' '})
			caps := make([]string, len(parts))
			for i, cap := range parts {
				caps[i] = string(cap)
			}
			r.CodeData = caps

		case "HIGHESTMODSEQ":
			fallthrough
		case "UIDNEXT":
			fallthrough
		case "UIDVALIDITY":
			fallthrough
		case "UNSEEN":
			n, err := strconv.ParseUint(string(codeData), 10, 32)
			if err != nil {
				return fmt.Errorf("invalid %s data: %v",
					r.Code, err)
			}
			if n == 0 {
				return fmt.Errorf("invalid zero value for %s",
					r.Code)
			}
			r.CodeData = n

		case "PERMANENTFLAGS":
			if len(codeData) < 2 ||
				codeData[0] != '(' || codeData[1] != ')' {
				return fmt.Errorf("invalid %q data", r.Code)
			}

			parts := bytes.Split(codeData[1:len(codeData)-1],
				[]byte{' '})
			flags := make([]string, len(parts))
			for i, flag := range parts {
				flags[i] = string(flag)
			}
			r.CodeData = flags
		}
	}

	// Response text
	text, err := s.ReadUntilAndSkip([]byte("\r\n"))
	if err != nil {
		return err
	}

	r.Text = string(text)
	return nil
}

func (r *ResponseText) GoString() string {
	if r.Code == "" {
		return fmt.Sprintf("#<response-text %q>", r.Text)
	} else {
		return fmt.Sprintf("#<response-text %s %v %q>",
			r.Code, r.CodeData, r.Text)
	}
}
