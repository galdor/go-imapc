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

var Tests = []struct {
	str  string
	estr string
}{
	{"", ""},
	{"foo", "foo"},
	{"FoO 42", "FoO 42"},
	{"&ab", "&-ab"},
	{"a&b", "a&-b"},
	{"ab&", "ab&-"},
	{"métalloïde", "m&AOk-tallo&AO8-de"},
	{"← →", "&IZAAoCGS-"},
}

func TestModifiedUTF7Encode(t *testing.T) {
	for _, test := range Tests {
		estr := ModifiedUTF7Encode(test.str)
		if estr != test.estr {
			t.Errorf("%q was encoded as %q instead of %q",
				test.str, estr, test.estr)
		}
	}
}

func TestModifiedUTF7Decode(t *testing.T) {
	for _, test := range Tests {
		str, err := ModifiedUTF7Decode(test.estr)
		if err != nil {
			t.Errorf("cannot decode %q: %v", test.estr, err)
			continue
		}

		if str != test.str {
			t.Errorf("%q was decoded as %q instead of %q",
				test.estr, str, test.str)
		}
	}
}

func TestModifiedUTF7DecodeInvalid(t *testing.T) {
	tests := []string{
		"&",
		"ab&AOk",
		"ab&A",
	}

	for _, test := range tests {
		_, err := ModifiedUTF7Decode(test)
		if err == nil {
			t.Errorf("decoded invalid string %q", test)
			continue
		}
	}
}
