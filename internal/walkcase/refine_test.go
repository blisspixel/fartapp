package walkcase

import (
	"bytes"
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/blisspixel/fartapp/internal/coupledblowdown"
)

func TestRefineIndependentClockAndLegacyProfile(t *testing.T) {
	input := readFixture(t, "ordinary-low-pressure.json")
	options := coupledblowdown.AccuracyOptions{RelativeTolerance: 1e-8, MaxEvaluations: 100000}
	legacyBefore, _ := json.Marshal(Run(input, "witness"))
	report := Refine(input, options)
	assertPredicted(t, report)
	// Independently integrated gamma=7/5 isothermal low-pressure primitive.
	const referenceSeconds = 0.05839420446440555
	if report.Operation != "refine" || report.ImplementationRevision != RefinementRevision ||
		math.Abs(*report.ElapsedSeconds-referenceSeconds) > 2e-12 || report.Witness != "" ||
		report.Accuracy == nil || report.Accuracy.Estimates == nil || !report.Accuracy.ToleranceSatisfied ||
		!report.Accuracy.DischargeComplete || report.Accuracy.Evaluations > options.MaxEvaluations {
		t.Fatalf("incomplete or inaccurate report: %#v", report)
	}
	if report.Accuracy.Estimates.TimeErrorSeconds > report.Accuracy.Estimates.RequestedTimeToleranceSeconds ||
		!strings.Contains(report.NumericalPolicy.Method, "gauss-kronrod") ||
		report.History[len(report.History)-1].TimeSeconds != *report.ElapsedSeconds {
		t.Fatal("accuracy evidence and retained account disagree")
	}
	coarse := Refine(editJSON(t, input, change{path: "step/max_withdrawal_fraction_per_step", value: 0.01}), options)
	assertPredicted(t, coarse)
	if *coarse.Steps != 1 || math.Abs(*coarse.ElapsedSeconds-*report.ElapsedSeconds) > 2e-13 ||
		coarse.Accuracy.Evaluations >= report.Accuracy.Evaluations {
		t.Fatal("history density controls the calculated clock")
	}
	legacyAfter, _ := json.Marshal(Run(input, "witness"))
	if !bytes.Equal(legacyBefore, legacyAfter) || bytes.Contains(legacyAfter, []byte(`"accuracy"`)) {
		t.Fatal("refinement changed the legacy witness profile")
	}
	if _, err := json.Marshal(report); err != nil {
		t.Fatal(err)
	}
}

func TestRefineCompletionBudgetAndRefusals(t *testing.T) {
	input := readFixture(t, "ordinary-low-pressure.json")
	options := coupledblowdown.AccuracyOptions{RelativeTolerance: 1e-8, MaxEvaluations: 100000}
	truncated := Refine(editJSON(t, input, change{path: "step/max_steps", value: 1}), options)
	assertPredicted(t, truncated)
	if truncated.Stop != "max-steps" || !truncated.Accuracy.ToleranceSatisfied || truncated.Accuracy.DischargeComplete {
		t.Fatal("truncated path was treated as completed discharge")
	}
	noFlow := Refine(editJSON(t, input, change{path: "restriction/area/prescribed_m2", value: 0}), options)
	assertPredicted(t, noFlow)
	if noFlow.Stop != "no-flow" || *noFlow.ElapsedSeconds != 0 || noFlow.Accuracy.Evaluations != 0 {
		t.Fatal("closed outlet did not preserve the no-flow identity")
	}
	tests := []struct {
		name   string
		data   []byte
		opts   coupledblowdown.AccuracyOptions
		reason string
	}{
		{"budget", input, coupledblowdown.AccuracyOptions{RelativeTolerance: 1e-8, MaxEvaluations: 15}, "accuracy_budget_exhausted"},
		{"time", editJSON(t, input, change{path: "step/max_time_s", value: 0.01}), options, "unsupported_accuracy_configuration"},
		{"syntax", []byte("{"), options, "malformed_json"},
		{"size", bytes.Repeat([]byte(" "), MaxInputBytes+1), options, "input_too_large"},
		{"nan", input, coupledblowdown.AccuracyOptions{RelativeTolerance: math.NaN(), MaxEvaluations: 100}, "invalid_accuracy_options"},
		{"absolute infinity", input, coupledblowdown.AccuracyOptions{RelativeTolerance: 1e-8, AbsoluteTimeToleranceSeconds: math.Inf(1), MaxEvaluations: 100}, "invalid_accuracy_options"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			report := Refine(test.data, test.opts)
			if report.Predicted() || report.ImplementationRevision != RefinementRevision ||
				len(report.Diagnostics) != 1 || report.Diagnostics[0].ReasonCode != test.reason {
				t.Fatalf("refusal = %#v", report)
			}
			if report.Accuracy != nil && report.Accuracy.Estimates != nil {
				t.Fatal("failed integration exported a usable error estimate")
			}
			if _, err := json.Marshal(report); err != nil {
				t.Fatal(err)
			}
		})
	}
}
