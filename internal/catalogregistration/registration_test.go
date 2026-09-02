package catalogregistration

import (
	"encoding/csv"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/blisspixel/fartapp/internal/evaluation"
)

const (
	maximumCorpusBytes = 64 * 1024
	maximumCorpusRows  = 256
	maximumFieldBytes  = 128
)

func TestRegistrationCorpus(t *testing.T) {
	rows := readRegistrationCorpus(t)
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		validateCorpusRow(t, row)
		if _, exists := seen[row[0]]; exists {
			t.Fatalf("case id %q is duplicated", row[0])
		}
		seen[row[0]] = struct{}{}
		t.Run(row[0], func(t *testing.T) {
			if got, want := runCorpusRow(row), row[9]; got != want {
				t.Fatalf("result = %q, want %q", got, want)
			}
		})
	}
}

func TestExactBindingAccessorsAndDistinctTypes(t *testing.T) {
	binding := validBinding(t)
	if binding.CatalogScope().ID() != "catalog.scope" ||
		binding.CatalogScope().Revision() != "r1" ||
		binding.DeclarationAuthority().ID() != "authority.a" ||
		binding.DeclarationAuthority().Revision() != "r1" ||
		binding.Subject().ID() != "subject.a" ||
		binding.Subject().Revision() != "r1" ||
		binding.Capability().Value() != "capability.a" {
		t.Fatalf("binding accessors = %#v", binding)
	}
}

func TestRegistrationResultStates(t *testing.T) {
	binding := validBinding(t)
	registered, err := EvaluatedRegistered(binding)
	if err != nil || !registered.Valid() ||
		registered.EvaluationKind() != evaluation.KindEvaluated ||
		registered.Binding() != binding {
		t.Fatalf("registered result = (%#v, %v)", registered, err)
	}
	if presence, ok := registered.Presence(); !ok || presence != PresenceRegistered {
		t.Fatalf("registered presence = (%d, %t)", presence, ok)
	}
	registration, ok := registered.Registration()
	if !ok || !registration.Valid() || registration.Binding() != binding {
		t.Fatalf("registration = (%#v, %t)", registration, ok)
	}

	notRegistered, err := EvaluatedNotRegistered(binding)
	if err != nil || !notRegistered.Valid() {
		t.Fatalf("not-registered result = (%#v, %v)", notRegistered, err)
	}
	if presence, present := notRegistered.Presence(); !present || presence != PresenceNotRegistered {
		t.Fatalf("not-registered presence = (%d, %t)", presence, present)
	}
	if registration, present := notRegistered.Registration(); present || registration.Valid() {
		t.Fatalf("not-registered registration = (%#v, %t)", registration, present)
	}

	notEvaluated, err := NotEvaluated(binding)
	if err != nil || !notEvaluated.Valid() ||
		notEvaluated.EvaluationKind() != evaluation.KindNotEvaluated {
		t.Fatalf("not-evaluated result = (%#v, %v)", notEvaluated, err)
	}
	if presence, present := notEvaluated.Presence(); present || presence != PresenceInvalid {
		t.Fatalf("not-evaluated presence = (%d, %t)", presence, present)
	}
	if registration, present := notEvaluated.Registration(); present || registration.Valid() {
		t.Fatalf("not-evaluated registration = (%#v, %t)", registration, present)
	}
}

