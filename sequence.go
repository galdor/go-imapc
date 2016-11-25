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
	"encoding"
	"fmt"
	"strconv"
)

// ---------------------------------------------------------------------------
//  Sequence number
// ---------------------------------------------------------------------------
type SequenceNumber uint32

func (n SequenceNumber) MarshalText() ([]byte, error) {
	return []byte(n.String()), nil
}

func (n SequenceNumber) String() string {
	if n == SequenceStar {
		return "*"
	} else {
		return strconv.FormatUint(uint64(n), 10)
	}
}

func (n SequenceNumber) GoString() string {
	return n.String()
}

// SequenceStar represents '*', i.e. the largest number in use in the current
// context. We use zero since it's available: sequence numbers are always
// greater than zero.
const SequenceStar SequenceNumber = 0

// ---------------------------------------------------------------------------
//  Sequence range
// ---------------------------------------------------------------------------
type SequenceRange struct {
	First SequenceNumber
	Last  SequenceNumber
}

func NewSequenceRange(first, last SequenceNumber) SequenceRange {
	return SequenceRange{
		First: first,
		Last:  last,
	}
}

func (r SequenceRange) String() string {
	return fmt.Sprintf("%v:%v", r.First, r.Last)
}

func (r SequenceRange) GoString() string {
	return fmt.Sprintf("#<sequence-range %#v:%#v>", r.First, r.Last)
}

func (r SequenceRange) MarshalText() ([]byte, error) {
	first, err := r.First.MarshalText()
	if err != nil {
		return nil, err
	}

	last, err := r.Last.MarshalText()
	if err != nil {
		return nil, err
	}

	str := fmt.Sprintf("%s:%s", first, last)
	return []byte(str), nil
}

// ---------------------------------------------------------------------------
//  Sequence set
// ---------------------------------------------------------------------------
type SequenceSetEntry interface {
	fmt.Stringer
	fmt.GoStringer
	encoding.TextMarshaler
}

type SequenceSet []SequenceSetEntry

func NewSequenceSet() SequenceSet {
	return SequenceSet{}
}

func (s SequenceSet) String() string {
	buf := bytes.NewBuffer([]byte{})

	for i, e := range s {
		if i > 0 {
			buf.WriteByte(',')
		}

		buf.WriteString(e.String())
	}

	return buf.String()
}

func (s SequenceSet) GoString() string {
	buf := bytes.NewBuffer([]byte{})
	buf.WriteString("#<sequence-set ")

	for i, e := range s {
		if i > 0 {
			buf.WriteByte(',')
		}

		buf.WriteString(e.GoString())
	}

	buf.WriteString(">")
	return buf.String()
}

func (s SequenceSet) MarshalText() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})

	for i, e := range s {
		if i > 0 {
			buf.WriteByte(',')
		}

		etext, err := e.MarshalText()
		if err != nil {
			return nil, err
		}

		buf.Write(etext)
	}

	return buf.Bytes(), nil
}

func (s *SequenceSet) Append(e SequenceSetEntry) {
	*s = append(*s, e)
}
