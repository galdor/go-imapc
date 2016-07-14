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
	"unicode/utf16"
	"unicode/utf8"
)

func ModifiedUTF7Encode(s string) string {
	buf := bytes.NewBuffer([]byte{})

	sb := []byte(s)

loop:
	for {
		// Find the next non-directly encodable character
		end := bytes.IndexFunc(sb, func(r rune) bool {
			return (r < 0x20 || r > 0x7e)
		})

		// Write all directly encodable characters
		if end == -1 {
			end = len(sb)
		}

		for i := 0; i < end; i++ {
			if sb[i] == '&' {
				buf.WriteString("&-")
			} else {
				buf.WriteByte(sb[i])
			}
		}

		if end == len(sb) {
			break loop
		}

		sb = sb[end:]

		// Find the next directly encodable character
		end = bytes.IndexFunc(sb, func(r rune) bool {
			return (r >= 0x20 && r <= 0x7e)
		})

		if end == -1 {
			end = len(sb)
		}

		// Encode and write non-directly encodable characters
		buf.WriteByte('&')
		buf.Write(ModifiedBase64Encode(sb[:end]))
		buf.WriteByte('-')

		// Skip to the end of the sequence
		if end == len(sb) {
			break loop
		}

		sb = sb[end:]
	}

	return buf.String()
}

func ModifiedUTF7Decode(s string) (string, error) {
	buf := bytes.NewBuffer([]byte{})

	sb := []byte(s)

loop:
	for len(sb) > 0 {
		// Find the next encoded sequence
		end := bytes.IndexByte(sb, byte('&'))

		if end == -1 {
			buf.Write(sb)
			break
		}

		// Write and skip directly encoded characters
		buf.Write(sb[:end])

		sb = sb[end:]

		// Find the end of the sequence
		end = bytes.IndexByte(sb, byte('-'))
		if end == -1 {
			// '&' without '-'
			return "", fmt.Errorf("invalid modified utf7 encoding")
		}

		if end == 1 {
			buf.WriteByte('&')
			sb = sb[2:]
			continue
		}

		// Decode and write the sequence
		seq, err := ModifiedBase64Decode(sb[1:end])
		if err != nil {
			return "", err
		}

		buf.Write(seq)

		// Skip to the end of the sequence
		if end == len(sb) {
			break loop
		}

		sb = sb[end+1:]
	}

	return buf.String(), nil
}

func ModifiedBase64Encode(data []byte) []byte {
	// Convert utf8 data to utf16-be
	u16data := UTF8ToUTF16BE(data)

	// Encode using base64 without padding
	const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+,"

	out := bytes.NewBuffer([]byte{})

	for len(u16data) >= 3 {
		a := ((u16data[0] & 0xfc) >> 2)
		b := ((u16data[0] & 0x03) << 4) | ((u16data[1] & 0xf0) >> 4)
		c := ((u16data[1] & 0x0f) << 2) | ((u16data[2] & 0xc0) >> 6)
		d := (u16data[2] & 0x3f)

		out.Write([]byte{base64Chars[a], base64Chars[b],
			base64Chars[c], base64Chars[d]})

		u16data = u16data[3:]
	}

	switch len(u16data) {
	case 2:
		a := ((u16data[0] & 0xfc) >> 2)
		b := ((u16data[0] & 0x03) << 4) | ((u16data[1] & 0xf0) >> 4)
		c := ((u16data[1] & 0x0f) << 2)

		out.Write([]byte{base64Chars[a], base64Chars[b],
			base64Chars[c]})

	case 1:
		a := (u16data[0] & 0xfc) >> 2
		b := (u16data[0] & 0x03) << 4

		out.Write([]byte{base64Chars[a], base64Chars[b]})
	}

	return out.Bytes()
}

