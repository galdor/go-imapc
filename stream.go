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
	"io"
	"math"
	"strconv"
)

type Stream struct {
	Reader io.Reader
	Buf    []byte
}

func NewStream(r io.Reader) *Stream {
	return &Stream{
		Reader: r,
		Buf:    []byte{},
	}
}

func (s *Stream) IsEmpty() (bool, error) {
	if len(s.Buf) > 0 {
		return false, nil
	}

	_, err := s.Peek(1)
	if err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return true, nil
		}

		return false, err
	}

	return false, nil
}

func (s *Stream) Peek(n int) ([]byte, error) {
	if len(s.Buf) < n {
		blen := len(s.Buf)
		rest := n - blen
		s.Buf = append(s.Buf, make([]byte, rest)...)

		nbRead, err := io.ReadAtLeast(s.Reader, s.Buf[blen:], rest)
		if err != nil {
			s.Buf = s.Buf[0 : blen+nbRead]
			return nil, err
		}
	}

	return dupBytes(s.Buf[0:n]), nil
}

func (s *Stream) PeekUpTo(n int) ([]byte, error) {
	_, err := s.Peek(n)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, err
	}

	if n >= len(s.Buf) {
		n = len(s.Buf)
	}

	return dupBytes(s.Buf[0:n]), nil
}

func (s *Stream) StartsWithBytes(bs []byte) (bool, error) {
	data, err := s.Peek(len(bs))
	if err != nil {
		return false, err
	}

	return bytes.Equal(data, bs), nil
}

func (s *Stream) StartsWithByte(b byte) (bool, error) {
	data, err := s.Peek(1)
	if err != nil {
		return false, err
	}

	return data[0] == b, nil
}

func (s *Stream) Skip(n int) error {
	_, err := s.Peek(n)
	if err != nil {
		return err
	}

	s.Buf = s.Buf[n:]
	return nil
}

func (s *Stream) SkipBytes(bs []byte) (bool, error) {
	data, err := s.Peek(len(bs))
	if err != nil {
		return false, err
	}

	if bytes.Equal(data, bs) {
		s.Buf = s.Buf[len(bs):]
		return true, nil
	}

	return false, nil
}

func (s *Stream) SkipByte(b byte) (bool, error) {
	return s.SkipBytes([]byte{b})
}

func (s *Stream) SkipWhile(fn func(byte) bool) error {
	_, err := s.ReadWhile(fn)
	return err
}

func (s *Stream) Read(n int) ([]byte, error) {
	data, err := s.Peek(n)
	if err != nil {
		return nil, err
	}

	s.Buf = s.Buf[n:]
	return data, nil
}

func (s *Stream) ReadAll() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})

loop:
	for {
		data, err := s.PeekUpTo(4096)
		if err != nil {
			return nil, err
		}

		buf.Write(data)
		s.Buf = s.Buf[len(data):]

		empty, err := s.IsEmpty()
		if err != nil {
			return nil, err
		} else if empty {
			break loop
		}
	}

	return buf.Bytes(), nil
}

func (s *Stream) ReadWhile(fn func(byte) bool) ([]byte, error) {
	const blockSize = 32

	data := []byte{}

loop:
	for {
		block, err := s.PeekUpTo(blockSize)
		if err != nil {
			return nil, err
		} else if len(block) == 0 {
			break loop
		}

		for i, b := range block {
			if !fn(b) {
				s.Buf = s.Buf[i:]
				break loop
			}

			data = append(data, b)
		}

		s.Buf = s.Buf[len(block):]
	}

	return data, nil
}

func (s *Stream) PeekUntil(delim []byte) ([]byte, error) {
	for {
		idx := bytes.Index(s.Buf, delim)
		if idx >= 0 {
			return dupBytes(s.Buf[0:idx]), nil
		}

		blen := len(s.Buf)
		s.Buf = append(s.Buf, make([]byte, 4096)...)

		nread, err := s.Reader.Read(s.Buf[blen:])
		if err != nil {
			return nil, err
		}
		s.Buf = s.Buf[0 : blen+nread]
	}
}

func (s *Stream) PeekUntilByte(delim byte) ([]byte, error) {
	return s.PeekUntil([]byte{delim})
}

func (s *Stream) ReadUntil(delim []byte) ([]byte, error) {
	data, err := s.PeekUntil(delim)
	if err != nil {
		return nil, err
	}

	s.Buf = s.Buf[len(data):]
	return data, nil
}

func (s *Stream) ReadUntilAndSkip(delim []byte) ([]byte, error) {
	data, err := s.PeekUntil(delim)
	if err != nil {
		return nil, err
	}

	s.Buf = s.Buf[len(data)+len(delim):]
	return data, nil
}

func (s *Stream) ReadUntilByte(delim byte) ([]byte, error) {
	return s.ReadUntil([]byte{delim})
}

func (s *Stream) ReadUntilByteAndSkip(delim byte) ([]byte, error) {
	return s.ReadUntilAndSkip([]byte{delim})
}

