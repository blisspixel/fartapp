package walkcase

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestNormalizeRequestFingerprintUsesExistingExactInputEnvelope(t *testing.T) {
	const input = `{"schema":"fart.walk-case/v0alpha1","model":{"id":"continuum.quasi-steady-coupled-blowdown","version":"v0alpha1"},"quantity_system":"si","closure":"rigid-isothermal","reservoir":{"components":[{"id":"b","mass_kg":1,"specific_gas_constant_j_per_kg_k":250,"isochoric_heat_capacity_j_per_kg_k":500},{"id":"a","mass_kg":1,"specific_gas_constant_j_per_kg_k":250,"isochoric_heat_capacity_j_per_kg_k":500}],"volume_m3":1,"temperature_k":300},"restriction":{"back_pressure_pa":50000,"discharge_coefficient":1,"area":{"law":"prescribed","prescribed_m2":0.01}},"step":{"max_withdrawal_fraction_per_step":0.01,"max_steps":1}}`
	// This literal fixes the intended normalization independently of the parser:
	// sorted component IDs and explicit default time, under the input envelope.
	const normalizedEnvelope = `{"schema":"fart.walk-normalized-input/v0alpha1","inputs":{"schema":"fart.walk-case/v0alpha1","model":{"id":"continuum.quasi-steady-coupled-blowdown","version":"v0alpha1"},"quantity_system":"si","closure":"rigid-isothermal","reservoir":{"components":[{"id":"a","mass_kg":1,"specific_gas_constant_j_per_kg_k":250,"isochoric_heat_capacity_j_per_kg_k":500},{"id":"b","mass_kg":1,"specific_gas_constant_j_per_kg_k":250,"isochoric_heat_capacity_j_per_kg_k":500}],"volume_m3":1,"temperature_k":300},"restriction":{"back_pressure_pa":50000,"discharge_coefficient":1,"area":{"law":"prescribed","prescribed_m2":0.01}},"step":{"max_withdrawal_fraction_per_step":0.01,"max_steps":1,"max_time_s":0}}}`
	want := retainedTestSHA([]byte(normalizedEnvelope))
	for _, raw := range [][]byte{
		[]byte(input),
		[]byte(" \n" + strings.ReplaceAll(input, `"mass_kg":1`, `"mass_kg":1.00e0`) + "\n"),
		editJSON(t, []byte(input), change{path: "expected_witness", value: strings.Repeat("a", 64)}, change{path: "step/max_time_s", value: 0}),
	} {
		before := bytes.Clone(raw)
		got, err := NormalizeRequestFingerprint(raw)
		if err != nil || got != want || !bytes.Equal(before, raw) {
			t.Fatalf("fingerprint=(%s,%v), want=%s; input changed=%t", got, err, want, !bytes.Equal(before, raw))
		}
		if witness := Run(raw, "witness"); !witness.Predicted() || witness.InputDigest != want {
			t.Fatal("normalization API diverged from the existing witness wire contract")
		}
	}
	changed := editJSON(t, []byte(input), change{path: "reservoir/components/0/id", value: "renamed"})
	if got, err := NormalizeRequestFingerprint(changed); err != nil || got == want {
		t.Fatalf("component identity did not change the fingerprint: %s, %v", got, err)
	}
	if _, err := NormalizeRequestFingerprint(bytes.Repeat([]byte{' '}, MaxInputBytes+1)); !errors.Is(err, ErrInputTooLarge) {
		t.Fatalf("oversized request: %v", err)
	}
	if _, err := NormalizeRequestFingerprint([]byte(`{"schema":"unsupported"}`)); err == nil {
		t.Fatal("invalid request acquired a fingerprint")
	}
}

