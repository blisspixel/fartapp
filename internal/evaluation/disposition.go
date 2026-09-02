// Package evaluation defines Lab-process evaluation primitives without owning
// any domain outcome vocabulary or product wire representation. A disposition
// alone does not bind an evaluator or its inputs; a future enclosing profile
// must own that binding.
package evaluation

import "errors"

// ErrOutcomeValidatorRequired reports an attempt to construct an evaluated
// disposition without the profile-owned validator that defines its outcome.
var ErrOutcomeValidatorRequired = errors.New("evaluated disposition requires an outcome validator")

// Kind identifies whether a declared Lab evaluator produced an outcome.
// It does not describe source-law time, uncertainty, applicability, or truth.
type Kind uint8

const (
	KindInvalid Kind = iota
	KindEvaluated
	KindNotEvaluated
)

// Outcome is a named, immutable string value owned by an enclosing profile. The
// marker excludes the predeclared string type. The string restriction is
// deliberate for this bounded internal candidate: it prevents a validated
// result from changing through a retained pointer.
type Outcome interface {
	~string
	EvaluationOutcomeMarker()
}

// OutcomeValidator belongs to the contract that owns O. The evaluation package
// deliberately defines no shared outcome vocabulary.
type OutcomeValidator[O Outcome] func(O) error

// Disposition contains either one validated, profile-owned outcome or no outcome.
// The fields are private so callers cannot bypass the constructors.
type Disposition[O Outcome] struct {
	kind             Kind
	outcome          O
	outcomeValidated bool
}

// Evaluated constructs a disposition only after the owner of O accepts the
// outcome. The validator runs exactly once.
func Evaluated[O Outcome](
	outcome O,
	validate OutcomeValidator[O],
) (Disposition[O], error) {
	if validate == nil {
		return Disposition[O]{}, ErrOutcomeValidatorRequired
	}
	if err := validate(outcome); err != nil {
		return Disposition[O]{}, err
	}
	return Disposition[O]{
		kind:             KindEvaluated,
		outcome:          outcome,
		outcomeValidated: true,
	}, nil
}

// NotEvaluated constructs a disposition with no outcome.
func NotEvaluated[O Outcome]() Disposition[O] {
	return Disposition[O]{kind: KindNotEvaluated}
}

// Kind returns the disposition kind. The zero value returns KindInvalid.
func (value Disposition[O]) Kind() Kind {
	return value.kind
}

// Outcome returns the validated outcome only for a valid evaluated value.
func (value Disposition[O]) Outcome() (O, bool) {
	if !value.Valid() || value.kind != KindEvaluated {
		var zero O
		return zero, false
	}
	return value.outcome, true
}

// Valid reports whether the value satisfies the disposition invariant.
func (value Disposition[O]) Valid() bool {
	var zero O
	switch value.kind {
	case KindEvaluated:
		return value.outcomeValidated
	case KindNotEvaluated:
		return !value.outcomeValidated && value.outcome == zero
	default:
		return false
	}
}
