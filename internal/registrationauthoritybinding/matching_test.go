package registrationauthoritybinding

import (
	"encoding/csv"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/blisspixel/fartapp/internal/authoritymatching"
	"github.com/blisspixel/fartapp/internal/authorityresolution"
	"github.com/blisspixel/fartapp/internal/catalogregistration"
)

const (
	maximumCorpusBytes = 64 * 1024
	maximumCorpusRows  = 256
	maximumFieldBytes  = 128
)

func TestRegistrationAuthorityBindingCorpus(t *testing.T) {
	snapshotRows := readCorpusTSV(t, "snapshots.tsv", []string{
		"snapshot_id", "catalog_id", "catalog_revision", "expect",
	})
	recordRows := readCorpusTSV(t, "authority-records.tsv", []string{
		"snapshot_id", "catalog_id", "catalog_revision", "authority_id", "authority_revision",
	})
	registrationRows := readCorpusTSV(t, "registrations.tsv", []string{
		"registration_id", "catalog_id", "catalog_revision", "authority_id", "authority_revision",
		"subject_id", "subject_revision", "capability", "expect",
	})
	caseRows := readCorpusTSV(t, "cases.tsv", []string{
		"case_id", "snapshot_id", "resolved_authority_id", "resolved_authority_revision",
		"registration_id", "expect", "positive_witness",
	})

	type snapshotFixture struct {
		scope  catalogregistration.CatalogScopeRef
		expect string
	}
	fixtures := make(map[string]snapshotFixture, len(snapshotRows))
	for _, row := range snapshotRows {
		if row[0] == "" {
			t.Fatal("snapshot id is empty")
		}
		if _, exists := fixtures[row[0]]; exists {
			t.Fatalf("snapshot id %q is duplicated", row[0])
		}
		if !slices.Contains([]string{
			"invalid_record", "record_scope_mismatch", "valid_snapshot",
		}, row[3]) {
			t.Fatalf("snapshot expectation %q is unsupported", row[3])
		}
		scope, err := catalogregistration.NewCatalogScopeRef(row[1], row[2])
		if err != nil {
			t.Fatalf("snapshot %q scope: %v", row[0], err)
		}
		fixtures[row[0]] = snapshotFixture{scope: scope, expect: row[3]}
	}

	records := make(map[string][]authoritymatching.DeclarationAuthorityRecord)
	recordInvalid := make(map[string]bool)
	for _, row := range recordRows {
		if _, exists := fixtures[row[0]]; !exists {
			t.Fatalf("record references unknown snapshot %q", row[0])
		}
		record, err := constructAuthorityRecord(row[1], row[2], row[3], row[4])
		if err != nil {
			recordInvalid[row[0]] = true
			continue
		}
		records[row[0]] = append(records[row[0]], record)
	}

	validSnapshots := make(map[string]authoritymatching.FiniteAuthoritySnapshot)
	for id, fixture := range fixtures {
		actual := "valid_snapshot"
		var snapshot authoritymatching.FiniteAuthoritySnapshot
		if recordInvalid[id] {
			actual = "invalid_record"
		} else {
			var err error
			snapshot, err = authoritymatching.NewFiniteAuthoritySnapshot(fixture.scope, records[id])
			switch {
			case errors.Is(err, authoritymatching.ErrRecordScopeMismatch):
				actual = "record_scope_mismatch"
			case err != nil:
				actual = "unexpected_snapshot_error"
			}
		}
		if actual != fixture.expect {
			t.Fatalf("snapshot %q result = %q, want %q", id, actual, fixture.expect)
		}
		if actual == "valid_snapshot" {
			validSnapshots[id] = snapshot
		}
	}

	registrations := make(map[string]catalogregistration.Registration)
	seenRegistrations := make(map[string]struct{}, len(registrationRows))
	for _, row := range registrationRows {
		if row[0] == "" {
			t.Fatal("registration id is empty")
		}
		if _, exists := seenRegistrations[row[0]]; exists {
			t.Fatalf("registration id %q is duplicated", row[0])
		}
		seenRegistrations[row[0]] = struct{}{}
		if !slices.Contains([]string{"invalid_registration", "valid_registration"}, row[8]) {
			t.Fatalf("registration expectation %q is unsupported", row[8])
		}
		registration, err := constructRegistration(row[1], row[2], row[3], row[4], row[5], row[6], row[7])
		actual := "valid_registration"
		if err != nil {
			actual = "invalid_registration"
		}
		if actual != row[8] {
			t.Fatalf("registration %q result = %q, want %q", row[0], actual, row[8])
		}
		if actual == "valid_registration" {
			registrations[row[0]] = registration
		}
	}

	rule := ExactAuthorityBindingV0Alpha1()
	seenCases := make(map[string]struct{}, len(caseRows))
	for _, row := range caseRows {
		if row[0] == "" {
			t.Fatal("case id is empty")
		}
		if _, exists := seenCases[row[0]]; exists {
			t.Fatalf("case id %q is duplicated", row[0])
		}
		seenCases[row[0]] = struct{}{}
		if !slices.Contains([]string{
			"authority_not_resolved", "catalog_scope_mismatch", "exact_authority_binding_match",
			"no_exact_authority_binding_match",
		}, row[5]) {
			t.Fatalf("case expectation %q is unsupported", row[5])
		}
		wantPositive, applicable := parsePositiveExpectation(t, row[0], row[5], row[6])
		snapshot, exists := validSnapshots[row[1]]
		if !exists {
			t.Fatalf("case %q references non-valid snapshot %q", row[0], row[1])
		}
		t.Run(row[0], func(t *testing.T) {
			authority, resolved := resolveAuthority(t, snapshot, row[2], row[3])
			if row[5] == "authority_not_resolved" {
				if resolved || row[4] != "not_applicable" {
					t.Fatalf("upstream non-resolution = (%#v, %t), registration %q", authority, resolved, row[4])
				}
				return
			}
			if !resolved {
				t.Fatal("expected positive upstream resolution")
			}
			registration, exists := registrations[row[4]]
			if !exists {
				t.Fatalf("unknown valid registration %q", row[4])
			}
			decision, err := Compare(rule, authority, registration)
			if row[5] == "catalog_scope_mismatch" {
				if !errors.Is(err, ErrCatalogScopeMismatch) || decision.Valid() || applicable {
					t.Fatalf("scope mismatch = (%#v, %v), applicable %t", decision, err, applicable)
				}
				return
			}
			if err != nil || !decision.Valid() || decision.Rule() != rule ||
				decision.ResolvedAuthority() != authority || decision.Registration() != registration {
				t.Fatalf("decision = (%#v, %v)", decision, err)
			}
			outcome, ok := decision.Outcome()
			if !ok || outcomeToken(outcome) != row[5] {
				t.Fatalf("outcome = (%d, %t), want %q", outcome, ok, row[5])
			}
			witness, positive := decision.ExactBindingWitness()
			if positive != wantPositive || witness.Valid() != wantPositive {
				t.Fatalf("witness = (%#v, %t), want %t", witness, positive, wantPositive)
			}
			if positive && (witness.Decision() != decision || witness.Rule() != rule ||
				witness.ResolvedAuthority() != authority || witness.Registration() != registration) {
				t.Fatalf("positive witness did not retain inputs: %#v", witness)
			}
		})
	}
}