func TestPresenceValidatorRejectsForeignOutcome(t *testing.T) {
	if err := validatePresence(presenceOutcome("unknown")); !errors.Is(err, errInvalidPresence) {
		t.Fatalf("validatePresence foreign outcome error = %v", err)
	}
	if value, err := evaluated(validBinding(t), presenceOutcome("unknown")); !errors.Is(err, errInvalidPresence) || value.Valid() {
		t.Fatalf("evaluated foreign outcome = (%#v, %v)", value, err)
	}

	foreign, err := evaluation.Evaluated(presenceOutcome("unknown"), func(presenceOutcome) error {
		return nil
	})
	if err != nil {
		t.Fatalf("construct foreign disposition: %v", err)
	}
	if presence, ok := presenceFromDisposition(foreign); ok || presence != PresenceInvalid {
		t.Fatalf("foreign disposition presence = (%d, %t)", presence, ok)
	}
	if presence, ok := presenceFromDisposition(evaluation.NotEvaluated[presenceOutcome]()); ok || presence != PresenceInvalid {
		t.Fatalf("non-evaluated disposition presence = (%d, %t)", presence, ok)
	}
	presenceRegistered.EvaluationOutcomeMarker()
}

func TestZeroAndInvalidValuesRemainInvalid(t *testing.T) {
	if (CatalogScopeRef{}).Valid() || (DeclarationAuthorityRef{}).Valid() ||
		(SubjectRef{}).Valid() || (CapabilityID{}).Valid() ||
		(Binding{}).Valid() || (Result{}).Valid() || (Registration{}).Valid() {
		t.Fatal("a zero value is valid")
	}
	if _, err := Bind(CatalogScopeRef{}, DeclarationAuthorityRef{}, SubjectRef{}, CapabilityID{}); !errors.Is(err, ErrInvalidBinding) {
		t.Fatalf("Bind zero values error = %v", err)
	}
	if _, err := EvaluatedRegistered(Binding{}); !errors.Is(err, ErrInvalidBinding) {
		t.Fatalf("EvaluatedRegistered zero binding error = %v", err)
	}
	if _, err := EvaluatedNotRegistered(Binding{}); !errors.Is(err, ErrInvalidBinding) {
		t.Fatalf("EvaluatedNotRegistered zero binding error = %v", err)
	}
	if _, err := NotEvaluated(Binding{}); !errors.Is(err, ErrInvalidBinding) {
		t.Fatalf("NotEvaluated zero binding error = %v", err)
	}

	forged := Result{binding: validBinding(t)}
	if forged.Valid() {
		t.Fatal("result with zero disposition is valid")
	}
	if presence, ok := forged.Presence(); ok || presence != PresenceInvalid {
		t.Fatalf("forged presence = (%d, %t)", presence, ok)
	}
}

func TestTokenCandidate(t *testing.T) {
	valid := []string{"a", "0", "catalog.inspect", "not_registered", "lab-application", strings.Repeat("a", 128)}
	for _, value := range valid {
		if err := validateToken(value); err != nil {
			t.Errorf("validateToken(%q) = %v", value, err)
		}
	}
	invalid := []string{
		"", ".", "-a", "a-", "a..b", "a_-b", "UPPER", "a:b", "é",
		strings.Repeat("a", 129),
	}
	for _, value := range invalid {
		if err := validateToken(value); !errors.Is(err, ErrInvalidToken) {
			t.Errorf("validateToken(%q) = %v", value, err)
		}
	}
}

func TestReferenceConstructorsRejectEachInvalidComponent(t *testing.T) {
	constructors := []struct {
		name string
		call func(string, string) error
	}{
		{name: "catalog scope", call: func(id, revision string) error {
			_, err := NewCatalogScopeRef(id, revision)
			return err
		}},
		{name: "declaration authority", call: func(id, revision string) error {
			_, err := NewDeclarationAuthorityRef(id, revision)
			return err
		}},
		{name: "subject", call: func(id, revision string) error {
			_, err := NewSubjectRef(id, revision)
			return err
		}},
	}
	for _, constructor := range constructors {
		t.Run(constructor.name, func(t *testing.T) {
			if err := constructor.call("", "r1"); !errors.Is(err, ErrInvalidToken) {
				t.Errorf("empty id error = %v", err)
			}
			if err := constructor.call("valid", ""); !errors.Is(err, ErrInvalidToken) {
				t.Errorf("empty revision error = %v", err)
			}
		})
	}
	if _, err := NewCapabilityID(""); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("empty capability error = %v", err)
	}
}

func validBinding(t *testing.T) Binding {
	t.Helper()
	binding, result := buildBinding(
		"catalog.scope", "r1", "authority.a", "r1", "subject.a", "r1", "capability.a",
	)
	if result != "valid_binding" {
		t.Fatalf("build valid binding = %s", result)
	}
	return binding
}

func buildBinding(
	catalogID, catalogRevision, authorityID, authorityRevision,
	subjectID, subjectRevision, capabilityValue string,
) (Binding, string) {
	catalog, err := NewCatalogScopeRef(catalogID, catalogRevision)
	if err != nil {
		return Binding{}, "invalid_token"
	}
	authority, err := NewDeclarationAuthorityRef(authorityID, authorityRevision)
	if err != nil {
		return Binding{}, "invalid_token"
	}
	subject, err := NewSubjectRef(subjectID, subjectRevision)
	if err != nil {
		return Binding{}, "invalid_token"
	}
	capability, err := NewCapabilityID(capabilityValue)
	if err != nil {
		return Binding{}, "invalid_token"
	}
	binding, err := Bind(catalog, authority, subject, capability)
	if err != nil {
		return Binding{}, "invalid_binding"
	}
	return binding, "valid_binding"
}

func runCorpusRow(row []string) string {
	binding, result := buildBinding(row[2], row[3], row[4], row[5], row[6], row[7], row[8])
	if result != "valid_binding" {
		return result
	}
	switch row[1] {
	case "registered":
		value, err := EvaluatedRegistered(binding)
		if err == nil && value.Valid() {
			if registration, ok := value.Registration(); ok && registration.Valid() {
				return "valid_registered"
			}
		}
	case "not_registered":
		value, err := EvaluatedNotRegistered(binding)
		if err == nil && value.Valid() {
			if registration, ok := value.Registration(); !ok && !registration.Valid() {
				return "valid_not_registered"
			}
		}
	case "not_evaluated":
		value, err := NotEvaluated(binding)
		if err == nil && value.Valid() {
			if _, ok := value.Presence(); !ok {
				return "valid_not_evaluated"
			}
		}
	case "authority_revision_distinct":
		other, otherResult := buildBinding(
			row[2], row[3], row[4], "different", row[6], row[7], row[8],
		)
		if otherResult == "valid_binding" && binding != other {
			return "distinct_binding"
		}
	case "authority_id_distinct":
		other, otherResult := buildBinding(
			row[2], row[3], "different", row[5], row[6], row[7], row[8],
		)
		if otherResult == "valid_binding" && binding != other {
			return "distinct_binding"
		}
	case "catalog_id_distinct":
		other, otherResult := buildBinding(
			"different", row[3], row[4], row[5], row[6], row[7], row[8],
		)
		if otherResult == "valid_binding" && binding != other {
			return "distinct_binding"
		}
	case "catalog_revision_distinct":
		other, otherResult := buildBinding(
			row[2], "different", row[4], row[5], row[6], row[7], row[8],
		)
		if otherResult == "valid_binding" && binding != other {
			return "distinct_binding"
		}
	case "capability_distinct":
		other, otherResult := buildBinding(
			row[2], row[3], row[4], row[5], row[6], row[7], "different",
		)
		if otherResult == "valid_binding" && binding != other {
			return "distinct_binding"
		}
	case "subject_revision_distinct":
		other, otherResult := buildBinding(
			row[2], row[3], row[4], row[5], row[6], "different", row[8],
		)
		if otherResult == "valid_binding" && binding != other {
			return "distinct_binding"
		}
	case "subject_id_distinct":
		other, otherResult := buildBinding(
			row[2], row[3], row[4], row[5], "different", row[7], row[8],
		)
		if otherResult == "valid_binding" && binding != other {
			return "distinct_binding"
		}
	case "binding":
		return "valid_binding"
	}
	return "unexpected_harness_result"
}

