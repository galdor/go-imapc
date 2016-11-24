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

const IMAPDateFormat = "02-Jan-2006"

type MailboxList struct {
	Flags              []string
	HierarchyDelimiter rune
	Name               string
}

func QuotedStringEncode(s string) []byte {
	return QuoteByteString(ModifiedUTF7Encode([]byte(s)))
}

func AStringEncode(s string) []byte {
	return AStringEncodeByteString([]byte(s))
}

func AStringEncodeByteString(bs []byte) []byte {
	if ByteStringAll(bs, IsAtomChar) {
		return bs
	} else {
		return QuoteByteString(ModifiedUTF7Encode(bs))
	}
}

func IsAstringChar(b byte) bool {
	return IsAtomChar(b) || IsRespSpecialChar(b)
}

func IsAtomChar(b byte) bool {
	return IsChar(b) && !IsAtomSpecialChar(b)
}

func IsAtomSpecialChar(b byte) bool {
	return b == '(' || b == ')' || b == '{' || b == ' ' ||
		IsCtlChar(b) || IsListWildcardChar(b) ||
		IsQuotedSpecialChar(b) || IsRespSpecialChar(b)
}

func IsListWildcardChar(b byte) bool {
	return b == '%' || b == '*'
}

func IsQuotedSpecialChar(b byte) bool {
	return b == '"' || b == '\\'
}

func IsRespSpecialChar(b byte) bool {
	return b == ']'
}

func IsChar(b byte) bool {
	return b < 128
}

func IsCtlChar(b byte) bool {
	return b < 32 || b == 127
}

func IsDigitChar(b byte) bool {
	return b >= '0' && b <= '9'
}
