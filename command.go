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
	Args() []interface{}
	Continue(*BufferedWriter, *ResponseContinuation) error
}

type CommandResponse struct {
	Data   []Response
	Status *ResponseStatus
	Error  error
}

type Literal []byte

// ---------------------------------------------------------------------------
//  Command: AUTHENTICATE
// ---------------------------------------------------------------------------
// PLAIN (RFC 4616)
type CommandAuthenticatePlain struct {
	Login    string
	Password string
}

func (c *CommandAuthenticatePlain) Args() []interface{} {
	return []interface{}{"AUTHENTICATE", "PLAIN"}
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

func (c *CommandAuthenticateCramMD5) Args() []interface{} {
	return []interface{}{"AUTHENTICATE", "CRAM-MD5"}
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

func (c *CommandCapability) Args() []interface{} {
	return []interface{}{"CAPABILITY"}
}

func (c *CommandCapability) Continue(w *BufferedWriter, r *ResponseContinuation) error {
	return nil
}

// ---------------------------------------------------------------------------
//  Command: LIST
// ---------------------------------------------------------------------------
type CommandList struct {
	Ref     string
	Pattern string
}

func (c *CommandList) Args() []interface{} {
	ref := QuotedStringEncode(c.Ref)
	pattern := QuotedStringEncode(c.Pattern)

	return []interface{}{"LIST", ref, pattern}
}

func (c *CommandList) Continue(w *BufferedWriter, r *ResponseContinuation) error {
	return nil
}

// ---------------------------------------------------------------------------
//  Command: LSUB
// ---------------------------------------------------------------------------
type CommandLSub struct {
	Ref     string
	Pattern string
}

func (c *CommandLSub) Args() []interface{} {
	ref := QuotedStringEncode(c.Ref)
	pattern := QuotedStringEncode(c.Pattern)

	return []interface{}{"LSUB", ref, pattern}
}

func (c *CommandLSub) Continue(w *BufferedWriter, r *ResponseContinuation) error {
	return nil
}

// ---------------------------------------------------------------------------
//  Command: CREATE
// ---------------------------------------------------------------------------
type CommandCreate struct {
	MailboxName string
}

func (c *CommandCreate) Args() []interface{} {
	mbox := QuotedStringEncode(c.MailboxName)

	return []interface{}{"CREATE", mbox}
}

func (c *CommandCreate) Continue(w *BufferedWriter, r *ResponseContinuation) error {
	return nil
}

// ---------------------------------------------------------------------------
//  Command: DELETE
// ---------------------------------------------------------------------------
type CommandDelete struct {
	MailboxName string
}

func (c *CommandDelete) Args() []interface{} {
	mbox := QuotedStringEncode(c.MailboxName)

	return []interface{}{"DELETE", mbox}
}

func (c *CommandDelete) Continue(w *BufferedWriter, r *ResponseContinuation) error {
	return nil
}

// ---------------------------------------------------------------------------
//  Command: RENAME
// ---------------------------------------------------------------------------
type CommandRename struct {
	MailboxName    string
	MailboxNewName string
}

func (c *CommandRename) Args() []interface{} {
	mbox := QuotedStringEncode(c.MailboxName)
	newName := QuotedStringEncode(c.MailboxNewName)

	return []interface{}{"RENAME", mbox, newName}
}

func (c *CommandRename) Continue(w *BufferedWriter, r *ResponseContinuation) error {
	return nil
}

// ---------------------------------------------------------------------------
//  Command: SUBSCRIBE
// ---------------------------------------------------------------------------
type CommandSubscribe struct {
	MailboxName string
}

func (c *CommandSubscribe) Args() []interface{} {
	mbox := QuotedStringEncode(c.MailboxName)

	return []interface{}{"SUBSCRIBE", mbox}
}

func (c *CommandSubscribe) Continue(w *BufferedWriter, r *ResponseContinuation) error {
	return nil
}

// ---------------------------------------------------------------------------
//  Command: UNSUBSCRIBE
// ---------------------------------------------------------------------------
type CommandUnsubscribe struct {
	MailboxName string
}

func (c *CommandUnsubscribe) Args() []interface{} {
	mbox := QuotedStringEncode(c.MailboxName)

	return []interface{}{"UNSUBSCRIBE", mbox}
}

func (c *CommandUnsubscribe) Continue(w *BufferedWriter, r *ResponseContinuation) error {
	return nil
}

// ---------------------------------------------------------------------------
//  Command: EXAMINE
// ---------------------------------------------------------------------------
type CommandExamine struct {
	MailboxName string
}

func (c *CommandExamine) Args() []interface{} {
	mailboxName := QuotedStringEncode(c.MailboxName)

	return []interface{}{"EXAMINE", mailboxName}
}

func (c *CommandExamine) Continue(w *BufferedWriter, r *ResponseContinuation) error {
	return nil
}

// ---------------------------------------------------------------------------
//  Command: SELECT
// ---------------------------------------------------------------------------
type CommandSelect struct {
	MailboxName string
}

func (c *CommandSelect) Args() []interface{} {
	mailboxName := QuotedStringEncode(c.MailboxName)

	return []interface{}{"SELECT", mailboxName}
}

func (c *CommandSelect) Continue(w *BufferedWriter, r *ResponseContinuation) error {
	return nil
}

// ---------------------------------------------------------------------------
//  Command: CLOSE
// ---------------------------------------------------------------------------
type CommandClose struct{}

func (c *CommandClose) Args() []interface{} {
	return []interface{}{"CLOSE"}
}

func (c *CommandClose) Continue(w *BufferedWriter, r *ResponseContinuation) error {
	return nil
}

// ---------------------------------------------------------------------------
//  Command: LOGOUT
// ---------------------------------------------------------------------------
type CommandLogout struct{}

func (c *CommandLogout) Args() []interface{} {
	return []interface{}{"LOGOUT"}
}

func (c *CommandLogout) Continue(w *BufferedWriter, r *ResponseContinuation) error {
	return nil
}

// ---------------------------------------------------------------------------
//  Command: SEARCH
// ---------------------------------------------------------------------------
type CommandSearch struct {
	Charset string
	Key     SearchKey
}

func (c *CommandSearch) Args() []interface{} {
	args := []interface{}{"SEARCH"}

	if c.Charset != "" {
		args = append(args, "CHARSET", AStringEncode(c.Charset))
	}

	return append(args, c.Key...)
}

func (c *CommandSearch) Continue(w *BufferedWriter, r *ResponseContinuation) error {
	return nil
}