func TestRuleInvalidAndForgedValues(t *testing.T) {
	rule := ExactAuthorityBindingV0Alpha1()
	if !rule.Valid() || rule.ID() != exactAuthorityBindingRuleID ||
		rule.Revision() != exactAuthorityBindingRuleRevision {
		t.Fatalf("exact rule = %#v", rule)
	}
	for index, forged := range []Rule{
		{},
		{id: rule.id, revision: "v0alpha2", marker: rule.marker},
		{id: "lab.other", revision: rule.revision, marker: rule.marker},
		{id: rule.id, revision: rule.revision, marker: 2},
	} {
		if forged.Valid() {
			t.Fatalf("forged rule %d is valid: %#v", index, forged)
		}
	}

	authority := mustResolvedAuthority(t, "catalog.scope", "r1", "authority.a", "r1")
	registration := mustRegistration(t, "catalog.scope", "r1", "authority.a", "r1", "subject.a", "r1", "capability.a")
	if value, err := Compare(Rule{}, authority, registration); !errors.Is(err, ErrInvalidRule) || value.Valid() {
		t.Fatalf("zero-rule comparison = (%#v, %v)", value, err)
	}
	if value, err := Compare(rule, authorityresolution.SnapshotResolvedAuthority{}, registration); !errors.Is(err, ErrInvalidResolvedAuthority) || value.Valid() {
		t.Fatalf("zero-authority comparison = (%#v, %v)", value, err)
	}
	if value, err := Compare(rule, authority, catalogregistration.Registration{}); !errors.Is(err, ErrInvalidRegistration) || value.Valid() {
		t.Fatalf("zero-registration comparison = (%#v, %v)", value, err)
	}

	var zero Decision
	if zero.Valid() || zero.Rule().Valid() || zero.ResolvedAuthority().Valid() || zero.Registration().Valid() {
		t.Fatalf("zero decision is observable as valid: %#v", zero)
	}
	if outcome, ok := zero.Outcome(); ok || outcome != OutcomeInvalid {
		t.Fatalf("zero outcome = (%d, %t)", outcome, ok)
	}
	if witness, ok := zero.ExactBindingWitness(); ok || witness.Valid() {
		t.Fatalf("zero witness = (%#v, %t)", witness, ok)
	}

	var zeroWitness ExactBindingWitness
	if zeroWitness.Valid() || zeroWitness.Decision().Valid() || zeroWitness.Rule().Valid() ||
		zeroWitness.ResolvedAuthority().Valid() || zeroWitness.Registration().Valid() {
		t.Fatalf("zero exact witness is observable as valid: %#v", zeroWitness)
	}
}

func TestExactComparisonAndForgedWitnesses(t *testing.T) {
	rule := ExactAuthorityBindingV0Alpha1()
	authority := mustResolvedAuthority(t, "catalog.scope", "r1", "authority.a", "r1")
	matching := mustRegistration(t, "catalog.scope", "r1", "authority.a", "r1", "subject.a", "r1", "capability.a")
	different := mustRegistration(t, "catalog.scope", "r1", "authority.b", "r1", "subject.a", "r1", "capability.a")
	crossScope := mustRegistration(t, "other.scope", "r1", "authority.a", "r1", "subject.a", "r1", "capability.a")

	positive, err := Compare(rule, authority, matching)
	if err != nil {
		t.Fatalf("positive comparison: %v", err)
	}
	witness, ok := positive.ExactBindingWitness()
	if !ok || !witness.Valid() || witness.Decision() != positive ||
		witness.ResolvedAuthority().Binding().Authority() != matching.Binding().DeclarationAuthority() {
		t.Fatalf("positive witness = (%#v, %t)", witness, ok)
	}

	negative, err := Compare(rule, authority, different)
	if err != nil {
		t.Fatalf("negative comparison: %v", err)
	}
	outcome, outcomeOK := negative.Outcome()
	if !outcomeOK || outcome != OutcomeNoExactAuthorityBindingMatch ||
		!negative.Registration().Valid() {
		t.Fatalf("negative decision = (%#v, %d, %t)", negative, outcome, outcomeOK)
	}
	if forged := (ExactBindingWitness{decision: negative}); forged.Valid() {
		t.Fatalf("negative decision forged a positive witness: %#v", forged)
	}

	if value, scopeErr := Compare(rule, authority, crossScope); !errors.Is(scopeErr, ErrCatalogScopeMismatch) || value.Valid() {
		t.Fatalf("cross-scope comparison = (%#v, %v)", value, scopeErr)
	}
	if forged := (Decision{rule: rule, authority: authority, registration: crossScope}); forged.Valid() {
		t.Fatalf("cross-scope forged decision is valid: %#v", forged)
	}
}

