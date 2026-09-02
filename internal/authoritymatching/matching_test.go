package authoritymatching

import (
	"encoding/csv"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strconv"
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

func TestAuthorityMatchingCorpus(t *testing.T) {
	snapshotRows := readCorpusTSV(t, "snapshots.tsv", []string{
		"snapshot_id", "catalog_id", "catalog_revision", "expect",
	})
	recordRows := readCorpusTSV(t, "records.tsv", []string{
		"snapshot_id", "catalog_id", "catalog_revision", "authority_id", "authority_revision",
	})
	caseRows := readCorpusTSV(t, "cases.tsv", []string{
		"case_id", "snapshot_id", "catalog_id", "catalog_revision",
		"authority_id", "authority_revision", "expect", "match_count",
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

	records := make(map[string][]DeclarationAuthorityRecord)
	recordInvalid := make(map[string]bool)
	for _, row := range recordRows {
		if _, exists := fixtures[row[0]]; !exists {
			t.Fatalf("record references unknown snapshot %q", row[0])
		}
		record, err := constructRecord(row[1], row[2], row[3], row[4])
		if err != nil {
			recordInvalid[row[0]] = true
			continue
		}
		records[row[0]] = append(records[row[0]], record)
	}

	validSnapshots := make(map[string]FiniteAuthoritySnapshot)
	for id, fixture := range fixtures {
		actual := "valid_snapshot"
		var snapshot FiniteAuthoritySnapshot
		if recordInvalid[id] {
			actual = "invalid_record"
		} else {
			var err error
			snapshot, err = NewFiniteAuthoritySnapshot(fixture.scope, records[id])
			switch {
			case errors.Is(err, ErrRecordScopeMismatch):
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
			"match_scope_mismatch", "multiple_matches_in_snapshot",
			"no_match_in_snapshot", "one_match_in_snapshot",
		}, row[6]) {
			t.Fatalf("case expectation %q is unsupported", row[6])
		}
		wantCount := 0
		if row[6] == "match_scope_mismatch" {
			if row[7] != "not_applicable" {
				t.Fatalf("error case %q must not encode a match count, got %q", row[0], row[7])
			}
		} else {
			var err error
			wantCount, err = strconv.Atoi(row[7])
			if err != nil || wantCount < 0 || wantCount > MaximumAuthorityRecords {
				t.Fatalf("case %q has invalid match count %q", row[0], row[7])
			}
		}
		snapshot, exists := validSnapshots[row[1]]
		if !exists {
			t.Fatalf("case %q references non-valid snapshot %q", row[0], row[1])
		}
		t.Run(row[0], func(t *testing.T) {
			binding, err := constructBinding(row[2], row[3], row[4], row[5])
			if err != nil {
				t.Fatalf("construct match binding: %v", err)
			}
			match, err := snapshot.Match(binding)
			if row[6] == "match_scope_mismatch" {
				if !errors.Is(err, ErrMatchScopeMismatch) || match.Valid() {
					t.Fatalf("scope mismatch = (%#v, %v)", match, err)
				}
				return
			}
			if err != nil || !match.Valid() || match.MatchCount() != wantCount {
				t.Fatalf("match = (%#v, %v), count = %d", match, err, match.MatchCount())
			}
			cardinality, ok := match.Cardinality()
			if !ok || cardinalityToken(cardinality) != row[6] {
				t.Fatalf("cardinality = (%d, %t), want %q", cardinality, ok, row[6])
			}
			witness, one := match.SnapshotAuthorityMatch()
			if row[6] == "one_match_in_snapshot" {
				if !one || !witness.Valid() || witness.Binding() != binding ||
					witness.Record().SelfBinding() != binding {
					t.Fatalf("one-match witness = (%#v, %t)", witness, one)
				}
			} else if one || witness.Valid() {
				t.Fatalf("non-unique witness = (%#v, %t)", witness, one)
			}
		})
	}
}

func TestFiniteAuthoritySnapshotBoundsAndInvalidInputs(t *testing.T) {
	scope := mustScope(t, "catalog.scope", "r1")
	empty, err := NewFiniteAuthoritySnapshot(scope, nil)
	if err != nil || !empty.Valid() || empty.CatalogScope() != scope ||
		empty.RecordCount() != 0 || empty.DistinctBindingCount() != 0 {
		t.Fatalf("empty snapshot = (%#v, %v)", empty, err)
	}

	record := mustRecord(t, "catalog.scope", "r1", "authority.a", "r1")
	records := make([]DeclarationAuthorityRecord, MaximumAuthorityRecords)
	for index := range records {
		records[index] = record
	}
	maximum, err := NewFiniteAuthoritySnapshot(scope, records)
	if err != nil || !maximum.Valid() || maximum.RecordCount() != MaximumAuthorityRecords ||
		maximum.DistinctBindingCount() != 1 {
		t.Fatalf("maximum snapshot = (%#v, %v)", maximum, err)
	}
	match, err := maximum.Match(record.SelfBinding())
	if err != nil || match.MatchCount() != MaximumAuthorityRecords {
		t.Fatalf("maximum match = (%#v, %v)", match, err)
	}

	tooMany := make([]DeclarationAuthorityRecord, MaximumAuthorityRecords+1)
	if value, err := NewFiniteAuthoritySnapshot(scope, tooMany); !errors.Is(err, ErrSnapshotLimit) || value.Valid() {
		t.Fatalf("oversized snapshot = (%#v, %v)", value, err)
	}
	if value, err := NewFiniteAuthoritySnapshot(catalogregistration.CatalogScopeRef{}, nil); !errors.Is(err, ErrInvalidSnapshot) || value.Valid() {
		t.Fatalf("zero-scope snapshot = (%#v, %v)", value, err)
	}
	if value, err := NewFiniteAuthoritySnapshot(scope, []DeclarationAuthorityRecord{{}}); !errors.Is(err, ErrInvalidRecord) || value.Valid() {
		t.Fatalf("zero-record snapshot = (%#v, %v)", value, err)
	}
	other := mustRecord(t, "other.scope", "r1", "authority.a", "r1")
	if value, err := NewFiniteAuthoritySnapshot(scope, []DeclarationAuthorityRecord{other}); !errors.Is(err, ErrRecordScopeMismatch) || value.Valid() {
		t.Fatalf("record scope mismatch = (%#v, %v)", value, err)
	}
	if _, err := NewDeclarationAuthorityRecord(AuthorityMatchBinding{}); !errors.Is(err, ErrInvalidRecord) {
		t.Fatalf("zero record error = %v", err)
	}
	if _, err := NewAuthorityMatchBinding(catalogregistration.CatalogScopeRef{}, catalogregistration.DeclarationAuthorityRef{}); !errors.Is(err, ErrInvalidMatchBinding) {
		t.Fatalf("zero binding error = %v", err)
	}
}

func TestSnapshotDefensiveCopyOrderAndTargetLocalDuplicates(t *testing.T) {
	a := mustRecord(t, "catalog.scope", "r1", "authority.a", "r1")
	b := mustRecord(t, "catalog.scope", "r1", "authority.b", "r1")
	c := mustRecord(t, "catalog.scope", "r1", "authority.c", "r1")
	input := []DeclarationAuthorityRecord{a, a, b}
	snapshot, err := NewFiniteAuthoritySnapshot(a.SelfBinding().CatalogScope(), input)
	if err != nil {
		t.Fatalf("construct snapshot: %v", err)
	}
	input[0], input[1], input[2] = c, c, c
	assertCardinality(t, snapshot, a.SelfBinding(), AuthorityMatchCardinalityMultipleMatchesInSnapshot, 2)
	assertCardinality(t, snapshot, b.SelfBinding(), AuthorityMatchCardinalityOneMatchInSnapshot, 1)
	assertCardinality(t, snapshot, c.SelfBinding(), AuthorityMatchCardinalityNoMatchInSnapshot, 0)
	for index, records := range [][]DeclarationAuthorityRecord{
		{a, b, a},
		{b, a, a},
	} {
		interleaved, interleavedErr := NewFiniteAuthoritySnapshot(a.SelfBinding().CatalogScope(), records)
		if interleavedErr != nil {
			t.Fatalf("construct interleaved snapshot %d: %v", index, interleavedErr)
		}
		assertCardinality(t, interleaved, a.SelfBinding(), AuthorityMatchCardinalityMultipleMatchesInSnapshot, 2)
		assertCardinality(t, interleaved, b.SelfBinding(), AuthorityMatchCardinalityOneMatchInSnapshot, 1)
	}

	forward, forwardErr := NewFiniteAuthoritySnapshot(a.SelfBinding().CatalogScope(), []DeclarationAuthorityRecord{a, b, c})
	reverse, reverseErr := NewFiniteAuthoritySnapshot(a.SelfBinding().CatalogScope(), []DeclarationAuthorityRecord{c, b, a})
	if forwardErr != nil || reverseErr != nil || forward.RecordCount() != reverse.RecordCount() ||
		forward.DistinctBindingCount() != reverse.DistinctBindingCount() {
		t.Fatalf("ordered snapshots = (%#v, %v), (%#v, %v)", forward, forwardErr, reverse, reverseErr)
	}
	for _, binding := range []AuthorityMatchBinding{a.SelfBinding(), b.SelfBinding(), c.SelfBinding()} {
		assertCardinality(t, forward, binding, AuthorityMatchCardinalityOneMatchInSnapshot, 1)
		assertCardinality(t, reverse, binding, AuthorityMatchCardinalityOneMatchInSnapshot, 1)
	}
}

func TestExactRevisionAndCrossScopeBehavior(t *testing.T) {
	record := mustRecord(t, "catalog.scope", "r1", "authority.a", "r1")
	snapshot, err := NewFiniteAuthoritySnapshot(record.SelfBinding().CatalogScope(), []DeclarationAuthorityRecord{record})
	if err != nil {
		t.Fatalf("construct snapshot: %v", err)
	}
	for _, binding := range []AuthorityMatchBinding{
		mustBinding(t, "catalog.scope", "r1", "authority.other", "r1"),
		mustBinding(t, "catalog.scope", "r1", "authority.a", "r2"),
	} {
		assertCardinality(t, snapshot, binding, AuthorityMatchCardinalityNoMatchInSnapshot, 0)
	}
	for _, binding := range []AuthorityMatchBinding{
		mustBinding(t, "other.scope", "r1", "authority.a", "r1"),
		mustBinding(t, "catalog.scope", "r2", "authority.a", "r1"),
	} {
		match, err := snapshot.Match(binding)
		if !errors.Is(err, ErrMatchScopeMismatch) || match.Valid() {
			t.Fatalf("cross-scope match = (%#v, %v)", match, err)
		}
	}
	if match, err := snapshot.Match(AuthorityMatchBinding{}); !errors.Is(err, ErrInvalidMatchBinding) || match.Valid() {
		t.Fatalf("zero match binding = (%#v, %v)", match, err)
	}

	sharedOne := mustRecord(t, "scope.one", "r1", "authority.shared", "r1")
	sharedTwo := mustRecord(t, "scope.two", "r1", "authority.shared", "r1")
	one, oneErr := NewFiniteAuthoritySnapshot(sharedOne.SelfBinding().CatalogScope(), []DeclarationAuthorityRecord{sharedOne})
	two, twoErr := NewFiniteAuthoritySnapshot(sharedTwo.SelfBinding().CatalogScope(), []DeclarationAuthorityRecord{sharedTwo})
	if oneErr != nil || twoErr != nil {
		t.Fatalf("cross-scope snapshots = (%v, %v)", oneErr, twoErr)
	}
	assertCardinality(t, one, sharedOne.SelfBinding(), AuthorityMatchCardinalityOneMatchInSnapshot, 1)
	assertCardinality(t, two, sharedTwo.SelfBinding(), AuthorityMatchCardinalityOneMatchInSnapshot, 1)
}

func TestWitnessAndForgedValues(t *testing.T) {
	var zeroSnapshot FiniteAuthoritySnapshot
	if zeroSnapshot.Valid() || zeroSnapshot.CatalogScope().Valid() ||
		zeroSnapshot.RecordCount() != 0 || zeroSnapshot.DistinctBindingCount() != 0 {
		t.Fatalf("zero snapshot is observable as valid: %#v", zeroSnapshot)
	}
	if match, err := zeroSnapshot.Match(AuthorityMatchBinding{}); !errors.Is(err, ErrInvalidSnapshot) || match.Valid() {
		t.Fatalf("zero snapshot match = (%#v, %v)", match, err)
	}

	record := mustRecord(t, "catalog.scope", "r1", "authority.a", "r1")
	snapshot, err := NewFiniteAuthoritySnapshot(record.SelfBinding().CatalogScope(), []DeclarationAuthorityRecord{record})
	if err != nil {
		t.Fatalf("construct snapshot: %v", err)
	}
	match, err := snapshot.Match(record.SelfBinding())
	if err != nil || !match.Valid() || match.SnapshotScope() != snapshot.CatalogScope() ||
		match.Binding() != record.SelfBinding() || match.SnapshotRecordCount() != 1 {
		t.Fatalf("authority match = (%#v, %v)", match, err)
	}
	if match.Binding().Authority().ID() != "authority.a" ||
		match.Binding().Authority().Revision() != "r1" {
		t.Fatalf("authority binding accessor = %#v", match.Binding().Authority())
	}
	witness, ok := match.SnapshotAuthorityMatch()
	if !ok || !witness.Valid() || witness.Binding() != record.SelfBinding() ||
		witness.SnapshotRecordCount() != 1 || !witness.Record().Valid() ||
		witness.Record().SelfBinding() != witness.Binding() {
		t.Fatalf("snapshot authority match = (%#v, %t)", witness, ok)
	}

	var zeroMatch AuthorityMatch
	if zeroMatch.Valid() || zeroMatch.MatchCount() != 0 {
		t.Fatal("zero authority match is valid")
	}
	if cardinality, present := zeroMatch.Cardinality(); present || cardinality != AuthorityMatchCardinalityInvalid {
		t.Fatalf("zero cardinality = (%d, %t)", cardinality, present)
	}
	if one, present := zeroMatch.SnapshotAuthorityMatch(); present || one.Valid() {
		t.Fatalf("zero one-match witness = (%#v, %t)", one, present)
	}
	var zeroWitness SnapshotAuthorityMatch
	if zeroWitness.Valid() || zeroWitness.Record().Valid() {
		t.Fatal("zero snapshot authority match is valid")
	}

	duplicate, duplicateErr := NewFiniteAuthoritySnapshot(
		record.SelfBinding().CatalogScope(),
		[]DeclarationAuthorityRecord{record, record},
	)
	if duplicateErr != nil {
		t.Fatalf("construct duplicate snapshot: %v", duplicateErr)
	}
	duplicateMatch, duplicateMatchErr := duplicate.Match(record.SelfBinding())
	if duplicateMatchErr != nil {
		t.Fatalf("construct duplicate match: %v", duplicateMatchErr)
	}
	if forged := (SnapshotAuthorityMatch{match: duplicateMatch}); forged.Valid() {
		t.Fatal("multiple-match forged witness is valid")
	}

	otherBinding := mustBinding(t, "other.scope", "r1", "authority.a", "r1")
	if forged := (AuthorityMatch{snapshot: snapshot, binding: otherBinding}); forged.Valid() {
		t.Fatal("cross-scope forged match is valid")
	}
	invalidSnapshots := []FiniteAuthoritySnapshot{
		{core: &snapshotCore{catalogScope: snapshot.CatalogScope()}},
		{core: &snapshotCore{
			catalogScope: snapshot.CatalogScope(),
			multiplicity: map[AuthorityMatchBinding]int{},
			recordCount:  -1,
		}},
		{core: &snapshotCore{
			catalogScope: snapshot.CatalogScope(),
			multiplicity: map[AuthorityMatchBinding]int{{}: 1},
			recordCount:  1,
		}},
		{core: &snapshotCore{
			catalogScope: snapshot.CatalogScope(),
			multiplicity: map[AuthorityMatchBinding]int{otherBinding: 1},
			recordCount:  1,
		}},
		{core: &snapshotCore{
			catalogScope: snapshot.CatalogScope(),
			multiplicity: map[AuthorityMatchBinding]int{record.SelfBinding(): 0},
		}},
		{core: &snapshotCore{
			catalogScope: snapshot.CatalogScope(),
			multiplicity: map[AuthorityMatchBinding]int{record.SelfBinding(): 1},
			recordCount:  2,
		}},
		{core: &snapshotCore{
			catalogScope: snapshot.CatalogScope(),
			multiplicity: map[AuthorityMatchBinding]int{
				record.SelfBinding(): int(^uint(0) >> 1),
			},
			recordCount: MaximumAuthorityRecords,
		}},
	}
	for index, invalid := range invalidSnapshots {
		if invalid.Valid() {
			t.Fatalf("forged snapshot %d is valid: %#v", index, invalid)
		}
	}
}

func FuzzFiniteAuthorityMatching(f *testing.F) {
	for _, seed := range [][3]string{
		{"a", "b", "a"},
		{"same", "same", "same"},
		{".", "valid", "valid"},
		{"é", "a", "b"},
	} {
		f.Add(seed[0], seed[1], seed[2])
	}
	f.Fuzz(func(t *testing.T, firstID, secondID, queryID string) {
		first, firstErr := constructRecord("fuzz.scope", "r1", firstID, "r1")
		second, secondErr := constructRecord("fuzz.scope", "r1", secondID, "r1")
		if firstErr != nil || secondErr != nil {
			return
		}
		snapshot, err := NewFiniteAuthoritySnapshot(
			first.SelfBinding().CatalogScope(),
			[]DeclarationAuthorityRecord{first, second},
		)
		if err != nil || !snapshot.Valid() {
			t.Fatalf("fuzz snapshot = (%#v, %v)", snapshot, err)
		}
		query, queryErr := constructBinding("fuzz.scope", "r1", queryID, "r1")
		if queryErr != nil {
			return
		}
		match, matchErr := snapshot.Match(query)
		if matchErr != nil || !match.Valid() {
			t.Fatalf("fuzz match = (%#v, %v)", match, matchErr)
		}
		wantCount := 0
		if query == first.SelfBinding() {
			wantCount++
		}
		if query == second.SelfBinding() {
			wantCount++
		}
		cardinality, ok := match.Cardinality()
		if !ok || match.MatchCount() != wantCount || cardinality != cardinalityForCount(wantCount) {
			t.Fatalf("fuzz cardinality = (%d, %t), count = %d, want %d", cardinality, ok, match.MatchCount(), wantCount)
		}
		witness, one := match.SnapshotAuthorityMatch()
		if one != (wantCount == 1) || one != witness.Valid() {
			t.Fatalf("fuzz witness = (%#v, %t), count = %d", witness, one, wantCount)
		}
	})
}

func assertCardinality(
	t *testing.T,
	snapshot FiniteAuthoritySnapshot,
	binding AuthorityMatchBinding,
	want AuthorityMatchCardinality,
	wantCount int,
) {
	t.Helper()
	match, err := snapshot.Match(binding)
	if err != nil || !match.Valid() || match.MatchCount() != wantCount {
		t.Fatalf("match = (%#v, %v), count = %d", match, err, match.MatchCount())
	}
	cardinality, ok := match.Cardinality()
	if !ok || cardinality != want {
		t.Fatalf("cardinality = (%d, %t), want %d", cardinality, ok, want)
	}
}

func cardinalityForCount(count int) AuthorityMatchCardinality {
	switch count {
	case 0:
		return AuthorityMatchCardinalityNoMatchInSnapshot
	case 1:
		return AuthorityMatchCardinalityOneMatchInSnapshot
	default:
		return AuthorityMatchCardinalityMultipleMatchesInSnapshot
	}
}

func cardinalityToken(value AuthorityMatchCardinality) string {
	switch value {
	case AuthorityMatchCardinalityNoMatchInSnapshot:
		return "no_match_in_snapshot"
	case AuthorityMatchCardinalityOneMatchInSnapshot:
		return "one_match_in_snapshot"
	case AuthorityMatchCardinalityMultipleMatchesInSnapshot:
		return "multiple_matches_in_snapshot"
	default:
		return "invalid"
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
	catalogID, catalogRevision, authorityID, authorityRevision string,
) AuthorityMatchBinding {
	t.Helper()
	value, err := constructBinding(catalogID, catalogRevision, authorityID, authorityRevision)
	if err != nil {
		t.Fatalf("construct binding: %v", err)
	}
	return value
}

func mustRecord(
	t *testing.T,
	catalogID, catalogRevision, authorityID, authorityRevision string,
) DeclarationAuthorityRecord {
	t.Helper()
	value, err := constructRecord(catalogID, catalogRevision, authorityID, authorityRevision)
	if err != nil {
		t.Fatalf("construct record: %v", err)
	}
	return value
}

func constructBinding(
	catalogID, catalogRevision, authorityID, authorityRevision string,
) (AuthorityMatchBinding, error) {
	scope, err := catalogregistration.NewCatalogScopeRef(catalogID, catalogRevision)
	if err != nil {
		return AuthorityMatchBinding{}, err
	}
	authority, err := catalogregistration.NewDeclarationAuthorityRef(authorityID, authorityRevision)
	if err != nil {
		return AuthorityMatchBinding{}, err
	}
	return NewAuthorityMatchBinding(scope, authority)
}

func constructRecord(
	catalogID, catalogRevision, authorityID, authorityRevision string,
) (DeclarationAuthorityRecord, error) {
	binding, err := constructBinding(catalogID, catalogRevision, authorityID, authorityRevision)
	if err != nil {
		return DeclarationAuthorityRecord{}, err
	}
	return NewDeclarationAuthorityRecord(binding)
}

func readCorpusTSV(t *testing.T, name string, header []string) [][]string {
	t.Helper()
	path := filepath.Join(
		"..", "..", "testdata", "conformance", "authority-matching-v0alpha1", name,
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
