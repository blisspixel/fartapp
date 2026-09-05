package walkcase

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/blisspixel/fartapp/internal/coupledblowdown"
)

func TestWalkOperationsOnIsothermalFixture(t *testing.T) {
	input := readFixture(t, "isothermal-choked.json")
	for _, operation := range []string{"predict", "simulate", "inspect", "explain", "certify", "branch", "witness"} {
		t.Run(operation, func(t *testing.T) {
			report := Run(input, operation)
			assertPredicted(t, report)
			if report.Operation != operation || report.Model == nil || report.Inputs == nil || report.NumericalPolicy == nil ||
				report.LawContext != "earth.continuum.si@v0alpha1" || len(report.Dimensions) != 6 || len(report.Claims) == 0 {
				t.Fatalf("incomplete %s report", operation)
			}
			for _, claim := range report.Claims {
				if claim.Status != "satisfied-within-roundoff" || math.Abs(claim.Residual) > claim.Tolerance {
					t.Fatalf("claim = %#v", claim)
				}
			}
			encoded, err := json.Marshal(report)
			if err != nil || !bytes.Contains(encoded, []byte(`"residual":`)) || !bytes.Contains(encoded, []byte(`"tolerance":`)) {
				t.Fatalf("residual evidence missing: %v", err)
			}
		})
	}
	predict := Run(input, "predict")
	if predict.EqualizationFraction == nil || *predict.EqualizationFraction <= 0 || predict.InitialRestriction.Regime != "choked" {
		t.Fatal("predict lost the endpoint or initial choking")
	}
	certify := Run(input, "certify")
	if !strings.Contains(strings.Join(certify.Explanation, " "), "does not establish time-step accuracy") {
		t.Fatal("arithmetic certification overstated its evidence")
	}
	branch := Run(input, "branch").Branch
	if branch.BothEqualized || branch.SameMassEndpoint || branch.BaselineStop != "max-time" || branch.VariantStop != "max-time" || branch.Variant == nil || branch.BaselineMassOutKg == branch.VariantMassOutKg {
		t.Fatal("time-budget branch incorrectly claimed a common endpoint")
	}
}

func TestOrdinarySyntheticWalkSupportsAnHonestAreaCounterfactual(t *testing.T) {
	report := Run(readFixture(t, "ordinary-low-pressure.json"), "branch")
	assertPredicted(t, report)
	if report.Initial.PressurePascals < 102254 || report.Initial.PressurePascals > 102256 || report.InitialRestriction.Regime != "subsonic" || report.Signature.ChokedOccurred {
		t.Fatal("fixture escaped its low-pressure, subsonic envelope")
	}
	comparison := report.Branch
	if comparison == nil || !comparison.BothEqualized || !comparison.SameMassEndpoint {
		t.Fatalf("incomparable endpoints: %#v", comparison)
	}
	ratio := comparison.VariantElapsedSeconds / comparison.BaselineElapsedSeconds
	if math.Abs(ratio-0.5) > 1e-10 {
		t.Fatalf("doubling-area duration ratio = %.17g, want 0.5", ratio)
	}
	if report.Nonclaims == nil || !contains(report.Nonclaims.Operation, "reference-pfft-ratification") {
		t.Fatal("synthetic fixture lost the Reference Pfft boundary")
	}
}

func TestHistoryRetainsInitialFinalAndComponentAccounts(t *testing.T) {
	input := editJSON(t, readFixture(t, "isothermal-choked.json"), change{path: "reservoir/components", value: []any{
		map[string]any{"id": "component.b", "mass_kg": 1.0625, "specific_gas_constant_j_per_kg_k": 200, "isochoric_heat_capacity_j_per_kg_k": 400},
		map[string]any{"id": "component.a", "mass_kg": 0.5, "specific_gas_constant_j_per_kg_k": 200, "isochoric_heat_capacity_j_per_kg_k": 400},
	}})
	report := Run(input, "simulate")
	assertPredicted(t, report)
	if len(report.History) != *report.Steps+1 || len(report.History) > coupledblowdown.MaxSteps+1 {
		t.Fatal("history is missing a boundary or exceeds its bound")
	}
	initial, final := report.History[0], report.History[len(report.History)-1]
	if initial.TimeSeconds != 0 || initial.MassKilograms != report.Initial.MassKilograms || final.TimeSeconds != *report.ElapsedSeconds || final.MassKilograms != report.Final.MassKilograms || final.PressurePascals != report.Final.PressurePascals || final.TemperatureKelvin != report.Final.TemperatureKelvin {
		t.Fatal("history lost initial or final state")
	}
	for _, sample := range report.History {
		if sample.Components[0].ID != "component.a" || sample.Components[1].ID != "component.b" || sample.ThrustNewtons != -sample.RecoilNewtons || sample.EffectiveAreaSquareMetres != 0.01 {
			t.Fatal("history lost identity, area, or force bookkeeping")
		}
		if math.Abs(sample.SourceTotalEnthalpyWatts-sample.MassFlowKilogramsPerSecond*600*sample.TemperatureKelvin) > 1e-8 {
			t.Fatal("source total enthalpy was replaced by exit static enthalpy")
		}
		mass := 0.0
		for index, component := range sample.Components {
			mass += component.MassKilograms
			if math.Abs(component.MassKilograms+component.MassOutKilograms-initial.Components[index].MassKilograms) > 1e-14 {
				t.Fatal("component transfer does not close")
			}
		}
		if math.Abs(mass-sample.MassKilograms) > 1e-14 {
			t.Fatal("component masses do not sum to reservoir mass")
		}
	}
	// Component order is nonsemantic and normalized before arithmetic or hashing.
	permuted := editJSON(t, input, change{path: "reservoir/components", value: []any{
		map[string]any{"id": "component.a", "mass_kg": 0.5, "specific_gas_constant_j_per_kg_k": 200, "isochoric_heat_capacity_j_per_kg_k": 400},
		map[string]any{"id": "component.b", "mass_kg": 1.0625, "specific_gas_constant_j_per_kg_k": 200, "isochoric_heat_capacity_j_per_kg_k": 400},
	}})
	if Run(permuted, "witness").Witness != Run(input, "witness").Witness {
		t.Fatal("component order changed normalized evidence")
	}
}

