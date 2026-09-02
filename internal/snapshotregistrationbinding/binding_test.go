package snapshotregistrationbinding

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

	"github.com/blisspixel/fartapp/internal/authoritymatching"
	"github.com/blisspixel/fartapp/internal/authorityresolution"
	"github.com/blisspixel/fartapp/internal/cataloglookup"
	"github.com/blisspixel/fartapp/internal/catalogregistration"
	"github.com/blisspixel/fartapp/internal/registrationauthoritybinding"
)

const (
	maximumCorpusBytes = 64 * 1024
	maximumCorpusRows  = 256
	maximumFieldBytes  = 128
)

func TestSnapshotRegistrationBindingCorpus(t *testing.T) {
	rows := readCorpusTSV(t, "cases.tsv", []string{
		"case_id", "lookup_subject_id", "lookup_subject_revision", "lookup_capability",
		"snapshot_state", "binding_catalog_id", "binding_catalog_revision",
		"binding_authority_id", "binding_authority_revision", "binding_subject_id",
		"binding_subject_revision", "binding_capability", "expect", "positive_witness",
	})
	rule := registrationauthoritybinding.ExactAuthorityBindingV0Alpha1()
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		if row[0] == "" {
			t.Fatal("case id is empty")
		}
		if _, exists := seen[row[0]]; exists {
			t.Fatalf("case id %q is duplicated", row[0])
		}
		seen[row[0]] = struct{}{}
		if !slices.Contains([]string{"absent", "present"}, row[4]) {
			t.Fatalf("case %q snapshot state %q is unsupported", row[0], row[4])
		}
		if !slices.Contains([]string{
			"positive_composition", "registration_mismatch", "snapshot_registration_absent",
		}, row[12]) {
			t.Fatalf("case %q expectation %q is unsupported", row[0], row[12])
		}
		wantPositive := false
		if row[12] == "positive_composition" {
			if row[13] != "true" {
				t.Fatalf("positive case %q witness = %q", row[0], row[13])
			}
			wantPositive = true
		} else if row[13] != "not_applicable" {
			t.Fatalf("error case %q must not encode a witness, got %q", row[0], row[13])
		}

		t.Run(row[0], func(t *testing.T) {
			lookupRegistration := mustRegistration(t, row[1], row[2], row[3])
			bindingRegistration := mustRegistrationFull(
				t, row[5], row[6], row[7], row[8], row[9], row[10], row[11],
			)
			lookup := mustLookup(t, lookupRegistration, row[4] == "present", nil)
			bindingAuthority := mustResolvedAuthorityFull(t, row[5], row[6], row[7], row[8])
			binding := mustExactBindingWitness(t, bindingAuthority, bindingRegistration, rule)
			composed, err := ComposePositive(lookup, binding)
			switch row[12] {
			case "registration_mismatch":
				if !errors.Is(err, ErrRegistrationMismatch) || composed.Valid() {
					t.Fatalf("registration mismatch = (%#v, %v)", composed, err)
				}
				return
			case "snapshot_registration_absent":
				if !errors.Is(err, ErrSnapshotRegistrationAbsent) || composed.Valid() {
					t.Fatalf("snapshot absence = (%#v, %v)", composed, err)
				}
				return
			}
			if err != nil || !composed.Valid() || !wantPositive ||
				composed.SnapshotLookup() != lookup || composed.ExactBindingWitness() != binding ||
				composed.Registration() != lookupRegistration {
				t.Fatalf("positive composition = (%#v, %v)", composed, err)
			}
		})
	}
}

func TestInvalidAndForgedCompositionValues(t *testing.T) {
	authority := mustResolvedAuthority(t)
	registration := mustRegistration(t, "subject.a", "r1", "capability.a")
	lookup := mustLookup(t, registration, true, nil)
	binding := mustExactBindingWitness(
		t, authority, registration, registrationauthoritybinding.ExactAuthorityBindingV0Alpha1(),
	)

	if value, err := ComposePositive(cataloglookup.SnapshotLookup{}, binding); !errors.Is(err, ErrInvalidSnapshotLookup) || value.Valid() {
		t.Fatalf("invalid lookup composition = (%#v, %v)", value, err)
	}
	if value, err := ComposePositive(lookup, registrationauthoritybinding.ExactBindingWitness{}); !errors.Is(err, ErrInvalidExactBindingWitness) || value.Valid() {
		t.Fatalf("invalid binding composition = (%#v, %v)", value, err)
	}

	var zero SnapshotMemberExactBindingWitness
	if zero.Valid() || zero.SnapshotLookup().Valid() || zero.ExactBindingWitness().Valid() ||
		zero.Registration().Valid() {
		t.Fatalf("zero composition is observable as valid: %#v", zero)
	}

	other := mustRegistration(t, "subject.other", "r1", "capability.a")
	otherBinding := mustExactBindingWitness(
		t, authority, other, registrationauthoritybinding.ExactAuthorityBindingV0Alpha1(),
	)
	forged := SnapshotMemberExactBindingWitness{lookup: lookup, binding: otherBinding}
	if forged.Valid() || forged.Registration() != registration {
		t.Fatalf("mismatched forged composition = %#v", forged)
	}

	differentAuthorityRegistration := mustRegistrationForAuthority(
		t, "authority.b", "r1", "subject.a", "r1", "capability.a",
	)
	decision, err := registrationauthoritybinding.Compare(
		registrationauthoritybinding.ExactAuthorityBindingV0Alpha1(),
		authority,
		differentAuthorityRegistration,
	)
	if err != nil {
		t.Fatalf("no-match binding decision: %v", err)
	}
	if witness, ok := decision.ExactBindingWitness(); ok || witness.Valid() {
		t.Fatalf("pairwise no-match produced positive witness: (%#v, %t)", witness, ok)
	}
}

