package restrictionhistoryprediction

import (
	"bytes"
	"errors"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestPredictConstantChokedHistory(t *testing.T) {
	report := Predict(readFixture(t, "gamma15-choked-history.json"))
	if !report.Predicted() {
		t.Fatalf("Predict: %#v", report)
	}
	if len(report.Samples) != 2 || report.Samples[0].Regime != "choked" || report.Totals == nil {
		t.Fatalf("projection = %#v", report)
	}
	mdot := 0.01 * math.Sqrt(96000)
	assertNear(t, "mass", report.Totals.MassOutKilograms, mdot*0.01, 1e-12)
	assertNear(t, "impulse", report.Totals.ImpulseNewtonSeconds, 11, 1e-12)
	if report.Claims[0].Status != "satisfied-within-roundoff" {
		t.Fatalf("claim = %#v", report.Claims[0])
	}
	if len(report.ValidationEnvironment.AmbientInputs) != 0 {
		t.Fatal("ambient inputs were invented")
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
		{name: "missing samples", input: `{
			"schema":"fart.restriction-history-request/v0alpha1",
			"model":{"id":"continuum.quasi-steady-isentropic-converging-restriction","version":"v0alpha1"},
			"quantity_system":"si","back_pressure_pa":1,"discharge_coefficient":1,
			"stagnation":{"pressure_pa":2,"temperature_k":3,"specific_gas_constant_j_per_kg_k":4,"heat_capacity_ratio":1.4}
		}`, reason: "missing_member"},
		{name: "repeated time", input: `{
			"schema":"fart.restriction-history-request/v0alpha1",
			"model":{"id":"continuum.quasi-steady-isentropic-converging-restriction","version":"v0alpha1"},
			"quantity_system":"si","back_pressure_pa":50000,"discharge_coefficient":1,
			"stagnation":{"pressure_pa":125000,"temperature_k":400,"specific_gas_constant_j_per_kg_k":200,"heat_capacity_ratio":1.5},
			"samples":[{"time_s":0,"prescribed_m2":0.01},{"time_s":0,"prescribed_m2":0.01}]
		}`, reason: "invalid_time"},
		{name: "model", input: `{
			"schema":"fart.restriction-history-request/v0alpha1",
			"model":{"id":"other","version":"v0alpha1"},
			"quantity_system":"si","back_pressure_pa":1,"discharge_coefficient":1,
			"stagnation":{"pressure_pa":2,"temperature_k":3,"specific_gas_constant_j_per_kg_k":4,"heat_capacity_ratio":1.4},
			"samples":[{"time_s":0,"prescribed_m2":0}]
		}`, reason: "unsupported_model_revision"},
		{name: "quantity", input: `{
			"schema":"fart.restriction-history-request/v0alpha1",
			"model":{"id":"continuum.quasi-steady-isentropic-converging-restriction","version":"v0alpha1"},
			"quantity_system":"cgs","back_pressure_pa":1,"discharge_coefficient":1,
			"stagnation":{"pressure_pa":2,"temperature_k":3,"specific_gas_constant_j_per_kg_k":4,"heat_capacity_ratio":1.4},
			"samples":[{"time_s":0,"prescribed_m2":0}]
		}`, reason: "unsupported_quantity_system"},
		{name: "negative area", input: `{
			"schema":"fart.restriction-history-request/v0alpha1",
			"model":{"id":"continuum.quasi-steady-isentropic-converging-restriction","version":"v0alpha1"},
			"quantity_system":"si","back_pressure_pa":1,"discharge_coefficient":1,
			"stagnation":{"pressure_pa":2,"temperature_k":3,"specific_gas_constant_j_per_kg_k":4,"heat_capacity_ratio":1.4},
			"samples":[{"time_s":0,"prescribed_m2":-1}]
		}`, reason: "negative_area"},
		{name: "missing sample time", input: `{
			"schema":"fart.restriction-history-request/v0alpha1",
			"model":{"id":"continuum.quasi-steady-isentropic-converging-restriction","version":"v0alpha1"},
			"quantity_system":"si","back_pressure_pa":1,"discharge_coefficient":1,
			"stagnation":{"pressure_pa":2,"temperature_k":3,"specific_gas_constant_j_per_kg_k":4,"heat_capacity_ratio":1.4},
			"samples":[{"prescribed_m2":0}]
		}`, reason: "missing_member"},
		{name: "gamma", input: `{
			"schema":"fart.restriction-history-request/v0alpha1",
			"model":{"id":"continuum.quasi-steady-isentropic-converging-restriction","version":"v0alpha1"},
			"quantity_system":"si","back_pressure_pa":1,"discharge_coefficient":1,
			"stagnation":{"pressure_pa":2,"temperature_k":3,"specific_gas_constant_j_per_kg_k":4,"heat_capacity_ratio":1},
			"samples":[{"time_s":0,"prescribed_m2":0}]
		}`, reason: "invalid_heat_capacity_ratio"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			report := Predict([]byte(test.input))
			if report.Predicted() || len(report.Diagnostics) != 1 || report.Diagnostics[0].ReasonCode != test.reason {
				t.Fatalf("report = %#v, want %q", report, test.reason)
			}
		})
	}
	if got := Predict(bytes.Repeat([]byte{' '}, MaxInputBytes+1)).Diagnostics[0].ReasonCode; got != "input_too_large" {
		t.Fatalf("oversized = %q", got)
	}
}

func TestReadBoundedAndFailureEnvelope(t *testing.T) {
	got, err := ReadBounded(bytes.NewReader([]byte("x")))
	if err != nil || string(got) != "x" {
		t.Fatalf("ReadBounded = (%q, %v)", got, err)
	}
	if _, err := ReadBounded(nil); !errors.Is(err, ErrNilInput) {
		t.Fatalf("nil = %v", err)
	}
	if (Report{}).Predicted() || InputFailure("input_unavailable", "input_stream").Predicted() {
		t.Fatal("invalid report was valid")
	}
	if _, err := ReadBounded(errorReader{}); err == nil {
		t.Fatal("ReadBounded accepted a reader failure")
	}
	if _, err := ReadBounded(bytes.NewReader(bytes.Repeat([]byte{'x'}, MaxInputBytes+1))); !errors.Is(err, ErrInputTooLarge) {
		t.Fatalf("large = %v", err)
	}
	failure := InputFailure("input_not_found", "input_source_reference")
	if len(failure.ValidationEnvironment.ConsultedInputs) != 1 {
		t.Fatalf("input failure provenance = %#v", failure)
	}
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }

func FuzzPredict(f *testing.F) {
	f.Add(readFixture(f, "gamma15-choked-history.json"))
	f.Add([]byte(`{"schema":"x"}`))
	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) > MaxInputBytes+32 {
			t.Skip()
		}
		first := Predict(input)
		second := Predict(input)
		if !reflect.DeepEqual(first, second) {
			t.Fatal("history prediction was not deterministic")
		}
		if first.Predicted() && len(first.Diagnostics) != 0 {
			t.Fatal("valid report has diagnostics")
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
	if math.Abs(got-want) > tolerance {
		t.Fatalf("%s = %g, want %g ± %g", name, got, want, tolerance)
	}
}