func TestWitnessRequiresRetainedExpectationAndDetectsChangedInputs(t *testing.T) {
	input := readFixture(t, "isothermal-choked.json")
	witness := Run(input, "witness")
	assertPredicted(t, witness)
	if !validDigest(witness.Witness) || !validDigest(witness.InputDigest) || witness.WitnessSchema != WitnessSchema || witness.InputDigestSchema != InputDigestSchema || witness.WitnessAlgorithm != "sha256" {
		t.Fatal("undeclared witness encoding")
	}
	missing := Run(input, "reconstruct")
	assertReason(t, missing, "missing_member")
	if missing.Diagnostics[0].Path != "/expected_witness" {
		t.Fatal("reconstruction silently compared newly generated runs")
	}
	retained := editJSON(t, input, change{path: "expected_witness", value: witness.Witness})
	reconstructed := Run(retained, "reconstruct")
	assertPredicted(t, reconstructed)
	if reconstructed.WitnessMatch == nil || !*reconstructed.WitnessMatch || reconstructed.ExpectedWitness != witness.Witness || reconstructed.ReconstructedWitness != witness.Witness || reconstructed.InputDigest != witness.InputDigest || reconstructed.Inputs.ExpectedWitness != nil {
		t.Fatal("retained expected evidence was not compared")
	}
	// IDs do not change the original seven-field numeric digest, but they must
	// change an input-bound witness even when every numerical result is equal.
	changed := editJSON(t, retained, change{path: "reservoir/components/0/id", value: "component.renamed"})
	mismatch := Run(changed, "reconstruct")
	assertReason(t, mismatch, "witness_mismatch")
	if mismatch.Status != "mismatch" || mismatch.WitnessMatch == nil || *mismatch.WitnessMatch || mismatch.ExpectedWitness != witness.Witness || !validDigest(mismatch.ReconstructedWitness) || mismatch.InputDigest == witness.InputDigest || len(mismatch.History) == 0 {
		t.Fatal("mismatch discarded comparison evidence")
	}
	var indented bytes.Buffer
	if err := json.Indent(&indented, input, "", "    "); err != nil {
		t.Fatal(err)
	}
	if Run(indented.Bytes(), "witness").Witness != witness.Witness {
		t.Fatal("whitespace changed a normalized witness")
	}
}

func TestWitnessBindsCompleteNumericalAccountAndRejectsNonfiniteEncoding(t *testing.T) {
	input := readFixture(t, "isothermal-choked.json")
	baseline := Run(input, "simulate")
	assertPredicted(t, baseline)
	want, wantInput, err := witnessOf(baseline)
	if err != nil {
		t.Fatal(err)
	}
	mutations := []struct {
		name   string
		mutate func(*Report)
	}{
		{"history", func(report *Report) { report.History[0].ExitPressurePascals++ }},
		{"claim", func(report *Report) { report.Claims[0].Residual++ }},
		{"enthalpy", func(report *Report) { *report.EnthalpyOutJoules++ }},
		{"model", func(report *Report) { report.Model.Version = "v0alpha2" }},
		{"implementation", func(report *Report) { report.ImplementationRevision = "another-build" }},
		{"runtime", func(report *Report) { report.NumericalPolicy.Architecture = "another-architecture" }},
		{"undefined-signature", func(report *Report) { report.Signature.FormationNumber = nil }},
	}
	for _, test := range mutations {
		t.Run(test.name, func(t *testing.T) {
			report := Run(input, "simulate")
			test.mutate(&report)
			got, gotInput, err := witnessOf(report)
			if err != nil || got == want || gotInput != wantInput {
				t.Fatalf("mutation did not change only account digest: %v", err)
			}
		})
	}
	bad := Run(input, "simulate")
	bad.History[0].MassKilograms = math.NaN()
	if _, _, err := witnessOf(bad); err == nil {
		t.Fatal("nonfinite account was hashed")
	}
	bad = Run(input, "simulate")
	*bad.Inputs.Reservoir.VolumeCubicMetres = math.Inf(1)
	if _, _, err := witnessOf(bad); err == nil {
		t.Fatal("nonfinite input was hashed")
	}
}

