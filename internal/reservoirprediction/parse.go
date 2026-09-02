package reservoirprediction

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/blisspixel/fartapp/internal/idealmixturereservoir"
	"github.com/blisspixel/fartapp/internal/lawcatalog"
	"github.com/blisspixel/fartapp/internal/strictjson"
)

type requestDocument struct {
	Schema             string              `json:"schema"`
	Model              requestModel        `json:"model"`
	QuantitySystem     string              `json:"quantity_system"`
	Closure            string              `json:"closure"`
	WithdrawalFraction *float64            `json:"withdrawal_fraction"`
	Initial            requestInitialState `json:"initial"`
}

type requestModel struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

type requestInitialState struct {
	Components        []requestComponent `json:"components"`
	VolumeCubicMetres *float64           `json:"volume_m3"`
	TemperatureKelvin *float64           `json:"temperature_k"`
}

type requestComponent struct {
	ID                                      string   `json:"id"`
	MassKilograms                           *float64 `json:"mass_kg"`
	SpecificGasConstantJoulesPerKilogramK   *float64 `json:"specific_gas_constant_j_per_kg_k"`
	IsochoricHeatCapacityJoulesPerKilogramK *float64 `json:"isochoric_heat_capacity_j_per_kg_k"`
	sourceIndex                             int
}

type parsedRequest struct {
	ids        []string
	state      idealmixturereservoir.State
	withdrawal idealmixturereservoir.WithdrawalFraction
	closure    idealmixturereservoir.Closure
}