func TestRetainedWitnessRoundTripAndImmutableReconstruction(t *testing.T) {
	raw := readFixture(t, "ordinary-low-pressure.json")
	report := Run(raw, "witness")
	assertPredicted(t, report)
	encoded := retainedTestMarshal(t, report)
	retained, err := VerifyRetainedWitnessReport(encoded)
	if err != nil || retained.Witness() != report.Witness || retained.InputDigest() != report.InputDigest {
		t.Fatalf("verification=(%#v,%v)", retained, err)
	}
	for index := range encoded {
		encoded[index] = 0
	}
	report.Witness = "changed"
	report.History[0].Components[0].ID = "changed"
	first := retained.Reconstruct(raw)
	assertPredicted(t, first)
	if first.Operation != "reconstruct" || first.WitnessMatch == nil || !*first.WitnessMatch || first.ExpectedWitness != retained.Witness() {
		t.Fatal("retained target was not used for the new calculation")
	}
	first.History[0].Components[0].ID = "changed-again"
	second := retained.Reconstruct(raw)
	assertPredicted(t, second)
	if second.History[0].Components[0].ID != "synthetic.air-like" || second.Witness != retained.Witness() {
		t.Fatal("returned report shared mutable state with retained comparison")
	}
	withAuthoredTarget := editJSON(t, raw, change{path: "expected_witness", value: strings.Repeat("0", 64)})
	before := bytes.Clone(withAuthoredTarget)
	third := retained.Reconstruct(withAuthoredTarget)
	if third.WitnessMatch == nil || !*third.WitnessMatch || !bytes.Equal(before, withAuthoredTarget) {
		t.Fatal("authored target replaced retained evidence or request bytes were mutated")
	}
	unrelated := editJSON(t, raw, change{path: "reservoir/components/0/id", value: "unrelated"})
	rejected := retained.Reconstruct(unrelated)
	assertReason(t, rejected, "request_fingerprint_mismatch")
	if rejected.History != nil || rejected.Final != nil {
		t.Fatal("an unrelated authored request was solved before rejection")
	}
	assertReason(t, retained.Reconstruct([]byte("{}")), "unsupported_schema")
	assertReason(t, retained.Reconstruct(bytes.Repeat([]byte{' '}, MaxInputBytes+1)), "input_too_large")
	assertReason(t, (RetainedWitness{}).Reconstruct(raw), "invalid_retained_witness")
}

func TestPreviousImplementationRetainsIntegrityWithoutPretendingReconstructionMatches(t *testing.T) {
	raw := readFixture(t, "ordinary-low-pressure.json")
	report := Run(raw, "witness")
	assertPredicted(t, report)
	report.ImplementationRevision = "go-oracle.walk/v0alpha3"
	retainedTestRehash(t, &report)
	encoded := retainedTestMarshal(t, report)
	before := bytes.Clone(encoded)
	retained, err := VerifyRetainedWitnessReport(encoded)
	if err != nil {
		t.Fatal(err)
	}
	fresh := retained.Reconstruct(raw)
	if fresh.WitnessMatch == nil || *fresh.WitnessMatch || fresh.ExpectedWitness != report.Witness ||
		fresh.ImplementationRevision != ImplementationRevision {
		t.Fatalf("previous implementation was silently upgraded: %#v", fresh)
	}
	if !bytes.Equal(before, encoded) {
		t.Fatal("verification changed retained bytes")
	}
	report.ImplementationRevision = "go-oracle.walk/v0alpha2"
	retainedTestRehash(t, &report)
	if _, err := VerifyRetainedWitnessReport(retainedTestMarshal(t, report)); err == nil {
		t.Fatal("unreviewed older profile was accepted")
	}
}

