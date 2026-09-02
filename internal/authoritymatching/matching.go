// Package authoritymatching defines an internal, non-wire experiment for exact
// authority-reference match cardinality inside a supplied finite snapshot. It
// makes no claim about legitimacy, trust, authorship, personhood, network
// access, source-law time, or external catalog completeness.
package authoritymatching

import (
	"errors"

	"github.com/blisspixel/fartapp/internal/catalogregistration"
)

// MaximumAuthorityRecords bounds one finite in-process authority snapshot.
const MaximumAuthorityRecords = 4096

var (
	// ErrInvalidMatchBinding reports an invalid scope-plus-authority binding.
	ErrInvalidMatchBinding = errors.New("invalid authority match binding")
	// ErrInvalidRecord reports an invalid declaration-authority record.
	ErrInvalidRecord = errors.New("invalid declaration-authority record")
	// ErrInvalidSnapshot reports an invalid finite authority snapshot.
	ErrInvalidSnapshot = errors.New("invalid finite authority snapshot")
	// ErrSnapshotLimit reports a snapshot larger than MaximumAuthorityRecords.
	ErrSnapshotLimit = errors.New("finite authority snapshot exceeds record limit")
	// ErrRecordScopeMismatch reports a record outside its snapshot scope.
	ErrRecordScopeMismatch = errors.New("authority record scope mismatch")
	// ErrMatchScopeMismatch reports a match binding outside its snapshot scope.
	ErrMatchScopeMismatch = errors.New("authority match scope mismatch")
)

// AuthorityMatchBinding identifies one exact authority reference inside one
// exact catalog scope. Revision fields are identity material, not chronology.
type AuthorityMatchBinding struct {
	catalogScope catalogregistration.CatalogScopeRef
	authority    catalogregistration.DeclarationAuthorityRef
}

// NewAuthorityMatchBinding constructs an exact scope-plus-authority binding.
func NewAuthorityMatchBinding(
	catalogScope catalogregistration.CatalogScopeRef,
	authority catalogregistration.DeclarationAuthorityRef,
) (AuthorityMatchBinding, error) {
	value := AuthorityMatchBinding{catalogScope: catalogScope, authority: authority}
	if !value.Valid() {
		return AuthorityMatchBinding{}, ErrInvalidMatchBinding
	}
	return value, nil
}

// CatalogScope returns the exact catalog scope in the binding.
func (value AuthorityMatchBinding) CatalogScope() catalogregistration.CatalogScopeRef {
	return value.catalogScope
}

// Authority returns the exact opaque authority reference in the binding.
func (value AuthorityMatchBinding) Authority() catalogregistration.DeclarationAuthorityRef {
	return value.authority
}

// Valid reports whether both exact references are valid.
func (value AuthorityMatchBinding) Valid() bool {
	return value.catalogScope.Valid() && value.authority.Valid()
}

// DeclarationAuthorityRecord is one opaque record with an exact self-binding.
// It contains no authority category, presentation, endpoint, or relationship.
type DeclarationAuthorityRecord struct {
	self  AuthorityMatchBinding
	valid bool
}

// NewDeclarationAuthorityRecord constructs a record with one exact self-binding.
func NewDeclarationAuthorityRecord(
	self AuthorityMatchBinding,
) (DeclarationAuthorityRecord, error) {
	if !self.Valid() {
		return DeclarationAuthorityRecord{}, ErrInvalidRecord
	}
	return DeclarationAuthorityRecord{self: self, valid: true}, nil
}

// SelfBinding returns the record's exact self-binding.
func (value DeclarationAuthorityRecord) SelfBinding() AuthorityMatchBinding {
	return value.self
}

// Valid reports whether the record has a valid exact self-binding.
func (value DeclarationAuthorityRecord) Valid() bool {
	return value.valid && value.self.Valid()
}

type snapshotCore struct {
	catalogScope catalogregistration.CatalogScopeRef
	// A record currently contains only its self-binding, so the binding is the
	// complete record identity. Any future record field must be retained here or
	// included in that identity before this representation remains sound.
	multiplicity map[AuthorityMatchBinding]int
	recordCount  int
}

// FiniteAuthoritySnapshot is an immutable, bounded multiset of supplied
// declaration-authority records. Completeness is relative only to this multiset.
type FiniteAuthoritySnapshot struct {
	core *snapshotCore
}

// NewFiniteAuthoritySnapshot defensively copies record multiplicities under
// one exact catalog scope. Duplicate records are retained, not deduplicated.
func NewFiniteAuthoritySnapshot(
	catalogScope catalogregistration.CatalogScopeRef,
	records []DeclarationAuthorityRecord,
) (FiniteAuthoritySnapshot, error) {
	if !catalogScope.Valid() {
		return FiniteAuthoritySnapshot{}, ErrInvalidSnapshot
	}
	if len(records) > MaximumAuthorityRecords {
		return FiniteAuthoritySnapshot{}, ErrSnapshotLimit
	}
	multiplicity := make(map[AuthorityMatchBinding]int, len(records))
	for _, record := range records {
		if !record.Valid() {
			return FiniteAuthoritySnapshot{}, ErrInvalidRecord
		}
		binding := record.SelfBinding()
		if binding.CatalogScope() != catalogScope {
			return FiniteAuthoritySnapshot{}, ErrRecordScopeMismatch
		}
		multiplicity[binding]++
	}
	return FiniteAuthoritySnapshot{core: &snapshotCore{
		catalogScope: catalogScope,
		multiplicity: multiplicity,
		recordCount:  len(records),
	}}, nil
}

// CatalogScope returns the exact scope named by the snapshot.
func (value FiniteAuthoritySnapshot) CatalogScope() catalogregistration.CatalogScopeRef {
	if value.core == nil {
		return catalogregistration.CatalogScopeRef{}
	}
	return value.core.catalogScope
}