func parseRequest(data []byte) (parsedRequest, *Diagnostic) {
	if diagnostic := preflightJSON(data); diagnostic != nil {
		return parsedRequest{}, diagnostic
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
	if document.WithdrawalFraction == nil {
		return parsedRequest{}, schemaDiagnostic("/withdrawal_fraction", "missing_member")
	}
	withdrawal, err := idealmixturereservoir.NewWithdrawalFraction(*document.WithdrawalFraction)
	if err != nil {
		return parsedRequest{}, modelDiagnostic("/withdrawal_fraction", classifyModelError(err))
	}
	closure, ok := parseClosure(document.Closure)
	if !ok {
		return parsedRequest{}, schemaDiagnostic("/closure", "unsupported_closure")
	}
	if document.Initial.VolumeCubicMetres == nil {
		return parsedRequest{}, schemaDiagnostic("/initial/volume_m3", "missing_member")
	}
	if document.Initial.TemperatureKelvin == nil {
		return parsedRequest{}, schemaDiagnostic("/initial/temperature_k", "missing_member")
	}
	volume, err := idealmixturereservoir.NewVolume(*document.Initial.VolumeCubicMetres)
	if err != nil {
		return parsedRequest{}, modelDiagnostic("/initial/volume_m3", classifyModelError(err))
	}
	temperature, err := idealmixturereservoir.NewTemperature(*document.Initial.TemperatureKelvin)
	if err != nil {
		return parsedRequest{}, modelDiagnostic("/initial/temperature_k", classifyModelError(err))
	}
	if len(document.Initial.Components) == 0 {
		return parsedRequest{}, schemaDiagnostic("/initial/components", "missing_component")
	}
	if len(document.Initial.Components) > idealmixturereservoir.MaxComponents {
		return parsedRequest{}, schemaDiagnostic("/initial/components", "collection_limit_exceeded")
	}
	seenIDs := make(map[string]struct{}, len(document.Initial.Components))
	for index := range document.Initial.Components {
		component := &document.Initial.Components[index]
		component.sourceIndex = index
		path := fmt.Sprintf("/initial/components/%d/id", index)
		if err := lawcatalog.ValidateMachineToken(component.ID); err != nil {
			return parsedRequest{}, schemaDiagnostic(path, "invalid_token")
		}
		if _, exists := seenIDs[component.ID]; exists {
			return parsedRequest{}, schemaDiagnostic(path, "duplicate_component_id")
		}
		seenIDs[component.ID] = struct{}{}
	}
	sort.Slice(document.Initial.Components, func(left, right int) bool {
		return document.Initial.Components[left].ID < document.Initial.Components[right].ID
	})
	components := make([]idealmixturereservoir.Component, len(document.Initial.Components))
	ids := make([]string, len(document.Initial.Components))
	for index, component := range document.Initial.Components {
		path := fmt.Sprintf("/initial/components/%d", component.sourceIndex)
		if component.MassKilograms == nil || component.SpecificGasConstantJoulesPerKilogramK == nil ||
			component.IsochoricHeatCapacityJoulesPerKilogramK == nil {
			return parsedRequest{}, schemaDiagnostic(path, "missing_member")
		}
		mass, massErr := idealmixturereservoir.NewMass(*component.MassKilograms)
		gasConstant, gasErr := idealmixturereservoir.NewSpecificGasConstant(
			*component.SpecificGasConstantJoulesPerKilogramK,
		)
		heatCV, heatErr := idealmixturereservoir.NewIsochoricHeatCapacity(
			*component.IsochoricHeatCapacityJoulesPerKilogramK,
		)
		if massErr != nil {
			return parsedRequest{}, modelDiagnostic(path+"/mass_kg", classifyModelError(massErr))
		}
		if gasErr != nil {
			return parsedRequest{}, modelDiagnostic(
				path+"/specific_gas_constant_j_per_kg_k",
				classifyModelError(gasErr),
			)
		}
		if heatErr != nil {
			return parsedRequest{}, modelDiagnostic(
				path+"/isochoric_heat_capacity_j_per_kg_k",
				classifyModelError(heatErr),
			)
		}
		components[index], err = idealmixturereservoir.NewComponent(mass, gasConstant, heatCV)
		if err != nil {
			return parsedRequest{}, modelDiagnostic(path, classifyModelError(err))
		}
		ids[index] = component.ID
	}
	state, err := idealmixturereservoir.NewState(components, volume, temperature)
	if err != nil {
		return parsedRequest{}, modelDiagnostic("/initial", classifyModelError(err))
	}
	return parsedRequest{
		ids: ids, state: state,
		withdrawal: withdrawal, closure: closure,
	}, nil
}

func parseClosure(value string) (idealmixturereservoir.Closure, bool) {
	switch value {
	case "rigid-adiabatic":
		return idealmixturereservoir.RigidAdiabatic, true
	case "rigid-isothermal":
		return idealmixturereservoir.RigidIsothermal, true
	default:
		return 0, false
	}
}

func classifyModelError(err error) string {
	switch {
	case errors.Is(err, idealmixturereservoir.ErrNonFiniteValue):
		return "nonfinite_quantity"
	case errors.Is(err, idealmixturereservoir.ErrNonPositiveValue):
		return "nonpositive_quantity"
	case errors.Is(err, idealmixturereservoir.ErrInvalidWithdrawal):
		return "invalid_withdrawal"
	case errors.Is(err, idealmixturereservoir.ErrReservoirExhausted):
		return "reservoir_depletion"
	case errors.Is(err, idealmixturereservoir.ErrNoRepresentableProgress):
		return "no_representable_progress"
	case errors.Is(err, idealmixturereservoir.ErrInvalidComponentSet):
		return "invalid_component_set"
	default:
		return "numerical_domain_error"
	}
}

func schemaDiagnostic(path, reason string) *Diagnostic {
	return &Diagnostic{Code: "FART-E-SCHEMA-0002", Stage: "schema", Path: path, ReasonCode: reason}
}

func modelDiagnostic(path, reason string) *Diagnostic {
	return &Diagnostic{Code: "FART-E-MODEL-0001", Stage: "model", Path: path, ReasonCode: reason}
}

func preflightJSON(data []byte) *Diagnostic {
	issue := strictjson.Inspect(data, strictjson.Limits{
		MaximumDepth: maximumJSONDepth, MaximumMemberNameBytes: maximumMemberNameBytes,
	})
	if issue == nil {
		return nil
	}
	reason := string(issue.Kind)
	return &Diagnostic{
		Code: "FART-E-JSON-0002", Stage: "syntax", Path: issue.Path, ReasonCode: reason,
	}
}
