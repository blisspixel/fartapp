package walkcase

import (
	"errors"

	"github.com/blisspixel/fartapp/internal/coupledblowdown"
)

const RefinementRevision = "go-oracle.walk-refine/v0alpha2"

// RefinementEvidence records the requested policy and estimates separately from
// thermodynamic balance claims. Failed runs expose work counters, not estimates.
type RefinementEvidence struct {
	RelativeTolerance            float64              `json:"relative_tolerance"`
	AbsoluteTimeToleranceSeconds float64              `json:"absolute_time_tolerance_s"`
	MaxEvaluations               int                  `json:"max_evaluations"`
	Evaluations                  int                  `json:"evaluations"`
	AcceptedIntervals            int                  `json:"accepted_intervals"`
	Refinements                  int                  `json:"refinements"`
	ToleranceSatisfied           bool                 `json:"tolerance_satisfied"`
	DischargeComplete            bool                 `json:"discharge_complete"`
	Estimates                    *RefinementEstimates `json:"estimates,omitempty"`
}

type RefinementEstimates struct {
	TimeErrorSeconds              float64 `json:"time_error_s"`
	ImpulseErrorNewtonSeconds     float64 `json:"impulse_error_n_s"`
	StrokeErrorMetres             float64 `json:"stroke_error_m"`
	RequestedTimeToleranceSeconds float64 `json:"requested_time_tolerance_s"`
}

// Refine calculates the same ideal thermodynamic path with bounded adaptive
// quadrature. It does not change Run's legacy witness implementation or profile.
func Refine(data []byte, options coupledblowdown.AccuracyOptions) Report {
	fail := func(code, stage, path, reason string) Report {
		report := failure(code, stage, path, reason)
		report.Operation = "refine"
		report.ImplementationRevision = RefinementRevision
		return report
	}
	if len(data) > MaxInputBytes {
		return fail("FART-E-INPUT-0005", "input", "/", "input_too_large")
	}
	parsed, diagnostic := parseCase(data)
	if diagnostic != nil {
		return fail(diagnostic.Code, diagnostic.Stage, diagnostic.Path, diagnostic.ReasonCode)
	}
	config, err := coupledblowdown.NewConfig(parsed.state, parsed.closure, parsed.back, parsed.area,
		parsed.cd, parsed.fraction, parsed.maxSteps, parsed.maxTime)
	if err != nil {
		return fail("FART-E-MODEL-0006", "model", "/", classify(err))
	}
	result, evidence, err := coupledblowdown.SimulateAccurate(config, options)
	accuracy := &RefinementEvidence{
		RelativeTolerance: options.RelativeTolerance, AbsoluteTimeToleranceSeconds: options.AbsoluteTimeToleranceSeconds,
		MaxEvaluations: options.MaxEvaluations, Evaluations: evidence.Evaluations,
		AcceptedIntervals: evidence.AcceptedIntervals, Refinements: evidence.Refinements,
		ToleranceSatisfied: evidence.ToleranceSatisfied, DischargeComplete: evidence.DischargeComplete,
	}
	if err != nil {
		reason := classify(err)
		switch {
		case errors.Is(err, coupledblowdown.ErrInvalidAccuracyOptions):
			// Invalid IEEE values must not contaminate the JSON refusal.
			return fail("FART-E-SCHEMA-0005", "schema", "/accuracy", "invalid_accuracy_options")
		case errors.Is(err, coupledblowdown.ErrUnsupportedAccuracyConfig):
			reason = "unsupported_accuracy_configuration"
		case errors.Is(err, coupledblowdown.ErrAccuracyBudgetExhausted):
			reason = "accuracy_budget_exhausted"
		case errors.Is(err, coupledblowdown.ErrAccuracyNotAchieved):
			reason = "accuracy_not_achieved"
		}
		report := fail("FART-E-NUMERICAL-0004", "model", "/accuracy", reason)
		report.Accuracy = accuracy
		return report
	}
	accuracy.Estimates = &RefinementEstimates{
		TimeErrorSeconds:              evidence.EstimatedTimeErrorSeconds,
		ImpulseErrorNewtonSeconds:     evidence.EstimatedImpulseErrorNewtonSeconds,
		StrokeErrorMetres:             evidence.EstimatedStrokeErrorMetres,
		RequestedTimeToleranceSeconds: evidence.RequestedTimeToleranceSeconds,
	}
	report := simulateReport(parsed, "refine", result)
	report.ImplementationRevision = RefinementRevision
	report.NumericalPolicy.Method = "regularized-mass-coordinate-adaptive-gauss-kronrod-15-7"
	report.Accuracy = accuracy
	report.Explanation = []string{
		"The ideal reservoir path is integrated in a regularized mass coordinate. Choking and compliance-cap transitions are split explicitly.",
		"The withdrawal fraction selects retained history samples; the declared tolerances and evaluation budget control quadrature work.",
		"Estimated quadrature errors exclude input uncertainty, floating-point representation error, and physical model error. They are not rigorous solution bounds or empirical validation.",
		"Tolerance satisfaction and completed discharge are separate. A maximum-step stop can satisfy the requested accuracy for an incomplete path.",
	}
	return finish(report, &result)
}