func ModifiedBase64Decode(data []byte) ([]byte, error) {
	// Decode base64-encoded to utf16-be
	var table = [256]int8{
		-1, -1, -1, -1, -1, -1, -1, -1, // 00 - 07
		-1, -1, -1, -1, -1, -1, -1, -1, // 08 - 0f
		-1, -1, -1, -1, -1, -1, -1, -1, // 10 - 17
		-1, -1, -1, -1, -1, -1, -1, -1, // 18 - 1f
		-1, -1, -1, -1, -1, -1, -1, -1, // 20 - 27
		-1, -1, -1, 62, 63, -1, -1, -1, // 28 - 2f
		52, 53, 54, 55, 56, 57, 58, 59, // 30 - 37
		60, 61, -1, -1, -1, -1, -1, -1, // 38 - 3f
		-1, 0, 1, 2, 3, 4, 5, 6, // 40 - 47
		7, 8, 9, 10, 11, 12, 13, 14, // 48 - 4f
		15, 16, 17, 18, 19, 20, 21, 22, // 50 - 57
		23, 24, 25, -1, -1, -1, -1, -1, // 58 - 5f
		-1, 26, 27, 28, 29, 30, 31, 32, // 60 - 67
		33, 34, 35, 36, 37, 38, 39, 40, // 68 - 6f
		41, 42, 43, 44, 45, 46, 47, 48, // 70 - 77
		49, 50, 51, -1, -1, -1, -1, -1, // 78 - 7f
		-1, -1, -1, -1, -1, -1, -1, -1, // 80 - 87
		-1, -1, -1, -1, -1, -1, -1, -1, // 88 - 8f
		-1, -1, -1, -1, -1, -1, -1, -1, // 90 - 97
		-1, -1, -1, -1, -1, -1, -1, -1, // 98 - 9f
		-1, -1, -1, -1, -1, -1, -1, -1, // a0 - a7
		-1, -1, -1, -1, -1, -1, -1, -1, // a8 - af
		-1, -1, -1, -1, -1, -1, -1, -1, // b0 - b7
		-1, -1, -1, -1, -1, -1, -1, -1, // b8 - bf
		-1, -1, -1, -1, -1, -1, -1, -1, // c0 - c7
		-1, -1, -1, -1, -1, -1, -1, -1, // c8 - cf
		-1, -1, -1, -1, -1, -1, -1, -1, // d0 - d7
		-1, -1, -1, -1, -1, -1, -1, -1, // d8 - df
		-1, -1, -1, -1, -1, -1, -1, -1, // e0 - e7
		-1, -1, -1, -1, -1, -1, -1, -1, // e8 - ef
		-1, -1, -1, -1, -1, -1, -1, -1, // f0 - f7
		-1, -1, -1, -1, -1, -1, -1, -1, // f8 - ff
	}

	u16buf := bytes.NewBuffer([]byte{})

	for len(data) > 0 {
		var a, b, c, d int8

		if len(data) >= 4 {
			a = table[data[0]]
			b = table[data[1]]
			c = table[data[2]]
			d = table[data[3]]

			if a == -1 || b == -1 || c == -1 || d == -1 {
				return nil,
					fmt.Errorf("invalid base64 character")
			}

			u16buf.Write([]byte{
				byte((a << 2) | (b >> 4)),
				byte((b << 4) | (c >> 2)),
				byte((c << 6) | d),
			})

			data = data[4:]
		} else if len(data) >= 3 {
			a = table[data[0]]
			b = table[data[1]]
			c = table[data[2]]

			if a == -1 || b == -1 || c == -1 {
				return nil,
					fmt.Errorf("invalid base64 character")
			}

			u16buf.Write([]byte{
				byte((a << 2) | (b >> 4)),
				byte((b << 4) | (c >> 2)),
			})

			data = data[3:]
		} else if len(data) >= 2 {
			a = table[data[0]]
			b = table[data[1]]

			if a == -1 || b == -1 {
				return nil,
					fmt.Errorf("invalid base64 character")
			}

			u16buf.Write([]byte{
				byte((a << 2) | (b >> 4)),
			})

			data = data[2:]
		} else {
			return nil, fmt.Errorf("invalid base64 encoding")
		}
	}

	// Decode utf16-be to utf8
	nbchars := len(data)
	if nbchars%2 != 0 {
		return nil, fmt.Errorf("invalid utf16 data")
	}

	return UTF16BEToUTF8(u16buf.Bytes()), nil
}

func UTF8ToUTF16BE(data []byte) []byte {
	u16buf := bytes.NewBuffer([]byte{})

	for len(data) > 0 {
		r, sz := utf8.DecodeRune(data)

		if r1, r2 := utf16.EncodeRune(r); r1 != 0xfffd && r2 != 0xfffd {
			u16buf.Write([]byte{byte(r1 >> 8), byte(r1),
				byte(r2 >> 8), byte(r2)})
		} else {
			u16buf.Write([]byte{byte(r >> 8), byte(r)})
		}

		data = data[sz:]
	}

	return u16buf.Bytes()
}

func UTF16BEToUTF8(data []byte) []byte {
	nbchars := len(data)

	u16s := make([]uint16, nbchars/2)
	for i := 0; i < nbchars; i += 2 {
		u16s[i/2] = (uint16(data[i]) << 8) | uint16(data[i+1])
	}

	runes := utf16.Decode(u16s)

	size := 0
	for _, r := range runes {
		size += utf8.RuneLen(r)
	}

	out := make([]byte, size)

	i := 0
	for _, r := range runes {
		i += utf8.EncodeRune(out[i:], r)
	}

	return out
}