func TestMaximumAuthoritySnapshotAndPayloadIndependence(t *testing.T) {
	target := mustAuthorityRecord(t, "catalog.scope", "r1", "authority.target", "r1")
	other := mustAuthorityRecord(t, "catalog.scope", "r1", "authority.other", "r1")
	records := make([]authoritymatching.DeclarationAuthorityRecord, authoritymatching.MaximumAuthorityRecords)
	for index := range records {
		records[index] = other
	}
	records[len(records)-1] = target
	snapshot, err := authoritymatching.NewFiniteAuthoritySnapshot(target.SelfBinding().CatalogScope(), records)
	if err != nil {
		t.Fatalf("maximum snapshot: %v", err)
	}
	match, err := snapshot.Match(target.SelfBinding())
	if err != nil {
		t.Fatalf("target match: %v", err)
	}
	resolution, err := authorityresolution.ResolveInSnapshot(match)
	if err != nil {
		t.Fatalf("target resolution: %v", err)
	}
	authority, ok := resolution.ResolvedAuthority()
	if !ok || authority.SnapshotRecordCount() != len(records) {
		t.Fatalf("maximum authority = (%#v, %t)", authority, ok)
	}

	for _, registration := range []catalogregistration.Registration{
		mustRegistration(t, "catalog.scope", "r1", "authority.target", "r1", "subject.a", "r1", "capability.a"),
		mustRegistration(t, "catalog.scope", "r1", "authority.target", "r1", "subject.other", "r9", "capability.other"),
	} {
		decision, compareErr := Compare(ExactAuthorityBindingV0Alpha1(), authority, registration)
		if compareErr != nil {
			t.Fatalf("payload-independent comparison: %v", compareErr)
		}
		outcome, outcomeOK := decision.Outcome()
		witness, positive := decision.ExactBindingWitness()
		if !outcomeOK || outcome != OutcomeExactAuthorityBindingMatch || !positive ||
			!witness.Valid() || witness.ResolvedAuthority().SnapshotRecordCount() != len(records) {
			t.Fatalf("payload-independent decision = (%#v, %d, %t), witness = (%#v, %t)", decision, outcome, outcomeOK, witness, positive)
		}
	}
}

func FuzzCompareExactAuthorityBinding(f *testing.F) {
	for _, seed := range [][8]string{
		{"scope", "r1", "scope", "r1", "a", "r1", "a", "r1"},
		{"scope", "r1", "scope", "r1", "a", "r1", "b", "r1"},
		{"one", "r1", "two", "r1", "a", "r1", "a", "r1"},
		{"scope", "r1", "scope", "r1", "é", "r1", "a", "r1"},
	} {
		f.Add(seed[0], seed[1], seed[2], seed[3], seed[4], seed[5], seed[6], seed[7], "subject", "capability")
	}
	f.Fuzz(func(
		t *testing.T,
		resolvedCatalogID, resolvedCatalogRevision, registrationCatalogID, registrationCatalogRevision,
		resolvedAuthorityID, resolvedAuthorityRevision, registrationAuthorityID, registrationAuthorityRevision,
		subjectID, capability string,
	) {
		authorityRecord, authorityErr := constructAuthorityRecord(
			resolvedCatalogID, resolvedCatalogRevision, resolvedAuthorityID, resolvedAuthorityRevision,
		)
		if authorityErr != nil {
			return
		}
		snapshot, snapshotErr := authoritymatching.NewFiniteAuthoritySnapshot(
			authorityRecord.SelfBinding().CatalogScope(),
			[]authoritymatching.DeclarationAuthorityRecord{authorityRecord},
		)
		if snapshotErr != nil {
			t.Fatalf("snapshot: %v", snapshotErr)
		}
		match, matchErr := snapshot.Match(authorityRecord.SelfBinding())
		if matchErr != nil {
			t.Fatalf("match: %v", matchErr)
		}
		resolution, resolutionErr := authorityresolution.ResolveInSnapshot(match)
		if resolutionErr != nil {
			t.Fatalf("resolution: %v", resolutionErr)
		}
		authority, resolved := resolution.ResolvedAuthority()
		if !resolved {
			t.Fatal("one-record authority did not resolve")
		}
		registration, registrationErr := constructRegistration(
			registrationCatalogID, registrationCatalogRevision,
			registrationAuthorityID, registrationAuthorityRevision,
			subjectID, "r1", capability,
		)
		if registrationErr != nil {
			return
		}

		decision, err := Compare(ExactAuthorityBindingV0Alpha1(), authority, registration)
		sameScope := authority.Binding().CatalogScope() == registration.Binding().CatalogScope()
		if !sameScope {
			if !errors.Is(err, ErrCatalogScopeMismatch) || decision.Valid() {
				t.Fatalf("cross-scope comparison = (%#v, %v)", decision, err)
			}
			return
		}
		if err != nil || !decision.Valid() {
			t.Fatalf("same-scope comparison = (%#v, %v)", decision, err)
		}
		wantPositive := authority.Binding().Authority() == registration.Binding().DeclarationAuthority()
		outcome, ok := decision.Outcome()
		witness, positive := decision.ExactBindingWitness()
		if !ok || outcome != outcomeForEquality(wantPositive) || positive != wantPositive ||
			witness.Valid() != wantPositive {
			t.Fatalf("comparison = (%#v, %d, %t), witness = (%#v, %t), want %t", decision, outcome, ok, witness, positive, wantPositive)
		}
	})
}