func (s *Stream) ReadIMAPQuotedChar() (byte, bool, error) {
	data, err := s.Read(1)
	if err != nil {
		return 0, false, err
	}

	var c byte
	quoted := false

	if data[0] == '\\' {
		data, err = s.Read(1)
		if err != nil {
			return 0, false, err
		}

		if !IsQuotedSpecialChar(data[0]) {
			return 0, false,
				fmt.Errorf("invalid quoted character %q", c)
		}

		quoted = true
	}

	if data[0] == '\r' || data[0] == '\n' {
		return c, quoted, fmt.Errorf("invalid quoted character %q", c)
	}

	return data[0], quoted, nil
}

func (s *Stream) ReadIMAPAstring() ([]byte, error) {
	if ok, err := s.StartsWithByte('"'); err != nil {
		return nil, err
	} else if ok {
		return s.ReadIMAPQuotedString()
	}

	if ok, err := s.StartsWithByte('{'); err != nil {
		return nil, err
	} else if ok {
		return s.ReadIMAPLiteralString()
	}

	data, err := s.ReadWhile(IsAstringChar)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *Stream) ReadIMAPQuotedString() ([]byte, error) {
	if found, err := s.SkipByte('"'); err != nil {
		return nil, err
	} else if !found {
		return nil, fmt.Errorf("missing first '\"' for quoted string")
	}

	data := []byte{}

loop:
	for {
		c, quoted, err := s.ReadIMAPQuotedChar()
		if err != nil {
			return nil, err
		}

		if c == '"' && !quoted {
			break loop
		}

		data = append(data, c)
	}

	return data, nil
}

func (s *Stream) ReadIMAPLiteralString() ([]byte, error) {
	if found, err := s.SkipByte('{'); err != nil {
		return nil, err
	} else if !found {
		return nil, fmt.Errorf("missing '{' for literal string")
	}

	countData, err := s.ReadUntilByteAndSkip('}')
	if err != nil {
		return nil, err
	}

	count, err := strconv.ParseUint(string(countData), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid literal string size")
	}

	// Go uses the 'int' type for all length, but it can be 32 bit on 32
	// bit platform.
	if count > math.MaxInt32 {
		return nil, fmt.Errorf("literal string size too large")
	}

	if found, err := s.SkipBytes([]byte("\r\n")); err != nil {
		return nil, err
	} else if !found {
		return nil,
			fmt.Errorf("missing \\r\\n after literal string size")
	}

	data, err := s.Read(int(count))
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *Stream) ReadIMAPFlagList() ([]string, error) {
	if found, err := s.SkipByte('('); err != nil {
		return nil, err
	} else if !found {
		return nil, fmt.Errorf("missing '(' for flag list")
	}

	flagsData, err := s.ReadUntilByteAndSkip(')')
	if err != nil {
		return nil, err
	}

	if len(flagsData) == 0 {
		return []string{}, nil
	}

	parts := bytes.Split(flagsData, []byte{' '})
	flags := make([]string, len(parts))

	for i, part := range parts {
		flags[i] = string(part)
	}

	return flags, nil
}

func (s *Stream) ReadIMAPNumber() (uint32, error) {
	data, err := s.ReadWhile(IsDigitChar)
	if err != nil {
		return 0, err
	}

	n, err := strconv.ParseUint(string(data), 10, 32)
	if err != nil {
		return 0, err
	}

	return uint32(n), err
}

func (s *Stream) ReadIMAPMailboxList() (*MailboxList, error) {
	mbox := &MailboxList{}

	// Mailbox flags
	flags, err := s.ReadIMAPFlagList()
	if err != nil {
		return nil, err
	}
	mbox.Flags = flags

	if found, err := s.SkipByte(' '); err != nil {
		return nil, err
	} else if !found {
		return nil, fmt.Errorf("missing space after mailbox flags")
	}

	// Hierarchy delimiter
	if found, err := s.SkipByte('"'); err != nil {
		return nil, err
	} else if !found {
		return nil,
			fmt.Errorf("missing first '\"' for hierarchy delimiter")
	}

	if found, err := s.SkipBytes([]byte("NIL")); err != nil {
		return nil, err
	} else if !found {
		c, _, err := s.ReadIMAPQuotedChar()
		if err != nil {
			return nil, err
		}

		mbox.HierarchyDelimiter = rune(c)
	}

	if found, err := s.SkipByte('"'); err != nil {
		return nil, err
	} else if !found {
		return nil,
			fmt.Errorf("missing last '\"' for hierarchy delimiter")
	}

	if found, err := s.SkipByte(' '); err != nil {
		return nil, err
	} else if !found {
		return nil,
			fmt.Errorf("missing space after hierarchy delimiter")
	}

	// Name
	encodedName, err := s.ReadIMAPAstring()
	if err != nil {
		return nil, err
	}
	name, err := ModifiedUTF7Decode(encodedName)
	if err != nil {
		return nil, fmt.Errorf("invalid mailbox name: %v", err)
	}
	mbox.Name = string(name)

	return mbox, nil
}

func dupBytes(data []byte) []byte {
	ndata := make([]byte, len(data))
	copy(ndata, data)
	return ndata
}
