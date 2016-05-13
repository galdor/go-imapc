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
)

type Command interface {
	Write(*BufferedWriter)
	Continue(*BufferedWriter, *ResponseContinuation)
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

func (c *CommandAuthenticatePlain) Write(w *BufferedWriter) {
	w.AppendString("AUTHENTICATE PLAIN\r\n")
}

func (c *CommandAuthenticatePlain) Continue(w *BufferedWriter, r *ResponseContinuation) {
	creds := fmt.Sprintf("\x00%s\x00%s", c.Login, c.Password)
	ecreds := base64.StdEncoding.EncodeToString([]byte(creds))
	w.AppendString(ecreds)
	w.AppendString("\r\n")
}

// CRAM-MD5 (RFC 2195)
// TODO

// DIGEST-MD5 (RFC 2831)
// TODO

// SCRAM-SHA-1 (RFC 5802)
// TODO
