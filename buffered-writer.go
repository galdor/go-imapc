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
	"io"
)

type BufferedWriter struct {
	Writer io.Writer
	Buffer []byte
}

func NewBufferedWriter(w io.Writer) *BufferedWriter {
	return &BufferedWriter{
		Writer: w,
	}
}

func (w *BufferedWriter) Write(data []byte) (int, error) {
	w.Append(data)
	return len(data), nil
}

func (w *BufferedWriter) Append(data []byte) {
	w.Buffer = append(w.Buffer, data...)
}

func (w *BufferedWriter) AppendString(s string) {
	w.Append([]byte(s))
}

func (w *BufferedWriter) Reset() {
	w.Buffer = w.Buffer[:0]
}

func (w *BufferedWriter) Flush() error {
	if _, err := w.Writer.Write(w.Buffer); err != nil {
		return err
	}

	w.Buffer = w.Buffer[:0]
	return nil
}
