package authorityresolution

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

	"github.com/blisspixel/fartapp/internal/authoritymatching"
	"github.com/blisspixel/fartapp/internal/catalogregistration"
)

const (
	maximumCorpusBytes = 64 * 1024
	maximumCorpusRows  = 256
	maximumFieldBytes  = 128
)

func TestAuthorityResolutionCorpus(t *testing.T) {
	snapshotRows := readCorpusTSV(t, "snapshots.tsv", []string{
		"snapshot_id", "catalog_id", "catalog_revision", "expect",
	})
	recordRows := readCorpusTSV(t, "records.tsv", []string{
		"snapshot_id", "catalog_id", "catalog_revision", "authority_id", "authority_revision",
	})
	caseRows := readCorpusTSV(t, "cases.tsv", []string{
		"case_id", "snapshot_id", "catalog_id", "catalog_revision", "authority_id",
		"authority_revision", "expect", "match_count", "snapshot_count", "positive_witness",
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
		record, err := constructRecord(row[1], row[2], row[3], row[4])
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
			"invalid_authority_match", "match_scope_mismatch", "not_resolved_multiple_matches_in_snapshot",
			"not_resolved_no_match_in_snapshot", "resolved_one_match_in_snapshot",
		}, row[6]) {
			t.Fatalf("case expectation %q is unsupported", row[6])
		}
		wantCount, wantSnapshotCount, wantPositive := parseCaseExpectations(t, row)
		if row[6] == "invalid_authority_match" {
			for index := 1; index <= 5; index++ {
				if row[index] != "not_applicable" {
					t.Fatalf("invalid-match case %q field %d must be not_applicable, got %q", row[0], index+1, row[index])
				}
			}
			t.Run(row[0], func(t *testing.T) {
				resolution, err := ResolveInSnapshot(authoritymatching.AuthorityMatch{})
				if !errors.Is(err, ErrInvalidAuthorityMatch) || resolution.Valid() {
					t.Fatalf("invalid-match resolution = (%#v, %v)", resolution, err)
				}
			})
			continue
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
				if !errors.Is(err, authoritymatching.ErrMatchScopeMismatch) || match.Valid() {
					t.Fatalf("scope mismatch = (%#v, %v)", match, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("match: %v", err)
			}
			resolution, err := ResolveInSnapshot(match)
			if err != nil || !resolution.Valid() || resolution.Match() != match ||
				resolution.Binding() != binding || resolution.MatchCount() != wantCount ||
				resolution.SnapshotRecordCount() != wantSnapshotCount {
				t.Fatalf("resolution = (%#v, %v)", resolution, err)
			}
			outcome, ok := resolution.Outcome()
			if !ok || outcomeToken(outcome) != row[6] {
				t.Fatalf("outcome = (%d, %t), want %q", outcome, ok, row[6])
			}
			resolved, positive := resolution.ResolvedAuthority()
			if positive != wantPositive || resolved.Valid() != wantPositive {
				t.Fatalf("positive witness = (%#v, %t), want %t", resolved, positive, wantPositive)
			}
			if positive {
				if resolved.Resolution() != resolution || resolved.Binding() != binding ||
					resolved.Record().SelfBinding() != binding ||
					resolved.MatchWitness().Binding() != binding ||
					resolved.SnapshotRecordCount() != wantSnapshotCount {
					t.Fatalf("resolved authority = %#v", resolved)
				}
			}
		})
	}
}

func TestInvalidAndForgedResolutionValues(t *testing.T) {
	var zeroMatch authoritymatching.AuthorityMatch
	if value, err := ResolveInSnapshot(zeroMatch); !errors.Is(err, ErrInvalidAuthorityMatch) || value.Valid() {
		t.Fatalf("zero-match resolution = (%#v, %v)", value, err)
	}

	var zero SnapshotResolution
	if zero.Valid() || zero.Match().Valid() || zero.Binding().Valid() || zero.MatchCount() != 0 ||
		zero.SnapshotRecordCount() != 0 {
		t.Fatalf("zero resolution is observable as valid: %#v", zero)
	}
	if outcome, ok := zero.Outcome(); ok || outcome != OutcomeInvalid {
		t.Fatalf("zero outcome = (%d, %t)", outcome, ok)
	}
	if resolved, ok := zero.ResolvedAuthority(); ok || resolved.Valid() {
		t.Fatalf("zero resolved authority = (%#v, %t)", resolved, ok)
	}

	var zeroResolved SnapshotResolvedAuthority
	if zeroResolved.Valid() || zeroResolved.Resolution().Valid() ||
		zeroResolved.MatchWitness().Valid() || zeroResolved.Binding().Valid() ||
		zeroResolved.Record().Valid() || zeroResolved.SnapshotRecordCount() != 0 {
		t.Fatalf("zero positive witness is observable as valid: %#v", zeroResolved)
	}

	scope := mustScope(t, "catalog.scope", "r1")
	query := mustBinding(t, "catalog.scope", "r1", "authority.a", "r1")
	for name, records := range map[string][]authoritymatching.DeclarationAuthorityRecord{
		"no_match": nil,
		"multiple_matches": {
			mustRecord(t, "catalog.scope", "r1", "authority.a", "r1"),
			mustRecord(t, "catalog.scope", "r1", "authority.a", "r1"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			snapshot, err := authoritymatching.NewFiniteAuthoritySnapshot(scope, records)
			if err != nil {
				t.Fatalf("snapshot: %v", err)
			}
			match, err := snapshot.Match(query)
			if err != nil {
				t.Fatalf("match: %v", err)
			}
			resolution, err := ResolveInSnapshot(match)
			if err != nil {
				t.Fatalf("resolution: %v", err)
			}
			forged := SnapshotResolvedAuthority{resolution: resolution}
			if forged.Valid() || forged.MatchWitness().Valid() || forged.Record().Valid() {
				t.Fatalf("forged positive witness is valid: %#v", forged)
			}
		})
	}
}