// RecordCount returns the total record count, including duplicates.
func (value FiniteAuthoritySnapshot) RecordCount() int {
	if value.core == nil {
		return 0
	}
	return value.core.recordCount
}

// DistinctBindingCount returns the number of distinct exact self-bindings.
func (value FiniteAuthoritySnapshot) DistinctBindingCount() int {
	if value.core == nil {
		return 0
	}
	return len(value.core.multiplicity)
}

// Valid reports whether the finite multiset satisfies its invariants.
func (value FiniteAuthoritySnapshot) Valid() bool {
	if value.core == nil || value.core.multiplicity == nil ||
		!value.core.catalogScope.Valid() || value.core.recordCount < 0 ||
		value.core.recordCount > MaximumAuthorityRecords {
		return false
	}
	total := 0
	for binding, count := range value.core.multiplicity {
		if !binding.Valid() || binding.CatalogScope() != value.core.catalogScope || count <= 0 {
			return false
		}
		if count > MaximumAuthorityRecords-total {
			return false
		}
		total += count
	}
	return total == value.core.recordCount
}

// Match counts exact records for one same-scope binding. A cross-scope binding
// is an error and never becomes a cardinality outcome.
func (value FiniteAuthoritySnapshot) Match(
	binding AuthorityMatchBinding,
) (AuthorityMatch, error) {
	if !value.Valid() {
		return AuthorityMatch{}, ErrInvalidSnapshot
	}
	if !binding.Valid() {
		return AuthorityMatch{}, ErrInvalidMatchBinding
	}
	if binding.CatalogScope() != value.core.catalogScope {
		return AuthorityMatch{}, ErrMatchScopeMismatch
	}
	return AuthorityMatch{snapshot: value, binding: binding}, nil
}

// AuthorityMatchCardinality is an exact multiplicity class inside one retained
// finite authority snapshot.
type AuthorityMatchCardinality uint8

const (
	AuthorityMatchCardinalityInvalid AuthorityMatchCardinality = iota
	AuthorityMatchCardinalityNoMatchInSnapshot
	AuthorityMatchCardinalityOneMatchInSnapshot
	AuthorityMatchCardinalityMultipleMatchesInSnapshot
)

// AuthorityMatch retains the full finite snapshot and exact match binding.
type AuthorityMatch struct {
	snapshot FiniteAuthoritySnapshot
	binding  AuthorityMatchBinding
}

// SnapshotScope returns the exact retained snapshot scope.
func (value AuthorityMatch) SnapshotScope() catalogregistration.CatalogScopeRef {
	return value.snapshot.CatalogScope()
}

// Binding returns the exact match binding.
func (value AuthorityMatch) Binding() AuthorityMatchBinding { return value.binding }

// SnapshotRecordCount returns the retained multiset size, including duplicates.
func (value AuthorityMatch) SnapshotRecordCount() int { return value.snapshot.RecordCount() }

// MatchCount returns the number of exact matching records.
func (value AuthorityMatch) MatchCount() int {
	if !value.Valid() {
		return 0
	}
	return value.snapshot.core.multiplicity[value.binding]
}

// Cardinality returns the exact match-count class for a valid match operation.
func (value AuthorityMatch) Cardinality() (AuthorityMatchCardinality, bool) {
	if !value.Valid() {
		return AuthorityMatchCardinalityInvalid, false
	}
	switch value.snapshot.core.multiplicity[value.binding] {
	case 0:
		return AuthorityMatchCardinalityNoMatchInSnapshot, true
	case 1:
		return AuthorityMatchCardinalityOneMatchInSnapshot, true
	default:
		return AuthorityMatchCardinalityMultipleMatchesInSnapshot, true
	}
}

// SnapshotAuthorityMatch returns a positive uniqueness witness only for one
// exact match in the retained snapshot.
func (value AuthorityMatch) SnapshotAuthorityMatch() (SnapshotAuthorityMatch, bool) {
	cardinality, ok := value.Cardinality()
	if !ok || cardinality != AuthorityMatchCardinalityOneMatchInSnapshot {
		return SnapshotAuthorityMatch{}, false
	}
	return SnapshotAuthorityMatch{match: value}, true
}

// Valid reports whether the match binding belongs to the retained snapshot.
func (value AuthorityMatch) Valid() bool {
	return value.snapshot.Valid() && value.binding.Valid() &&
		value.binding.CatalogScope() == value.snapshot.core.catalogScope
}

// SnapshotAuthorityMatch retains the complete one-match snapshot witness.
type SnapshotAuthorityMatch struct {
	match AuthorityMatch
}

// Binding returns the exact uniquely matched binding.
func (value SnapshotAuthorityMatch) Binding() AuthorityMatchBinding {
	return value.match.Binding()
}

// Record returns the unique opaque record with exact self-reference.
func (value SnapshotAuthorityMatch) Record() DeclarationAuthorityRecord {
	if !value.Valid() {
		return DeclarationAuthorityRecord{}
	}
	return DeclarationAuthorityRecord{self: value.match.binding, valid: true}
}

// SnapshotRecordCount returns the retained witness multiset size.
func (value SnapshotAuthorityMatch) SnapshotRecordCount() int {
	return value.match.SnapshotRecordCount()
}

// Valid reports whether the retained snapshot still contains exactly one match
// and the record self-binding equals the requested match binding.
func (value SnapshotAuthorityMatch) Valid() bool {
	if !value.match.Valid() || value.match.MatchCount() != 1 {
		return false
	}
	record := DeclarationAuthorityRecord{self: value.match.binding, valid: true}
	return record.Valid() && record.SelfBinding() == value.match.binding
}
