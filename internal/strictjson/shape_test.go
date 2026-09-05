package strictjson

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

type shapeRow struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
}

type shapeDocument struct {
	Name        string      `json:"name"`
	Count       *int        `json:"count,omitempty"`
	Amount      json.Number `json:"amount"`
	Enabled     bool        `json:"enabled"`
	Nested      *shapeRow   `json:"nested,omitzero"`
	Rows        []shapeRow  `json:"rows"`
	Numbers     []float64   `json:"numbers"`
	PointerName string      `json:"a/b~c"`
	Ignored     any         `json:"-"`
}

func TestInspectShapeAcceptsClosedStructuresWithoutInventingPresence(t *testing.T) {
	for _, input := range []string{
		`{}`,
		`{"nested":{},"rows":[],"numbers":[]}`,
		`{"name":"x","count":0,"amount":1e999,"enabled":true,"nested":{"label":"a","value":2},"rows":[{"value":0}],"numbers":[0,1.5],"a/b~c":"x"}`,
		`{"count":1.5}`,
	} {
		checkShapeSyntax(t, []byte(input))
		if issue := InspectShape[shapeDocument]([]byte(input)); issue != nil {
			t.Fatalf("shape(%s) = %#v", input, issue)
		}
	}
	// Shape conformance does not erase the typed decoder's numeric obligations.
	var document shapeDocument
	if err := json.Unmarshal([]byte(`{"count":1.5}`), &document); err == nil {
		t.Fatal("fractional integer unexpectedly decoded")
	}
	if issue := InspectShape[*shapeDocument]([]byte(`{}`)); issue != nil {
		t.Fatalf("pointer shape = %#v", issue)
	}
	if issue := InspectShape[[]shapeRow]([]byte(`[{}]`)); issue != nil {
		t.Fatalf("slice shape = %#v", issue)
	}
	for _, inspect := range []func([]byte) *Issue{InspectShape[string], InspectShape[bool], InspectShape[float64]} {
		if issue := inspect([]byte(`null`)); issue == nil || issue.Kind != NullValue {
			t.Fatal("scalar shape accepted null")
		}
	}
}

func TestInspectShapeReportsExactPathsAndScalarCategories(t *testing.T) {
	for _, test := range []struct {
		input string
		kind  Kind
		path  string
	}{
		{`{"Name":"x"}`, UnknownMember, "/Name"},
		{`{"name":"x","NAME":"y"}`, UnknownMember, "/NAME"},
		{`{"nested":{"other":1}}`, UnknownMember, "/nested/other"},
		{`{"rows":[{"Label":"x"}]}`, UnknownMember, "/rows/0/Label"},
		{`{"a/b~c":null}`, NullValue, "/a~1b~0c"},
		{`{"ignored":1}`, UnknownMember, "/ignored"},
		{`null`, NullValue, "/"},
		{`{"count":null}`, NullValue, "/count"},
		{`{"nested":null}`, NullValue, "/nested"},
		{`{"rows":[null]}`, NullValue, "/rows/0"},
		{`[]`, ShapeMismatch, "/"},
		{`0`, ShapeMismatch, "/"},
		{`{"nested":0}`, ShapeMismatch, "/nested"},
		{`{"rows":{}}`, ShapeMismatch, "/rows"},
		{`{"rows":[0]}`, ShapeMismatch, "/rows/0"},
		{`{"name":{}}`, ShapeMismatch, "/name"},
		{`{"count":{}}`, ShapeMismatch, "/count"},
		{`{"count":"1"}`, ShapeMismatch, "/count"},
		{`{"amount":"1"}`, ShapeMismatch, "/amount"},
		{`{"enabled":0}`, ShapeMismatch, "/enabled"},
		{`{"numbers":[false]}`, ShapeMismatch, "/numbers/0"},
	} {
		checkShapeSyntax(t, []byte(test.input))
		issue := InspectShape[shapeDocument]([]byte(test.input))
		if issue == nil || issue.Kind != test.kind || issue.Path != test.path || issue.ByteOffset != 0 {
			t.Fatalf("shape(%s)=%#v, want %s at %s", test.input, issue, test.kind, test.path)
		}
	}
	for _, input := range []string{`{"z":null,"A":{},"name":1}`, `{"name":1,"A":{},"z":null}`} {
		issue := InspectShape[shapeDocument]([]byte(input))
		if issue == nil || issue.Kind != UnknownMember || issue.Path != "/A" {
			t.Fatalf("object-order-dependent diagnostic: %#v", issue)
		}
	}
}

type recursiveShape struct {
	Next *recursiveShape `json:"next"`
}
type hookedShape struct{}

