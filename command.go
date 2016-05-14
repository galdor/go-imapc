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
	"crypto/hmac"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

type Command interface {
	Write(*BufferedWriter)
	Continue(*BufferedWriter, *ResponseContinuation) error
}

// ---------------------------------------------------------------------------
//  Command: AUTHENTICATE
// ---------------------------------------------------------------------------
// PLAIN (RFC 4616)
type CommandAuthenticatePlain struct {
	Login    string
	Password string
}

func (c *CommandAuthenticatePlain) Write(w *BufferedWriter) {
	w.AppendString("AUTHENTICATE PLAIN\r\n")
}

func (c *CommandAuthenticatePlain) Continue(w *BufferedWriter, r *ResponseContinuation) error {
	creds := fmt.Sprintf("\x00%s\x00%s", c.Login, c.Password)
	ecreds := base64.StdEncoding.EncodeToString([]byte(creds))

	w.AppendString(ecreds)
	w.AppendString("\r\n")

	return nil
}

// CRAM-MD5 (RFC 2195)
type CommandAuthenticateCramMD5 struct {
	Login    string
	Password string
}

func (c *CommandAuthenticateCramMD5) Write(w *BufferedWriter) {
	w.AppendString("AUTHENTICATE CRAM-MD5\r\n")
}

func (c *CommandAuthenticateCramMD5) Continue(w *BufferedWriter, r *ResponseContinuation) error {
	challenge, err := base64.StdEncoding.DecodeString(r.Text)
	if err != nil {
		return fmt.Errorf("cannot decode challenge: %v", err)
	}

	enc := hmac.New(md5.New, []byte(c.Password))
	enc.Write(challenge)
	hashedPassword := hex.EncodeToString(enc.Sum(nil))

	creds := c.Login + " " + hashedPassword
	ecreds := base64.StdEncoding.EncodeToString([]byte(creds))

	w.AppendString(ecreds)
	w.AppendString("\r\n")

	return nil
}

// DIGEST-MD5 (RFC 2831)
// TODO

// SCRAM-SHA-1 (RFC 5802)
// TODO

// ---------------------------------------------------------------------------
//  Command: CAPABILITY
// ---------------------------------------------------------------------------
type CommandCapability struct {
}

func (c *CommandCapability) Write(w *BufferedWriter) {
	w.AppendString("CAPABILITY\r\n")
}

func (c *CommandCapability) Continue(w *BufferedWriter, r *ResponseContinuation) error {
	return nil
}