func resolveAuthority(
	t *testing.T,
	snapshot authoritymatching.FiniteAuthoritySnapshot,
	authorityID, authorityRevision string,
) (authorityresolution.SnapshotResolvedAuthority, bool) {
	t.Helper()
	authority, err := catalogregistration.NewDeclarationAuthorityRef(authorityID, authorityRevision)
	if err != nil {
		t.Fatalf("construct authority query: %v", err)
	}
	binding, err := authoritymatching.NewAuthorityMatchBinding(snapshot.CatalogScope(), authority)
	if err != nil {
		t.Fatalf("construct match binding: %v", err)
	}
	match, err := snapshot.Match(binding)
	if err != nil {
		t.Fatalf("match authority: %v", err)
	}
	resolution, err := authorityresolution.ResolveInSnapshot(match)
	if err != nil {
		t.Fatalf("resolve authority: %v", err)
	}
	return resolution.ResolvedAuthority()
}

func parsePositiveExpectation(t *testing.T, caseID, expect, raw string) (bool, bool) {
	t.Helper()
	if expect == "authority_not_resolved" || expect == "catalog_scope_mismatch" {
		if raw != "not_applicable" {
			t.Fatalf("error case %q must not encode a witness, got %q", caseID, raw)
		}
		return false, false
	}
	switch raw {
	case "true":
		return true, true
	case "false":
		return false, true
	default:
		t.Fatalf("case %q has invalid witness flag %q", caseID, raw)
		return false, false
	}
}

func outcomeForEquality(equal bool) Outcome {
	if equal {
		return OutcomeExactAuthorityBindingMatch
	}
	return OutcomeNoExactAuthorityBindingMatch
}

func outcomeToken(value Outcome) string {
	switch value {
	case OutcomeExactAuthorityBindingMatch:
		return "exact_authority_binding_match"
	case OutcomeNoExactAuthorityBindingMatch:
		return "no_exact_authority_binding_match"
	default:
		return "invalid"
	}
}

func mustResolvedAuthority(
	t *testing.T,
	catalogID, catalogRevision, authorityID, authorityRevision string,
) authorityresolution.SnapshotResolvedAuthority {
	t.Helper()
	record := mustAuthorityRecord(t, catalogID, catalogRevision, authorityID, authorityRevision)
	snapshot, err := authoritymatching.NewFiniteAuthoritySnapshot(
		record.SelfBinding().CatalogScope(),
		[]authoritymatching.DeclarationAuthorityRecord{record},
	)
	if err != nil {
		t.Fatalf("construct authority snapshot: %v", err)
	}
	resolved, ok := resolveAuthority(t, snapshot, authorityID, authorityRevision)
	if !ok {
		t.Fatal("one-record authority did not resolve")
	}
	return resolved
}

func mustAuthorityRecord(
	t *testing.T,
	catalogID, catalogRevision, authorityID, authorityRevision string,
) authoritymatching.DeclarationAuthorityRecord {
	t.Helper()
	value, err := constructAuthorityRecord(catalogID, catalogRevision, authorityID, authorityRevision)
	if err != nil {
		t.Fatalf("construct authority record: %v", err)
	}
	return value
}

