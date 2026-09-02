package reservoirprediction

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestPredictSyntheticClosedForms(t *testing.T) {
	tests := []struct {
		fixture     string
		closure     string
		temperature float64
		pressure    float64
		energy      float64
		enthalpy    float64
		heat        float64
	}{
		{
			fixture: "synthetic-mixture-adiabatic.json", closure: "rigid-adiabatic",
			temperature: 200, pressure: 70000, energy: 140000, enthalpy: 980000,
		},
		{
			fixture: "synthetic-mixture-isothermal.json", closure: "rigid-isothermal",
			temperature: 400, pressure: 140000, energy: 280000, enthalpy: 1260000, heat: 420000,
		},
	}
	for _, test := range tests {
		t.Run(test.closure, func(t *testing.T) {
			report := Predict(readFixture(t, test.fixture))
			if !report.Predicted() {
				t.Fatalf("Predict: %#v", report)
			}
			if report.RequestSchema != RequestSchema || report.Model.ID != ModelID ||
				report.Model.Version != ModelVersion || report.ImplementationRevision != ImplementationRevision ||
				report.QuantitySystem != QuantitySystem || report.Closure != test.closure ||
				report.WithdrawalFraction == nil || *report.WithdrawalFraction != 0.75 {
				t.Fatalf("identity fields = %#v", report)
			}
			assertNear(t, "initial mass", report.Initial.TotalMassKilograms, 4, 0)
			assertNear(t, "initial pressure", report.Initial.PressurePascals, 560000, 0)
			assertNear(t, "initial energy", report.Initial.InternalEnergyJoules, 1120000, 0)
			assertNear(t, "final mass", report.Final.TotalMassKilograms, 1, 0)
			assertNear(t, "final temperature", report.Final.TemperatureKelvin, test.temperature, 1e-12)
			assertNear(t, "final pressure", report.Final.PressurePascals, test.pressure, 1e-9)
			assertNear(t, "final energy", report.Final.InternalEnergyJoules, test.energy, 1e-9)
			assertNear(t, "enthalpy", report.Transfers.IntegratedEnthalpyOutJoules, test.enthalpy, 1e-8)
			assertNear(t, "heat", report.Transfers.HeatIntoReservoirJoules, test.heat, 1e-8)
			if len(report.Initial.Components) != 2 || report.Initial.Components[0].ID != "component.a" ||
				len(report.Transfers.Components) != 2 || report.Transfers.Components[1].ID != "component.b" ||
				len(report.Balances.Components) != 2 || len(report.Claims) != 4 {
				t.Fatalf("component or claim projection = %#v", report)
			}
			for _, claim := range report.Claims {
				if claim.Status != "satisfied-within-roundoff" || math.Abs(claim.Residual) > claim.Tolerance {
					t.Errorf("claim = %#v", claim)
				}
			}
			if len(report.ValidationEnvironment.ConsultedInputs) != 1 ||
				report.ValidationEnvironment.ConsultedInputs[0] != "document_bytes" ||
				len(report.ValidationEnvironment.AmbientInputs) != 0 ||
				len(report.Assumptions) != 8 || report.Nonclaims == nil ||
				len(report.Nonclaims.Model) != 6 || len(report.Nonclaims.Operation) != 2 ||
				len(report.Nonclaims.Evidence) != 1 {
				t.Fatalf("scope disclosure = %#v", report)
			}
		})
	}
}

func TestPredictionCanonicalizesComponentOrderAndMemberOrder(t *testing.T) {
	first := Predict(readFixture(t, "synthetic-mixture-adiabatic.json"))
	reordered := []byte(`{
      "initial":{"temperature_k":400,"volume_m3":1,"components":[
        {"isochoric_heat_capacity_j_per_kg_k":800,"specific_gas_constant_j_per_kg_k":400,"mass_kg":3,"id":"component.b"},
        {"mass_kg":1,"id":"component.a","isochoric_heat_capacity_j_per_kg_k":400,"specific_gas_constant_j_per_kg_k":200}]},
      "withdrawal_fraction":0.75,"closure":"rigid-adiabatic","quantity_system":"si",
      "model":{"version":"v0alpha1","id":"continuum.rigid-calorically-perfect-ideal-mixture"},
      "schema":"fart.reservoir-prediction-request/v0alpha1"}`)
	second := Predict(reordered)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("input order changed report:\n%#v\n%#v", first, second)
	}
	firstJSON, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := json.Marshal(Predict(reordered))
	if err != nil || !bytes.Equal(firstJSON, secondJSON) {
		t.Fatalf("JSON is not deterministic: %v", err)
	}
}

