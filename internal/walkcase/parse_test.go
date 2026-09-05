package walkcase

import (
	"bytes"
	"errors"
	"io"
	"math"
	"strings"
	"testing"

	"github.com/blisspixel/fartapp/internal/coupledblowdown"
	"github.com/blisspixel/fartapp/internal/idealmixturereservoir"
	"github.com/blisspixel/fartapp/internal/restrictionflow"
)

func TestWalkRejectsInvalidDocumentSemantics(t *testing.T) {
	base := readFixture(t, "isothermal-choked.json")
	tests := []struct {
		name   string
		change change
		reason string
	}{
		{"schema", change{path: "schema", value: "wrong"}, "unsupported_schema"},
		{"model missing", change{path: "model", remove: true}, "unsupported_model_revision"},
		{"model revision", change{path: "model/version", value: "v0alpha2"}, "unsupported_model_revision"},
		{"quantity", change{path: "quantity_system", value: "cgs"}, "unsupported_quantity_system"},
		{"law unresolved", change{path: "law_context/id", value: "nope"}, "unresolved_law_context"},
		{"law exact revision", change{path: "law_context/version", remove: true}, "exact_law_revision_required"},
		{"closure", change{path: "closure", value: "elastic"}, "unsupported_closure"},
		{"volume missing", change{path: "reservoir/volume_m3", remove: true}, "missing_member"},
		{"volume negative", change{path: "reservoir/volume_m3", value: -1}, "nonpositive_quantity"},
		{"temperature", change{path: "reservoir/temperature_k", value: 0}, "nonpositive_quantity"},
		{"empty components", change{path: "reservoir/components", value: []any{}}, "missing_component"},
		{"component ID", change{path: "reservoir/components/0/id", value: "bad token"}, "invalid_token"},
		{"mass missing", change{path: "reservoir/components/0/mass_kg", remove: true}, "missing_member"},
		{"mass", change{path: "reservoir/components/0/mass_kg", value: -1}, "nonpositive_quantity"},
		{"R", change{path: "reservoir/components/0/specific_gas_constant_j_per_kg_k", value: 0}, "nonpositive_quantity"},
		{"cv", change{path: "reservoir/components/0/isochoric_heat_capacity_j_per_kg_k", value: 0}, "nonpositive_quantity"},
		{"back pressure missing", change{path: "restriction/back_pressure_pa", remove: true}, "missing_member"},
		{"back pressure", change{path: "restriction/back_pressure_pa", value: 0}, "nonpositive_quantity"},
		{"Cd", change{path: "restriction/discharge_coefficient", value: 2}, "invalid_discharge_coefficient"},
		{"area missing", change{path: "restriction/area/prescribed_m2", remove: true}, "missing_member"},
		{"area negative", change{path: "restriction/area/prescribed_m2", value: -1}, "negative_area"},
		{"area law", change{path: "restriction/area/law", value: "magic"}, "unsupported_area_law"},
		{"area inactive member", change{path: "restriction/area/compliance_m2_per_pa", value: 0}, "unexpected_area_member"},
		{"compliance missing", change{path: "restriction/area/law", value: "linear-compliance"}, "missing_member"},
		{"step missing", change{path: "step/max_steps", remove: true}, "missing_member"},
		{"zero steps", change{path: "step/max_steps", value: 0}, "invalid_step_policy"},
		{"too many steps", change{path: "step/max_steps", value: 4097}, "invalid_step_policy"},
		{"invalid fraction", change{path: "step/max_withdrawal_fraction_per_step", value: 1}, "invalid_step_policy"},
		{"negative time", change{path: "step/max_time_s", value: -1}, "invalid_step_policy"},
		{"branch missing area", change{path: "branch/prescribed_area_m2", remove: true}, "missing_member"},
		{"negative branch", change{path: "branch/prescribed_area_m2", value: -1}, "negative_area"},
		{"expected empty", change{path: "expected_witness", value: ""}, "invalid_witness_digest"},
		{"expected uppercase", change{path: "expected_witness", value: strings.Repeat("A", 64)}, "invalid_witness_digest"},
		{"unknown member", change{path: "surprise", value: 0}, "unknown_member"},
		{"case alias", change{path: "Schema", value: RequestSchema}, "unknown_member"},
		{"nested alias", change{path: "reservoir/Volume_m3", value: 1}, "unknown_member"},
		{"null optional context", change{path: "law_context", value: nil}, "null_not_allowed"},
		{"null optional time", change{path: "step/max_time_s", value: nil}, "null_not_allowed"},
		{"null required", change{path: "reservoir/temperature_k", value: nil}, "null_not_allowed"},
		{"wrong reservoir type", change{path: "reservoir", value: 1}, "document_shape_invalid"},
		{"wrong component type", change{path: "reservoir/components", value: "nope"}, "document_shape_invalid"},
		{"wrong scalar type", change{path: "step/max_steps", value: "64"}, "document_shape_invalid"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) { assertReason(t, Run(editJSON(t, base, test.change), "predict"), test.reason) })
	}
	assertReason(t, Run(editJSON(t, base, change{path: "branch", remove: true}), "branch"), "missing_member")
	assertReason(t, Run(base, "dance"), "unsupported_operation")
	for _, operation := range []string{"predict", "simulate", "branch", "witness"} {
		assertReason(t, Run(editJSON(t, base, change{path: "restriction/back_pressure_pa", value: 125001}), operation), "adverse_pressure")
	}
	for _, area := range []struct {
		value  map[string]any
		reason string
	}{
		{map[string]any{"law": "linear-compliance", "prescribed_m2": 0.01, "compliance_m2_per_pa": -1, "maximum_m2": 0.02}, "negative_compliance"},
		{map[string]any{"law": "linear-compliance", "prescribed_m2": 0.01, "compliance_m2_per_pa": 1, "maximum_m2": -1}, "negative_area"},
		{map[string]any{"law": "linear-compliance", "prescribed_m2": 0.01, "compliance_m2_per_pa": 1, "maximum_m2": 0.001}, "invalid_area_law"},
	} {
		assertReason(t, Run(editJSON(t, base, change{path: "restriction/area", value: area.value}), "simulate"), area.reason)
	}
	component := map[string]any{"id": "a", "mass_kg": 1, "specific_gas_constant_j_per_kg_k": 200, "isochoric_heat_capacity_j_per_kg_k": 400}
	assertReason(t, Run(editJSON(t, base, change{path: "reservoir/components", value: []any{component, component}}), "simulate"), "duplicate_component_id")
	components := make([]any, 65)
	for index := range components {
		components[index] = component
	}
	assertReason(t, Run(editJSON(t, base, change{path: "reservoir/components", value: components}), "simulate"), "too_many_components")
	assertReason(t, Run(editJSON(t, base, change{path: "reservoir/volume_m3", value: 1e-310}), "simulate"), "invalid_state")
}

