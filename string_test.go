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
	"testing"
)

func TestQuoteString(t *testing.T) {
	tests := []struct {
		str  string
		qstr string
	}{
		{``, `""`},
		{`foo`, `"foo"`},
		{`"`, `"\""`},
		{`\`, `"\\"`},
		{`"foo"`, `"\"foo\""`},
		{`\foo\`, `"\\foo\\"`},
		{`""`, `"\"\""`},
	}

	for _, test := range tests {
		qstr := QuoteString(test.str)
		if qstr != test.qstr {
			t.Errorf("%s was quoted as %s instead of %s",
				test.str, qstr, test.qstr)
		}
	}
}
