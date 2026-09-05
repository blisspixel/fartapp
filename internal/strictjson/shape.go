package strictjson

import (
	"bytes"
	"encoding"
	"encoding/json"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const (
	UnknownMember    Kind = "unknown_member"
	NullValue        Kind = "null_not_allowed"
	ShapeMismatch    Kind = "document_shape_invalid"
	UnsupportedShape Kind = "unsupported_shape"
)

// InspectShape checks exact JSON member names, non-null values, and structural
// and scalar categories against T. Callers MUST first enforce an encoded-byte
// limit and successfully call Inspect with their depth and member-name limits.
// This second pass does not replace syntax, duplicate, Unicode, or limit checks.
//
// Supported shapes are nonrecursive structs, slices (except byte slices), scalar
// strings, booleans and numbers, and pointers to those types. Struct fields must
// be exported, nonembedded, and have explicit valid JSON names; omitempty and
// omitzero are allowed and json:"-" fields are ignored. Duplicate names, maps,
// interfaces, arrays, custom JSON/text decoders, and other shapes are refused,
// even when the corresponding member is absent. json.Number is a numeric leaf.
//
// Missing members, empty objects/arrays, numeric range and integral spelling,
// tagged-variant rules, and model invariants belong to the caller's typed
// decoder and semantic validation. Unknown members and explicit null never
// acquire default values. Object keys are checked in bytewise sorted order;
// array elements in index order. Shape issues have JSON-pointer paths and no
// byte offset. Inspect's behavior and syntax-neutral contract are unchanged.
func InspectShape[T any](data []byte) *Issue {
	shape, issue := compileShape(reflect.TypeFor[T](), make(map[reflect.Type]bool), "")
	if issue != nil {
		return issue
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return &Issue{Kind: ShapeMismatch, Path: "/"}
	}
	return shape.inspect(value, "")
}

type valueShape struct {
	kind   reflect.Kind
	fields map[string]*valueShape
	item   *valueShape
}

func compileShape(expected reflect.Type, visiting map[reflect.Type]bool, path string) (*valueShape, *Issue) {
	failure := func() (*valueShape, *Issue) {
		return nil, &Issue{Kind: UnsupportedShape, Path: pathOrRoot(path)}
	}
	if visiting[expected] || hasDecoderHook(expected) {
		return failure()
	}
	visiting[expected] = true
	defer delete(visiting, expected)
	if expected.Kind() == reflect.Pointer {
		return compileShape(expected.Elem(), visiting, path)
	}
	shape := &valueShape{kind: expected.Kind()}
	if expected == reflect.TypeFor[json.Number]() {
		shape.kind = reflect.Float64
		return shape, nil
	}
	switch expected.Kind() {
	case reflect.Struct:
		shape.fields = make(map[string]*valueShape, expected.NumField())
		for index := range expected.NumField() {
			field := expected.Field(index)
			tag, explicit := field.Tag.Lookup("json")
			if tag == "-" {
				continue
			}
			if !explicit || !field.IsExported() || field.Anonymous {
				return failure()
			}
			parts := strings.Split(tag, ",")
			name := parts[0]
			if !validShapeMemberName(name) {
				return failure()
			}
			for _, option := range parts[1:] {
				if option != "omitempty" && option != "omitzero" {
					return failure()
				}
			}
			if _, duplicate := shape.fields[name]; duplicate {
				return failure()
			}
			child, issue := compileShape(field.Type, visiting, joinPointer(path, name))
			if issue != nil {
				return nil, issue
			}
			shape.fields[name] = child
		}
	case reflect.Slice:
		// encoding/json represents a byte slice as a base64 string, not an array.
		if expected.Elem().Kind() == reflect.Uint8 {
			return failure()
		}
		var issue *Issue
		shape.item, issue = compileShape(expected.Elem(), visiting, path)
		if issue != nil {
			return nil, issue
		}
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
	default:
		return failure()
	}
	return shape, nil
}

func (shape *valueShape) inspect(value any, path string) *Issue {
	if value == nil {
		return &Issue{Kind: NullValue, Path: pathOrRoot(path)}
	}
	valid := false
	switch shape.kind {
	case reflect.Struct:
		members, ok := value.(map[string]any)
		valid = ok
		if ok {
			keys := make([]string, 0, len(members))
			for key := range members {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				next := joinPointer(path, key)
				field, found := shape.fields[key]
				if !found {
					return &Issue{Kind: UnknownMember, Path: next}
				}
				if issue := field.inspect(members[key], next); issue != nil {
					return issue
				}
			}
		}
	case reflect.Slice:
		elements, ok := value.([]any)
		valid = ok
		if ok {
			for index, element := range elements {
				if issue := shape.item.inspect(element, joinPointer(path, strconv.Itoa(index))); issue != nil {
					return issue
				}
			}
		}
	case reflect.String:
		_, valid = value.(string)
	case reflect.Bool:
		_, valid = value.(bool)
	default:
		_, valid = value.(json.Number)
	}
	if !valid {
		return &Issue{Kind: ShapeMismatch, Path: pathOrRoot(path)}
	}
	return nil
}

func hasDecoderHook(value reflect.Type) bool {
	for _, decoder := range []reflect.Type{reflect.TypeFor[json.Unmarshaler](), reflect.TypeFor[encoding.TextUnmarshaler]()} {
		if value.Implements(decoder) || reflect.PointerTo(value).Implements(decoder) {
			return true
		}
	}
	return false
}

func validShapeMemberName(value string) bool {
	if value == "" {
		return false
	}
	for _, char := range value {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && !strings.ContainsRune("!#$%&()*+-./:;<=>?@[]^_{|}~ ", char) {
			return false
		}
	}
	return true
}
