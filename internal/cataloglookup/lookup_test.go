package cataloglookup

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/blisspixel/fartapp/internal/catalogregistration"
)

const (
	maximumCorpusBytes = 64 * 1024
	maximumCorpusRows  = 256
	maximumFieldBytes  = 128
)

func TestLookupCorpus(t *testing.T) {
	snapshotRows := readCorpusTSV(t, "snapshots.tsv", []string{
		"snapshot_id", "catalog_id", "catalog_revision", "expect",
	})
	memberRows := readCorpusTSV(t, "members.tsv", []string{
		"snapshot_id", "catalog_id", "catalog_revision", "authority_id",
		"authority_revision", "subject_id", "subject_revision", "capability",
	})
	caseRows := readCorpusTSV(t, "cases.tsv", []string{
		"case_id", "snapshot_id", "catalog_id", "catalog_revision", "authority_id",
		"authority_revision", "subject_id", "subject_revision", "capability", "expect",
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
			"duplicate_registration", "invalid_registration",
			"member_scope_mismatch", "valid_snapshot",
		}, row[3]) {
			t.Fatalf("snapshot expectation %q is unsupported", row[3])
		}
		scope, err := catalogregistration.NewCatalogScopeRef(row[1], row[2])
		if err != nil {
			t.Fatalf("snapshot %q scope: %v", row[0], err)
		}
		fixtures[row[0]] = snapshotFixture{scope: scope, expect: row[3]}
	}

	members := make(map[string][]catalogregistration.Registration)
	memberInvalid := make(map[string]bool)
	for _, row := range memberRows {
		if _, exists := fixtures[row[0]]; !exists {
			t.Fatalf("member references unknown snapshot %q", row[0])
		}
		registration, err := constructRegistration(row[1], row[2], row[3], row[4], row[5], row[6], row[7])
		if err != nil {
			memberInvalid[row[0]] = true
			continue
		}
		members[row[0]] = append(members[row[0]], registration)
	}

	validSnapshots := make(map[string]FiniteSnapshot)
	for id, fixture := range fixtures {
		actual := "valid_snapshot"
		var snapshot FiniteSnapshot
		if memberInvalid[id] {
			actual = "invalid_registration"
		} else {
			var err error
			snapshot, err = NewFiniteSnapshot(fixture.scope, members[id])
			switch {
			case errors.Is(err, ErrDuplicateRegistration):
				actual = "duplicate_registration"
			case errors.Is(err, ErrScopeMismatch):
				actual = "member_scope_mismatch"
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

	seenCases := make(map[string]struct{}, len(caseRows))
	for _, row := range caseRows {
		if row[0] == "" {
			t.Fatal("case id is empty")
		}
		if _, exists := seenCases[row[0]]; exists {
			t.Fatalf("case id %q is duplicated", row[0])
		}
		seenCases[row[0]] = struct{}{}
		if !slices.Contains([]string{"not_registered", "registered", "scope_mismatch"}, row[9]) {
			t.Fatalf("case expectation %q is unsupported", row[9])
		}
		snapshot, exists := validSnapshots[row[1]]
		if !exists {
			t.Fatalf("case %q references non-valid snapshot %q", row[0], row[1])
		}
		t.Run(row[0], func(t *testing.T) {
			binding, err := constructBinding(row[2], row[3], row[4], row[5], row[6], row[7], row[8])
			if err != nil {
				t.Fatalf("construct query: %v", err)
			}
			lookup, err := snapshot.Lookup(binding)
			if row[9] == "scope_mismatch" {
				if !errors.Is(err, ErrScopeMismatch) || lookup.Valid() {
					t.Fatalf("scope mismatch = (%#v, %v)", lookup, err)
				}
				return
			}
			if err != nil || !lookup.Valid() {
				t.Fatalf("lookup = (%#v, %v)", lookup, err)
			}
			presence, ok := lookup.Presence()
			if !ok {
				t.Fatal("valid lookup has no presence")
			}
			if row[9] == "registered" {
				registration, registered := lookup.Registration()
				if presence != catalogregistration.PresenceRegistered ||
					!registered || !registration.Valid() || registration.Binding() != binding {
					t.Fatalf("registered lookup = (%d, %#v, %t)", presence, registration, registered)
				}
				return
			}
			registration, registered := lookup.Registration()
			if presence != catalogregistration.PresenceNotRegistered || registered || registration.Valid() {
				t.Fatalf("negative lookup = (%d, %#v, %t)", presence, registration, registered)
			}
		})
	}
}

func TestFiniteSnapshotBoundsAndInvalidInputs(t *testing.T) {
	scope := mustScope(t, "catalog.scope", "r1")
	empty, err := NewFiniteSnapshot(scope, nil)
	if err != nil || !empty.Valid() || empty.Scope() != scope || empty.RegistrationCount() != 0 {
		t.Fatalf("empty snapshot = (%#v, %v)", empty, err)
	}

	registrations := make([]catalogregistration.Registration, MaximumRegistrations)
	for index := range registrations {
		registrations[index] = mustRegistration(
			t, "catalog.scope", "r1", "authority.a", "r1", "subject.a", "r1",
			fmt.Sprintf("capability.%04d", index),
		)
	}
	maximum, err := NewFiniteSnapshot(scope, registrations)
	if err != nil || !maximum.Valid() || maximum.RegistrationCount() != MaximumRegistrations {
		t.Fatalf("maximum snapshot = (%#v, %v)", maximum, err)
	}

	tooMany := make([]catalogregistration.Registration, MaximumRegistrations+1)
	if value, err := NewFiniteSnapshot(scope, tooMany); !errors.Is(err, ErrSnapshotLimit) || value.Valid() {
		t.Fatalf("oversized snapshot = (%#v, %v)", value, err)
	}
	if value, err := NewFiniteSnapshot(catalogregistration.CatalogScopeRef{}, nil); !errors.Is(err, ErrInvalidSnapshot) || value.Valid() {
		t.Fatalf("zero-scope snapshot = (%#v, %v)", value, err)
	}
	if value, err := NewFiniteSnapshot(scope, []catalogregistration.Registration{{}}); !errors.Is(err, ErrInvalidSnapshot) || value.Valid() {
		t.Fatalf("zero-member snapshot = (%#v, %v)", value, err)
	}

	member := mustRegistration(t, "other.scope", "r1", "authority.a", "r1", "subject.a", "r1", "capability.a")
	if value, err := NewFiniteSnapshot(scope, []catalogregistration.Registration{member}); !errors.Is(err, ErrScopeMismatch) || value.Valid() {
		t.Fatalf("member scope mismatch = (%#v, %v)", value, err)
	}
	duplicate := mustRegistration(t, "catalog.scope", "r1", "authority.a", "r1", "subject.a", "r1", "capability.a")
	if value, err := NewFiniteSnapshot(scope, []catalogregistration.Registration{duplicate, duplicate}); !errors.Is(err, ErrDuplicateRegistration) || value.Valid() {
		t.Fatalf("duplicate snapshot = (%#v, %v)", value, err)
	}
}

func TestFiniteSnapshotDefensivelyCopiesInput(t *testing.T) {
	original := mustRegistration(t, "catalog.scope", "r1", "authority.a", "r1", "subject.a", "r1", "capability.a")
	replacement := mustRegistration(t, "catalog.scope", "r1", "authority.b", "r1", "subject.b", "r1", "capability.b")
	input := []catalogregistration.Registration{original}
	snapshot, err := NewFiniteSnapshot(original.Binding().CatalogScope(), input)
	if err != nil {
		t.Fatalf("construct snapshot: %v", err)
	}
	input[0] = replacement
	assertPresence(t, snapshot, original.Binding(), catalogregistration.PresenceRegistered)
	assertPresence(t, snapshot, replacement.Binding(), catalogregistration.PresenceNotRegistered)
}

func TestExactNegativeAndScopeMismatch(t *testing.T) {
	member := mustRegistration(t, "catalog.scope", "r1", "authority.a", "r1", "subject.a", "r1", "capability.a")
	snapshot, err := NewFiniteSnapshot(member.Binding().CatalogScope(), []catalogregistration.Registration{member})
	if err != nil {
		t.Fatalf("construct snapshot: %v", err)
	}
	variants := []catalogregistration.Binding{
		mustBinding(t, "catalog.scope", "r1", "authority.other", "r1", "subject.a", "r1", "capability.a"),
		mustBinding(t, "catalog.scope", "r1", "authority.a", "r2", "subject.a", "r1", "capability.a"),
		mustBinding(t, "catalog.scope", "r1", "authority.a", "r1", "subject.other", "r1", "capability.a"),
		mustBinding(t, "catalog.scope", "r1", "authority.a", "r1", "subject.a", "r2", "capability.a"),
		mustBinding(t, "catalog.scope", "r1", "authority.a", "r1", "subject.a", "r1", "capability.other"),
	}
	for index, binding := range variants {
		t.Run(fmt.Sprintf("variant_%d", index), func(t *testing.T) {
			assertPresence(t, snapshot, binding, catalogregistration.PresenceNotRegistered)
		})
	}
	for _, binding := range []catalogregistration.Binding{
		mustBinding(t, "other.scope", "r1", "authority.a", "r1", "subject.a", "r1", "capability.a"),
		mustBinding(t, "catalog.scope", "r2", "authority.a", "r1", "subject.a", "r1", "capability.a"),
	} {
		lookup, err := snapshot.Lookup(binding)
		if !errors.Is(err, ErrScopeMismatch) || lookup.Valid() {
			t.Fatalf("scope mismatch = (%#v, %v)", lookup, err)
		}
	}
	if lookup, err := snapshot.Lookup(catalogregistration.Binding{}); !errors.Is(err, catalogregistration.ErrInvalidBinding) || lookup.Valid() {
		t.Fatalf("zero query = (%#v, %v)", lookup, err)
	}
}

func TestInputOrderAndOpaqueAuthorities(t *testing.T) {
	first := mustRegistration(t, "catalog.scope", "r1", "authority.a", "r1", "subject.a", "r1", "capability.a")
	second := mustRegistration(t, "catalog.scope", "r1", "authority.b", "r1", "subject.a", "r1", "capability.a")
	forward, forwardErr := NewFiniteSnapshot(first.Binding().CatalogScope(), []catalogregistration.Registration{first, second})
	reverse, reverseErr := NewFiniteSnapshot(first.Binding().CatalogScope(), []catalogregistration.Registration{second, first})
	if forwardErr != nil || reverseErr != nil || forward.RegistrationCount() != reverse.RegistrationCount() {
		t.Fatalf("ordered snapshots = (%#v, %v), (%#v, %v)", forward, forwardErr, reverse, reverseErr)
	}
	for _, binding := range []catalogregistration.Binding{first.Binding(), second.Binding()} {
		assertPresence(t, forward, binding, catalogregistration.PresenceRegistered)
		assertPresence(t, reverse, binding, catalogregistration.PresenceRegistered)
	}
}

func TestSnapshotLookupWitnessAndInvalidValues(t *testing.T) {
	var zero FiniteSnapshot
	if zero.Valid() || zero.Scope().Valid() || zero.RegistrationCount() != 0 {
		t.Fatalf("zero snapshot is observable as valid: %#v", zero)
	}
	if lookup, err := zero.Lookup(catalogregistration.Binding{}); !errors.Is(err, ErrInvalidSnapshot) || lookup.Valid() {
		t.Fatalf("zero snapshot lookup = (%#v, %v)", lookup, err)
	}

	member := mustRegistration(t, "catalog.scope", "r1", "authority.a", "r1", "subject.a", "r1", "capability.a")
	snapshot, err := NewFiniteSnapshot(member.Binding().CatalogScope(), []catalogregistration.Registration{member})
	if err != nil {
		t.Fatalf("construct snapshot: %v", err)
	}
	lookup, err := snapshot.Lookup(member.Binding())
	if err != nil || !lookup.Valid() || lookup.SnapshotScope() != snapshot.Scope() ||
		lookup.RegistrationCount() != 1 || lookup.Binding() != member.Binding() {
		t.Fatalf("snapshot lookup = (%#v, %v)", lookup, err)
	}

	var zeroLookup SnapshotLookup
	if zeroLookup.Valid() {
		t.Fatal("zero lookup is valid")
	}
	if presence, ok := zeroLookup.Presence(); ok || presence != catalogregistration.PresenceInvalid {
		t.Fatalf("zero lookup presence = (%d, %t)", presence, ok)
	}
	if registration, ok := zeroLookup.Registration(); ok || registration.Valid() {
		t.Fatalf("zero lookup registration = (%#v, %t)", registration, ok)
	}

	notRegistered, notRegisteredErr := catalogregistration.EvaluatedNotRegistered(member.Binding())
	if notRegisteredErr != nil {
		t.Fatalf("construct negative result: %v", notRegisteredErr)
	}
	if forged := (SnapshotLookup{snapshot: snapshot, result: notRegistered}); forged.Valid() {
		t.Fatal("member with negative result is valid")
	}

	absentBinding := mustBinding(t, "catalog.scope", "r1", "authority.a", "r1", "subject.a", "r1", "capability.absent")
	registeredAbsent, registeredAbsentErr := catalogregistration.EvaluatedRegistered(absentBinding)
	if registeredAbsentErr != nil {
		t.Fatalf("construct false-positive result: %v", registeredAbsentErr)
	}
	if forged := (SnapshotLookup{snapshot: snapshot, result: registeredAbsent}); forged.Valid() {
		t.Fatal("absent member with positive result is valid")
	}

	notEvaluated, notEvaluatedErr := catalogregistration.NotEvaluated(absentBinding)
	if notEvaluatedErr != nil {
		t.Fatalf("construct non-evaluated result: %v", notEvaluatedErr)
	}
	if forged := (SnapshotLookup{snapshot: snapshot, result: notEvaluated}); forged.Valid() {
		t.Fatal("non-evaluated snapshot lookup is valid")
	}

	otherBinding := mustBinding(t, "other.scope", "r1", "authority.a", "r1", "subject.a", "r1", "capability.a")
	otherResult, otherErr := catalogregistration.EvaluatedRegistered(otherBinding)
	if otherErr != nil {
		t.Fatalf("construct other-scope result: %v", otherErr)
	}
	if forged := (SnapshotLookup{snapshot: snapshot, result: otherResult}); forged.Valid() {
		t.Fatal("other-scope result is valid")
	}

	invalidCore := FiniteSnapshot{core: &snapshotCore{
		scope: snapshot.Scope(),
		members: map[catalogregistration.Binding]struct{}{
			{}: {},
		},
	}}
	if invalidCore.Valid() {
		t.Fatal("snapshot with invalid member is valid")
	}
	if nilMembers := (FiniteSnapshot{core: &snapshotCore{scope: snapshot.Scope()}}); nilMembers.Valid() {
		t.Fatal("snapshot with nil members map is valid")
	}
}

func FuzzFiniteSnapshotAndLookup(f *testing.F) {
	for _, seed := range [][3]string{
		{"a", "b", "a"},
		{"same", "same", "other"},
		{".", "valid", "valid"},
		{"é", "a", "b"},
	} {
		f.Add(seed[0], seed[1], seed[2])
	}
	f.Fuzz(func(t *testing.T, firstValue, secondValue, queryValue string) {
		scope, scopeErr := catalogregistration.NewCatalogScopeRef("fuzz.scope", "r1")
		if scopeErr != nil {
			t.Fatalf("construct static scope: %v", scopeErr)
		}
		first, firstErr := constructRegistration("fuzz.scope", "r1", "authority.a", "r1", "subject.a", "r1", firstValue)
		second, secondErr := constructRegistration("fuzz.scope", "r1", "authority.a", "r1", "subject.a", "r1", secondValue)
		if firstErr != nil || secondErr != nil {
			return
		}
		snapshot, err := NewFiniteSnapshot(scope, []catalogregistration.Registration{first, second})
		if first.Binding() == second.Binding() {
			if !errors.Is(err, ErrDuplicateRegistration) || snapshot.Valid() {
				t.Fatalf("duplicate snapshot = (%#v, %v)", snapshot, err)
			}
			return
		}
		if err != nil || !snapshot.Valid() {
			t.Fatalf("fuzz snapshot = (%#v, %v)", snapshot, err)
		}
		query, queryErr := constructBinding("fuzz.scope", "r1", "authority.a", "r1", "subject.a", "r1", queryValue)
		if queryErr != nil {
			return
		}
		lookup, lookupErr := snapshot.Lookup(query)
		if lookupErr != nil || !lookup.Valid() {
			t.Fatalf("fuzz lookup = (%#v, %v)", lookup, lookupErr)
		}
		presence, ok := lookup.Presence()
		wantRegistered := query == first.Binding() || query == second.Binding()
		if !ok || wantRegistered != (presence == catalogregistration.PresenceRegistered) {
			t.Fatalf("fuzz presence = (%d, %t), registered = %t", presence, ok, wantRegistered)
		}
	})
}

func assertPresence(
	t *testing.T,
	snapshot FiniteSnapshot,
	binding catalogregistration.Binding,
	want catalogregistration.Presence,
) {
	t.Helper()
	lookup, err := snapshot.Lookup(binding)
	if err != nil || !lookup.Valid() {
		t.Fatalf("lookup = (%#v, %v)", lookup, err)
	}
	presence, ok := lookup.Presence()
	if !ok || presence != want {
		t.Fatalf("presence = (%d, %t), want %d", presence, ok, want)
	}
}

func mustScope(t *testing.T, id, revision string) catalogregistration.CatalogScopeRef {
	t.Helper()
	value, err := catalogregistration.NewCatalogScopeRef(id, revision)
	if err != nil {
		t.Fatalf("construct scope: %v", err)
	}
	return value
}

func mustBinding(
	t *testing.T,
	catalogID, catalogRevision, authorityID, authorityRevision,
	subjectID, subjectRevision, capability string,
) catalogregistration.Binding {
	t.Helper()
	value, err := constructBinding(
		catalogID, catalogRevision, authorityID, authorityRevision,
		subjectID, subjectRevision, capability,
	)
	if err != nil {
		t.Fatalf("construct binding: %v", err)
	}
	return value
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

func constructBinding(
	catalogID, catalogRevision, authorityID, authorityRevision,
	subjectID, subjectRevision, capabilityValue string,
) (catalogregistration.Binding, error) {
	scope, err := catalogregistration.NewCatalogScopeRef(catalogID, catalogRevision)
	if err != nil {
		return catalogregistration.Binding{}, err
	}
	authority, err := catalogregistration.NewDeclarationAuthorityRef(authorityID, authorityRevision)
	if err != nil {
		return catalogregistration.Binding{}, err
	}
	subject, err := catalogregistration.NewSubjectRef(subjectID, subjectRevision)
	if err != nil {
		return catalogregistration.Binding{}, err
	}
	capability, err := catalogregistration.NewCapabilityID(capabilityValue)
	if err != nil {
		return catalogregistration.Binding{}, err
	}
	return catalogregistration.Bind(scope, authority, subject, capability)
}

func constructRegistration(
	catalogID, catalogRevision, authorityID, authorityRevision,
	subjectID, subjectRevision, capabilityValue string,
) (catalogregistration.Registration, error) {
	binding, err := constructBinding(
		catalogID, catalogRevision, authorityID, authorityRevision,
		subjectID, subjectRevision, capabilityValue,
	)
	if err != nil {
		return catalogregistration.Registration{}, err
	}
	result, err := catalogregistration.EvaluatedRegistered(binding)
	if err != nil {
		return catalogregistration.Registration{}, err
	}
	registration, ok := result.Registration()
	if !ok {
		return catalogregistration.Registration{}, ErrInvalidSnapshot
	}
	return registration, nil
}

func readCorpusTSV(t *testing.T, name string, header []string) [][]string {
	t.Helper()
	path := filepath.Join(
		"..", "..", "testdata", "conformance", "catalog-lookup-v0alpha1", name,
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
