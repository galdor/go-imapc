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
	"strconv"
	"time"
)

type SearchKey []interface{}

func SearchKeyAll() SearchKey {
	return SearchKey{"ALL"}
}

func SearchKeyAnswered() SearchKey {
	return SearchKey{"ANSWERED"}
}

func SearchKeyBCC(String string) SearchKey {
	return SearchKey{"BCC", AStringEncode(String)}
}

func SearchKeyBefore(Date time.Time) SearchKey {
	return SearchKey{"BEFORE", Date.Format(IMAPDateFormat)}
}

func SearchKeyBody(String string) SearchKey {
	return SearchKey{"BODY", AStringEncode(String)}
}

func SearchKeyCC(String string) SearchKey {
	return SearchKey{"CC", AStringEncode(String)}
}

func SearchKeyDeleted() SearchKey {
	return SearchKey{"DELETED"}
}

func SearchKeyFlagged() SearchKey {
	return SearchKey{"FLAGGED"}
}

func SearchKeyFrom(String string) SearchKey {
	return SearchKey{"FROM", AStringEncode(String)}
}

func SearchKeyKeyword(Keyword string) SearchKey {
	return SearchKey{"KEYWORD", AStringEncode(Keyword)}
}

func SearchKeyNew() SearchKey {
	return SearchKey{"NEW"}
}

func SearchKeyOld() SearchKey {
	return SearchKey{"OLD"}
}

func SearchKeyOn(Date time.Time) SearchKey {
	return SearchKey{"ON", Date.Format(IMAPDateFormat)}
}

func SearchKeyRecent() SearchKey {
	return SearchKey{"RECENT"}
}

func SearchKeySeen() SearchKey {
	return SearchKey{"SEEN"}
}

func SearchKeySince(Date time.Time) SearchKey {
	return SearchKey{"SINCE", Date.Format(IMAPDateFormat)}
}

func SearchKeySubject(String string) SearchKey {
	return SearchKey{"SUBJECT", AStringEncode(String)}
}

func SearchKeyText(String string) SearchKey {
	return SearchKey{"TEXT", AStringEncode(String)}
}

func SearchKeyTo(String string) SearchKey {
	return SearchKey{"TO", AStringEncode(String)}
}

func SearchKeyUnanswered() SearchKey {
	return SearchKey{"UNANSWERED"}
}

func SearchKeyUndeleted() SearchKey {
	return SearchKey{"UNDELETED"}
}

func SearchKeyUnflagged() SearchKey {
	return SearchKey{"UNFLAGGED"}
}

func SearchKeyUnkeyword(Keyword string) SearchKey {
	return SearchKey{"UNKEYWORD", AStringEncode(Keyword)}
}

func SearchKeyUnseen() SearchKey {
	return SearchKey{"UNSEEN"}
}

func SearchKeyDraft() SearchKey {
	return SearchKey{"DRAFT"}
}

func SearchKeyHeader(Name, String string) SearchKey {
	return SearchKey{"HEADER", AStringEncode(Name), AStringEncode(String)}
}

func SearchKeyLarger(Size uint32) SearchKey {
	return SearchKey{"LARGER", strconv.FormatInt(int64(Size), 10)}
}

func SearchKeyNot(Key SearchKey) SearchKey {
	return append(SearchKey{"NOT"}, Key...)
}

func SearchKeyOr(Key1, Key2 SearchKey) SearchKey {
	return append(append(append(SearchKey{}, "OR"), Key1...), Key2...)
}

func SearchKeySentBefore(Date time.Time) SearchKey {
	return SearchKey{"BEFORE", Date.Format(IMAPDateFormat)}
}

func SearchKeySentOn(Date time.Time) SearchKey {
	return SearchKey{"SENTON", Date.Format(IMAPDateFormat)}
}

func SearchKeySentSince(Date time.Time) SearchKey {
	return SearchKey{"SENTSINCE", Date.Format(IMAPDateFormat)}
}

func SearchKeySmaller(Size uint32) SearchKey {
	return SearchKey{"SMALLER", strconv.FormatInt(int64(Size), 10)}
}

// TODO UID

func SearchKeyUndraft() SearchKey {
	return SearchKey{"UNDRAFT"}
}

// TODO sequence-set

func SearchKeyList(list []SearchKey) SearchKey {
	key := SearchKey{"("}
	for _, k := range list {
		key = append(key, k...)
	}
	return append(key, ")")
}

func ParseSearchString(str string) (SearchKey, error) {
	data := bytes.TrimSpace([]byte(str))
	if len(data) == 0 {
		return SearchKeyAll(), nil
	}

	key := SearchKey{}

loop:
	for data != nil {
		k, rest, err := parseSearchStringKey(data)
		if err != nil {
			return nil, err
		} else if k == nil {
			break loop
		}

		key = append(key, k...)
		data = rest
	}

	return key, nil
}

