// Package catalogregistration defines an internal structural experiment for
// exact catalog registration records. It has no product wire representation and
// makes no claim about source-law truth or declaration-authority standing.
package catalogregistration

import (
	"errors"
	"fmt"

	"github.com/blisspixel/fartapp/internal/evaluation"
)

const maximumTokenBytes = 128

var (
	ErrInvalidToken    = errors.New("invalid catalog registration token")
	ErrInvalidBinding  = errors.New("invalid catalog registration binding")
	errInvalidPresence = errors.New("invalid catalog registration presence")
)

type exactRef struct {
	id       string
	revision string
}

func newExactRef(id, revision string) (exactRef, error) {
	if err := validateToken(id); err != nil {
		return exactRef{}, fmt.Errorf("id: %w", err)
	}
	if err := validateToken(revision); err != nil {
		return exactRef{}, fmt.Errorf("revision: %w", err)
	}
	return exactRef{id: id, revision: revision}, nil
}

func (value exactRef) valid() bool {
	return validateToken(value.id) == nil && validateToken(value.revision) == nil
}

// CatalogScopeRef identifies the exact catalog scope named by a registration
// binding. It does not prove that the scope was consulted or closed. Revision
// is identity material, not source-law time.
type CatalogScopeRef struct {
	value exactRef
}

func NewCatalogScopeRef(id, revision string) (CatalogScopeRef, error) {
	value, err := newExactRef(id, revision)
	return CatalogScopeRef{value: value}, err
}

func (value CatalogScopeRef) ID() string       { return value.value.id }
func (value CatalogScopeRef) Revision() string { return value.value.revision }
func (value CatalogScopeRef) Valid() bool      { return value.value.valid() }

// DeclarationAuthorityRef identifies the exact authority to which a catalog
// registration is attributed. It does not establish legitimacy or ownership.
type DeclarationAuthorityRef struct {
	value exactRef
}

func NewDeclarationAuthorityRef(id, revision string) (DeclarationAuthorityRef, error) {
	value, err := newExactRef(id, revision)
	return DeclarationAuthorityRef{value: value}, err
}

func (value DeclarationAuthorityRef) ID() string       { return value.value.id }
func (value DeclarationAuthorityRef) Revision() string { return value.value.revision }
func (value DeclarationAuthorityRef) Valid() bool      { return value.value.valid() }

// SubjectRef identifies the exact catalog subject for which capability
// registration is queried. The shared model defines no subject categories.
type SubjectRef struct {
	value exactRef
}

func NewSubjectRef(id, revision string) (SubjectRef, error) {
	value, err := newExactRef(id, revision)
	return SubjectRef{value: value}, err
}

func (value SubjectRef) ID() string       { return value.value.id }
func (value SubjectRef) Revision() string { return value.value.revision }
func (value SubjectRef) Valid() bool      { return value.value.valid() }

// CapabilityID is one opaque catalog capability token.
type CapabilityID struct {
	value string
}

func NewCapabilityID(value string) (CapabilityID, error) {
	if err := validateToken(value); err != nil {
		return CapabilityID{}, err
	}
	return CapabilityID{value: value}, nil
}

func (value CapabilityID) Value() string { return value.value }
func (value CapabilityID) Valid() bool   { return validateToken(value.value) == nil }

// Binding identifies one exact structural registration question. It does not
// establish that a catalog lookup was performed or closed.
type Binding struct {
	catalogScope CatalogScopeRef
	authority    DeclarationAuthorityRef
	subject      SubjectRef
	capability   CapabilityID
}

func Bind(
	catalogScope CatalogScopeRef,
	authority DeclarationAuthorityRef,
	subject SubjectRef,
	capability CapabilityID,
) (Binding, error) {
	value := Binding{
		catalogScope: catalogScope,
		authority:    authority,
		subject:      subject,
		capability:   capability,
	}
	if !value.Valid() {
		return Binding{}, ErrInvalidBinding
	}
	return value, nil
}

func (value Binding) CatalogScope() CatalogScopeRef                 { return value.catalogScope }
func (value Binding) DeclarationAuthority() DeclarationAuthorityRef { return value.authority }
func (value Binding) Subject() SubjectRef                           { return value.subject }
func (value Binding) Capability() CapabilityID                      { return value.capability }
func (value Binding) Valid() bool {
	return value.catalogScope.Valid() && value.authority.Valid() &&
		value.subject.Valid() && value.capability.Valid()
}

