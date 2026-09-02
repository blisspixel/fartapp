// Package strictjson performs bounded, schema-neutral JSON syntax inspection.
package strictjson

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Kind string

const (
	EmptyInput        Kind = "empty_input"
	Malformed         Kind = "malformed_json"
	TrailingValue     Kind = "trailing_json_value"
	MaximumDepth      Kind = "maximum_depth_exceeded"
	MemberNameTooLong Kind = "member_name_too_long"
	DuplicateMember   Kind = "duplicate_member"
)

type Limits struct {
	MaximumDepth           int
	MaximumMemberNameBytes int
}

type Issue struct {
	Kind       Kind
	Path       string
	ByteOffset int64
}

func Inspect(data []byte, limits Limits) *Issue {
	if len(TrimSpace(data)) == 0 {
		return &Issue{Kind: EmptyInput, Path: "/"}
	}
	if !utf8.Valid(data) || !validUnicodeEscapes(data) {
		return &Issue{Kind: Malformed, Path: "/"}
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if issue := scanValue(decoder, "", 0, limits); issue != nil {
		return issue
	}
	offset := decoder.InputOffset()
	if offset < int64(len(data)) && len(TrimSpace(data[offset:])) != 0 {
		return &Issue{Kind: TrailingValue, Path: "/", ByteOffset: offset}
	}
	return nil
}

func scanValue(decoder *json.Decoder, path string, depth int, limits Limits) *Issue {
	if depth > limits.MaximumDepth {
		return &Issue{Kind: MaximumDepth, Path: pathOrRoot(path), ByteOffset: decoder.InputOffset()}
	}
	token, err := decoder.Token()
	if err != nil {
		return &Issue{Kind: Malformed, Path: "/", ByteOffset: decoder.InputOffset()}
	}
	delimiter, isDelimiter := token.(json.Delim)
	if !isDelimiter {
		return nil
	}
	switch delimiter {
	case '{':
		seen := make(map[string]struct{})
		for decoder.More() {
			keyToken, keyErr := decoder.Token()
			key, ok := keyToken.(string)
			if keyErr != nil || !ok {
				return &Issue{Kind: Malformed, Path: "/", ByteOffset: decoder.InputOffset()}
			}
			if len(key) > limits.MaximumMemberNameBytes {
				return &Issue{Kind: MemberNameTooLong, Path: pathOrRoot(path), ByteOffset: decoder.InputOffset()}
			}
			childPath := joinPointer(path, key)
			if _, exists := seen[key]; exists {
				return &Issue{Kind: DuplicateMember, Path: childPath, ByteOffset: decoder.InputOffset()}
			}
			seen[key] = struct{}{}
			if issue := scanValue(decoder, childPath, depth+1, limits); issue != nil {
				return issue
			}
		}
		if _, err = decoder.Token(); err != nil {
			return &Issue{Kind: Malformed, Path: "/", ByteOffset: decoder.InputOffset()}
		}
	case '[':
		for index := 0; decoder.More(); index++ {
			childPath := joinPointer(path, strconv.Itoa(index))
			if issue := scanValue(decoder, childPath, depth+1, limits); issue != nil {
				return issue
			}
		}
		if _, err = decoder.Token(); err != nil {
			return &Issue{Kind: Malformed, Path: "/", ByteOffset: decoder.InputOffset()}
		}
	default:
		return &Issue{Kind: Malformed, Path: "/", ByteOffset: decoder.InputOffset()}
	}
	return nil
}

// TrimSpace removes only the four whitespace bytes permitted by JSON.
func TrimSpace(value []byte) []byte {
	start := 0
	for start < len(value) && isSpace(value[start]) {
		start++
	}
	end := len(value)
	for end > start && isSpace(value[end-1]) {
		end--
	}
	return value[start:end]
}

func isSpace(value byte) bool {
	return value == ' ' || value == '\t' || value == '\n' || value == '\r'
}

func validUnicodeEscapes(data []byte) bool {
	inString := false
	for index := 0; index < len(data); index++ {
		switch data[index] {
		case '"':
			inString = !inString
		case '\\':
			if !inString || index+1 >= len(data) {
				continue
			}
			if data[index+1] != 'u' {
				index++
				continue
			}
			value, valid := decodeHexQuad(data, index+2)
			if !valid {
				continue
			}
			if value >= 0xdc00 && value <= 0xdfff {
				return false
			}
			if value >= 0xd800 && value <= 0xdbff {
				if index+12 > len(data) || data[index+6] != '\\' || data[index+7] != 'u' {
					return false
				}
				low, lowValid := decodeHexQuad(data, index+8)
				if !lowValid || low < 0xdc00 || low > 0xdfff {
					return false
				}
				index += 11
				continue
			}
			index += 5
		}
	}
	return true
}

func decodeHexQuad(data []byte, start int) (uint16, bool) {
	if start+4 > len(data) {
		return 0, false
	}
	var value uint16
	for _, character := range data[start : start+4] {
		value <<= 4
		switch {
		case character >= '0' && character <= '9':
			value += uint16(character - '0')
		case character >= 'a' && character <= 'f':
			value += uint16(character-'a') + 10
		case character >= 'A' && character <= 'F':
			value += uint16(character-'A') + 10
		default:
			return 0, false
		}
	}
	return value, true
}

func pathOrRoot(path string) string {
	if path == "" {
		return "/"
	}
	return path
}

func joinPointer(path, segment string) string {
	escaped := strings.ReplaceAll(strings.ReplaceAll(segment, "~", "~0"), "/", "~1")
	return path + "/" + escaped
}
