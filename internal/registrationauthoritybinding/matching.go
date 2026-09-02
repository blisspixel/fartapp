// Package registrationauthoritybinding defines an internal, non-wire exact relation
// between one snapshot-resolved opaque authority reference and one positive
// structural catalog registration. It does not define attribution, provenance,
// responsibility, trust, ownership, delegation, identity, or catalog closure.
package registrationauthoritybinding

import (
	"errors"

	"github.com/blisspixel/fartapp/internal/authorityresolution"
	"github.com/blisspixel/fartapp/internal/catalogregistration"
)

const (
	exactAuthorityBindingRuleID       = "lab.exact-authority-binding"
	exactAuthorityBindingRuleRevision = "v0alpha1"
	exactAuthorityBindingRuleMarker   = uint8(1)
)

var (
	// ErrInvalidRule reports a missing or forged exact-binding rule identity.
	ErrInvalidRule = errors.New("invalid exact authority-binding rule")
	// ErrInvalidResolvedAuthority reports missing or forged positive resolution evidence.
	ErrInvalidResolvedAuthority = errors.New("invalid snapshot-resolved authority")
	// ErrInvalidRegistration reports a missing or forged positive structural registration.
	ErrInvalidRegistration = errors.New("invalid positive structural registration")
	// ErrCatalogScopeMismatch reports inputs from different exact catalog scopes.
	ErrCatalogScopeMismatch = errors.New("exact authority-binding catalog scope mismatch")
)

// Rule is the closed identity of one implemented pairwise comparison rule.
// Revision is identity material, not chronology, maturity, or precedence.
type Rule struct {
	id       string
	revision string
	marker   uint8
}

// ExactAuthorityBindingV0Alpha1 returns the only rule implemented by this
// candidate. It performs exact Go value equality and no normalization.
func ExactAuthorityBindingV0Alpha1() Rule {
	return Rule{
		id:       exactAuthorityBindingRuleID,
		revision: exactAuthorityBindingRuleRevision,
		marker:   exactAuthorityBindingRuleMarker,
	}
}

// ID returns the rule's exact Lab engineering identifier.
func (value Rule) ID() string { return value.id }

// Revision returns the rule's exact revision identity.
func (value Rule) Revision() string { return value.revision }

// Valid reports whether this value names the exact implemented rule revision.
func (value Rule) Valid() bool {
	return value.id == exactAuthorityBindingRuleID &&
		value.revision == exactAuthorityBindingRuleRevision &&
		value.marker == exactAuthorityBindingRuleMarker
}

// Outcome is the pairwise exact-reference comparison result for valid inputs.
type Outcome uint8

const (
	OutcomeInvalid Outcome = iota
	OutcomeExactAuthorityBindingMatch
	OutcomeNoExactAuthorityBindingMatch
)

// Decision retains the exact rule, snapshot-resolved authority witness, and
// positive structural registration. It stores no redundant comparison result.
type Decision struct {
	rule         Rule
	authority    authorityresolution.SnapshotResolvedAuthority
	registration catalogregistration.Registration
}

// Compare applies one exact pairwise authority-reference rule. Catalog-scope
// mismatch is a composition error and never becomes a no-match outcome.
func Compare(
	rule Rule,
	authority authorityresolution.SnapshotResolvedAuthority,
	registration catalogregistration.Registration,
) (Decision, error) {
	if !rule.Valid() {
		return Decision{}, ErrInvalidRule
	}
	if !authority.Valid() {
		return Decision{}, ErrInvalidResolvedAuthority
	}
	if !registration.Valid() {
		return Decision{}, ErrInvalidRegistration
	}
	if authority.Binding().CatalogScope() != registration.Binding().CatalogScope() {
		return Decision{}, ErrCatalogScopeMismatch
	}
	return Decision{
		rule:         rule,
		authority:    authority,
		registration: registration,
	}, nil
}

// Rule returns the exact retained comparison-rule identity.
func (value Decision) Rule() Rule { return value.rule }

// ResolvedAuthority returns the complete retained resolution witness.
func (value Decision) ResolvedAuthority() authorityresolution.SnapshotResolvedAuthority {
	return value.authority
}

// Registration returns the retained positive structural registration. Product
// registration presence still requires separately ratified catalog closure.
func (value Decision) Registration() catalogregistration.Registration {
	return value.registration
}

// Outcome compares only the two exact declaration-authority references.
func (value Decision) Outcome() (Outcome, bool) {
	if !value.Valid() {
		return OutcomeInvalid, false
	}
	if value.authority.Binding().Authority() ==
		value.registration.Binding().DeclarationAuthority() {
		return OutcomeExactAuthorityBindingMatch, true
	}
	return OutcomeNoExactAuthorityBindingMatch, true
}

// ExactBindingWitness returns a positive pairwise witness only for exact
// authority-reference equality under the retained rule.
func (value Decision) ExactBindingWitness() (ExactBindingWitness, bool) {
	outcome, ok := value.Outcome()
	if !ok || outcome != OutcomeExactAuthorityBindingMatch {
		return ExactBindingWitness{}, false
	}
	return ExactBindingWitness{decision: value}, true
}

// Valid reports whether the complete retained inputs are valid and share the
// exact same catalog scope.
func (value Decision) Valid() bool {
	return value.rule.Valid() && value.authority.Valid() && value.registration.Valid() &&
		value.authority.Binding().CatalogScope() == value.registration.Binding().CatalogScope()
}

// ExactBindingWitness is the positive exact-reference-equality refinement. It
// is not an attribution, provenance, identity-equivalence, or trust assertion.
type ExactBindingWitness struct {
	decision Decision
}

// Decision returns the complete retained pairwise comparison witness.
func (value ExactBindingWitness) Decision() Decision { return value.decision }

// Rule returns the exact retained comparison-rule identity.
func (value ExactBindingWitness) Rule() Rule { return value.decision.Rule() }

// ResolvedAuthority returns the complete retained resolution witness.
func (value ExactBindingWitness) ResolvedAuthority() authorityresolution.SnapshotResolvedAuthority {
	return value.decision.ResolvedAuthority()
}

// Registration returns the retained positive structural registration.
func (value ExactBindingWitness) Registration() catalogregistration.Registration {
	return value.decision.Registration()
}

// Valid reports whether both exact authority references remain equal under the
// retained same-scope decision and exact implemented rule revision.
func (value ExactBindingWitness) Valid() bool {
	if !value.decision.Valid() {
		return false
	}
	outcome, ok := value.decision.Outcome()
	return ok && outcome == OutcomeExactAuthorityBindingMatch &&
		value.decision.authority.Binding().Authority() ==
			value.decision.registration.Binding().DeclarationAuthority()
}