func TestRetainedVerificationIsSelfConsistencyNotProofOfExecution(t *testing.T) {
	raw := readFixture(t, "ordinary-low-pressure.json")
	for _, mutation := range []struct {
		name   string
		mutate func(*Report)
	}{
		{"changed interior result", func(report *Report) { report.History[1].ExitSpeedMetresPerSecond++ }},
		{"different retained runtime", func(report *Report) { report.NumericalPolicy.GoVersion = "go1.0-retained-profile" }},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			report := Run(raw, "witness")
			original := report.Witness
			mutation.mutate(&report)
			if _, err := VerifyRetainedWitnessReport(retainedTestMarshal(t, report)); err == nil {
				t.Fatal("changed account retained a stale accepted witness")
			}
			// Anyone can recompute unkeyed hashes. This must not become a fake
			// execution proof or a hidden solver run during archive verification.
			retainedTestRehash(t, &report)
			retained, err := VerifyRetainedWitnessReport(retainedTestMarshal(t, report))
			if err != nil || retained.Witness() == original {
				t.Fatalf("internally consistent account rejected or hash unbound: %v", err)
			}
			reconstruction := retained.Reconstruct(raw)
			assertReason(t, reconstruction, "witness_mismatch")
			if reconstruction.ExpectedWitness != retained.Witness() || reconstruction.ReconstructedWitness != original || len(reconstruction.History) == 0 {
				t.Fatal("explicit reconstruction did not preserve both comparison values")
			}
		})
	}
}

func TestRetainedReportsAdmitCurrentClosureAndStoppingProfiles(t *testing.T) {
	raw := readFixture(t, "isothermal-choked.json")
	for _, test := range []struct {
		name    string
		changes []change
	}{
		{"time budget", nil},
		{"standalone", []change{{path: "law_context", remove: true}}},
		{"adiabatic", []change{{path: "closure", value: "rigid-adiabatic"}}},
		{"no flow", []change{{path: "restriction/area/prescribed_m2", value: 0}}},
		{"already equal", []change{{path: "restriction/back_pressure_pa", value: 125000}}},
		{"step budget", []change{{path: "step/max_time_s", value: 0}}},
		{"no representable progress", []change{{path: "step/max_withdrawal_fraction_per_step", value: 1e-300}}},
		{"asymptotic", []change{{path: "restriction/area", value: map[string]any{"law": "linear-compliance", "prescribed_m2": 0, "compliance_m2_per_pa": 1e-8, "maximum_m2": 0.01}}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			report := Run(editJSON(t, raw, test.changes...), "witness")
			assertPredicted(t, report)
			if _, err := VerifyRetainedWitnessReport(retainedTestMarshal(t, report)); err != nil {
				t.Fatalf("current %s report rejected: %v", report.Stop, err)
			}
		})
	}
}

func TestRetainedHistoryBoundariesAndNormalizedComponentOrder(t *testing.T) {
	raw := readFixture(t, "isothermal-choked.json")
	components := make([]map[string]any, 64)
	for index := range components {
		components[index] = map[string]any{
			"id": fmt.Sprintf("component.%02d", 63-index), "mass_kg": 1.5625 / 64,
			"specific_gas_constant_j_per_kg_k": 200, "isochoric_heat_capacity_j_per_kg_k": 400,
		}
	}
	withMaximumComponents := editJSON(t, raw, change{path: "reservoir/components", value: components})
	report := Run(withMaximumComponents, "witness")
	assertPredicted(t, report)
	if _, err := VerifyRetainedWitnessReport(retainedTestMarshal(t, report)); err != nil {
		t.Fatalf("64-component history rejected: %v", err)
	}
	report.Inputs.Reservoir.Components[0], report.Inputs.Reservoir.Components[1] = report.Inputs.Reservoir.Components[1], report.Inputs.Reservoir.Components[0]
	retainedTestRehash(t, &report)
	if _, err := VerifyRetainedWitnessReport(retainedTestMarshal(t, report)); err == nil || !strings.Contains(err.Error(), "inputs_not_normalized") {
		t.Fatalf("unordered retained input: %v", err)
	}
	withMaximumSamples := editJSON(t, raw,
		change{path: "step/max_steps", value: 4096},
		change{path: "step/max_time_s", value: 0},
		change{path: "step/max_withdrawal_fraction_per_step", value: 1e-7},
	)
	report = Run(withMaximumSamples, "witness")
	assertPredicted(t, report)
	if len(report.History) != 4097 || report.Stop != "max-steps" {
		t.Fatal("fixture did not exercise the inclusive history limit")
	}
	if _, err := VerifyRetainedWitnessReport(retainedTestMarshal(t, report)); err != nil {
		t.Fatalf("4097-sample history rejected: %v", err)
	}
}