func TestZeroPredictionIsExactAndDisclosed(t *testing.T) {
	input := strings.Replace(
		string(readFixture(t, "synthetic-mixture-adiabatic.json")),
		`"withdrawal_fraction": 0.75`,
		`"withdrawal_fraction": 0`,
		1,
	)
	report := Predict([]byte(input))
	if !report.Predicted() || !reflect.DeepEqual(*report.Initial, *report.Final) ||
		report.Transfers.TotalMassOutKilograms != 0 ||
		report.Transfers.IntegratedEnthalpyOutJoules != 0 ||
		report.Transfers.HeatIntoReservoirJoules != 0 ||
		report.Balances.TotalMassResidualKilograms != 0 || report.Balances.EnergyResidualJoules != 0 {
		t.Fatalf("zero prediction = %#v", report)
	}
	encoded, err := json.Marshal(report)
	if err != nil || !bytes.Contains(encoded, []byte(`"withdrawal_fraction":0`)) {
		t.Fatalf("zero withdrawal was omitted from JSON: (%s, %v)", encoded, err)
	}
}

func TestExtremeFiniteInputCannotCreateInfiniteTolerance(t *testing.T) {
	input := []byte(`{
      "schema":"fart.reservoir-prediction-request/v0alpha1",
      "model":{"id":"continuum.rigid-calorically-perfect-ideal-mixture","version":"v0alpha1"},
      "quantity_system":"si","closure":"rigid-isothermal","withdrawal_fraction":0.1,
      "initial":{"components":[{"id":"component.a","mass_kg":1e308,
      "specific_gas_constant_j_per_kg_k":1e-308,
      "isochoric_heat_capacity_j_per_kg_k":1e-308}],"volume_m3":1,"temperature_k":1}}`)
	report := Predict(input)
	if !report.Predicted() {
		t.Fatalf("extreme finite prediction = %#v", report)
	}
	for _, claim := range report.Claims {
		if math.IsNaN(claim.Residual) || math.IsInf(claim.Residual, 0) ||
			math.IsNaN(claim.Tolerance) || math.IsInf(claim.Tolerance, 0) {
			t.Fatalf("nonfinite claim = %#v", claim)
		}
	}
	if _, err := json.Marshal(report); err != nil {
		t.Fatalf("marshal extreme finite prediction: %v", err)
	}
}