func TestWalkRejectsMalformedAndUnboundedJSON(t *testing.T) {
	for _, test := range []struct {
		input  []byte
		reason string
	}{
		{nil, "empty_input"},
		{[]byte(`{"schema":"a","schema":"b"}`), "duplicate_member"},
		{[]byte(`{"schema":"\ud800"}`), "malformed_json"},
		{[]byte{'{', 0xff, '}'}, "malformed_json"},
		{[]byte("{}\v"), "trailing_json_value"},
		{[]byte(`{} {}`), "trailing_json_value"},
		{[]byte(`[]`), "document_shape_invalid"},
		{[]byte(`null`), "null_not_allowed"},
		{[]byte(strings.Repeat("[", 34) + strings.Repeat("]", 34)), "maximum_depth_exceeded"},
		{[]byte(`{"` + strings.Repeat("x", 129) + `":0}`), "member_name_too_long"},
		{bytes.Repeat([]byte{' '}, MaxInputBytes+1), "input_too_large"},
	} {
		assertReason(t, Run(test.input, "simulate"), test.reason)
	}
}

func TestStepLimitsAreNotHiddenPhysicalConclusions(t *testing.T) {
	base := readFixture(t, "isothermal-choked.json")
	for _, test := range []struct {
		change change
		want   string
	}{
		{change{path: "step/max_time_s", remove: true}, "max-steps"},
		{change{path: "closure", value: "rigid-adiabatic"}, "max-time"},
	} {
		report := Run(editJSON(t, base, test.change), "explain")
		assertPredicted(t, report)
		if report.Stop != test.want || !strings.Contains(strings.Join(report.Explanation, " "), "budget") {
			t.Fatal("budget stop was not explained")
		}
	}
}

func TestReadBoundedFailureAndClaimDefenses(t *testing.T) {
	got, err := ReadBounded(bytes.NewReader([]byte("x")))
	if err != nil || string(got) != "x" {
		t.Fatalf("read=(%q,%v)", got, err)
	}
	if _, err := ReadBounded(nil); !errors.Is(err, ErrNilInput) {
		t.Fatal("nil input accepted")
	}
	if _, err := ReadBounded(bytes.NewReader(bytes.Repeat([]byte{'x'}, MaxInputBytes+1))); !errors.Is(err, ErrInputTooLarge) {
		t.Fatal("large input accepted")
	}
	if _, err := ReadBounded(errorReader{}); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatal("reader error hidden")
	}
	if (Report{}).Predicted() || InputFailure("input_unavailable", "input_stream").Predicted() {
		t.Fatal("invalid report accepted")
	}
	for _, value := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if !math.IsNaN(roundoffTolerance(value)) || newClaim("x", "method", "kg", value, 1).Status != "failed" {
			t.Fatal("nonfinite claim accepted")
		}
	}
	report := finish(Report{Schema: ReportSchema, Status: "predicted", Claims: []Claim{newClaim("x", "method", "kg", 2, 1)}}, nil)
	assertReason(t, report, "invariant_violation")
	if len(report.Claims) != 1 || report.Claims[0].Residual != 2 {
		t.Fatal("failed claim evidence discarded")
	}
	if claims(nil) != nil || inspectDocumentShape([]byte("{")) == nil {
		t.Fatal("invalid helper input accepted")
	}
	for _, test := range []struct {
		err  error
		want string
	}{
		{restrictionflow.ErrAdversePressure, "adverse_pressure"},
		{idealmixturereservoir.ErrReservoirExhausted, "reservoir_depletion"},
		{coupledblowdown.ErrInvalidStepPolicy, "invalid_step_policy"},
		{errors.New("unknown"), "numerical_domain_error"},
	} {
		if classify(test.err) != test.want {
			t.Fatal("incorrect error classification")
		}
	}
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }
