package strictjson

import (
	"reflect"
	"strings"
	"testing"
)

var testLimits = Limits{MaximumDepth: 4, MaximumMemberNameBytes: 8}

func TestInspectAcceptsBoundedJSON(t *testing.T) {
	for _, input := range []string{
		`null`, `true`, `123`, `"text"`, `[]`, `{}`,
		" {\"a\":[1,{\"b\":\"\\ud83d\\ude00\"}]} \r\n",
	} {
		if issue := Inspect([]byte(input), testLimits); issue != nil {
			t.Errorf("Inspect(%q) = %#v", input, issue)
		}
	}
}

func TestInspectClassifiesFailures(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		kind  Kind
		path  string
	}{
		{name: "empty", input: []byte(" \r\n"), kind: EmptyInput, path: "/"},
		{name: "malformed", input: []byte("{"), kind: Malformed, path: "/"},
		{name: "invalid UTF-8", input: []byte{'{', '"', 0xff, '"', ':', '0', '}'}, kind: Malformed, path: "/"},
		{name: "lone high surrogate", input: []byte(`{"a":"\ud800"}`), kind: Malformed, path: "/"},
		{name: "lone low surrogate", input: []byte(`{"a":"\udc00"}`), kind: Malformed, path: "/"},
		{name: "trailing", input: []byte(`{} []`), kind: TrailingValue, path: "/"},
		{name: "duplicate", input: []byte(`{"a":1,"a":2}`), kind: DuplicateMember, path: "/a"},
		{name: "escaped pointer", input: []byte(`{"a/b":{"~c":1,"~c":2}}`), kind: DuplicateMember, path: "/a~1b/~0c"},
		{name: "long member", input: []byte(`{"123456789":0}`), kind: MemberNameTooLong, path: "/"},
		{name: "deep", input: []byte(`[[[[[[0]]]]]]`), kind: MaximumDepth, path: "/0/0/0/0/0"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			issue := Inspect(test.input, testLimits)
			if issue == nil || issue.Kind != test.kind || issue.Path != test.path {
				t.Fatalf("Inspect = %#v, want (%s, %s)", issue, test.kind, test.path)
			}
		})
	}
}

func TestLimitsAreCallerOwned(t *testing.T) {
	if issue := Inspect([]byte(`{"":{}}`), Limits{}); issue == nil || issue.Kind != MaximumDepth {
		t.Fatalf("zero depth limit = %#v", issue)
	}
	if issue := Inspect([]byte(`{"":1}`), Limits{MaximumDepth: 1}); issue != nil {
		t.Fatalf("empty member within zero-byte limit = %#v", issue)
	}
	long := strings.Repeat("x", 9)
	if issue := Inspect([]byte(`{"`+long+`":1}`), testLimits); issue == nil || issue.ByteOffset == 0 {
		t.Fatalf("member limit issue = %#v", issue)
	}
}

func FuzzInspect(f *testing.F) {
	f.Add([]byte(`{"a":[1,{"b":"\ud83d\ude00"}]}`))
	f.Add([]byte(`{"a":1,"a":2}`))
	f.Add([]byte{0xff})
	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) > 131_072 {
			t.Skip()
		}
		first := Inspect(input, Limits{MaximumDepth: 32, MaximumMemberNameBytes: 128})
		second := Inspect(input, Limits{MaximumDepth: 32, MaximumMemberNameBytes: 128})
		if !reflect.DeepEqual(first, second) {
			t.Fatalf("inspection is not deterministic: %#v != %#v", first, second)
		}
		if first != nil && first.Path == "" {
			t.Fatalf("issue has an empty path: %#v", first)
		}
	})
}