// Presence is an evaluated profile-owned outcome for one exact binding.
type Presence uint8

const (
	PresenceInvalid Presence = iota
	PresenceRegistered
	PresenceNotRegistered
)

type presenceOutcome string

const (
	presenceRegistered    presenceOutcome = "registered"
	presenceNotRegistered presenceOutcome = "not_registered"
)

func (presenceOutcome) EvaluationOutcomeMarker() {}

func validatePresence(value presenceOutcome) error {
	switch value {
	case presenceRegistered, presenceNotRegistered:
		return nil
	default:
		return errInvalidPresence
	}
}

// Result records process disposition separately from the profile-owned
// registration presence outcome.
type Result struct {
	binding     Binding
	disposition evaluation.Disposition[presenceOutcome]
}

func EvaluatedRegistered(binding Binding) (Result, error) {
	return evaluated(binding, presenceRegistered)
}

func EvaluatedNotRegistered(binding Binding) (Result, error) {
	return evaluated(binding, presenceNotRegistered)
}

func evaluated(binding Binding, presence presenceOutcome) (Result, error) {
	if !binding.Valid() {
		return Result{}, ErrInvalidBinding
	}
	disposition, err := evaluation.Evaluated(presence, validatePresence)
	if err != nil {
		return Result{}, err
	}
	return Result{binding: binding, disposition: disposition}, nil
}

func NotEvaluated(binding Binding) (Result, error) {
	if !binding.Valid() {
		return Result{}, ErrInvalidBinding
	}
	return Result{
		binding:     binding,
		disposition: evaluation.NotEvaluated[presenceOutcome](),
	}, nil
}

func (value Result) Binding() Binding { return value.binding }

func (value Result) EvaluationKind() evaluation.Kind {
	return value.disposition.Kind()
}

func (value Result) Presence() (Presence, bool) {
	if !value.binding.Valid() || !value.disposition.Valid() ||
		value.disposition.Kind() != evaluation.KindEvaluated {
		return PresenceInvalid, false
	}
	return presenceFromDisposition(value.disposition)
}

func presenceFromDisposition(
	disposition evaluation.Disposition[presenceOutcome],
) (Presence, bool) {
	outcome, ok := disposition.Outcome()
	if !ok {
		return PresenceInvalid, false
	}
	switch outcome {
	case presenceRegistered:
		return PresenceRegistered, true
	case presenceNotRegistered:
		return PresenceNotRegistered, true
	default:
		return PresenceInvalid, false
	}
}

func (value Result) Registration() (Registration, bool) {
	presence, ok := value.Presence()
	if !ok || presence != PresenceRegistered || !value.Valid() {
		return Registration{}, false
	}
	return Registration{binding: value.binding, valid: true}, true
}

func (value Result) Valid() bool {
	if !value.binding.Valid() || !value.disposition.Valid() {
		return false
	}
	if value.disposition.Kind() == evaluation.KindNotEvaluated {
		return true
	}
	_, ok := presenceFromDisposition(value.disposition)
	return ok
}

// Registration is a positive structural value produced only by an evaluated
// registered result. It does not establish that the declaration is true or
// valid outside the exact catalog binding.
type Registration struct {
	binding Binding
	valid   bool
}

func (value Registration) Binding() Binding { return value.binding }
func (value Registration) Valid() bool      { return value.valid && value.binding.Valid() }

func validateToken(value string) error {
	if len(value) == 0 || len(value) > maximumTokenBytes {
		return fmt.Errorf("%w: token must contain 1 to %d ASCII bytes", ErrInvalidToken, maximumTokenBytes)
	}
	separator := false
	for index := 0; index < len(value); index++ {
		character := value[index]
		alphanumeric := (character >= 'a' && character <= 'z') ||
			(character >= '0' && character <= '9')
		if alphanumeric {
			separator = false
			continue
		}
		if character != '.' && character != '_' && character != '-' {
			return fmt.Errorf("%w: token must use lowercase ASCII", ErrInvalidToken)
		}
		if index == 0 || index == len(value)-1 || separator {
			return fmt.Errorf("%w: separators require alphanumeric neighbors", ErrInvalidToken)
		}
		separator = true
	}
	return nil
}
