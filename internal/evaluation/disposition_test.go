package evaluation

import (
	"encoding/csv"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"
)

const (
	maximumCorpusBytes = 64 * 1024
	maximumCorpusRows  = 256
	maximumFieldBytes  = 128
)

type implementationOutcome string
type authorizationOutcome string

func (implementationOutcome) EvaluationOutcomeMarker() {}

func (authorizationOutcome) EvaluationOutcomeMarker() {}

var errInvalidOutcome = errors.New("invalid profile-owned outcome")

func validateImplementationOutcome(value implementationOutcome) error {
	switch value {
	case "available", "unavailable":
		return nil
	default:
		return errInvalidOutcome
	}
}

func validateAuthorizationOutcome(value authorizationOutcome) error {
	switch value {
	case "permitted", "refused":
		return nil
	default:
		return errInvalidOutcome
	}
}

func TestDispositionCorpus(t *testing.T) {
	rows := readDispositionCorpus(t)
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		caseID := row[0]
		validateCorpusInstruction(t, row)
		if _, exists := seen[caseID]; exists {
			t.Fatalf("case id %q is duplicated", caseID)
		}
		seen[caseID] = struct{}{}
		t.Run(caseID, func(t *testing.T) {
			if got, want := runCorpusInstruction(row), row[4]; got != want {
				t.Fatalf("result = %q, want %q", got, want)
			}
		})
	}
}

func TestEvaluatedRequiresProfileValidator(t *testing.T) {
	value, err := Evaluated(implementationOutcome("available"), nil)
	if !errors.Is(err, ErrOutcomeValidatorRequired) || value.Valid() {
		t.Fatalf("Evaluated without validator = (%#v, %v)", value, err)
	}
}

func TestEvaluatedCallsProfileValidatorExactlyOnce(t *testing.T) {
	calls := 0
	value, err := Evaluated(implementationOutcome("available"), func(outcome implementationOutcome) error {
		calls++
		return validateImplementationOutcome(outcome)
	})
	if err != nil || !value.Valid() || calls != 1 {
		t.Fatalf("Evaluated = (%#v, %v), calls = %d", value, err, calls)
	}
	outcome, ok := value.Outcome()
	if !ok || outcome != "available" || value.Kind() != KindEvaluated {
		t.Fatalf("outcome = (%q, %t), kind = %d", outcome, ok, value.Kind())
	}
}

func TestEvaluatedRejectsInvalidProfileOutcome(t *testing.T) {
	value, err := Evaluated(implementationOutcome("unknown"), validateImplementationOutcome)
	if !errors.Is(err, errInvalidOutcome) || value.Valid() {
		t.Fatalf("Evaluated invalid outcome = (%#v, %v)", value, err)
	}
}

func TestNotEvaluatedHasNoOutcome(t *testing.T) {
	value := NotEvaluated[implementationOutcome]()
	outcome, ok := value.Outcome()
	if !value.Valid() || value.Kind() != KindNotEvaluated || ok || outcome != "" {
		t.Fatalf("not evaluated = (%#v, %q, %t)", value, outcome, ok)
	}
}

func TestZeroAndForgedDispositionsAreInvalid(t *testing.T) {
	values := []Disposition[implementationOutcome]{
		{},
		{kind: Kind(255)},
		{kind: KindNotEvaluated, outcome: "available"},
		{kind: KindNotEvaluated, outcomeValidated: true},
		{kind: KindEvaluated, outcome: "available"},
	}
	for index, value := range values {
		if value.Valid() {
			t.Errorf("forged value %d is valid: %#v", index, value)
		}
		if outcome, ok := value.Outcome(); ok || outcome != "" {
			t.Errorf("forged value %d exposed outcome (%q, %t)", index, outcome, ok)
		}
	}
}

func TestOutcomeTypesRemainProfileOwned(t *testing.T) {
	implementation, implementationErr := Evaluated(
		implementationOutcome("available"),
		validateImplementationOutcome,
	)
	authorization, authorizationErr := Evaluated(
		authorizationOutcome("permitted"),
		validateAuthorizationOutcome,
	)
	if implementationErr != nil || authorizationErr != nil ||
		!implementation.Valid() || !authorization.Valid() {
		t.Fatalf(
			"profile dispositions = (%#v, %v), (%#v, %v)",
			implementation,
			implementationErr,
			authorization,
			authorizationErr,
		)
	}
}

