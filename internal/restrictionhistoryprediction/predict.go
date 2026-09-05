package restrictionhistoryprediction

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/blisspixel/fartapp/internal/restrictionflow"
	"github.com/blisspixel/fartapp/internal/restrictionhistory"
	"github.com/blisspixel/fartapp/internal/strictjson"
)

var (
	ErrInputTooLarge = errors.New("restriction history input exceeds the byte limit")
	ErrNilInput      = errors.New("restriction history input reader is nil")
)

type requestDocument struct {
	Schema               string            `json:"schema"`
	Model                requestModel      `json:"model"`
	QuantitySystem       string            `json:"quantity_system"`
	Stagnation           requestStagnation `json:"stagnation"`
	BackPressurePascals  *float64          `json:"back_pressure_pa"`
	DischargeCoefficient *float64          `json:"discharge_coefficient"`
	Samples              []requestSample   `json:"samples"`
}

type requestModel struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

type requestStagnation struct {
	PressurePascals     *float64 `json:"pressure_pa"`
	TemperatureKelvin   *float64 `json:"temperature_k"`
	SpecificGasConstant *float64 `json:"specific_gas_constant_j_per_kg_k"`
	HeatCapacityRatio   *float64 `json:"heat_capacity_ratio"`
}

type requestSample struct {
	TimeSeconds            *float64 `json:"time_s"`
	PrescribedSquareMetres *float64 `json:"prescribed_m2"`
}