func TestMaximumSnapshotAndTargetLocalMultiplicity(t *testing.T) {
	a := mustRecord(t, "catalog.scope", "r1", "authority.a", "r1")
	b := mustRecord(t, "catalog.scope", "r1", "authority.b", "r1")
	records := make([]authoritymatching.DeclarationAuthorityRecord, authoritymatching.MaximumAuthorityRecords)
	for index := range records {
		records[index] = a
	}
	records[len(records)-1] = b
	snapshot, err := authoritymatching.NewFiniteAuthoritySnapshot(a.SelfBinding().CatalogScope(), records)
	if err != nil {
		t.Fatalf("maximum snapshot: %v", err)
	}

	for _, test := range []struct {
		binding  authoritymatching.AuthorityMatchBinding
		outcome  Outcome
		count    int
		positive bool
	}{
		{a.SelfBinding(), OutcomeNotResolvedMultipleMatchesInSnapshot, len(records) - 1, false},
		{b.SelfBinding(), OutcomeResolvedOneMatchInSnapshot, 1, true},
		{mustBinding(t, "catalog.scope", "r1", "authority.c", "r1"), OutcomeNotResolvedNoMatchInSnapshot, 0, false},
	} {
		match, matchErr := snapshot.Match(test.binding)
		if matchErr != nil {
			t.Fatalf("match: %v", matchErr)
		}
		resolution, resolutionErr := ResolveInSnapshot(match)
		if resolutionErr != nil {
			t.Fatalf("resolve: %v", resolutionErr)
		}
		outcome, ok := resolution.Outcome()
		resolved, positive := resolution.ResolvedAuthority()
		if !ok || outcome != test.outcome || resolution.MatchCount() != test.count ||
			resolution.SnapshotRecordCount() != len(records) || positive != test.positive ||
			resolved.Valid() != test.positive {
			t.Fatalf("resolution = (%#v, %d, %t), positive = (%#v, %t)", resolution, outcome, ok, resolved, positive)
		}
	}
}

func FuzzResolveInSnapshot(f *testing.F) {
	for _, seed := range [][4]string{
		{"a", "b", "a", "r1"},
		{"same", "same", "same", "r1"},
		{"a", "a", "a", "r2"},
		{"é", "a", "b", "r1"},
	} {
		f.Add(seed[0], seed[1], seed[2], seed[3])
	}
	f.Fuzz(func(t *testing.T, firstID, secondID, queryID, queryRevision string) {
		first, firstErr := constructRecord("fuzz.scope", "r1", firstID, "r1")
		second, secondErr := constructRecord("fuzz.scope", "r1", secondID, "r1")
		if firstErr != nil || secondErr != nil {
			return
		}
		snapshot, err := authoritymatching.NewFiniteAuthoritySnapshot(
			first.SelfBinding().CatalogScope(),
			[]authoritymatching.DeclarationAuthorityRecord{first, second},
		)
		if err != nil {
			t.Fatalf("snapshot: %v", err)
		}
		query, queryErr := constructBinding("fuzz.scope", "r1", queryID, queryRevision)
		if queryErr != nil {
			return
		}
		match, matchErr := snapshot.Match(query)
		if matchErr != nil {
			t.Fatalf("match: %v", matchErr)
		}
		resolution, resolutionErr := ResolveInSnapshot(match)
		if resolutionErr != nil || !resolution.Valid() {
			t.Fatalf("resolution = (%#v, %v)", resolution, resolutionErr)
		}
		wantCount := 0
		if query == first.SelfBinding() {
			wantCount++
		}
		if query == second.SelfBinding() {
			wantCount++
		}
		outcome, ok := resolution.Outcome()
		resolved, positive := resolution.ResolvedAuthority()
		if !ok || outcome != outcomeForCount(wantCount) || resolution.MatchCount() != wantCount ||
			positive != (wantCount == 1) || resolved.Valid() != positive {
			t.Fatalf("resolution = (%#v, %d, %t), positive = (%#v, %t), want count %d", resolution, outcome, ok, resolved, positive, wantCount)
		}

		reversed, reverseErr := authoritymatching.NewFiniteAuthoritySnapshot(
			first.SelfBinding().CatalogScope(),
			[]authoritymatching.DeclarationAuthorityRecord{second, first},
		)
		if reverseErr != nil {
			t.Fatalf("reversed snapshot: %v", reverseErr)
		}
		reverseMatch, reverseMatchErr := reversed.Match(query)
		if reverseMatchErr != nil {
			t.Fatalf("reverse match: %v", reverseMatchErr)
		}
		reverseResolution, reverseResolutionErr := ResolveInSnapshot(reverseMatch)
		if reverseResolutionErr != nil {
			t.Fatalf("reverse resolution: %v", reverseResolutionErr)
		}
		reverseOutcome, reverseOK := reverseResolution.Outcome()
		if !reverseOK || reverseOutcome != outcome || reverseResolution.MatchCount() != wantCount {
			t.Fatalf("reverse resolution = (%#v, %d, %t)", reverseResolution, reverseOutcome, reverseOK)
		}
	})
}

