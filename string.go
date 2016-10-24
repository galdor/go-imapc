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

func QuoteString(str string) []byte {
	return QuoteByteString([]byte(str))
}

func QuoteByteString(str []byte) []byte {
	qstr := make([]byte, len(str)+2)

	qstr[0] = '"'
	i := 1
	for _, b := range str {
		if b == '"' || b == '\\' {
			qstr = append(qstr, 0)
			qstr[i] = '\\'
			i++
		}

		qstr[i] = b
		i++
	}
	qstr[i] = '"'

	return qstr
}

func ByteStringAll(str []byte, fn func(byte) bool) bool {
	for _, b := range str {
		if !fn(b) {
			return false
		}
	}

	return true
}