func TestPredictRejectsSyntaxSchemaAndModelErrors(t *testing.T) {
	base := string(readFixture(t, "synthetic-mixture-adiabatic.json"))
	tooManyComponents := strings.Repeat(`{"id":"x","mass_kg":1,"specific_gas_constant_j_per_kg_k":1,"isochoric_heat_capacity_j_per_kg_k":1},`, 65)
	tests := []struct {
		name   string
		input  string
		stage  string
		path   string
		reason string
	}{
		{name: "empty", input: "", stage: "syntax", path: "/", reason: "empty_input"},
		{name: "malformed", input: "{", stage: "syntax", path: "/", reason: "malformed_json"},
		{name: "trailing", input: "{}{}", stage: "syntax", path: "/", reason: "trailing_json_value"},
		{name: "duplicate", input: `{"schema":1,"schema":2}`, stage: "syntax", path: "/schema", reason: "duplicate_member"},
		{name: "unknown", input: strings.Replace(base, `"schema":`, `"unknown":0,"schema":`, 1), stage: "schema", path: "/", reason: "document_shape_invalid"},
		{name: "schema", input: strings.Replace(base, RequestSchema, "fart.reservoir-prediction-request/v9", 1), stage: "schema", path: "/schema", reason: "unsupported_schema"},
		{name: "model", input: strings.Replace(base, ModelID, "continuum.other", 1), stage: "schema", path: "/model", reason: "unsupported_model_revision"},
		{name: "units", input: strings.Replace(base, `"quantity_system": "si"`, `"quantity_system": "other"`, 1), stage: "schema", path: "/quantity_system", reason: "unsupported_quantity_system"},
		{name: "closure", input: strings.Replace(base, "rigid-adiabatic", "magic", 1), stage: "schema", path: "/closure", reason: "unsupported_closure"},
		{name: "missing fraction", input: strings.Replace(base, `"withdrawal_fraction": 0.75,`, "", 1), stage: "schema", path: "/withdrawal_fraction", reason: "missing_member"},
		{name: "depletion", input: strings.Replace(base, `"withdrawal_fraction": 0.75`, `"withdrawal_fraction": 1`, 1), stage: "model", path: "/withdrawal_fraction", reason: "reservoir_depletion"},
		{name: "negative withdrawal", input: strings.Replace(base, `"withdrawal_fraction": 0.75`, `"withdrawal_fraction": -1`, 1), stage: "model", path: "/withdrawal_fraction", reason: "invalid_withdrawal"},
		{name: "missing volume", input: strings.Replace(base, `"volume_m3": 1,`, "", 1), stage: "schema", path: "/initial/volume_m3", reason: "missing_member"},
		{name: "bad temperature", input: strings.Replace(base, `"temperature_k": 400`, `"temperature_k": 0`, 1), stage: "model", path: "/initial/temperature_k", reason: "nonpositive_quantity"},
		{name: "no components", input: strings.Replace(base, componentArray(base), "[]", 1), stage: "schema", path: "/initial/components", reason: "missing_component"},
		{name: "too many components", input: minimalRequest("[" + strings.TrimSuffix(tooManyComponents, ",") + "]"), stage: "schema", path: "/initial/components", reason: "collection_limit_exceeded"},
		{name: "duplicate id", input: strings.Replace(base, "component.b", "component.a", 1), stage: "schema", path: "/initial/components/1/id", reason: "duplicate_component_id"},
		{name: "invalid id", input: strings.Replace(base, "component.a", "bad id", 1), stage: "schema", path: "/initial/components/0/id", reason: "invalid_token"},
		{name: "missing component field", input: strings.Replace(base, `"mass_kg": 1,`, "", 1), stage: "schema", path: "/initial/components/0", reason: "missing_member"},
		{name: "bad component mass", input: strings.Replace(base, `"mass_kg": 1`, `"mass_kg": 0`, 1), stage: "model", path: "/initial/components/0/mass_kg", reason: "nonpositive_quantity"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			report := Predict([]byte(test.input))
			if report.Predicted() || report.Status != "invalid" || len(report.Diagnostics) != 1 {
				t.Fatalf("report = %#v", report)
			}
			if report.WithdrawalFraction != nil {
				t.Fatalf("invalid report fabricated a withdrawal fraction: %#v", report)
			}
			diagnostic := report.Diagnostics[0]
			if diagnostic.Stage != test.stage || diagnostic.Path != test.path || diagnostic.ReasonCode != test.reason {
				t.Fatalf("diagnostic = %#v, want (%s, %s, %s)", diagnostic, test.stage, test.path, test.reason)
			}
			if len(report.ValidationEnvironment.AmbientInputs) != 0 {
				t.Fatalf("failure consulted ambient inputs: %#v", report.ValidationEnvironment)
			}
		})
	}
}

func TestPreflightLimits(t *testing.T) {
	deep := strings.Repeat("[", maximumJSONDepth+2) + "0" + strings.Repeat("]", maximumJSONDepth+2)
	if got := Predict([]byte(deep)).Diagnostics[0].ReasonCode; got != "maximum_depth_exceeded" {
		t.Fatalf("deep reason = %q", got)
	}
	longName := strings.Repeat("x", maximumMemberNameBytes+1)
	if got := Predict([]byte(`{"` + longName + `":0}`)).Diagnostics[0].ReasonCode; got != "member_name_too_long" {
		t.Fatalf("long member reason = %q", got)
	}
	oversized := bytes.Repeat([]byte{' '}, MaxInputBytes+1)
	if got := Predict(oversized).Diagnostics[0].ReasonCode; got != "input_too_large" {
		t.Fatalf("oversized reason = %q", got)
	}
}