func readRegistrationCorpus(t *testing.T) [][]string {
	t.Helper()
	path := filepath.Join(
		"..", "..", "testdata", "conformance",
		"catalog-registration-v0alpha1", "cases.tsv",
	)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read corpus: %v", err)
	}
	if len(data) > maximumCorpusBytes || !utf8.Valid(data) {
		t.Fatalf("corpus must be valid UTF-8 within %d bytes", maximumCorpusBytes)
	}
	for _, character := range string(data) {
		if (character < 0x20 && character != '\t' && character != '\r' && character != '\n') ||
			(character >= 0x7f && character <= 0x9f) {
			t.Fatalf("corpus contains control character U+%04X", character)
		}
	}
	reader := csv.NewReader(strings.NewReader(string(data)))
	reader.Comma = '\t'
	reader.FieldsPerRecord = 10
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("parse corpus: %v", err)
	}
	if len(records) < 2 || len(records)-1 > maximumCorpusRows {
		t.Fatalf("corpus row count = %d", len(records)-1)
	}
	wantHeader := []string{
		"case_id", "action", "catalog_id", "catalog_revision", "authority_id",
		"authority_revision", "subject_id", "subject_revision", "capability", "expect",
	}
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

func validateCorpusRow(t *testing.T, row []string) {
	t.Helper()
	if row[0] == "" {
		t.Fatal("corpus case id is empty")
	}
	allowedActions := []string{
		"authority_id_distinct", "authority_revision_distinct", "binding",
		"capability_distinct", "catalog_id_distinct", "catalog_revision_distinct",
		"not_evaluated", "not_registered", "registered", "subject_id_distinct",
		"subject_revision_distinct",
	}
	if !slices.Contains(allowedActions, row[1]) {
		t.Fatalf("corpus action %q is not supported", row[1])
	}
	allowedExpectations := []string{
		"distinct_binding", "invalid_binding", "invalid_token", "valid_binding",
		"valid_not_evaluated", "valid_not_registered", "valid_registered",
	}
	if !slices.Contains(allowedExpectations, row[9]) {
		t.Fatalf("corpus expectation %q is not supported", row[9])
	}
}

func FuzzRegistrationConstructors(f *testing.F) {
	for _, seed := range []string{"", "a", "catalog.inspect", ".", "A", "a..b", "é"} {
		f.Add(seed, seed, seed)
	}
	f.Fuzz(func(t *testing.T, id, revision, capabilityValue string) {
		catalog, catalogErr := NewCatalogScopeRef(id, revision)
		authority, authorityErr := NewDeclarationAuthorityRef(id, revision)
		subject, subjectErr := NewSubjectRef(id, revision)
		capability, capabilityErr := NewCapabilityID(capabilityValue)
		if (catalogErr == nil) != catalog.Valid() ||
			(authorityErr == nil) != authority.Valid() ||
			(subjectErr == nil) != subject.Valid() ||
			(capabilityErr == nil) != capability.Valid() {
			t.Fatalf("constructor validity mismatch for (%q, %q, %q)", id, revision, capabilityValue)
		}
		binding, err := Bind(catalog, authority, subject, capability)
		allValid := catalog.Valid() && authority.Valid() && subject.Valid() && capability.Valid()
		if allValid != (err == nil && binding.Valid()) {
			t.Fatalf("Bind = (%#v, %v), all valid = %t", binding, err, allValid)
		}
		if !allValid {
			return
		}
		registered, registeredErr := EvaluatedRegistered(binding)
		notRegistered, notRegisteredErr := EvaluatedNotRegistered(binding)
		notEvaluated, notEvaluatedErr := NotEvaluated(binding)
		if registeredErr != nil || notRegisteredErr != nil || notEvaluatedErr != nil ||
			!registered.Valid() || !notRegistered.Valid() || !notEvaluated.Valid() {
			t.Fatal("a valid binding did not construct all result states")
		}
	})
}