func readDispositionCorpus(t *testing.T) [][]string {
	t.Helper()
	path := filepath.Join(
		"..", "..", "testdata", "conformance",
		"evaluation-disposition-v0alpha1", "cases.tsv",
	)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read corpus: %v", err)
	}
	if len(data) > maximumCorpusBytes {
		t.Fatalf("corpus has %d bytes, limit is %d", len(data), maximumCorpusBytes)
	}
	if !utf8.Valid(data) {
		t.Fatal("corpus is not valid UTF-8")
	}
	for _, character := range string(data) {
		if (character < 0x20 && character != '\t' && character != '\r' && character != '\n') ||
			(character >= 0x7f && character <= 0x9f) {
			t.Fatalf("corpus contains control character U+%04X", character)
		}
	}
	reader := csv.NewReader(strings.NewReader(string(data)))
	reader.Comma = '\t'
	reader.FieldsPerRecord = 5
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("parse corpus: %v", err)
	}
	if len(records) < 2 || len(records)-1 > maximumCorpusRows {
		t.Fatalf("corpus row count = %d", len(records)-1)
	}
	wantHeader := []string{"case_id", "action", "owner", "outcome", "expect"}
	if !slices.Equal(records[0], wantHeader) {
		t.Fatalf("corpus header = %q, want %q", records[0], wantHeader)
	}
	for rowIndex, record := range records {
		for columnIndex, field := range record {
			if len(field) > maximumFieldBytes {
				t.Fatalf("field %d:%d has %d bytes", rowIndex+1, columnIndex+1, len(field))
			}
		}
	}
	return records[1:]
}

func validateCorpusInstruction(t *testing.T, row []string) {
	t.Helper()
	if row[0] == "" {
		t.Fatal("corpus case id is empty")
	}
	allowedActions := []string{
		"evaluated",
		"evaluated_no_validator",
		"forged",
		"forged_kind",
		"not_evaluated",
		"zero",
	}
	if !slices.Contains(allowedActions, row[1]) {
		t.Fatalf("corpus action %q is not supported", row[1])
	}
	if row[2] != "implementation" && row[2] != "authorization" {
		t.Fatalf("corpus owner %q is not supported", row[2])
	}
	allowedExpectations := []string{
		"invalid_disposition",
		"invalid_outcome",
		"missing_validator",
		"valid",
	}
	if !slices.Contains(allowedExpectations, row[4]) {
		t.Fatalf("corpus expectation %q is not supported", row[4])
	}
}

func runCorpusInstruction(row []string) string {
	action, owner, outcome := row[1], row[2], row[3]
	switch action {
	case "evaluated":
		switch owner {
		case "implementation":
			value, err := Evaluated(
				implementationOutcome(outcome),
				validateImplementationOutcome,
			)
			if err != nil {
				return "invalid_outcome"
			}
			if value.Valid() {
				return "valid"
			}
		case "authorization":
			value, err := Evaluated(
				authorizationOutcome(outcome),
				validateAuthorizationOutcome,
			)
			if err != nil {
				return "invalid_outcome"
			}
			if value.Valid() {
				return "valid"
			}
		}
		return "invalid_disposition"
	case "evaluated_no_validator":
		value, err := Evaluated(implementationOutcome(outcome), nil)
		if errors.Is(err, ErrOutcomeValidatorRequired) && !value.Valid() {
			return "missing_validator"
		}
		return "invalid_disposition"
	case "not_evaluated":
		if outcome != "-" {
			return "invalid_disposition"
		}
		value := NotEvaluated[implementationOutcome]()
		if value.Valid() {
			return "valid"
		}
		return "invalid_disposition"
	case "zero":
		var value Disposition[implementationOutcome]
		if !value.Valid() {
			return "invalid_disposition"
		}
	case "forged":
		value := Disposition[implementationOutcome]{
			kind:    KindNotEvaluated,
			outcome: implementationOutcome(outcome),
		}
		if !value.Valid() {
			return "invalid_disposition"
		}
	case "forged_kind":
		value := Disposition[implementationOutcome]{kind: Kind(255)}
		if !value.Valid() {
			return "invalid_disposition"
		}
	}
	return "unexpected_harness_result"
}

func FuzzDispositionConstructors(f *testing.F) {
	for _, seed := range []string{"", "available", "unavailable", "unknown", "undetermined"} {
		f.Add(seed, uint8(0), false)
	}
	f.Fuzz(func(t *testing.T, outcome string, rawKind uint8, validated bool) {
		value, err := Evaluated(implementationOutcome(outcome), validateImplementationOutcome)
		accepted := outcome == "available" || outcome == "unavailable"
		if accepted != (err == nil && value.Valid()) {
			t.Fatalf("Evaluated(%q) = (%#v, %v)", outcome, value, err)
		}

		notEvaluated := NotEvaluated[implementationOutcome]()
		if exposed, ok := notEvaluated.Outcome(); !notEvaluated.Valid() || ok || exposed != "" {
			t.Fatalf("NotEvaluated exposed (%q, %t)", exposed, ok)
		}

		forged := Disposition[implementationOutcome]{
			kind:             Kind(rawKind),
			outcome:          implementationOutcome(outcome),
			outcomeValidated: validated,
		}
		if forged.Valid() {
			switch forged.kind {
			case KindEvaluated:
				if !forged.outcomeValidated {
					t.Fatal("evaluated forged value lacks validation")
				}
			case KindNotEvaluated:
				if forged.outcomeValidated || forged.outcome != "" {
					t.Fatal("not-evaluated forged value contains an outcome")
				}
			default:
				t.Fatal("unknown kind is valid")
			}
		}
	})
}