func parseSearchStringKey(data []byte) (SearchKey, []byte, error) {
	data = bytes.TrimLeft(data, " \t")
	if len(data) == 0 {
		return nil, nil, nil
	}

	var tag []byte
	idx := bytes.IndexAny(data, " \t")
	if idx >= 0 {
		tag = data[:idx]
		data = data[idx+1:]
	} else {
		tag = data[:]
		data = nil
	}
	tag = bytes.ToUpper(tag)

	// TODO UID sequence-set
	// TODO sequence-set
	// TODO ( key, ... )

	const (
		argString = iota
		argDate
		argSize
		argKey
	)

	var specs []int

	switch string(tag) {
	case "ALL":
		specs = []int{}
	case "ANSWERED":
		specs = []int{}
	case "BCC":
		specs = []int{argString}
	case "BEFORE":
		specs = []int{argDate}
	case "BODY":
		specs = []int{argString}
	case "CC":
		specs = []int{argString}
	case "DELETED":
		specs = []int{}
	case "FLAGGED":
		specs = []int{}
	case "FROM":
		specs = []int{argString}
	case "KEYWORD":
		specs = []int{argString}
	case "NEW":
		specs = []int{}
	case "OLD":
		specs = []int{}
	case "ON":
		specs = []int{argDate}
	case "RECENT":
		specs = []int{}
	case "SEEN":
		specs = []int{}
	case "SINCE":
		specs = []int{argDate}
	case "SUBJECT":
		specs = []int{argString}
	case "TEXT":
		specs = []int{argString}
	case "TO":
		specs = []int{argString}
	case "UNANSWERED":
		specs = []int{}
	case "UNDELETED":
		specs = []int{}
	case "UNFLAGGED":
		specs = []int{}
	case "UNKEYWORD":
		specs = []int{argString}
	case "UNSEEN":
		specs = []int{}
	case "DRAFT":
		specs = []int{}
	case "HEADER":
		specs = []int{argString, argString}
	case "LARGER":
		specs = []int{argSize}
	case "NOT":
		specs = []int{argKey}
	case "OR":
		specs = []int{argKey, argKey}
	case "SENTBEFORE":
		specs = []int{argDate}
	case "SENTON":
		specs = []int{argDate}
	case "SENTSINCE":
		specs = []int{argDate}
	case "SMALLER":
		specs = []int{argSize}
	case "UNDRAFT":
		specs = []int{}
	default:
		return nil, nil, fmt.Errorf("unknown key %q", tag)
	}

	args := []interface{}{}
	for _, spec := range specs {
		data = bytes.TrimLeft(data, " \t")
		if len(data) == 0 {
			return nil, nil, fmt.Errorf("missing argument for "+
				"%s key", tag)
		}

		var fn func([]byte) ([]byte, []byte, error)

		if spec == argKey {
			key, rest, err := parseSearchStringKey(data)
			if err != nil {
				return nil, nil, err
			} else if key == nil {
				return nil, nil,
					fmt.Errorf("missing argument for "+
						"%s key", tag)
			}

			args = append(args, key...)
			data = rest
		} else {
			switch spec {
			case argString:
				fn = parseSearchStringArgString
			case argDate:
				fn = parseSearchStringArgDate
			case argSize:
				fn = parseSearchStringArgSize
			}

			arg, rest, err := fn(data)
			if err != nil {
				return nil, nil,
					fmt.Errorf("invalid argument for "+
						"%s key: %v", tag, err)
			}

			args = append(args, arg)
			data = rest
		}
	}

	key := append(SearchKey{tag}, args...)
	return key, data, nil
}

func parseSearchStringArgString(data []byte) ([]byte, []byte, error) {
	var arg []byte

	if data[0] == '"' {
		data = data[1:]

	loop:
		for {
			if len(data) == 0 {
				return nil, nil,
					fmt.Errorf("truncated quoted string")
			} else if data[0] == '"' {
				data = data[1:]
				break loop
			} else if data[0] == '\\' {
				data = data[1:]

				if data[0] == '\\' || data[0] == '"' {
					arg = append(arg, data[0])
				} else {
					return nil, nil,
						fmt.Errorf("invalid quoted " +
							"character")
				}
			} else {
				arg = append(arg, data[0])
			}

			data = data[1:]
		}

	} else {
		idx := bytes.IndexAny(data, " \t")
		if idx >= 0 {
			arg = data[:idx]
			data = data[idx+1:]
		} else {
			arg = data[:]
			data = nil
		}
	}

	return AStringEncodeByteString(arg), data, nil
}

func parseSearchStringArgDate(data []byte) ([]byte, []byte, error) {
	var arg []byte

	idx := bytes.IndexAny(data, " \t")
	if idx >= 0 {
		arg = data[:idx]
		data = data[idx+1:]
	} else {
		arg = data[:]
		data = nil
	}

	date, err := time.Parse("2006-01-02", string(arg))
	if err != nil {
		return nil, nil, fmt.Errorf("invalid date")
	}

	return []byte(date.Format(IMAPDateFormat)), data, nil
}

func parseSearchStringArgSize(data []byte) ([]byte, []byte, error) {
	var arg []byte

	idx := bytes.IndexAny(data, " \t")
	if idx >= 0 {
		arg = data[:idx]
		data = data[idx+1:]
	} else {
		arg = data[:]
		data = nil
	}

	n, err := strconv.ParseUint(string(arg), 10, 32)
	if err != nil || n < 0 {
		return nil, nil, fmt.Errorf("invalid size")
	}

	return []byte(strconv.FormatUint(n, 10)), data, nil
}