func TestReadBounded(t *testing.T) {
	input := []byte("bounded")
	got, err := ReadBounded(bytes.NewReader(input))
	if err != nil || !bytes.Equal(got, input) {
		t.Fatalf("ReadBounded = (%q, %v)", got, err)
	}
	if _, err := ReadBounded(io.LimitReader(strings.NewReader(strings.Repeat("x", MaxInputBytes+1)), MaxInputBytes+1)); !errors.Is(err, ErrInputTooLarge) {
		t.Fatalf("large error = %v", err)
	}
	if _, err := ReadBounded(errorReader{}); err == nil {
		t.Fatal("ReadBounded accepted a reader failure")
	}
}

func TestHelpersAndFailureValidation(t *testing.T) {
	if (Report{}).Predicted() || failure("c", "s", "/p", "r").Predicted() {
		t.Fatal("invalid report was valid")
	}
	inputFailure := InputFailure("input_unavailable", "input_stream")
	if len(inputFailure.ValidationEnvironment.ConsultedInputs) != 1 ||
		inputFailure.ValidationEnvironment.ConsultedInputs[0] != "input_stream" {
		t.Fatalf("input failure provenance = %#v", inputFailure.ValidationEnvironment)
	}
	claim := newClaim("id", "method", "u", 2, 1)
	if claim.Status != "failed" {
		t.Fatalf("failed claim = %#v", claim)
	}
	for _, claim := range []Claim{
		newClaim("id", "method", "u", math.Inf(1), math.Inf(1)),
		newClaim("id", "method", "u", 0, math.NaN()),
	} {
		if claim.Status != "failed" {
			t.Fatalf("nonfinite claim passed = %#v", claim)
		}
	}
	if tolerance := roundoffTolerance(2); tolerance <= 0 {
		t.Fatalf("tolerance = %g", tolerance)
	}
	if _, ok := parseClosure("other"); ok {
		t.Fatal("unsupported closure parsed")
	}
}

func FuzzPredict(f *testing.F) {
	f.Add(readFixture(f, "synthetic-mixture-adiabatic.json"))
	f.Add([]byte(`{"schema":"x"}`))
	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) > MaxInputBytes+32 {
			t.Skip()
		}
		first := Predict(input)
		second := Predict(input)
		if !reflect.DeepEqual(first, second) {
			t.Fatal("prediction was not deterministic")
		}
		if first.Schema != ReportSchema || first.ImplementationRevision != ImplementationRevision ||
			len(first.ValidationEnvironment.AmbientInputs) != 0 {
			t.Fatalf("invalid envelope: %#v", first)
		}
		if first.Predicted() {
			if len(first.Diagnostics) != 0 {
				t.Fatal("valid report has diagnostics")
			}
		} else if len(first.Diagnostics) != 1 {
			t.Fatalf("invalid report diagnostics = %#v", first.Diagnostics)
		}
	})
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }

func readFixture(tb testing.TB, name string) []byte {
	tb.Helper()
	data, err := os.ReadFile("../../testdata/reservoir/" + name)
	if err != nil {
		tb.Fatalf("read fixture: %v", err)
	}
	return data
}

func componentArray(input string) string {
	start := strings.Index(input, `"components": [`)
	if start < 0 {
		return ""
	}
	start += len(`"components": `)
	end := strings.Index(input[start:], "\n    ],")
	if end < 0 {
		return ""
	}
	return input[start : start+end+6]
}

func minimalRequest(components string) string {
	return `{"schema":"` + RequestSchema + `","model":{"id":"` + ModelID + `","version":"` + ModelVersion + `"},"quantity_system":"si","closure":"rigid-adiabatic","withdrawal_fraction":0.5,"initial":{"components":` + components + `,"volume_m3":1,"temperature_k":1}}`
}

func assertNear(t *testing.T, label string, got, want, tolerance float64) {
	t.Helper()
	if math.Abs(got-want) > tolerance {
		t.Fatalf("%s = %.17g, want %.17g within %.3g", label, got, want, tolerance)
	}
}