func TestRetainedReportRejectsNoncanonicalOrHostileJSON(t *testing.T) {
	base := retainedTestMarshal(t, Run(readFixture(t, "ordinary-low-pressure.json"), "witness"))
	for _, test := range []struct {
		name string
		raw  []byte
	}{
		{"empty", nil},
		{"empty object", []byte(`{}`)},
		{"array", []byte(`[]`)},
		{"null", []byte(`null`)},
		{"scalar", []byte(`1`)},
		{"trailing value", append(bytes.Clone(base), []byte(` {}`)...)},
		{"trailing newline", append(bytes.Clone(base), '\n')},
		{"space", append([]byte{' '}, base...)},
		{"case alias", bytes.Replace(base, []byte(`"schema":`), []byte(`"Schema":`), 1)},
		{"unknown nested member", bytes.Replace(base, []byte(`"model":{`), []byte(`"model":{"surprise":0,`), 1)},
		{"duplicate", bytes.Replace(base, []byte(`"schema":`), []byte(`"schema":"anything","schema":`), 1)},
		{"missing scalar", bytes.Replace(base, []byte(`"time_s":0,`), nil, 1)},
		{"null scalar", bytes.Replace(base, []byte(`"time_s":0`), []byte(`"time_s":null`), 1)},
		{"null optional", bytes.Replace(base, []byte(`"status":`), []byte(`"branch":null,"status":`), 1)},
		{"invalid UTF8", []byte{'{', '"', 0xff, '"', ':', '0', '}'}},
		{"lone surrogate", []byte(`{"schema":"\ud800"}`)},
		{"nesting", []byte(strings.Repeat("[", 34) + strings.Repeat("]", 34))},
		{"long key", []byte(`{"` + strings.Repeat("x", 129) + `":0}`)},
		{"number overflow", bytes.Replace(base, []byte(`"time_s":0`), []byte(`"time_s":1e999`), 1)},
	} {
		t.Run(test.name, func(t *testing.T) {
			if retained, err := VerifyRetainedWitnessReport(test.raw); err == nil || retained != (RetainedWitness{}) {
				t.Fatalf("invalid report was retained: %#v, %v", retained, err)
			}
		})
	}
	if _, err := VerifyRetainedWitnessReport(bytes.Repeat([]byte{' '}, MaxRetainedReportBytes+1)); !errors.Is(err, ErrRetainedReportTooLarge) {
		t.Fatalf("report byte bound: %v", err)
	}
}

func TestRetainedReportRequiresCompleteCurrentAccountEvenAfterRehash(t *testing.T) {
	base := retainedTestMarshal(t, Run(readFixture(t, "ordinary-low-pressure.json"), "witness"))
	tests := []struct {
		name   string
		mutate func(*Report)
		reason string
	}{
		{"schema", func(r *Report) { r.Schema = "future" }, "unsupported_witness_profile"},
		{"operation", func(r *Report) { r.Operation = "simulate" }, "unsupported_witness_profile"},
		{"status", func(r *Report) { r.Status = "mismatch" }, "unsupported_witness_profile"},
		{"request schema", func(r *Report) { r.RequestSchema = "future" }, "unsupported_witness_profile"},
		{"implementation", func(r *Report) { r.ImplementationRevision = "future" }, "unsupported_witness_profile"},
		{"model", func(r *Report) { r.Model = nil }, "unsupported_witness_profile"},
		{"model revision", func(r *Report) { r.Model.Version = "future" }, "unsupported_witness_profile"},
		{"numerical policy", func(r *Report) { r.NumericalPolicy = nil }, "unsupported_witness_profile"},
		{"inputs", func(r *Report) { r.Inputs = nil }, "unsupported_witness_profile"},
		{"quantity system", func(r *Report) { r.QuantitySystem = "atemporal" }, "unsupported_witness_profile"},
		{"branch", func(r *Report) { r.Branch = &BranchComparison{} }, "unexpected_operation_fields"},
		{"accuracy evidence", func(r *Report) { r.Accuracy = &RefinementEvidence{} }, "unexpected_operation_fields"},
		{"expected target", func(r *Report) { r.ExpectedWitness = strings.Repeat("a", 64) }, "unexpected_operation_fields"},
		{"reconstructed target", func(r *Report) { r.ReconstructedWitness = strings.Repeat("a", 64) }, "unexpected_operation_fields"},
		{"match", func(r *Report) { value := true; r.WitnessMatch = &value }, "unexpected_operation_fields"},
		{"predicted endpoint", func(r *Report) { value := 0.0; r.EqualizationFraction = &value }, "unexpected_operation_fields"},
		{"reachability", func(r *Report) { r.EndpointReachability = "finite" }, "unexpected_operation_fields"},
		{"runtime absent", func(r *Report) { r.NumericalPolicy.GoVersion = "" }, "invalid_runtime_profile"},
		{"runtime blank", func(r *Report) { r.NumericalPolicy.Architecture = " " }, "invalid_runtime_profile"},
		{"runtime controls", func(r *Report) { r.NumericalPolicy.OperatingSystem = "windows\x1b[2J" }, "invalid_runtime_profile"},
		{"runtime oversized", func(r *Report) { r.NumericalPolicy.GoVersion = strings.Repeat("x", 129) }, "invalid_runtime_profile"},
		{"missing final", func(r *Report) { r.Final = nil }, "incomplete_numerical_account"},
		{"missing elapsed", func(r *Report) { r.ElapsedSeconds = nil }, "incomplete_numerical_account"},
		{"missing steps", func(r *Report) { r.Steps = nil }, "incomplete_numerical_account"},
		{"missing enthalpy", func(r *Report) { r.EnthalpyOutJoules = nil }, "incomplete_numerical_account"},
		{"missing heat", func(r *Report) { r.HeatInJoules = nil }, "incomplete_numerical_account"},
		{"missing impulse", func(r *Report) { r.ImpulseNewtonSeconds = nil }, "incomplete_numerical_account"},
		{"missing recoil", func(r *Report) { r.RecoilImpulseNewtonSeconds = nil }, "incomplete_numerical_account"},
		{"missing mass out", func(r *Report) { r.MassOutKilograms = nil }, "incomplete_numerical_account"},
		{"missing initial", func(r *Report) { r.Initial = nil }, "incomplete_numerical_account"},
		{"missing initial restriction", func(r *Report) { r.InitialRestriction = nil }, "incomplete_numerical_account"},
		{"missing signature", func(r *Report) { r.Signature = nil }, "incomplete_numerical_account"},
		{"missing pressure tolerance", func(r *Report) { r.EqualizationPressureTolerancePascals = nil }, "incomplete_numerical_account"},
		{"missing history", func(r *Report) { r.History = nil }, "invalid_history_count"},
		{"negative steps", func(r *Report) { *r.Steps = -1 }, "invalid_history_count"},
		{"too many steps", func(r *Report) { *r.Steps = 4097 }, "invalid_history_count"},
		{"unsupported stop", func(r *Report) { r.Stop = "converged-exactly" }, "unsupported_stop"},
		{"negative mass", func(r *Report) { *r.MassOutKilograms = -1 }, "invalid_account_quantity"},
		{"negative formation", func(r *Report) { *r.Signature.FormationNumber = -1 }, "invalid_account_quantity"},
		{"explanation injection", func(r *Report) { r.Explanation[0] = "untrusted\x1b[2J" }, "unsupported_witness_explanation"},
		{"invalid request", func(r *Report) { r.Inputs.Reservoir.Components[0].ID = "wrong token" }, "invalid_retained_inputs"},
		{"unnormalized expected", func(r *Report) { value := strings.Repeat("a", 64); r.Inputs.ExpectedWitness = &value }, "inputs_not_normalized"},
		{"law mismatch", func(r *Report) { r.LawContext = "atemporal" }, "inconsistent_report_references"},
		{"closure mismatch", func(r *Report) { r.Closure = "rigid-adiabatic" }, "inconsistent_report_references"},
		{"method mismatch", func(r *Report) { r.NumericalPolicy.Method = "exact" }, "inconsistent_report_references"},
		{"dimension missing", func(r *Report) { r.Dimensions = nil }, "inconsistent_report_references"},
		{"assumption missing", func(r *Report) { r.Assumptions = nil }, "inconsistent_report_references"},
		{"nonclaims missing", func(r *Report) { r.Nonclaims = nil }, "inconsistent_report_references"},
		{"validation array null", func(r *Report) { r.ValidationEnvironment.AmbientInputs = nil }, "inconsistent_report_references"},
		{"step budget exceeded", func(r *Report) { *r.Inputs.Step.MaxSteps = 1 }, "account_exceeds_request_budget"},
		{"time budget exceeded", func(r *Report) { *r.Inputs.Step.MaxTimeSeconds = 1e-9 }, "account_exceeds_request_budget"},
		{"claims absent", func(r *Report) { r.Claims = nil }, "incomplete_balance_claims"},
		{"claim identity", func(r *Report) { r.Claims[0].ID = "empirical-proof" }, "invalid_balance_claim"},
		{"claim status", func(r *Report) { r.Claims[0].Status = "not-applicable" }, "invalid_balance_claim"},
		{"negative tolerance", func(r *Report) { r.Claims[0].Tolerance = -1 }, "invalid_balance_claim"},
		{"unsatisfied residual", func(r *Report) { r.Claims[0].Residual = 1 }, "invalid_balance_claim"},
		{"initial time", func(r *Report) { r.History[0].TimeSeconds = 1 }, "inconsistent_history_endpoints"},
		{"missing final sample", func(r *Report) { r.History[len(r.History)-1].TimeSeconds = 0 }, "inconsistent_history_endpoints"},
		{"final summary mismatch", func(r *Report) { r.Final.PressurePascals++ }, "inconsistent_history_endpoints"},
		{"initial restriction mismatch", func(r *Report) { r.InitialRestriction.MassFlowKilogramsPerS++ }, "inconsistent_initial_restriction"},
		{"critical ratio", func(r *Report) { r.InitialRestriction.CriticalPressureRatio = 0 }, "inconsistent_initial_restriction"},
		{"unsupported sample regime", func(r *Report) { r.History[1].Regime = "incompressible" }, "invalid_history_quantity"},
		{"negative sample rate", func(r *Report) { r.History[1].MassFlowKilogramsPerSecond = -1 }, "invalid_history_quantity"},
		{"time reversal", func(r *Report) { r.History[1].TimeSeconds = 0 }, "invalid_history_progress"},
		{"mass increase", func(r *Report) { r.History[1].MassKilograms = 1 }, "invalid_history_progress"},
		{"component missing", func(r *Report) { r.History[1].Components = nil }, "invalid_history_component_count"},
		{"component identity", func(r *Report) { r.History[1].Components[0].ID = "different" }, "invalid_history_component"},
		{"initial component mass", func(r *Report) { r.History[0].Components[0].MassKilograms++ }, "invalid_history_component"},
		{"initial component withdrawal", func(r *Report) { r.History[0].Components[0].MassOutKilograms++ }, "invalid_history_component"},
		{"negative component out", func(r *Report) { r.History[1].Components[0].MassOutKilograms = -1 }, "invalid_history_component"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var report Report
			if err := json.Unmarshal(base, &report); err != nil {
				t.Fatal(err)
			}
			test.mutate(&report)
			retainedTestRehash(t, &report)
			_, err := VerifyRetainedWitnessReport(retainedTestMarshal(t, report))
			if err == nil || !strings.Contains(err.Error(), test.reason) {
				t.Fatalf("verification error=%v, want=%s", err, test.reason)
			}
		})
	}
}

func TestRetainedDigestProfilesAndCollectionLimits(t *testing.T) {
	base := Run(readFixture(t, "ordinary-low-pressure.json"), "witness")
	for _, test := range []struct {
		name   string
		mutate func(*Report)
		reason string
	}{
		{"witness schema", func(r *Report) { r.WitnessSchema = "future" }, "unsupported_witness_digest"},
		{"input schema", func(r *Report) { r.InputDigestSchema = "future" }, "unsupported_witness_digest"},
		{"algorithm", func(r *Report) { r.WitnessAlgorithm = "md5" }, "unsupported_witness_digest"},
		{"uppercase digest", func(r *Report) { r.Witness = strings.Repeat("A", 64) }, "unsupported_witness_digest"},
		{"stale input digest", func(r *Report) { r.InputDigest = strings.Repeat("0", 64) }, "input_digest_mismatch"},
		{"stale witness", func(r *Report) { r.Witness = strings.Repeat("0", 64) }, "account_witness_mismatch"},
		{"too many samples", func(r *Report) { r.History = make([]HistorySample, 4098) }, "collection_limit_exceeded"},
		{"too many components", func(r *Report) { r.History[0].Components = make([]ComponentMass, 65) }, "collection_limit_exceeded"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var report Report
			if err := json.Unmarshal(retainedTestMarshal(t, base), &report); err != nil {
				t.Fatal(err)
			}
			test.mutate(&report)
			_, err := VerifyRetainedWitnessReport(retainedTestMarshal(t, report))
			if err == nil || !strings.Contains(err.Error(), test.reason) {
				t.Fatalf("verification error=%v, want=%s", err, test.reason)
			}
		})
	}
	// The count gate must run before typed decoding: these compact empty
	// objects are small on the wire but would expand into large report structs.
	largeArray := []byte(`{"history":[` + strings.Repeat(`{},`, 4097) + `{}` + `]}`)
	if _, err := VerifyRetainedWitnessReport(largeArray); err == nil || !strings.Contains(err.Error(), "collection_limit_exceeded") {
		t.Fatalf("compact allocation attack: %v", err)
	}
}