func TestIncompatibleLawCannotRunSICalculation(t *testing.T) {
	for _, operation := range []string{"predict", "simulate", "inspect", "explain", "branch", "certify", "witness", "reconstruct"} {
		report := Run(readFixture(t, "atemporal-no-dimension.json"), operation)
		assertReason(t, report, "incompatible_law_context")
		if report.Initial != nil || report.Final != nil || report.Dimensions != nil || report.Claims != nil || report.History != nil {
			t.Fatalf("%s fabricated an atemporal physical account", operation)
		}
	}
	standalone := editJSON(t, readFixture(t, "isothermal-choked.json"), change{path: "law_context", remove: true})
	report := Run(standalone, "inspect")
	assertPredicted(t, report)
	if report.Dimensions != nil || report.LawContext != "" || report.QuantitySystem != "si" || report.Model == nil {
		t.Fatal("standalone model fabricated a context")
	}
}

func TestClosedAndAsymptoticEndpointsAreDistinguished(t *testing.T) {
	base := readFixture(t, "isothermal-choked.json")
	closed := editJSON(t, base, change{path: "restriction/area/prescribed_m2", value: 0})
	predict := Run(closed, "predict")
	assertPredicted(t, predict)
	if predict.EndpointReachability != "unreachable-closed-restriction" || *predict.EqualizationFraction != 0 || *predict.MassOutKilograms != 0 || !reflect.DeepEqual(predict.Initial, predict.Final) {
		t.Fatal("closed restriction claimed reachable equalization")
	}
	simulation := Run(closed, "explain")
	assertPredicted(t, simulation)
	if simulation.Stop != "no-flow" || len(simulation.History) != 1 || simulation.Signature.FormationNumber != nil {
		t.Fatal("zero-flow identity lost its sample or undefined signature")
	}
	equal := editJSON(t, base, change{path: "restriction/back_pressure_pa", value: 125000})
	predict = Run(equal, "predict")
	assertPredicted(t, predict)
	if predict.EndpointReachability != "already-equalized" || *predict.MassOutKilograms != 0 {
		t.Fatal("equal pressure was not an identity")
	}
	compliant := editJSON(t, base, change{path: "restriction/area", value: map[string]any{"law": "linear-compliance", "prescribed_m2": 0, "compliance_m2_per_pa": 1e-7, "maximum_m2": 0.01}})
	predict = Run(compliant, "predict")
	assertPredicted(t, predict)
	if predict.EndpointReachability != "asymptotic-limit" || !strings.Contains(strings.Join(predict.Explanation, " "), "infinite-time") {
		t.Fatal("compliance claimed finite-time equalization")
	}
	assertPredicted(t, Run(compliant, "simulate"))
}

func FuzzRun(f *testing.F) {
	f.Add(readFixture(f, "isothermal-choked.json"), "simulate")
	f.Add([]byte(`{"schema":"x"}`), "predict")
	f.Fuzz(func(t *testing.T, input []byte, operation string) {
		if len(input) > MaxInputBytes+32 {
			t.Skip()
		}
		first, second := Run(input, operation), Run(input, operation)
		if !reflect.DeepEqual(first, second) {
			t.Fatal("walk was not deterministic")
		}
		if _, err := json.Marshal(first); err != nil {
			t.Fatalf("report cannot be encoded: %v", err)
		}
		if len(first.History) > coupledblowdown.MaxSteps+1 {
			t.Fatal("unbounded history")
		}
	})
}

type change struct {
	path   string
	value  any
	remove bool
}

func editJSON(t testing.TB, data []byte, changes ...change) []byte {
	t.Helper()
	var document map[string]any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	for _, edit := range changes {
		parts := strings.Split(edit.path, "/")
		var parent any = document
		for _, part := range parts[:len(parts)-1] {
			switch node := parent.(type) {
			case map[string]any:
				parent = node[part]
			case []any:
				index, err := strconv.Atoi(part)
				if err != nil {
					t.Fatal(err)
				}
				parent = node[index]
			default:
				t.Fatalf("invalid test path %s", edit.path)
			}
		}
		object, ok := parent.(map[string]any)
		if !ok {
			t.Fatalf("test path parent is not an object: %s", edit.path)
		}
		key := parts[len(parts)-1]
		if edit.remove {
			delete(object, key)
		} else {
			object[key] = edit.value
		}
	}
	encoded, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func readFixture(t testing.TB, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "walk", name))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return data
}

func assertPredicted(t *testing.T, report Report) {
	t.Helper()
	if !report.Predicted() {
		t.Fatalf("operation failed: status=%s diagnostics=%#v", report.Status, report.Diagnostics)
	}
}

func assertReason(t *testing.T, report Report, reason string) {
	t.Helper()
	if report.Predicted() || len(report.Diagnostics) != 1 || report.Diagnostics[0].ReasonCode != reason {
		t.Fatalf("status=%s diagnostics=%#v, want %s", report.Status, report.Diagnostics, reason)
	}
}

func contains(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}
