package restrictionprediction

import (
	"bytes"
	"errors"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestPredictChokedClosedForm(t *testing.T) {
	report := Predict(readFixture(t, "gamma15-choked.json"))
	if !report.Predicted() {
		t.Fatalf("Predict: %#v", report)
	}
	if report.RequestSchema != RequestSchema || report.Model.ID != ModelID ||
		report.Model.Version != ModelVersion || report.ImplementationRevision != ImplementationRevision ||
		report.QuantitySystem != QuantitySystem || report.Flow.Regime != "choked" {
		t.Fatalf("identity fields = %#v", report)
	}
	assertNear(t, "exit pressure", report.Flow.ExitPressurePascals, 64000, 1e-10)
	assertNear(t, "exit temperature", report.Flow.ExitTemperatureKelvin, 320, 1e-12)
	assertNear(t, "thrust", report.Flow.ThrustNewtons, 1100, 1e-9)
	assertNear(t, "recoil", report.Flow.RecoilNewtons, -1100, 1e-9)
	assertNear(t, "area", report.Area.EffectiveSquareMetres, 0.01, 0)
	if report.Area.Law != "prescribed" || report.Area.ComplianceSquareMetresPa != nil {
		t.Fatalf("area projection = %#v", report.Area)
	}
	for _, claim := range report.Claims {
		if claim.Status != "satisfied-within-roundoff" || math.Abs(claim.Residual) > claim.Tolerance {
			t.Errorf("claim = %#v", claim)
		}
	}
	if len(report.Assumptions) != 8 || report.Nonclaims == nil ||
		len(report.Nonclaims.Model) != 6 || len(report.ValidationEnvironment.AmbientInputs) != 0 {
		t.Fatalf("scope disclosure = %#v", report)
	}
}

func TestPredictSubsonicAndComplianceFixtures(t *testing.T) {
	subsonic := Predict(readFixture(t, "gamma15-subsonic.json"))
	if !subsonic.Predicted() || subsonic.Flow.Regime != "subsonic" {
		t.Fatalf("subsonic = %#v", subsonic)
	}
	assertNear(t, "mach", subsonic.Flow.ThroatMach, 2.0/3.0, 1e-15)
	assertNear(t, "exit pressure", subsonic.Flow.ExitPressurePascals, 729000, 0)
	if subsonic.Flow.MassFlowKilogramsPerSecond >= subsonic.Flow.SonicMassFlowKilogramsPerS {
		t.Fatal("subsonic mass flow reached the sonic bound")
	}

	compliant := Predict(readFixture(t, "linear-compliance-choked.json"))
	if !compliant.Predicted() || compliant.Area.Law != "linear-compliance" ||
		compliant.Area.ComplianceSquareMetresPa == nil || compliant.Area.MaximumSquareMetres == nil {
		t.Fatalf("compliance = %#v", compliant)
	}
	assertNear(t, "effective area", compliant.Area.EffectiveSquareMetres, 0.0085, 1e-15)
	if compliant.Flow.Regime != "choked" {
		t.Fatalf("compliance regime = %s", compliant.Flow.Regime)
	}

	ordinary := Predict(readFixture(t, "ordinary-pressure-subsonic.json"))
	if !ordinary.Predicted() || ordinary.Flow.Regime != "subsonic" {
		t.Fatalf("ordinary = %#v", ordinary)
	}
	if ordinary.Flow.BackPressureRatio <= ordinary.Flow.CriticalPressureRatio {
		t.Fatal("ordinary Earth-biological comparison drifted into choking")
	}
}

func TestPredictRejectsInvalidDocuments(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		reason string
	}{
		{name: "empty", input: "", reason: "empty_input"},
		{name: "schema", input: `{"schema":"wrong"}`, reason: "unsupported_schema"},
		{name: "model", input: `{
			"schema":"fart.restriction-prediction-request/v0alpha1",
			"model":{"id":"other","version":"v0alpha1"},
			"quantity_system":"si","back_pressure_pa":1,"discharge_coefficient":1,
			"stagnation":{"pressure_pa":2,"temperature_k":3,"specific_gas_constant_j_per_kg_k":4,"heat_capacity_ratio":1.4},
			"area":{"law":"prescribed","prescribed_m2":0.01}
		}`, reason: "unsupported_model_revision"},
		{name: "quantity", input: `{
			"schema":"fart.restriction-prediction-request/v0alpha1",
			"model":{"id":"continuum.quasi-steady-isentropic-converging-restriction","version":"v0alpha1"},
			"quantity_system":"cgs","back_pressure_pa":1,"discharge_coefficient":1,
			"stagnation":{"pressure_pa":2,"temperature_k":3,"specific_gas_constant_j_per_kg_k":4,"heat_capacity_ratio":1.4},
			"area":{"law":"prescribed","prescribed_m2":0.01}
		}`, reason: "unsupported_quantity_system"},
		{name: "missing back", input: `{
			"schema":"fart.restriction-prediction-request/v0alpha1",
			"model":{"id":"continuum.quasi-steady-isentropic-converging-restriction","version":"v0alpha1"},
			"quantity_system":"si","discharge_coefficient":1,
			"stagnation":{"pressure_pa":2,"temperature_k":3,"specific_gas_constant_j_per_kg_k":4,"heat_capacity_ratio":1.4},
			"area":{"law":"prescribed","prescribed_m2":0.01}
		}`, reason: "missing_member"},
		{name: "missing Cd", input: `{
			"schema":"fart.restriction-prediction-request/v0alpha1",
			"model":{"id":"continuum.quasi-steady-isentropic-converging-restriction","version":"v0alpha1"},
			"quantity_system":"si","back_pressure_pa":1,
			"stagnation":{"pressure_pa":2,"temperature_k":3,"specific_gas_constant_j_per_kg_k":4,"heat_capacity_ratio":1.4},
			"area":{"law":"prescribed","prescribed_m2":0.01}
		}`, reason: "missing_member"},
		{name: "missing stagnation pressure", input: `{
			"schema":"fart.restriction-prediction-request/v0alpha1",
			"model":{"id":"continuum.quasi-steady-isentropic-converging-restriction","version":"v0alpha1"},
			"quantity_system":"si","back_pressure_pa":1,"discharge_coefficient":1,
			"stagnation":{"temperature_k":3,"specific_gas_constant_j_per_kg_k":4,"heat_capacity_ratio":1.4},
			"area":{"law":"prescribed","prescribed_m2":0.01}
		}`, reason: "missing_member"},
		{name: "nonpositive pressure", input: `{
			"schema":"fart.restriction-prediction-request/v0alpha1",
			"model":{"id":"continuum.quasi-steady-isentropic-converging-restriction","version":"v0alpha1"},
			"quantity_system":"si","back_pressure_pa":1,"discharge_coefficient":1,
			"stagnation":{"pressure_pa":0,"temperature_k":3,"specific_gas_constant_j_per_kg_k":4,"heat_capacity_ratio":1.4},
			"area":{"law":"prescribed","prescribed_m2":0.01}
		}`, reason: "nonpositive_quantity"},
		{name: "prescribed extra compliance", input: `{
			"schema":"fart.restriction-prediction-request/v0alpha1",
			"model":{"id":"continuum.quasi-steady-isentropic-converging-restriction","version":"v0alpha1"},
			"quantity_system":"si","back_pressure_pa":1,"discharge_coefficient":1,
			"stagnation":{"pressure_pa":2,"temperature_k":3,"specific_gas_constant_j_per_kg_k":4,"heat_capacity_ratio":1.4},
			"area":{"law":"prescribed","prescribed_m2":0.01,"compliance_m2_per_pa":1e-8}
		}`, reason: "unexpected_member"},
		{name: "compliance missing coefficient", input: `{
			"schema":"fart.restriction-prediction-request/v0alpha1",
			"model":{"id":"continuum.quasi-steady-isentropic-converging-restriction","version":"v0alpha1"},
			"quantity_system":"si","back_pressure_pa":1,"discharge_coefficient":1,
			"stagnation":{"pressure_pa":2,"temperature_k":3,"specific_gas_constant_j_per_kg_k":4,"heat_capacity_ratio":1.4},
			"area":{"law":"linear-compliance","prescribed_m2":0.01,"maximum_m2":0.02}
		}`, reason: "missing_member"},
		{name: "maximum below prescribed", input: `{
			"schema":"fart.restriction-prediction-request/v0alpha1",
			"model":{"id":"continuum.quasi-steady-isentropic-converging-restriction","version":"v0alpha1"},
			"quantity_system":"si","back_pressure_pa":1,"discharge_coefficient":1,
			"stagnation":{"pressure_pa":2,"temperature_k":3,"specific_gas_constant_j_per_kg_k":4,"heat_capacity_ratio":1.4},
			"area":{"law":"linear-compliance","prescribed_m2":0.02,"compliance_m2_per_pa":0,"maximum_m2":0.01}
		}`, reason: "invalid_area_law"},
		{name: "unknown member", input: `{
			"schema":"fart.restriction-prediction-request/v0alpha1",
			"model":{"id":"continuum.quasi-steady-isentropic-converging-restriction","version":"v0alpha1"},
			"quantity_system":"si","back_pressure_pa":1,"discharge_coefficient":1,
			"stagnation":{"pressure_pa":2,"temperature_k":3,"specific_gas_constant_j_per_kg_k":4,"heat_capacity_ratio":1.4},
			"area":{"law":"prescribed","prescribed_m2":0.01},"earth":true
		}`, reason: "document_shape_invalid"},
		{name: "area law", input: `{
			"schema":"fart.restriction-prediction-request/v0alpha1",
			"model":{"id":"continuum.quasi-steady-isentropic-converging-restriction","version":"v0alpha1"},
			"quantity_system":"si","back_pressure_pa":1,"discharge_coefficient":1,
			"stagnation":{"pressure_pa":2,"temperature_k":3,"specific_gas_constant_j_per_kg_k":4,"heat_capacity_ratio":1.4},
			"area":{"law":"oscillator","prescribed_m2":0.01}
		}`, reason: "unsupported_area_law"},
		{name: "prescribed extra", input: `{
			"schema":"fart.restriction-prediction-request/v0alpha1",
			"model":{"id":"continuum.quasi-steady-isentropic-converging-restriction","version":"v0alpha1"},
			"quantity_system":"si","back_pressure_pa":1,"discharge_coefficient":1,
			"stagnation":{"pressure_pa":2,"temperature_k":3,"specific_gas_constant_j_per_kg_k":4,"heat_capacity_ratio":1.4},
			"area":{"law":"prescribed","prescribed_m2":0.01,"maximum_m2":0.02}
		}`, reason: "unexpected_member"},
		{name: "compliance missing max", input: `{
			"schema":"fart.restriction-prediction-request/v0alpha1",
			"model":{"id":"continuum.quasi-steady-isentropic-converging-restriction","version":"v0alpha1"},
			"quantity_system":"si","back_pressure_pa":1,"discharge_coefficient":1,
			"stagnation":{"pressure_pa":2,"temperature_k":3,"specific_gas_constant_j_per_kg_k":4,"heat_capacity_ratio":1.4},
			"area":{"law":"linear-compliance","prescribed_m2":0.01,"compliance_m2_per_pa":1e-8}
		}`, reason: "missing_member"},
		{name: "adverse", input: `{
			"schema":"fart.restriction-prediction-request/v0alpha1",
			"model":{"id":"continuum.quasi-steady-isentropic-converging-restriction","version":"v0alpha1"},
			"quantity_system":"si","back_pressure_pa":2000,"discharge_coefficient":1,
			"stagnation":{"pressure_pa":1000,"temperature_k":300,"specific_gas_constant_j_per_kg_k":200,"heat_capacity_ratio":1.4},
			"area":{"law":"prescribed","prescribed_m2":0.01}
		}`, reason: "adverse_pressure"},
		{name: "gamma", input: `{
			"schema":"fart.restriction-prediction-request/v0alpha1",
			"model":{"id":"continuum.quasi-steady-isentropic-converging-restriction","version":"v0alpha1"},
			"quantity_system":"si","back_pressure_pa":1,"discharge_coefficient":1,
			"stagnation":{"pressure_pa":2,"temperature_k":3,"specific_gas_constant_j_per_kg_k":4,"heat_capacity_ratio":1},
			"area":{"law":"prescribed","prescribed_m2":0.01}
		}`, reason: "invalid_heat_capacity_ratio"},
		{name: "Cd", input: `{
			"schema":"fart.restriction-prediction-request/v0alpha1",
			"model":{"id":"continuum.quasi-steady-isentropic-converging-restriction","version":"v0alpha1"},
			"quantity_system":"si","back_pressure_pa":1,"discharge_coefficient":2,
			"stagnation":{"pressure_pa":2,"temperature_k":3,"specific_gas_constant_j_per_kg_k":4,"heat_capacity_ratio":1.4},
			"area":{"law":"prescribed","prescribed_m2":0.01}
		}`, reason: "invalid_discharge_coefficient"},
		{name: "negative area", input: `{
			"schema":"fart.restriction-prediction-request/v0alpha1",
			"model":{"id":"continuum.quasi-steady-isentropic-converging-restriction","version":"v0alpha1"},
			"quantity_system":"si","back_pressure_pa":1,"discharge_coefficient":1,
			"stagnation":{"pressure_pa":2,"temperature_k":3,"specific_gas_constant_j_per_kg_k":4,"heat_capacity_ratio":1.4},
			"area":{"law":"prescribed","prescribed_m2":-1}
		}`, reason: "negative_area"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			report := Predict([]byte(test.input))
			if report.Predicted() || len(report.Diagnostics) != 1 || report.Diagnostics[0].ReasonCode != test.reason {
				t.Fatalf("report = %#v, want reason %q", report, test.reason)
			}
			if report.Flow != nil || report.BackPressurePascals != nil {
				t.Fatalf("invalid report fabricated flow: %#v", report)
			}
		})
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
	if _, err := ReadBounded(nil); !errors.Is(err, ErrNilInput) {
		t.Fatalf("nil error = %v", err)
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
	if got := classifyModelError(errors.New("other")); got != "numerical_domain_error" {
		t.Fatalf("fallback classification = %q", got)
	}
}

func FuzzPredict(f *testing.F) {
	f.Add(readFixture(f, "gamma15-choked.json"))
	f.Add(readFixture(f, "linear-compliance-choked.json"))
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

func readFixture(t testing.TB, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "restriction", name))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return data
}

func assertNear(t *testing.T, name string, got, want, tolerance float64) {
	t.Helper()
	if math.IsNaN(got) || math.IsInf(got, 0) || math.Abs(got-want) > tolerance {
		t.Fatalf("%s = %g, want %g ± %g", name, got, want, tolerance)
	}
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }
