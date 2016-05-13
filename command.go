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
	"encoding/base64"
	"fmt"
	"io"
)

type Command interface {
	Write(io.Writer) error
	Continue(io.Writer, *ResponseContinuation) error
}

// ---------------------------------------------------------------------------
//  Command: AUTHENTICATE
//  RFC 3501 6.2.2.
// ---------------------------------------------------------------------------
// PLAIN (RFC 4616)
type CommandAuthenticatePlain struct {
	Login    string
	Password string
}

func (c *CommandAuthenticatePlain) Write(w io.Writer) error {
	_, err := io.WriteString(w, "AUTHENTICATE PLAIN\r\n")
	return err
}

func (c *CommandAuthenticatePlain) Continue(w io.Writer, r *ResponseContinuation) error {
	creds := fmt.Sprintf("\x00%s\x00%s", c.Login, c.Password)
	ecreds := base64.StdEncoding.EncodeToString([]byte(creds))
	_, err := io.WriteString(w, ecreds+"\r\n")
	return err
}

// CRAM-MD5 (RFC 2195)
// TODO

// DIGEST-MD5 (RFC 2831)
// TODO

// SCRAM-SHA-1 (RFC 5802)
// TODO
