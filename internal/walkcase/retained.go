package walkcase

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/blisspixel/fartapp/internal/coupledblowdown"
	"github.com/blisspixel/fartapp/internal/idealmixturereservoir"
	"github.com/blisspixel/fartapp/internal/strictjson"
)

// MaxRetainedReportBytes limits this experimental storage profile. A valid walk
// calculation may produce a larger report and consequently cannot be retained
// in this profile. This is not a limit on the scientific meaning of a case.
const MaxRetainedReportBytes = 16 << 20

var ErrRetainedReportTooLarge = errors.New("retained walk report exceeds the byte limit")

// RetainedWitness is an immutable, internally consistent software comparison.
// Its zero value is invalid. It establishes neither that a solver produced the
// report nor an occurrence identity, signature, empirical validity, or trust.
type RetainedWitness struct {
	inputDigest string
	witness     string
}

func (retained RetainedWitness) InputDigest() string { return retained.inputDigest }

func (retained RetainedWitness) Witness() string { return retained.witness }

// NormalizeRequestFingerprint validates and normalizes an authored request,
// then computes the existing versioned input digest without time integration.
// Authored whitespace, numeric spelling, component order, and expected_witness
// are not identities. Preserve the original bytes separately when retaining
// evidence. Validation evaluates the initial model boundary, but runs no solver.
func NormalizeRequestFingerprint(raw []byte) (string, error) {
	if len(raw) > MaxInputBytes {
		return "", ErrInputTooLarge
	}
	parsed, diagnostic := parseCase(raw)
	if diagnostic != nil {
		return "", fmt.Errorf("walk request: %s at %q", diagnostic.ReasonCode, diagnostic.Path)
	}
	return normalizedFingerprint(parsed)
}

func normalizedFingerprint(parsed parsedCase) (string, error) {
	// Reuse the existing envelope encoder so this API cannot silently introduce
	// a second definition of the normalized-input wire bytes.
	_, digest, err := witnessOf(Report{Inputs: &parsed.document})
	return digest, err
}

