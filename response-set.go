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

type ResponseSet interface {
	Init([]Response, *ResponseStatus) error
}

// ---------------------------------------------------------------------------
//  Response set: LIST
// ---------------------------------------------------------------------------
type ResponseSetList struct {
	Mailboxes []MailboxList
}

func (rs *ResponseSetList) Init(resps []Response, status *ResponseStatus) error {
	rs.Mailboxes = []MailboxList{}

	for _, resp := range resps {
		switch tresp := resp.(type) {
		case *ResponseList:
			mbox := MailboxList{
				Flags:              tresp.Flags,
				HierarchyDelimiter: tresp.HierarchyDelimiter,
				Name:               tresp.Name,
			}

			rs.Mailboxes = append(rs.Mailboxes, mbox)
		}
	}

	return nil
}

// ---------------------------------------------------------------------------
//  Response set: LSUB
// ---------------------------------------------------------------------------
type ResponseSetLSub struct {
	Mailboxes []MailboxList
}

func (rs *ResponseSetLSub) Init(resps []Response, status *ResponseStatus) error {
	rs.Mailboxes = []MailboxList{}

	for _, resp := range resps {
		switch tresp := resp.(type) {
		case *ResponseLSub:
			mbox := MailboxList{
				Flags:              tresp.Flags,
				HierarchyDelimiter: tresp.HierarchyDelimiter,
				Name:               tresp.Name,
			}

			rs.Mailboxes = append(rs.Mailboxes, mbox)
		}
	}

	return nil
}

// ---------------------------------------------------------------------------
//  Response set: EXAMINE
// ---------------------------------------------------------------------------
type ResponseSetExamine struct {
	Flags          []string
	Exists         uint32
	Recent         uint32
	Unseen         uint32
	PermanentFlags []string
	UIDNext        uint32
	UIDValidity    uint32
}

func (rs *ResponseSetExamine) Init(resps []Response, status *ResponseStatus) error {
	for _, resp := range resps {
		switch tresp := resp.(type) {
		case *ResponseFlags:
			rs.Flags = tresp.Flags
		case *ResponseExists:
			rs.Exists = tresp.Count
		case *ResponseRecent:
			rs.Recent = tresp.Count
		case *ResponseOk:
			codeData := tresp.Text.CodeData

			switch tresp.Text.Code {
			case "UNSEEN":
				rs.Unseen = codeData.(uint32)
			case "PERMANENTFLAGS":
				rs.PermanentFlags = codeData.([]string)
			case "UIDNEXT":
				rs.UIDNext = codeData.(uint32)
			case "UIDVALIDITY":
				rs.UIDValidity = codeData.(uint32)
			}
		}
	}

	return nil
}

// ---------------------------------------------------------------------------
//  Response set: SELECT
// ---------------------------------------------------------------------------
type ResponseSetSelect struct {
	Flags          []string
	Exists         uint32
	Recent         uint32
	Unseen         uint32
	PermanentFlags []string
	UIDNext        uint32
	UIDValidity    uint32
}

func (rs *ResponseSetSelect) Init(resps []Response, status *ResponseStatus) error {
	for _, resp := range resps {
		switch tresp := resp.(type) {
		case *ResponseFlags:
			rs.Flags = tresp.Flags
		case *ResponseExists:
			rs.Exists = tresp.Count
		case *ResponseRecent:
			rs.Recent = tresp.Count
		case *ResponseOk:
			codeData := tresp.Text.CodeData

			switch tresp.Text.Code {
			case "UNSEEN":
				rs.Unseen = codeData.(uint32)
			case "PERMANENTFLAGS":
				rs.PermanentFlags = codeData.([]string)
			case "UIDNEXT":
				rs.UIDNext = codeData.(uint32)
			case "UIDVALIDITY":
				rs.UIDValidity = codeData.(uint32)
			}
		}
	}

	return nil
}

// ---------------------------------------------------------------------------
//  Response set: SEARCH
// ---------------------------------------------------------------------------
type ResponseSetSearch struct {
	MessageSequenceNumbers []uint32
}

func (rs *ResponseSetSearch) Init(resps []Response, status *ResponseStatus) error {
	for _, resp := range resps {
		switch tresp := resp.(type) {
		case *ResponseSearch:
			seqNums := append(rs.MessageSequenceNumbers,
				tresp.MessageSequenceNumbers...)
			rs.MessageSequenceNumbers = seqNums
		}
	}

	return nil
}
