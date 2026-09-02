// Package authorityresolution defines an internal, non-wire profile candidate
// for snapshot-relative declaration-authority resolution. It makes no claim
// about external existence, identity, legitimacy, trust, attribution, biology,
// location, dimensionality, language, network access, or source-law time.
package authorityresolution

import (
	"errors"

	"github.com/blisspixel/fartapp/internal/authoritymatching"
)

var (
	// ErrInvalidAuthorityMatch reports missing or forged retained match evidence.
	ErrInvalidAuthorityMatch = errors.New("invalid retained authority match")
)

// Outcome is the profile-owned decision for one valid retained match. Every
// non-invalid outcome is qualified by the supplied finite snapshot.
type Outcome uint8

const (
	OutcomeInvalid Outcome = iota
	OutcomeNotResolvedNoMatchInSnapshot
	OutcomeResolvedOneMatchInSnapshot
	OutcomeNotResolvedMultipleMatchesInSnapshot
)

// SnapshotResolution retains the complete exact-match evidence from which its
// snapshot-relative outcome is derived. It stores no redundant outcome field.
type SnapshotResolution struct {
	match authoritymatching.AuthorityMatch
}

// ResolveInSnapshot applies the total three-way candidate rule to a valid
// retained authority match. It performs no lookup, dereference, or I/O.
func ResolveInSnapshot(
	match authoritymatching.AuthorityMatch,
) (SnapshotResolution, error) {
	if !match.Valid() {
		return SnapshotResolution{}, ErrInvalidAuthorityMatch
	}
	return SnapshotResolution{match: match}, nil
}

// Match returns the complete retained exact-match evidence.
func (value SnapshotResolution) Match() authoritymatching.AuthorityMatch {
	return value.match
}

// Binding returns the exact scope-plus-authority query binding.
func (value SnapshotResolution) Binding() authoritymatching.AuthorityMatchBinding {
	return value.match.Binding()
}

// MatchCount returns the number of exact record entries in the retained
// snapshot. It returns zero for an invalid resolution.
func (value SnapshotResolution) MatchCount() int {
	if !value.Valid() {
		return 0
	}
	return value.match.MatchCount()
}

// SnapshotRecordCount returns the complete retained snapshot size, including
// entries for other bindings and repeated entries.
func (value SnapshotResolution) SnapshotRecordCount() int {
	if !value.Valid() {
		return 0
	}
	return value.match.SnapshotRecordCount()
}

// Outcome returns the deterministic snapshot-relative decision.
func (value SnapshotResolution) Outcome() (Outcome, bool) {
	if !value.Valid() {
		return OutcomeInvalid, false
	}
	switch value.match.MatchCount() {
	case 0:
		return OutcomeNotResolvedNoMatchInSnapshot, true
	case 1:
		return OutcomeResolvedOneMatchInSnapshot, true
	default:
		return OutcomeNotResolvedMultipleMatchesInSnapshot, true
	}
}

// ResolvedAuthority returns a positive refinement only when exactly one record
// entry matches in the retained snapshot.
func (value SnapshotResolution) ResolvedAuthority() (SnapshotResolvedAuthority, bool) {
	outcome, ok := value.Outcome()
	if !ok || outcome != OutcomeResolvedOneMatchInSnapshot {
		return SnapshotResolvedAuthority{}, false
	}
	return SnapshotResolvedAuthority{resolution: value}, true
}

// Valid reports whether complete valid upstream match evidence is retained.
func (value SnapshotResolution) Valid() bool {
	return value.match.Valid()
}

// SnapshotResolvedAuthority is the positive one-match refinement. It denotes
// only resolution under this supplied-snapshot profile, never external status.
type SnapshotResolvedAuthority struct {
	resolution SnapshotResolution
}

// Resolution returns the complete retained snapshot-resolution witness.
func (value SnapshotResolvedAuthority) Resolution() SnapshotResolution {
	return value.resolution
}

// MatchWitness returns the upstream positive exact-match witness.
func (value SnapshotResolvedAuthority) MatchWitness() authoritymatching.SnapshotAuthorityMatch {
	witness, ok := value.resolution.match.SnapshotAuthorityMatch()
	if !ok {
		return authoritymatching.SnapshotAuthorityMatch{}
	}
	return witness
}

// Binding returns the exact resolved scope-plus-authority binding.
func (value SnapshotResolvedAuthority) Binding() authoritymatching.AuthorityMatchBinding {
	return value.resolution.Binding()
}

// Record returns the single exact matching record in the retained snapshot.
func (value SnapshotResolvedAuthority) Record() authoritymatching.DeclarationAuthorityRecord {
	return value.MatchWitness().Record()
}

// SnapshotRecordCount returns the complete retained snapshot size.
func (value SnapshotResolvedAuthority) SnapshotRecordCount() int {
	return value.resolution.SnapshotRecordCount()
}

// Valid reports whether exactly one record entry still supports the retained
// snapshot-relative resolution.
func (value SnapshotResolvedAuthority) Valid() bool {
	if !value.resolution.Valid() {
		return false
	}
	outcome, ok := value.resolution.Outcome()
	if !ok || outcome != OutcomeResolvedOneMatchInSnapshot {
		return false
	}
	witness := value.MatchWitness()
	return witness.Valid() && witness.Binding() == value.resolution.Binding() &&
		witness.Record().SelfBinding() == value.resolution.Binding()
}
