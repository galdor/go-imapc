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
	"io"
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

func (s *Stream) Peek(n int) ([]byte, error) {
	if len(s.Buf) < n {
		blen := len(s.Buf)
		rest := n - blen
		s.Buf = append(s.Buf, make([]byte, rest)...)

		_, err := io.ReadAtLeast(s.Reader, s.Buf[blen:], rest)
		if err != nil {
			return nil, err
		}
	}

	return dupBytes(s.Buf[0:n]), nil
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

func (s *Stream) Read(n int) ([]byte, error) {
	data, err := s.Peek(n)
	if err != nil {
		return nil, err
	}

	s.Buf = s.Buf[n:]
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

func dupBytes(data []byte) []byte {
	ndata := make([]byte, len(data))
	copy(ndata, data)
	return ndata
}