func (*hookedShape) UnmarshalJSON([]byte) error { panic("shape inspection invoked decoder") }

type textHookedShape string

func (*textHookedShape) UnmarshalText([]byte) error { panic("shape inspection invoked text decoder") }

func TestInspectShapeRejectsUnsupportedSchemasEvenWhenAbsent(t *testing.T) {
	tests := []struct {
		name    string
		inspect func([]byte) *Issue
		path    string
	}{
		{"map", InspectShape[map[string]int], "/"},
		{"interface", InspectShape[any], "/"},
		{"byte slice", InspectShape[[]byte], "/"},
		{"array", InspectShape[[2]int], "/"},
		{"complex", InspectShape[complex128], "/"},
		{"recursive", InspectShape[recursiveShape], "/next"},
		{"JSON decoder", InspectShape[hookedShape], "/"},
		{"text decoder", InspectShape[textHookedShape], "/"},
		{"missing tag", InspectShape[struct{ Field int }], "/"},
		{"empty tag", InspectShape[struct {
			Field int `json:",omitempty"`
		}], "/"},
		{"private field", inspectCompiledShape(reflect.StructOf([]reflect.StructField{
			{Name: "field", PkgPath: "test", Type: reflect.TypeFor[int](), Tag: `json:"field"`},
		})), "/"},
		{"embedded field", InspectShape[struct {
			shapeRow `json:"row"`
		}], "/"},
		{"duplicate names", inspectCompiledShape(reflect.StructOf([]reflect.StructField{
			{Name: "A", Type: reflect.TypeFor[int](), Tag: `json:"x"`},
			{Name: "B", Type: reflect.TypeFor[int](), Tag: `json:"x"`},
		})), "/"},
		{"string option", InspectShape[struct {
			A int `json:"a,string"`
		}], "/"},
		{"invalid tag name", inspectCompiledShape(reflect.StructOf([]reflect.StructField{
			{Name: "A", Type: reflect.TypeFor[int](), Tag: `json:"a\\b"`},
		})), "/"},
		{"absent map field", InspectShape[struct {
			Extra map[string]int `json:"extra,omitempty"`
		}], "/extra"},
		{"slice of maps", InspectShape[struct {
			Rows []map[string]int `json:"rows"`
		}], "/rows"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			issue := test.inspect([]byte(`{}`))
			if issue == nil || issue.Kind != UnsupportedShape || issue.Path != test.path {
				t.Fatalf("unsupported schema = %#v", issue)
			}
		})
	}
}

func TestShapeUsesExistingSyntaxAndNestingGate(t *testing.T) {
	// Syntax inspection is deliberately still schema-neutral, including null.
	if issue := Inspect([]byte(`null`), testLimits); issue != nil {
		t.Fatal("syntax inspector acquired schema policy")
	}
	for _, input := range []string{
		`{"name":"a","name":"b"}`,
		strings.Repeat("[", 34) + strings.Repeat("]", 34),
		`{"name":"\ud800"}`,
	} {
		if issue := Inspect([]byte(input), Limits{MaximumDepth: 32, MaximumMemberNameBytes: 128}); issue == nil {
			t.Fatal("syntax precondition incorrectly passed")
		}
	}
	if issue := InspectShape[shapeDocument]([]byte("{")); issue == nil || issue.Kind != ShapeMismatch {
		t.Fatal("malformed defensive call accepted")
	}
}

func FuzzInspectShape(f *testing.F) {
	f.Add([]byte(`{"name":"x","rows":[{"value":1}]}`))
	f.Add([]byte(`{"nested":{"Label":"x"},"count":null}`))
	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) > 65536 {
			t.Skip()
		}
		if Inspect(input, Limits{MaximumDepth: 32, MaximumMemberNameBytes: 128}) != nil {
			return
		}
		first, second := InspectShape[shapeDocument](input), InspectShape[shapeDocument](input)
		if !reflect.DeepEqual(first, second) {
			t.Fatal("shape issues were nondeterministic")
		}
		if first != nil && first.Path == "" {
			t.Fatal("empty shape issue path")
		}
	})
}

func checkShapeSyntax(t *testing.T, input []byte) {
	t.Helper()
	if issue := Inspect(input, Limits{MaximumDepth: 32, MaximumMemberNameBytes: 128}); issue != nil {
		t.Fatalf("invalid test syntax: %#v", issue)
	}
}

// Construct deliberately invalid reflection schemas dynamically so the test
// corpus does not itself declare invalid JSON tags in compiled Go source.
func inspectCompiledShape(expected reflect.Type) func([]byte) *Issue {
	return func([]byte) *Issue {
		_, issue := compileShape(expected, make(map[reflect.Type]bool), "")
		return issue
	}
}
