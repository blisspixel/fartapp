// Package cataloglookup defines an internal, non-wire experiment for bounded
// catalog membership decisions. A SnapshotLookup records only that membership
// was decided against every registration stored in one exact FiniteSnapshot.
// It does not establish fully supplied external content, authority-reference
// resolution, or declaration-attribution scope.
package cataloglookup

import (
	"errors"

	"github.com/blisspixel/fartapp/internal/catalogregistration"
)

// MaximumRegistrations bounds the number of positive registrations retained by
// one finite in-process snapshot.
const MaximumRegistrations = 4096

var (
	// ErrInvalidSnapshot reports invalid snapshot identity or member input.
	ErrInvalidSnapshot = errors.New("invalid finite catalog snapshot")
	// ErrSnapshotLimit reports a snapshot larger than MaximumRegistrations.
	ErrSnapshotLimit = errors.New("finite catalog snapshot exceeds registration limit")
	// ErrDuplicateRegistration reports a repeated exact registration binding.
	ErrDuplicateRegistration = errors.New("duplicate catalog registration")
	// ErrScopeMismatch reports a member or query bound to another catalog scope.
	ErrScopeMismatch = errors.New("catalog lookup scope mismatch")
)

type snapshotCore struct {
	scope   catalogregistration.CatalogScopeRef
	members map[catalogregistration.Binding]struct{}
}

// FiniteSnapshot is an immutable, bounded set of positive structural
// registrations supplied by its caller. Closure is relative to this set only.
type FiniteSnapshot struct {
	core *snapshotCore
}

// NewFiniteSnapshot copies registrations into a private membership set.
func NewFiniteSnapshot(
	scope catalogregistration.CatalogScopeRef,
	registrations []catalogregistration.Registration,
) (FiniteSnapshot, error) {
	if !scope.Valid() {
		return FiniteSnapshot{}, ErrInvalidSnapshot
	}
	if len(registrations) > MaximumRegistrations {
		return FiniteSnapshot{}, ErrSnapshotLimit
	}
	members := make(map[catalogregistration.Binding]struct{}, len(registrations))
	for _, registration := range registrations {
		if !registration.Valid() {
			return FiniteSnapshot{}, ErrInvalidSnapshot
		}
		binding := registration.Binding()
		if binding.CatalogScope() != scope {
			return FiniteSnapshot{}, ErrScopeMismatch
		}
		if _, exists := members[binding]; exists {
			return FiniteSnapshot{}, ErrDuplicateRegistration
		}
		members[binding] = struct{}{}
	}
	return FiniteSnapshot{core: &snapshotCore{scope: scope, members: members}}, nil
}

// Scope returns the exact catalog scope named by the snapshot.
func (value FiniteSnapshot) Scope() catalogregistration.CatalogScopeRef {
	if value.core == nil {
		return catalogregistration.CatalogScopeRef{}
	}
	return value.core.scope
}

// RegistrationCount returns the number of distinct retained registrations.
func (value FiniteSnapshot) RegistrationCount() int {
	if value.core == nil {
		return 0
	}
	return len(value.core.members)
}

// Valid reports whether the snapshot satisfies its structural invariants.
func (value FiniteSnapshot) Valid() bool {
	if value.core == nil || value.core.members == nil || !value.core.scope.Valid() ||
		len(value.core.members) > MaximumRegistrations {
		return false
	}
	for binding := range value.core.members {
		if !binding.Valid() || binding.CatalogScope() != value.core.scope {
			return false
		}
	}
	return true
}

// Lookup decides exact binding membership in this finite snapshot. A binding
// for another scope is an error and never becomes a negative result.
func (value FiniteSnapshot) Lookup(
	binding catalogregistration.Binding,
) (SnapshotLookup, error) {
	if !value.Valid() {
		return SnapshotLookup{}, ErrInvalidSnapshot
	}
	if !binding.Valid() {
		return SnapshotLookup{}, catalogregistration.ErrInvalidBinding
	}
	if binding.CatalogScope() != value.core.scope {
		return SnapshotLookup{}, ErrScopeMismatch
	}
	_, registered := value.core.members[binding]
	var (
		result catalogregistration.Result
		err    error
	)
	if registered {
		result, err = catalogregistration.EvaluatedRegistered(binding)
	} else {
		result, err = catalogregistration.EvaluatedNotRegistered(binding)
	}
	if err != nil {
		return SnapshotLookup{}, err
	}
	return SnapshotLookup{snapshot: value, result: result}, nil
}

// SnapshotLookup retains both a structural registration result and the exact
// finite snapshot against which it was decided.
type SnapshotLookup struct {
	snapshot FiniteSnapshot
	result   catalogregistration.Result
}

// SnapshotScope returns the scope of the retained finite snapshot.
func (value SnapshotLookup) SnapshotScope() catalogregistration.CatalogScopeRef {
	return value.snapshot.Scope()
}

// RegistrationCount returns the size of the retained finite snapshot.
func (value SnapshotLookup) RegistrationCount() int {
	return value.snapshot.RegistrationCount()
}

// Binding returns the exact queried registration binding.
func (value SnapshotLookup) Binding() catalogregistration.Binding {
	return value.result.Binding()
}

// Presence returns the evaluated structural presence outcome for a valid
// snapshot lookup.
func (value SnapshotLookup) Presence() (catalogregistration.Presence, bool) {
	if !value.Valid() {
		return catalogregistration.PresenceInvalid, false
	}
	return value.result.Presence()
}

// Registration returns a positive structural registration only when the exact
// binding is a member of the retained snapshot.
func (value SnapshotLookup) Registration() (catalogregistration.Registration, bool) {
	if !value.Valid() {
		return catalogregistration.Registration{}, false
	}
	return value.result.Registration()
}

// Valid reports whether the retained result exactly matches membership in the
// retained finite snapshot.
func (value SnapshotLookup) Valid() bool {
	if !value.snapshot.Valid() || !value.result.Valid() {
		return false
	}
	binding := value.result.Binding()
	if binding.CatalogScope() != value.snapshot.core.scope {
		return false
	}
	_, member := value.snapshot.core.members[binding]
	presence, ok := value.result.Presence()
	if !ok {
		return false
	}
	if member {
		return presence == catalogregistration.PresenceRegistered
	}
	return presence == catalogregistration.PresenceNotRegistered
}