func ReadBounded(reader io.Reader) ([]byte, error) {
	if reader == nil {
		return nil, ErrNilInput
	}
	data, err := io.ReadAll(io.LimitReader(reader, MaxInputBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > MaxInputBytes {
		return nil, ErrInputTooLarge
	}
	return data, nil
}

func Predict(data []byte) Report {
	if len(data) > MaxInputBytes {
		return failure("FART-E-INPUT-0004", "input", "/", "input_too_large")
	}
	if issue := strictjson.Inspect(data, strictjson.Limits{
		MaximumDepth: maximumJSONDepth, MaximumMemberNameBytes: maximumMemberNameBytes,
	}); issue != nil {
		return failure("FART-E-JSON-0004", "syntax", issue.Path, string(issue.Kind))
	}
	if issue := strictjson.InspectShape[requestDocument](data); issue != nil {
		return failure("FART-E-SCHEMA-0004", "schema", issue.Path, "document_shape_invalid")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var document requestDocument
	if err := decoder.Decode(&document); err != nil {
		return failure("FART-E-SCHEMA-0004", "schema", "/", "document_shape_invalid")
	}
	history, diagnostic := interpret(document)
	if diagnostic != nil {
		return failure(diagnostic.Code, diagnostic.Stage, diagnostic.Path, diagnostic.ReasonCode)
	}
	return buildReport(history)
}

func InputFailure(reason string, consultedInputs ...string) Report {
	report := failure("FART-E-IO-0004", "input", "/", reason)
	report.ValidationEnvironment.ConsultedInputs = append([]string(nil), consultedInputs...)
	return report
}

func interpret(document requestDocument) (restrictionhistory.History, *Diagnostic) {
	if document.Schema != RequestSchema {
		return restrictionhistory.History{}, schema("/", "unsupported_schema")
	}
	if document.Model.ID != ModelID || document.Model.Version != ModelVersion {
		return restrictionhistory.History{}, schema("/model", "unsupported_model_revision")
	}
	if document.QuantitySystem != QuantitySystem {
		return restrictionhistory.History{}, schema("/quantity_system", "unsupported_quantity_system")
	}
	if document.BackPressurePascals == nil {
		return restrictionhistory.History{}, schema("/back_pressure_pa", "missing_member")
	}
	if document.DischargeCoefficient == nil {
		return restrictionhistory.History{}, schema("/discharge_coefficient", "missing_member")
	}
	if len(document.Samples) == 0 {
		return restrictionhistory.History{}, schema("/samples", "missing_member")
	}
	if len(document.Samples) > restrictionhistory.MaxSamples {
		return restrictionhistory.History{}, schema("/samples", "invalid_sample_count")
	}
	stagnation, diagnostic := parseStagnation(document.Stagnation)
	if diagnostic != nil {
		return restrictionhistory.History{}, diagnostic
	}
	back, err := restrictionflow.NewPressure(*document.BackPressurePascals)
	if err != nil {
		return restrictionhistory.History{}, model("/back_pressure_pa", "nonpositive_quantity")
	}
	cd, err := restrictionflow.NewDischargeCoefficient(*document.DischargeCoefficient)
	if err != nil {
		return restrictionhistory.History{}, model("/discharge_coefficient", "invalid_discharge_coefficient")
	}
	samples := make([]restrictionhistory.Sample, len(document.Samples))
	for index, sample := range document.Samples {
		path := fmt.Sprintf("/samples/%d", index)
		if sample.TimeSeconds == nil || sample.PrescribedSquareMetres == nil {
			return restrictionhistory.History{}, schema(path, "missing_member")
		}
		time, err := restrictionhistory.NewSeconds(*sample.TimeSeconds)
		if err != nil {
			return restrictionhistory.History{}, model(path+"/time_s", "invalid_time")
		}
		area, err := restrictionflow.NewArea(*sample.PrescribedSquareMetres)
		if err != nil {
			return restrictionhistory.History{}, model(path+"/prescribed_m2", "negative_area")
		}
		built, err := restrictionhistory.NewSample(time, area)
		if err != nil {
			return restrictionhistory.History{}, model(path, "invalid_sample")
		}
		samples[index] = built
	}
	history, err := restrictionhistory.Integrate(stagnation, back, cd, samples)
	if err != nil {
		reason := "numerical_domain_error"
		if errors.Is(err, restrictionhistory.ErrInvalidTime) {
			reason = "invalid_time"
		}
		if errors.Is(err, restrictionhistory.ErrInvalidSampleCount) {
			reason = "invalid_sample_count"
		}
		if errors.Is(err, restrictionflow.ErrAdversePressure) {
			reason = "adverse_pressure"
		}
		return restrictionhistory.History{}, model("/samples", reason)
	}
	return history, nil
}

func parseStagnation(document requestStagnation) (restrictionflow.Stagnation, *Diagnostic) {
	if document.PressurePascals == nil || document.TemperatureKelvin == nil ||
		document.SpecificGasConstant == nil || document.HeatCapacityRatio == nil {
		return restrictionflow.Stagnation{}, schema("/stagnation", "missing_member")
	}
	pressure, err := restrictionflow.NewPressure(*document.PressurePascals)
	if err != nil {
		return restrictionflow.Stagnation{}, model("/stagnation/pressure_pa", "nonpositive_quantity")
	}
	temperature, err := restrictionflow.NewTemperature(*document.TemperatureKelvin)
	if err != nil {
		return restrictionflow.Stagnation{}, model("/stagnation/temperature_k", "nonpositive_quantity")
	}
	gas, err := restrictionflow.NewSpecificGasConstant(*document.SpecificGasConstant)
	if err != nil {
		return restrictionflow.Stagnation{}, model("/stagnation/specific_gas_constant_j_per_kg_k", "nonpositive_quantity")
	}
	gamma, err := restrictionflow.NewHeatCapacityRatio(*document.HeatCapacityRatio)
	if err != nil {
		return restrictionflow.Stagnation{}, model("/stagnation/heat_capacity_ratio", "invalid_heat_capacity_ratio")
	}
	stagnation, err := restrictionflow.NewStagnation(pressure, temperature, gas, gamma)
	if err != nil {
		return restrictionflow.Stagnation{}, model("/stagnation", "invalid_stagnation")
	}
	return stagnation, nil
}

func buildReport(history restrictionhistory.History) Report {
	samples := history.Samples()
	reportSamples := make([]Sample, len(samples))
	for index, sample := range samples {
		result := sample.Result()
		reportSamples[index] = Sample{
			TimeSeconds:            sample.Time().Value(),
			PrescribedSquareMetres: result.Request().AreaLaw().Prescribed().SquareMetres(),
			EffectiveSquareMetres:  result.EffectiveArea().SquareMetres(),
			Regime:                 result.Regime().String(),
			ExitPressurePascals:    result.ExitPressure().Pascals(),
			MassFlowKilogramsPerS:  result.MassFlow().KilogramsPerSecond(),
			ThrustNewtons:          result.Thrust().Newtons(),
			RecoilNewtons:          result.Recoil().Newtons(),
		}
	}
	totals := Totals{
		MassOutKilograms:            history.MassOutKilograms(),
		EnthalpyOutJoules:           history.EnthalpyOutJoules(),
		KineticEnergyOutJoules:      history.KineticEnergyOutJoules(),
		TotalEnthalpyOutJoules:      history.TotalEnthalpyOutJoules(),
		ImpulseNewtonSeconds:        history.ImpulseNewtonSeconds(),
		RecoilImpulseNewtonSeconds:  history.RecoilImpulseNewtonSeconds(),
		RecoilResidualNewtonSeconds: history.RecoilResidualNewtonSeconds(),
	}
	claim := Claim{
		ID: "restriction-history.recoil-action-reaction", Status: "failed",
		Method: "equal-and-opposite-impulse", EquationRevision: ModelID + "@" + ModelVersion,
		Residual: totals.RecoilResidualNewtonSeconds, ResidualUnit: "N s",
		Tolerance: roundoffTolerance(totals.ImpulseNewtonSeconds, totals.RecoilImpulseNewtonSeconds),
	}
	if finite(claim.Residual) && finite(claim.Tolerance) && math.Abs(claim.Residual) <= claim.Tolerance {
		claim.Status = "satisfied-within-roundoff"
	}
	if claim.Status != "satisfied-within-roundoff" {
		return failure("FART-E-NUMERICAL-0003", "model", "/", "invariant_violation")
	}
	return Report{
		Schema: ReportSchema, Status: "predicted", RequestSchema: RequestSchema,
		Model:                  &ModelReference{ID: ModelID, Version: ModelVersion},
		ImplementationRevision: ImplementationRevision, QuantitySystem: QuantitySystem,
		Samples: reportSamples, Totals: &totals,
		Assumptions: []string{
			"frozen-stagnation-state",
			"prescribed-area-history",
			"trapezoidal-rate-integration",
			"quasi-steady-samples",
			"single-calorically-perfect-gas",
			"enthalpy-out-is-static-exit-enthalpy",
			"total-enthalpy-includes-exit-kinetic-energy",
		},
		Nonclaims: &Nonclaims{
			Model: []string{
				"reservoir-coupling-and-blowdown",
				"species-resolved-composition-history",
				"plume-and-acoustics",
				"elapsed-source-depletion",
			},
			Operation: []string{"case-commitment", "certificate-issuance"},
			Evidence:  []string{"empirical-validation"},
		},
		Claims: []Claim{claim},
		ValidationEnvironment: ValidationEnvironment{
			ConsultedInputs: []string{"document_bytes"}, AmbientInputs: []string{},
		},
	}
}

func failure(code, stage, path, reason string) Report {
	return Report{
		Schema: ReportSchema, Status: "invalid", ImplementationRevision: ImplementationRevision,
		ValidationEnvironment: ValidationEnvironment{
			ConsultedInputs: []string{"document_bytes"}, AmbientInputs: []string{},
		},
		Diagnostics: []Diagnostic{{Code: code, Stage: stage, Path: path, ReasonCode: reason}},
	}
}

func schema(path, reason string) *Diagnostic {
	return &Diagnostic{Code: "FART-E-SCHEMA-0004", Stage: "schema", Path: path, ReasonCode: reason}
}

func model(path, reason string) *Diagnostic {
	return &Diagnostic{Code: "FART-E-MODEL-0005", Stage: "model", Path: path, ReasonCode: reason}
}

func roundoffTolerance(terms ...float64) float64 {
	epsilon := math.Nextafter(1, 2) - 1
	maximumMagnitude := 0.0
	for _, term := range terms {
		if !finite(term) {
			return math.NaN()
		}
		maximumMagnitude = math.Max(maximumMagnitude, math.Abs(term))
	}
	return 64*epsilon*maximumMagnitude + math.SmallestNonzeroFloat64
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