func TestMaximumSnapshotWitnessRetention(t *testing.T) {
	authority := mustResolvedAuthority(t)
	target := mustRegistration(t, "subject.target", "r1", "capability.target")
	registrations := make([]catalogregistration.Registration, cataloglookup.MaximumRegistrations)
	registrations[0] = target
	for index := 1; index < len(registrations); index++ {
		registrations[index] = mustRegistration(
			t,
			fmt.Sprintf("subject.%04d", index),
			"r1",
			"capability.other",
		)
	}
	lookup := mustLookup(t, target, true, registrations)
	binding := mustExactBindingWitness(
		t, authority, target, registrationauthoritybinding.ExactAuthorityBindingV0Alpha1(),
	)
	witness, err := ComposePositive(lookup, binding)
	if err != nil || !witness.Valid() ||
		witness.SnapshotLookup().RegistrationCount() != cataloglookup.MaximumRegistrations ||
		witness.ExactBindingWitness().ResolvedAuthority() != authority ||
		witness.Registration() != target {
		t.Fatalf("maximum composition = (%#v, %v)", witness, err)
	}
}

func FuzzComposePositive(f *testing.F) {
	for _, seed := range [][11]string{
		{"subject", "r1", "capability", "catalog.scope", "r1", "authority.a", "r1", "subject", "r1", "capability", "present"},
		{"subject", "r1", "capability", "other.scope", "r1", "authority.a", "r1", "subject", "r1", "capability", "present"},
		{"subject", "r1", "capability", "catalog.scope", "r1", "authority.b", "r1", "subject", "r1", "capability", "present"},
		{"subject.a", "r1", "capability", "catalog.scope", "r1", "authority.a", "r1", "subject.b", "r1", "capability", "present"},
		{"subject", "r1", "capability", "catalog.scope", "r1", "authority.a", "r1", "subject", "r1", "capability", "absent"},
	} {
		f.Add(
			seed[0], seed[1], seed[2], seed[3], seed[4], seed[5],
			seed[6], seed[7], seed[8], seed[9], seed[10],
		)
	}
	rule := registrationauthoritybinding.ExactAuthorityBindingV0Alpha1()
	f.Fuzz(func(
		t *testing.T,
		lookupSubject, lookupRevision, lookupCapability,
		bindingCatalogID, bindingCatalogRevision,
		bindingAuthorityID, bindingAuthorityRevision,
		bindingSubject, bindingRevision, bindingCapability, snapshotState string,
	) {
		lookupRegistration, lookupErr := constructRegistration(
			"authority.a", "r1", lookupSubject, lookupRevision, lookupCapability,
		)
		bindingRegistration, bindingErr := constructRegistrationFull(
			bindingCatalogID, bindingCatalogRevision,
			bindingAuthorityID, bindingAuthorityRevision,
			bindingSubject, bindingRevision, bindingCapability,
		)
		if lookupErr != nil || bindingErr != nil ||
			(snapshotState != "present" && snapshotState != "absent") {
			return
		}
		lookup := mustLookup(t, lookupRegistration, snapshotState == "present", nil)
		bindingAuthority := mustResolvedAuthorityFull(
			t, bindingCatalogID, bindingCatalogRevision,
			bindingAuthorityID, bindingAuthorityRevision,
		)
		binding := mustExactBindingWitness(t, bindingAuthority, bindingRegistration, rule)
		witness, err := ComposePositive(lookup, binding)
		if snapshotState == "absent" {
			if !errors.Is(err, ErrSnapshotRegistrationAbsent) || witness.Valid() {
				t.Fatalf("absent composition = (%#v, %v)", witness, err)
			}
			return
		}
		if lookupRegistration != bindingRegistration {
			if !errors.Is(err, ErrRegistrationMismatch) || witness.Valid() {
				t.Fatalf("mismatch composition = (%#v, %v)", witness, err)
			}
			return
		}
		if err != nil || !witness.Valid() || witness.Registration() != lookupRegistration {
			t.Fatalf("positive composition = (%#v, %v)", witness, err)
		}
	})
}