// VerifyRetainedWitnessReport accepts only the compact encoding/json bytes of
// a complete current-profile successful witness report. This narrow storage
// encoding is not RFC 8785, a universal .fart format, or a migration mechanism.
// It checks required structure, normalized inputs, and the account's own hash
// without running the solver. Unkeyed hashes are forgeable by anyone who can
// replace the report; a successful check is internal consistency, not evidence
// of authorship or execution. The retained runtime is kept when hashing.
func VerifyRetainedWitnessReport(reportBytes []byte) (RetainedWitness, error) {
	if len(reportBytes) > MaxRetainedReportBytes {
		return RetainedWitness{}, ErrRetainedReportTooLarge
	}
	if issue := strictjson.Inspect(reportBytes, strictjson.Limits{
		MaximumDepth: maximumJSONDepth, MaximumMemberNameBytes: maximumMemberNameBytes,
	}); issue != nil {
		return RetainedWitness{}, fmt.Errorf("retained walk report: %s at %q", issue.Kind, issue.Path)
	}
	if err := inspectRetainedCollections(reportBytes); err != nil {
		return RetainedWitness{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(reportBytes))
	decoder.DisallowUnknownFields()
	var report Report
	if err := decoder.Decode(&report); err != nil {
		return RetainedWitness{}, retainedFailure("report_shape_invalid")
	}
	encoded, err := json.Marshal(report)
	if err != nil || !bytes.Equal(encoded, reportBytes) {
		return RetainedWitness{}, retainedFailure("unsupported_report_encoding")
	}
	if err := validateRetainedProfile(report); err != nil {
		return RetainedWitness{}, err
	}
	inputBytes, err := json.Marshal(report.Inputs)
	if err != nil || len(inputBytes) > MaxInputBytes {
		return RetainedWitness{}, retainedFailure("invalid_retained_inputs")
	}
	parsed, diagnostic := parseCase(inputBytes)
	if diagnostic != nil {
		return RetainedWitness{}, retainedFailure("invalid_retained_inputs")
	}
	normalized, err := json.Marshal(parsed.document)
	if err != nil || !bytes.Equal(normalized, inputBytes) {
		return RetainedWitness{}, retainedFailure("inputs_not_normalized")
	}
	if err := validateRetainedAccount(report, parsed); err != nil {
		return RetainedWitness{}, err
	}
	expectedWitness, expectedInput := report.Witness, report.InputDigest
	// runWitness decorates a completed simulate report only after hashing it.
	// These fields, and only these fields, must be restored to that account.
	report.Operation = "simulate"
	report.Explanation = nil
	report.Witness = ""
	report.WitnessSchema = ""
	report.WitnessAlgorithm = ""
	report.InputDigest = ""
	report.InputDigestSchema = ""
	witness, digest, err := witnessOf(report)
	if err != nil {
		return RetainedWitness{}, retainedFailure("account_encoding_failure")
	}
	if digest != expectedInput {
		return RetainedWitness{}, retainedFailure("input_digest_mismatch")
	}
	if witness != expectedWitness {
		return RetainedWitness{}, retainedFailure("account_witness_mismatch")
	}
	return RetainedWitness{inputDigest: digest, witness: witness}, nil
}

// Reconstruct performs a new calculation using this implementation and runtime
// and compares its witness with the retained witness. An unrelated normalized
// request is rejected before solving. A valid authored expected_witness member
// is replaced by the retained comparison target, without modifying the caller's
// bytes. The returned Report belongs to this calculation and is not retained.
func (retained RetainedWitness) Reconstruct(requestBytes []byte) Report {
	if !validDigest(retained.inputDigest) || !validDigest(retained.witness) {
		return failure("FART-E-SCHEMA-0005", "schema", "/", "invalid_retained_witness")
	}
	if len(requestBytes) > MaxInputBytes {
		return failure("FART-E-INPUT-0005", "input", "/", "input_too_large")
	}
	parsed, diagnostic := parseCase(requestBytes)
	if diagnostic != nil {
		return failure(diagnostic.Code, diagnostic.Stage, diagnostic.Path, diagnostic.ReasonCode)
	}
	digest, err := normalizedFingerprint(parsed)
	if err != nil {
		return failure("FART-E-NUMERICAL-0004", "model", "/inputs", "input_encoding_failure")
	}
	if digest != retained.inputDigest {
		return failure("FART-E-NUMERICAL-0004", "comparison", "/inputs", "request_fingerprint_mismatch")
	}
	parsed.expectedWitness = retained.witness
	return runWitness(parsed, "reconstruct")
}

func validateRetainedProfile(report Report) error {
	if !report.Predicted() || report.Operation != "witness" || report.RequestSchema != RequestSchema ||
		report.ImplementationRevision != ImplementationRevision || report.QuantitySystem != QuantitySystem ||
		report.Model == nil || *report.Model != (ModelReference{ID: ModelID, Version: ModelVersion}) ||
		report.Inputs == nil || report.NumericalPolicy == nil {
		return retainedFailure("unsupported_witness_profile")
	}
	if report.Branch != nil || report.Accuracy != nil || report.ExpectedWitness != "" || report.ReconstructedWitness != "" ||
		report.WitnessMatch != nil || report.EqualizationFraction != nil || report.EndpointReachability != "" {
		return retainedFailure("unexpected_operation_fields")
	}
	if report.WitnessSchema != WitnessSchema || report.InputDigestSchema != InputDigestSchema ||
		report.WitnessAlgorithm != "sha256" || !validDigest(report.Witness) || !validDigest(report.InputDigest) {
		return retainedFailure("unsupported_witness_digest")
	}
	if !retainedRuntimeString(report.NumericalPolicy.GoVersion) ||
		!retainedRuntimeString(report.NumericalPolicy.OperatingSystem) ||
		!retainedRuntimeString(report.NumericalPolicy.Architecture) {
		return retainedFailure("invalid_runtime_profile")
	}
	if report.ElapsedSeconds == nil || report.Steps == nil || report.Initial == nil || report.Final == nil ||
		report.MassOutKilograms == nil || report.EnthalpyOutJoules == nil || report.HeatInJoules == nil ||
		report.ImpulseNewtonSeconds == nil || report.RecoilImpulseNewtonSeconds == nil ||
		report.EqualizationPressureTolerancePascals == nil || report.InitialRestriction == nil || report.Signature == nil {
		return retainedFailure("incomplete_numerical_account")
	}
	if *report.Steps < 0 || *report.Steps > coupledblowdown.MaxSteps || len(report.History) != 1+*report.Steps {
		return retainedFailure("invalid_history_count")
	}
	switch report.Stop {
	case "no-flow", "equalized", "max-steps", "max-time", "no-progress", "pressure-tolerance":
	default:
		return retainedFailure("unsupported_stop")
	}
	if !positiveEndpoint(*report.Initial) || !positiveEndpoint(*report.Final) ||
		!nonnegative(*report.ElapsedSeconds, *report.MassOutKilograms, *report.EnthalpyOutJoules,
			*report.HeatInJoules, *report.ImpulseNewtonSeconds, -*report.RecoilImpulseNewtonSeconds,
			*report.EqualizationPressureTolerancePascals, report.Signature.EquivalentDiameterMetres,
			report.Signature.StrokeLengthMetres) ||
		(report.Signature.FormationNumber != nil && !nonnegative(*report.Signature.FormationNumber)) {
		return retainedFailure("invalid_account_quantity")
	}
	if len(report.Explanation) != 1 || report.Explanation[0] != "This digest binds the normalized inputs and the full numerical account, including component identities, history, balances, model revision, and runtime profile. It is a software comparison, not an occurrence identity, signature, or empirical proof." {
		return retainedFailure("unsupported_witness_explanation")
	}
	return nil
}

func validateRetainedAccount(report Report, parsed parsedCase) error {
	expected := baseReport(parsed, "witness")
	policy := *report.NumericalPolicy
	policy.GoVersion = expected.NumericalPolicy.GoVersion
	policy.OperatingSystem = expected.NumericalPolicy.OperatingSystem
	policy.Architecture = expected.NumericalPolicy.Architecture
	if policy != *expected.NumericalPolicy || report.LawContext != expected.LawContext || report.Closure != expected.Closure ||
		!reflect.DeepEqual(report.Dimensions, expected.Dimensions) || !reflect.DeepEqual(report.Assumptions, expected.Assumptions) ||
		!reflect.DeepEqual(report.Nonclaims, expected.Nonclaims) || !reflect.DeepEqual(report.ValidationEnvironment, expected.ValidationEnvironment) {
		return retainedFailure("inconsistent_report_references")
	}
	if *report.Steps > parsed.maxSteps || (parsed.maxTime > 0 && *report.ElapsedSeconds > parsed.maxTime) {
		return retainedFailure("account_exceeds_request_budget")
	}
	claimSpecs := []struct{ id, method, unit string }{
		{"walk.mass-ledger", "double-entry-balance", "kg"},
		{"walk.energy-ledger", "double-entry-balance", "J"},
		{"walk.impulse-ledger", "equal-and-opposite-force-accounting", "N s"},
	}
	if len(report.Claims) != len(claimSpecs) {
		return retainedFailure("incomplete_balance_claims")
	}
	for index, claim := range report.Claims {
		spec := claimSpecs[index]
		if claim.ID != spec.id || claim.Method != spec.method || claim.ResidualUnit != spec.unit ||
			claim.EquationRevision != ModelID+"@"+ModelVersion || claim.Status != "satisfied-within-roundoff" ||
			!finite(claim.Residual) || !nonnegative(claim.Tolerance) || math.Abs(claim.Residual) > claim.Tolerance {
			return retainedFailure("invalid_balance_claim")
		}
	}
	initial, final := report.History[0], report.History[len(report.History)-1]
	if initial.TimeSeconds != 0 || final.TimeSeconds != *report.ElapsedSeconds ||
		!sampleMatchesEndpoint(initial, *report.Initial) || !sampleMatchesEndpoint(final, *report.Final) {
		return retainedFailure("inconsistent_history_endpoints")
	}
	flow := report.InitialRestriction
	if !retainedRegime(flow.Regime) || !nonnegative(flow.MassFlowKilogramsPerS, flow.ThrustNewtons) ||
		flow.CriticalPressureRatio <= 0 || flow.CriticalPressureRatio >= 1 || flow.BackPressureRatio <= 0 || flow.BackPressureRatio > 1 ||
		flow.Regime != initial.Regime || flow.MassFlowKilogramsPerS != initial.MassFlowKilogramsPerSecond || flow.ThrustNewtons != initial.ThrustNewtons {
		return retainedFailure("inconsistent_initial_restriction")
	}
	for index, sample := range report.History {
		if !retainedRegime(sample.Regime) || sample.MassKilograms <= 0 || sample.PressurePascals <= 0 || sample.TemperatureKelvin <= 0 ||
			!nonnegative(sample.TimeSeconds, sample.MassFlowKilogramsPerSecond, sample.SourceTotalEnthalpyWatts,
				sample.ExitSpeedMetresPerSecond, sample.ExitPressurePascals, sample.ExitTemperatureKelvin,
				sample.EffectiveAreaSquareMetres, sample.ThrustNewtons, -sample.RecoilNewtons) {
			return retainedFailure("invalid_history_quantity")
		}
		if index > 0 && (sample.TimeSeconds <= report.History[index-1].TimeSeconds || sample.MassKilograms > report.History[index-1].MassKilograms) {
			return retainedFailure("invalid_history_progress")
		}
		if len(sample.Components) != len(parsed.document.Reservoir.Components) {
			return retainedFailure("invalid_history_component_count")
		}
		for componentIndex, component := range sample.Components {
			input := parsed.document.Reservoir.Components[componentIndex]
			if component.ID != input.ID || component.MassKilograms <= 0 || !nonnegative(component.MassOutKilograms) ||
				(index == 0 && (component.MassKilograms != *input.MassKilograms || component.MassOutKilograms != 0)) {
				return retainedFailure("invalid_history_component")
			}
		}
	}
	return nil
}

func retainedRuntimeString(value string) bool {
	if len(value) == 0 || len(value) > 128 || strings.TrimSpace(value) == "" {
		return false
	}
	for _, character := range value {
		if character < ' ' || character > '~' {
			return false
		}
	}
	return true
}

func positiveEndpoint(endpoint Endpoint) bool {
	return endpoint.MassKilograms > 0 && endpoint.PressurePascals > 0 &&
		endpoint.TemperatureKelvin > 0 && endpoint.InternalEnergyJoules > 0
}

func nonnegative(values ...float64) bool {
	for _, value := range values {
		if !finite(value) || value < 0 {
			return false
		}
	}
	return true
}

func retainedRegime(regime string) bool {
	return regime == "no-flow" || regime == "subsonic" || regime == "choked"
}

func sampleMatchesEndpoint(sample HistorySample, endpoint Endpoint) bool {
	return sample.MassKilograms == endpoint.MassKilograms && sample.PressurePascals == endpoint.PressurePascals &&
		sample.TemperatureKelvin == endpoint.TemperatureKelvin
}

func retainedFailure(reason string) error { return errors.New("retained walk report: " + reason) }

// Enforce collection bounds before decoding into large typed report slices.
// Syntax, duplicate members, key length, and depth were already inspected.
// All arrays in this narrow profile have at most 64 entries except the sole
// top-level history array. Unknown or differently shaped fields fail later.
func inspectRetainedCollections(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	return inspectRetainedValue(decoder, "")
}

func inspectRetainedValue(decoder *json.Decoder, path string) error {
	token, err := decoder.Token()
	if err != nil {
		return retainedFailure("report_shape_invalid")
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		for decoder.More() {
			key, err := decoder.Token()
			if err != nil {
				return retainedFailure("report_shape_invalid")
			}
			if err := inspectRetainedValue(decoder, path+"/"+key.(string)); err != nil {
				return err
			}
		}
	case '[':
		maximum := idealmixturereservoir.MaxComponents
		if path == "/history" {
			maximum = coupledblowdown.MaxSteps + 1
		}
		for index := 0; decoder.More(); index++ {
			if index >= maximum {
				return retainedFailure("collection_limit_exceeded")
			}
			if err := inspectRetainedValue(decoder, path+"/"+strconv.Itoa(index)); err != nil {
				return err
			}
		}
	}
	if _, err := decoder.Token(); err != nil {
		return retainedFailure("report_shape_invalid")
	}
	return nil
}