func FuzzVerifyRetainedWitnessReport(f *testing.F) {
	closed := editJSON(f, readFixture(f, "ordinary-low-pressure.json"), change{path: "restriction/area/prescribed_m2", value: 0})
	f.Add(retainedTestMarshal(f, Run(closed, "witness")))
	f.Add([]byte(`{"history":[{}]}`))
	f.Add([]byte(`null`))
	f.Fuzz(func(t *testing.T, raw []byte) {
		if len(raw) > MaxRetainedReportBytes+1 {
			return
		}
		before := bytes.Clone(raw)
		retained, err := VerifyRetainedWitnessReport(raw)
		if !bytes.Equal(before, raw) {
			t.Fatal("verification mutated retained source bytes")
		}
		if err != nil {
			if retained != (RetainedWitness{}) {
				t.Fatal("invalid input acquired a retained comparison")
			}
			return
		}
		again, err := VerifyRetainedWitnessReport(raw)
		if err != nil || !reflect.DeepEqual(again, retained) || !validDigest(retained.Witness()) || !validDigest(retained.InputDigest()) {
			t.Fatal("accepted retained account is nondeterministic or incomplete")
		}
	})
}

func retainedTestMarshal(t testing.TB, value any) []byte {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func retainedTestSHA(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func retainedTestRehash(t testing.TB, report *Report) {
	t.Helper()
	account := *report
	account.Operation = "simulate"
	account.Explanation = nil
	account.Witness = ""
	account.WitnessSchema = ""
	account.WitnessAlgorithm = ""
	account.InputDigest = ""
	account.InputDigestSchema = ""
	// Build the documented envelopes directly, without calling witnessOf or
	// the verification API, so these tests catch accidental restoration changes.
	inputEnvelope := append([]byte(`{"schema":"fart.walk-normalized-input/v0alpha1","inputs":`), retainedTestMarshal(t, account.Inputs)...)
	inputEnvelope = append(inputEnvelope, '}')
	accountEnvelope := append([]byte(`{"schema":"fart.walk-witness/v0alpha1","account":`), retainedTestMarshal(t, account)...)
	accountEnvelope = append(accountEnvelope, '}')
	report.InputDigest = retainedTestSHA(inputEnvelope)
	report.Witness = retainedTestSHA(accountEnvelope)
}