func parseCaseExpectations(t *testing.T, row []string) (int, int, bool) {
	t.Helper()
	if row[6] == "invalid_authority_match" || row[6] == "match_scope_mismatch" {
		for index := 7; index <= 9; index++ {
			if row[index] != "not_applicable" {
				t.Fatalf("error case %q field %d must be not_applicable, got %q", row[0], index+1, row[index])
			}
		}
		return 0, 0, false
	}
	wantCount, countErr := strconv.Atoi(row[7])
	wantSnapshotCount, snapshotErr := strconv.Atoi(row[8])
	wantPositive, positiveErr := strconv.ParseBool(row[9])
	if countErr != nil || wantCount < 0 || wantCount > authoritymatching.MaximumAuthorityRecords {
		t.Fatalf("case %q has invalid match count %q", row[0], row[7])
	}
	if snapshotErr != nil || wantSnapshotCount < wantCount ||
		wantSnapshotCount > authoritymatching.MaximumAuthorityRecords {
		t.Fatalf("case %q has invalid snapshot count %q", row[0], row[8])
	}
	if positiveErr != nil || wantPositive != (wantCount == 1) {
		t.Fatalf("case %q has invalid positive witness %q", row[0], row[9])
	}
	return wantCount, wantSnapshotCount, wantPositive
}

func outcomeForCount(count int) Outcome {
	switch count {
	case 0:
		return OutcomeNotResolvedNoMatchInSnapshot
	case 1:
		return OutcomeResolvedOneMatchInSnapshot
	default:
		return OutcomeNotResolvedMultipleMatchesInSnapshot
	}
}

func outcomeToken(value Outcome) string {
	switch value {
	case OutcomeNotResolvedNoMatchInSnapshot:
		return "not_resolved_no_match_in_snapshot"
	case OutcomeResolvedOneMatchInSnapshot:
		return "resolved_one_match_in_snapshot"
	case OutcomeNotResolvedMultipleMatchesInSnapshot:
		return "not_resolved_multiple_matches_in_snapshot"
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
) authoritymatching.AuthorityMatchBinding {
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
) authoritymatching.DeclarationAuthorityRecord {
	t.Helper()
	value, err := constructRecord(catalogID, catalogRevision, authorityID, authorityRevision)
	if err != nil {
		t.Fatalf("construct record: %v", err)
	}
	return value
}

func constructBinding(
	catalogID, catalogRevision, authorityID, authorityRevision string,
) (authoritymatching.AuthorityMatchBinding, error) {
	scope, err := catalogregistration.NewCatalogScopeRef(catalogID, catalogRevision)
	if err != nil {
		return authoritymatching.AuthorityMatchBinding{}, err
	}
	authority, err := catalogregistration.NewDeclarationAuthorityRef(authorityID, authorityRevision)
	if err != nil {
		return authoritymatching.AuthorityMatchBinding{}, err
	}
	return authoritymatching.NewAuthorityMatchBinding(scope, authority)
}

func constructRecord(
	catalogID, catalogRevision, authorityID, authorityRevision string,
) (authoritymatching.DeclarationAuthorityRecord, error) {
	binding, err := constructBinding(catalogID, catalogRevision, authorityID, authorityRevision)
	if err != nil {
		return authoritymatching.DeclarationAuthorityRecord{}, err
	}
	return authoritymatching.NewDeclarationAuthorityRecord(binding)
}

func readCorpusTSV(t *testing.T, name string, header []string) [][]string {
	t.Helper()
	path := filepath.Join(
		"..", "..", "testdata", "conformance", "authority-resolution-v0alpha1", name,
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
