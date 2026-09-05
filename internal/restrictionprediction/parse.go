package restrictionprediction

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/blisspixel/fartapp/internal/restrictionflow"
	"github.com/blisspixel/fartapp/internal/strictjson"
)

type requestDocument struct {
	Schema               string            `json:"schema"`
	Model                requestModel      `json:"model"`
	QuantitySystem       string            `json:"quantity_system"`
	Stagnation           requestStagnation `json:"stagnation"`
	BackPressurePascals  *float64          `json:"back_pressure_pa"`
	DischargeCoefficient *float64          `json:"discharge_coefficient"`
	Area                 requestArea       `json:"area"`
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

type requestArea struct {
	Law                    string   `json:"law"`
	PrescribedSquareMetres *float64 `json:"prescribed_m2"`
	ComplianceSquareMetres *float64 `json:"compliance_m2_per_pa"`
	MaximumSquareMetres    *float64 `json:"maximum_m2"`
}

type parsedRequest struct {
	request restrictionflow.Request
}

func parseRequest(data []byte) (parsedRequest, *Diagnostic) {
	if diagnostic := preflightJSON(data); diagnostic != nil {
		return parsedRequest{}, diagnostic
	}
	if issue := strictjson.InspectShape[requestDocument](data); issue != nil {
		return parsedRequest{}, schemaDiagnostic(issue.Path, "document_shape_invalid")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var document requestDocument
	if err := decoder.Decode(&document); err != nil {
		return parsedRequest{}, schemaDiagnostic("/", "document_shape_invalid")
	}
	if document.Schema != RequestSchema {
		return parsedRequest{}, schemaDiagnostic("/schema", "unsupported_schema")
	}
	if document.Model.ID != ModelID || document.Model.Version != ModelVersion {
		return parsedRequest{}, schemaDiagnostic("/model", "unsupported_model_revision")
	}
	if document.QuantitySystem != QuantitySystem {
		return parsedRequest{}, schemaDiagnostic("/quantity_system", "unsupported_quantity_system")
	}
	if document.BackPressurePascals == nil {
		return parsedRequest{}, schemaDiagnostic("/back_pressure_pa", "missing_member")
	}
	if document.DischargeCoefficient == nil {
		return parsedRequest{}, schemaDiagnostic("/discharge_coefficient", "missing_member")
	}
	stagnation, diagnostic := parseStagnation(document.Stagnation)
	if diagnostic != nil {
		return parsedRequest{}, diagnostic
	}
	area, diagnostic := parseArea(document.Area)
	if diagnostic != nil {
		return parsedRequest{}, diagnostic
	}
	back, err := restrictionflow.NewPressure(*document.BackPressurePascals)
	if err != nil {
		return parsedRequest{}, modelDiagnostic("/back_pressure_pa", classifyModelError(err))
	}
	cd, err := restrictionflow.NewDischargeCoefficient(*document.DischargeCoefficient)
	if err != nil {
		return parsedRequest{}, modelDiagnostic("/discharge_coefficient", classifyModelError(err))
	}
	request, err := restrictionflow.NewRequest(stagnation, back, area, cd)
	if err != nil {
		return parsedRequest{}, modelDiagnostic("/", classifyModelError(err))
	}
	return parsedRequest{request: request}, nil
}

func parseStagnation(document requestStagnation) (restrictionflow.Stagnation, *Diagnostic) {
	if document.PressurePascals == nil {
		return restrictionflow.Stagnation{}, schemaDiagnostic("/stagnation/pressure_pa", "missing_member")
	}
	if document.TemperatureKelvin == nil {
		return restrictionflow.Stagnation{}, schemaDiagnostic("/stagnation/temperature_k", "missing_member")
	}
	if document.SpecificGasConstant == nil {
		return restrictionflow.Stagnation{}, schemaDiagnostic("/stagnation/specific_gas_constant_j_per_kg_k", "missing_member")
	}
	if document.HeatCapacityRatio == nil {
		return restrictionflow.Stagnation{}, schemaDiagnostic("/stagnation/heat_capacity_ratio", "missing_member")
	}
	pressure, err := restrictionflow.NewPressure(*document.PressurePascals)
	if err != nil {
		return restrictionflow.Stagnation{}, modelDiagnostic("/stagnation/pressure_pa", classifyModelError(err))
	}
	temperature, err := restrictionflow.NewTemperature(*document.TemperatureKelvin)
	if err != nil {
		return restrictionflow.Stagnation{}, modelDiagnostic("/stagnation/temperature_k", classifyModelError(err))
	}
	gasConstant, err := restrictionflow.NewSpecificGasConstant(*document.SpecificGasConstant)
	if err != nil {
		return restrictionflow.Stagnation{}, modelDiagnostic(
			"/stagnation/specific_gas_constant_j_per_kg_k",
			classifyModelError(err),
		)
	}
	gamma, err := restrictionflow.NewHeatCapacityRatio(*document.HeatCapacityRatio)
	if err != nil {
		return restrictionflow.Stagnation{}, modelDiagnostic(
			"/stagnation/heat_capacity_ratio",
			classifyModelError(err),
		)
	}
	stagnation, err := restrictionflow.NewStagnation(pressure, temperature, gasConstant, gamma)
	if err != nil {
		return restrictionflow.Stagnation{}, modelDiagnostic("/stagnation", classifyModelError(err))
	}
	return stagnation, nil
}

func parseArea(document requestArea) (restrictionflow.AreaLaw, *Diagnostic) {
	if document.PrescribedSquareMetres == nil {
		return restrictionflow.AreaLaw{}, schemaDiagnostic("/area/prescribed_m2", "missing_member")
	}
	prescribed, err := restrictionflow.NewArea(*document.PrescribedSquareMetres)
	if err != nil {
		return restrictionflow.AreaLaw{}, modelDiagnostic("/area/prescribed_m2", classifyModelError(err))
	}
	switch document.Law {
	case "prescribed":
		if document.ComplianceSquareMetres != nil {
			return restrictionflow.AreaLaw{}, schemaDiagnostic("/area/compliance_m2_per_pa", "unexpected_member")
		}
		if document.MaximumSquareMetres != nil {
			return restrictionflow.AreaLaw{}, schemaDiagnostic("/area/maximum_m2", "unexpected_member")
		}
		law, err := restrictionflow.NewPrescribedArea(prescribed)
		if err != nil {
			return restrictionflow.AreaLaw{}, modelDiagnostic("/area", classifyModelError(err))
		}
		return law, nil
	case "linear-compliance":
		if document.ComplianceSquareMetres == nil {
			return restrictionflow.AreaLaw{}, schemaDiagnostic("/area/compliance_m2_per_pa", "missing_member")
		}
		if document.MaximumSquareMetres == nil {
			return restrictionflow.AreaLaw{}, schemaDiagnostic("/area/maximum_m2", "missing_member")
		}
		compliance, err := restrictionflow.NewAreaCompliance(*document.ComplianceSquareMetres)
		if err != nil {
			return restrictionflow.AreaLaw{}, modelDiagnostic("/area/compliance_m2_per_pa", classifyModelError(err))
		}
		maximum, err := restrictionflow.NewArea(*document.MaximumSquareMetres)
		if err != nil {
			return restrictionflow.AreaLaw{}, modelDiagnostic("/area/maximum_m2", classifyModelError(err))
		}
		law, err := restrictionflow.NewLinearComplianceArea(prescribed, compliance, maximum)
		if err != nil {
			return restrictionflow.AreaLaw{}, modelDiagnostic("/area", classifyModelError(err))
		}
		return law, nil
	default:
		return restrictionflow.AreaLaw{}, schemaDiagnostic("/area/law", "unsupported_area_law")
	}
}

func classifyModelError(err error) string {
	switch {
	case errors.Is(err, restrictionflow.ErrNonFiniteValue):
		return "nonfinite_quantity"
	case errors.Is(err, restrictionflow.ErrNonPositiveValue):
		return "nonpositive_quantity"
	case errors.Is(err, restrictionflow.ErrNegativeArea):
		return "negative_area"
	case errors.Is(err, restrictionflow.ErrNegativeCompliance):
		return "negative_compliance"
	case errors.Is(err, restrictionflow.ErrInvalidDischargeCoefficient):
		return "invalid_discharge_coefficient"
	case errors.Is(err, restrictionflow.ErrInvalidHeatCapacityRatio):
		return "invalid_heat_capacity_ratio"
	case errors.Is(err, restrictionflow.ErrInvalidAreaLaw):
		return "invalid_area_law"
	case errors.Is(err, restrictionflow.ErrInvalidStagnation):
		return "invalid_stagnation"
	case errors.Is(err, restrictionflow.ErrAdversePressure):
		return "adverse_pressure"
	case errors.Is(err, restrictionflow.ErrNoRepresentableFlow):
		return "no_representable_flow"
	default:
		return "numerical_domain_error"
	}
}

func schemaDiagnostic(path, reason string) *Diagnostic {
	return &Diagnostic{Code: "FART-E-SCHEMA-0003", Stage: "schema", Path: path, ReasonCode: reason}
}

func modelDiagnostic(path, reason string) *Diagnostic {
	return &Diagnostic{Code: "FART-E-MODEL-0003", Stage: "model", Path: path, ReasonCode: reason}
}

func preflightJSON(data []byte) *Diagnostic {
	issue := strictjson.Inspect(data, strictjson.Limits{
		MaximumDepth: maximumJSONDepth, MaximumMemberNameBytes: maximumMemberNameBytes,
	})
	if issue == nil {
		return nil
	}
	return &Diagnostic{
		Code: "FART-E-JSON-0003", Stage: "syntax", Path: issue.Path, ReasonCode: string(issue.Kind),
	}
}