func constructAuthorityRecord(
	catalogID, catalogRevision, authorityID, authorityRevision string,
) (authoritymatching.DeclarationAuthorityRecord, error) {
	scope, err := catalogregistration.NewCatalogScopeRef(catalogID, catalogRevision)
	if err != nil {
		return authoritymatching.DeclarationAuthorityRecord{}, err
	}
	authority, err := catalogregistration.NewDeclarationAuthorityRef(authorityID, authorityRevision)
	if err != nil {
		return authoritymatching.DeclarationAuthorityRecord{}, err
	}
	binding, err := authoritymatching.NewAuthorityMatchBinding(scope, authority)
	if err != nil {
		return authoritymatching.DeclarationAuthorityRecord{}, err
	}
	return authoritymatching.NewDeclarationAuthorityRecord(binding)
}

func mustRegistration(
	t *testing.T,
	catalogID, catalogRevision, authorityID, authorityRevision,
	subjectID, subjectRevision, capability string,
) catalogregistration.Registration {
	t.Helper()
	value, err := constructRegistration(
		catalogID, catalogRevision, authorityID, authorityRevision,
		subjectID, subjectRevision, capability,
	)
	if err != nil {
		t.Fatalf("construct registration: %v", err)
	}
	return value
}

func constructRegistration(
	catalogID, catalogRevision, authorityID, authorityRevision,
	subjectID, subjectRevision, capability string,
) (catalogregistration.Registration, error) {
	scope, err := catalogregistration.NewCatalogScopeRef(catalogID, catalogRevision)
	if err != nil {
		return catalogregistration.Registration{}, err
	}
	authority, err := catalogregistration.NewDeclarationAuthorityRef(authorityID, authorityRevision)
	if err != nil {
		return catalogregistration.Registration{}, err
	}
	subject, err := catalogregistration.NewSubjectRef(subjectID, subjectRevision)
	if err != nil {
		return catalogregistration.Registration{}, err
	}
	capabilityID, err := catalogregistration.NewCapabilityID(capability)
	if err != nil {
		return catalogregistration.Registration{}, err
	}
	binding, err := catalogregistration.Bind(scope, authority, subject, capabilityID)
	if err != nil {
		return catalogregistration.Registration{}, err
	}
	result, err := catalogregistration.EvaluatedRegistered(binding)
	if err != nil {
		return catalogregistration.Registration{}, err
	}
	registration, ok := result.Registration()
	if !ok {
		return catalogregistration.Registration{}, catalogregistration.ErrInvalidBinding
	}
	return registration, nil
}

func readCorpusTSV(t *testing.T, name string, header []string) [][]string {
	t.Helper()
	path := filepath.Join(
		"..", "..", "testdata", "conformance", "registration-authority-binding-v0alpha1", name,
	)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	if len(data) > maximumCorpusBytes || !utf8.Valid(data) {
		t.Fatalf("%s must be valid UTF-8 within %d bytes", name, maximumCorpusBytes)
	}
	for _, character := range string(data) {
		if (character < 0x20 && character != '\t' && character != '\r' && character != '\n') ||
			(character >= 0x7f && character <= 0x9f) {
			t.Fatalf("%s contains control character U+%04X", name, character)
		}
	}
	reader := csv.NewReader(strings.NewReader(string(data)))
	reader.Comma = '\t'
	reader.FieldsPerRecord = len(header)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	if len(records) < 2 || len(records)-1 > maximumCorpusRows {
		t.Fatalf("%s row count = %d", name, len(records)-1)
	}
	if !slices.Equal(records[0], header) {
		t.Fatalf("%s header = %q, want %q", name, records[0], header)
	}
	for rowIndex, record := range records {
		for columnIndex, field := range record {
			if len(field) > maximumFieldBytes {
				t.Fatalf("%s field %d:%d has %d bytes", name, rowIndex+1, columnIndex+1, len(field))
			}
		}
	}
	return records[1:]
}
