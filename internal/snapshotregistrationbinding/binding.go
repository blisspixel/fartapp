// Package snapshotregistrationbinding composes two existing positive internal
// witnesses for the same structural registration. It creates no new lookup or
// authority-comparison outcome and defines no attribution, provenance, or
// catalog-closure decision.
package snapshotregistrationbinding

import (
	"errors"

	"github.com/blisspixel/fartapp/internal/cataloglookup"
	"github.com/blisspixel/fartapp/internal/catalogregistration"
	"github.com/blisspixel/fartapp/internal/registrationauthoritybinding"
)

var (
	// ErrInvalidSnapshotLookup reports missing or forged lookup evidence.
	ErrInvalidSnapshotLookup = errors.New("invalid snapshot lookup")
	// ErrSnapshotRegistrationAbsent reports a valid negative snapshot lookup.
	ErrSnapshotRegistrationAbsent = errors.New("snapshot lookup has no positive registration")
	// ErrInvalidExactBindingWitness reports missing or forged exact-binding evidence.
	ErrInvalidExactBindingWitness = errors.New("invalid exact registration-authority binding witness")
	// ErrRegistrationMismatch reports positive witnesses for different registrations.
	ErrRegistrationMismatch = errors.New("positive witnesses retain different registrations")
)

// SnapshotMemberExactBindingWitness retains positive finite-snapshot membership
// and positive exact authority-binding evidence for the same registration.
type SnapshotMemberExactBindingWitness struct {
	lookup  cataloglookup.SnapshotLookup
	binding registrationauthoritybinding.ExactBindingWitness
}

// ComposePositive joins two already-produced positive witnesses. It invokes no
// upstream constructor and defines no new membership, comparison, or negative
// composition outcome. It defensively revalidates both retained witnesses.
func ComposePositive(
	lookup cataloglookup.SnapshotLookup,
	binding registrationauthoritybinding.ExactBindingWitness,
) (SnapshotMemberExactBindingWitness, error) {
	if !lookup.Valid() {
		return SnapshotMemberExactBindingWitness{}, ErrInvalidSnapshotLookup
	}
	registration, present := lookup.Registration()
	if !present {
		return SnapshotMemberExactBindingWitness{}, ErrSnapshotRegistrationAbsent
	}
	if !binding.Valid() {
		return SnapshotMemberExactBindingWitness{}, ErrInvalidExactBindingWitness
	}
	if binding.Registration() != registration {
		return SnapshotMemberExactBindingWitness{}, ErrRegistrationMismatch
	}
	return SnapshotMemberExactBindingWitness{lookup: lookup, binding: binding}, nil
}

// SnapshotLookup returns the complete retained positive membership witness.
func (value SnapshotMemberExactBindingWitness) SnapshotLookup() cataloglookup.SnapshotLookup {
	return value.lookup
}

// ExactBindingWitness returns the complete retained authority-binding witness.
func (value SnapshotMemberExactBindingWitness) ExactBindingWitness() registrationauthoritybinding.ExactBindingWitness {
	return value.binding
}

// Registration returns the exact positive structural registration shared by
// both retained witnesses.
func (value SnapshotMemberExactBindingWitness) Registration() catalogregistration.Registration {
	registration, present := value.lookup.Registration()
	if !present {
		return catalogregistration.Registration{}
	}
	return registration
}

// Valid reports whether both retained positive witnesses expose the same exact
// structural registration.
func (value SnapshotMemberExactBindingWitness) Valid() bool {
	if !value.lookup.Valid() || !value.binding.Valid() {
		return false
	}
	registration, present := value.lookup.Registration()
	return present && registration.Valid() && value.binding.Registration() == registration
}