type testHelper interface {
	Helper()
	Fatal(args ...any)
	Fatalf(format string, args ...any)
}

func mustResolvedAuthority(t testHelper) authorityresolution.SnapshotResolvedAuthority {
	t.Helper()
	return mustResolvedAuthorityFull(t, "catalog.scope", "r1", "authority.a", "r1")
}

func mustResolvedAuthorityFull(
	t testHelper,
	catalogID, catalogRevision, authorityID, authorityRevision string,
) authorityresolution.SnapshotResolvedAuthority {
	t.Helper()
	scope, err := catalogregistration.NewCatalogScopeRef(catalogID, catalogRevision)
	if err != nil {
		t.Fatalf("construct scope: %v", err)
	}
	authority, err := catalogregistration.NewDeclarationAuthorityRef(authorityID, authorityRevision)
	if err != nil {
		t.Fatalf("construct authority: %v", err)
	}
	binding, err := authoritymatching.NewAuthorityMatchBinding(scope, authority)
	if err != nil {
		t.Fatalf("construct authority binding: %v", err)
	}
	record, err := authoritymatching.NewDeclarationAuthorityRecord(binding)
	if err != nil {
		t.Fatalf("construct authority record: %v", err)
	}
	snapshot, err := authoritymatching.NewFiniteAuthoritySnapshot(
		scope, []authoritymatching.DeclarationAuthorityRecord{record},
	)
	if err != nil {
		t.Fatalf("construct authority snapshot: %v", err)
	}
	match, err := snapshot.Match(binding)
	if err != nil {
		t.Fatalf("match authority: %v", err)
	}
	resolution, err := authorityresolution.ResolveInSnapshot(match)
	if err != nil {
		t.Fatalf("resolve authority: %v", err)
	}
	resolved, ok := resolution.ResolvedAuthority()
	if !ok {
		t.Fatal("one-record authority did not resolve")
	}
	return resolved
}

func mustRegistration(
	t testHelper,
	subjectID, subjectRevision, capability string,
) catalogregistration.Registration {
	t.Helper()
	return mustRegistrationForAuthority(
		t, "authority.a", "r1", subjectID, subjectRevision, capability,
	)
}

func mustRegistrationForAuthority(
	t testHelper,
	authorityID, authorityRevision, subjectID, subjectRevision, capability string,
) catalogregistration.Registration {
	t.Helper()
	registration, err := constructRegistration(
		authorityID, authorityRevision, subjectID, subjectRevision, capability,
	)
	if err != nil {
		t.Fatalf("construct registration: %v", err)
	}
	return registration
}

func mustRegistrationFull(
	t testHelper,
	catalogID, catalogRevision, authorityID, authorityRevision,
	subjectID, subjectRevision, capability string,
) catalogregistration.Registration {
	t.Helper()
	registration, err := constructRegistrationFull(
		catalogID, catalogRevision, authorityID, authorityRevision,
		subjectID, subjectRevision, capability,
	)
	if err != nil {
		t.Fatalf("construct full registration: %v", err)
	}
	return registration
}

func constructRegistration(
	authorityID, authorityRevision, subjectID, subjectRevision, capability string,
) (catalogregistration.Registration, error) {
	return constructRegistrationFull(
		"catalog.scope", "r1", authorityID, authorityRevision,
		subjectID, subjectRevision, capability,
	)
}

func constructRegistrationFull(
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

func mustLookup(
	t testHelper,
	query catalogregistration.Registration,
	present bool,
	members []catalogregistration.Registration,
) cataloglookup.SnapshotLookup {
	t.Helper()
	if members == nil && present {
		members = []catalogregistration.Registration{query}
	}
	snapshot, err := cataloglookup.NewFiniteSnapshot(query.Binding().CatalogScope(), members)
	if err != nil {
		t.Fatalf("construct registration snapshot: %v", err)
	}
	lookup, err := snapshot.Lookup(query.Binding())
	if err != nil {
		t.Fatalf("lookup registration: %v", err)
	}
	return lookup
}

func mustExactBindingWitness(
	t testHelper,
	authority authorityresolution.SnapshotResolvedAuthority,
	registration catalogregistration.Registration,
	rule registrationauthoritybinding.Rule,
) registrationauthoritybinding.ExactBindingWitness {
	t.Helper()
	decision, err := registrationauthoritybinding.Compare(rule, authority, registration)
	if err != nil {
		t.Fatalf("compare authority binding: %v", err)
	}
	witness, ok := decision.ExactBindingWitness()
	if !ok {
		t.Fatal("exact authority reference did not produce a binding witness")
	}
	return witness
}

func readCorpusTSV(t *testing.T, name string, header []string) [][]string {
	t.Helper()
	path := filepath.Join(
		"..", "..", "testdata", "conformance", "snapshot-registration-binding-v0alpha1", name,
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
