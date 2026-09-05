package walkcase

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"sort"
	"strconv"

	"github.com/blisspixel/fartapp/internal/coupledblowdown"
	"github.com/blisspixel/fartapp/internal/idealmixturereservoir"
	"github.com/blisspixel/fartapp/internal/lawcatalog"
	"github.com/blisspixel/fartapp/internal/restrictionflow"
	"github.com/blisspixel/fartapp/internal/strictjson"
)

var (
	ErrInputTooLarge = errors.New("walk-case input exceeds the byte limit")
	ErrNilInput      = errors.New("walk-case input reader is nil")
)

type requestDocument struct {
	Schema          string             `json:"schema"`
	Model           ModelReference     `json:"model"`
	QuantitySystem  string             `json:"quantity_system"`
	LawContext      *lawContextRef     `json:"law_context,omitempty"`
	Closure         string             `json:"closure"`
	Reservoir       requestReservoir   `json:"reservoir"`
	Restriction     requestRestriction `json:"restriction"`
	Step            requestStep        `json:"step"`
	Branch          *requestBranch     `json:"branch,omitempty"`
	ExpectedWitness *string            `json:"expected_witness,omitempty"`
}

type lawContextRef struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

type requestReservoir struct {
	Components        []requestComponent `json:"components"`
	VolumeCubicMetres *float64           `json:"volume_m3"`
	TemperatureKelvin *float64           `json:"temperature_k"`
}

type requestComponent struct {
	ID                                      string   `json:"id"`
	MassKilograms                           *float64 `json:"mass_kg"`
	SpecificGasConstantJoulesPerKilogramK   *float64 `json:"specific_gas_constant_j_per_kg_k"`
	IsochoricHeatCapacityJoulesPerKilogramK *float64 `json:"isochoric_heat_capacity_j_per_kg_k"`
}

type requestRestriction struct {
	BackPressurePascals  *float64    `json:"back_pressure_pa"`
	DischargeCoefficient *float64    `json:"discharge_coefficient"`
	Area                 requestArea `json:"area"`
}

type requestArea struct {
	Law                    string   `json:"law"`
	PrescribedSquareMetres *float64 `json:"prescribed_m2"`
	ComplianceSquareMetres *float64 `json:"compliance_m2_per_pa,omitempty"`
	MaximumSquareMetres    *float64 `json:"maximum_m2,omitempty"`
}

type requestStep struct {
	MaxWithdrawalFraction *float64 `json:"max_withdrawal_fraction_per_step"`
	MaxSteps              *int     `json:"max_steps"`
	MaxTimeSeconds        *float64 `json:"max_time_s"`
}

type requestBranch struct {
	PrescribedAreaSquareMetres *float64 `json:"prescribed_area_m2"`
}

type parsedCase struct {
	document        requestDocument
	expectedWitness string
	lawContext      string
	closure         idealmixturereservoir.Closure
	state           idealmixturereservoir.State
	back            restrictionflow.Pressure
	area            restrictionflow.AreaLaw
	cd              restrictionflow.DischargeCoefficient
	fraction        float64
	maxSteps        int
	maxTime         float64
	branchArea      *float64
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

func parseCase(data []byte) (parsedCase, *Diagnostic) {
	if issue := strictjson.Inspect(data, strictjson.Limits{
		MaximumDepth: maximumJSONDepth, MaximumMemberNameBytes: maximumMemberNameBytes,
	}); issue != nil {
		return parsedCase{}, &Diagnostic{
			Code: "FART-E-JSON-0005", Stage: "syntax", Path: issue.Path, ReasonCode: string(issue.Kind),
		}
	}
	if diagnostic := inspectDocumentShape(data); diagnostic != nil {
		return parsedCase{}, diagnostic
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var document requestDocument
	if err := decoder.Decode(&document); err != nil {
		return parsedCase{}, schema("/", "document_shape_invalid")
	}
	if document.Schema != RequestSchema {
		return parsedCase{}, schema("/schema", "unsupported_schema")
	}
	if document.Model.ID != ModelID || document.Model.Version != ModelVersion {
		return parsedCase{}, schema("/model", "unsupported_model_revision")
	}
	if document.QuantitySystem != QuantitySystem {
		return parsedCase{}, schema("/quantity_system", "unsupported_quantity_system")
	}
	lawContext := ""
	if document.LawContext != nil {
		if document.LawContext.Version == "" {
			return parsedCase{}, schema("/law_context/version", "exact_law_revision_required")
		}
		reference := document.LawContext.ID
		if document.LawContext.Version != "" {
			reference += "@" + document.LawContext.Version
		}
		if _, found := lawcatalog.Inspect(reference); !found {
			return parsedCase{}, schema("/law_context", "unresolved_law_context")
		}
		if reference != "earth.continuum.si@v0alpha1" {
			return parsedCase{}, model("/law_context", "incompatible_law_context")
		}
		lawContext = reference
	}
	closure, ok := parseClosure(document.Closure)
	if !ok {
		return parsedCase{}, schema("/closure", "unsupported_closure")
	}
	sort.Slice(document.Reservoir.Components, func(i, j int) bool {
		return document.Reservoir.Components[i].ID < document.Reservoir.Components[j].ID
	})
	state, diagnostic := parseReservoir(document.Reservoir)
	if diagnostic != nil {
		return parsedCase{}, diagnostic
	}
	if document.Restriction.BackPressurePascals == nil || document.Restriction.DischargeCoefficient == nil {
		return parsedCase{}, schema("/restriction", "missing_member")
	}
	back, err := restrictionflow.NewPressure(*document.Restriction.BackPressurePascals)
	if err != nil {
		return parsedCase{}, model("/restriction/back_pressure_pa", "nonpositive_quantity")
	}
	cd, err := restrictionflow.NewDischargeCoefficient(*document.Restriction.DischargeCoefficient)
	if err != nil {
		return parsedCase{}, model("/restriction/discharge_coefficient", "invalid_discharge_coefficient")
	}
	area, diagnostic := parseArea(document.Restriction.Area)
	if diagnostic != nil {
		return parsedCase{}, diagnostic
	}
	if document.Step.MaxWithdrawalFraction == nil || document.Step.MaxSteps == nil {
		return parsedCase{}, schema("/step", "missing_member")
	}
	maxTime := 0.0
	if document.Step.MaxTimeSeconds != nil {
		maxTime = *document.Step.MaxTimeSeconds
	}
	document.Step.MaxTimeSeconds = &maxTime
	if _, err := coupledblowdown.NewConfig(state, closure, back, area, cd,
		*document.Step.MaxWithdrawalFraction, *document.Step.MaxSteps, maxTime); err != nil {
		return parsedCase{}, model("/step", classify(err))
	}
	var branchArea *float64
	if document.Branch != nil {
		if document.Branch.PrescribedAreaSquareMetres == nil {
			return parsedCase{}, schema("/branch/prescribed_area_m2", "missing_member")
		}
		value := *document.Branch.PrescribedAreaSquareMetres
		if _, err := restrictionflow.NewArea(value); err != nil {
			return parsedCase{}, model("/branch/prescribed_area_m2", "negative_area")
		}
		branchArea = &value
	}
	expected := ""
	if document.ExpectedWitness != nil {
		expected = *document.ExpectedWitness
		if !validDigest(expected) {
			return parsedCase{}, schema("/expected_witness", "invalid_witness_digest")
		}
	}
	document.ExpectedWitness = nil
	return parsedCase{
		document:        document,
		expectedWitness: expected,
		lawContext:      lawContext,
		closure:         closure,
		state:           state,
		back:            back,
		area:            area,
		cd:              cd,
		fraction:        *document.Step.MaxWithdrawalFraction,
		maxSteps:        *document.Step.MaxSteps,
		maxTime:         maxTime,
		branchArea:      branchArea,
	}, nil
}

func parseReservoir(document requestReservoir) (idealmixturereservoir.State, *Diagnostic) {
	if document.VolumeCubicMetres == nil || document.TemperatureKelvin == nil {
		return idealmixturereservoir.State{}, schema("/reservoir", "missing_member")
	}
	if len(document.Components) == 0 {
		return idealmixturereservoir.State{}, schema("/reservoir/components", "missing_component")
	}
	if len(document.Components) > idealmixturereservoir.MaxComponents {
		return idealmixturereservoir.State{}, schema("/reservoir/components", "too_many_components")
	}
	volume, err := idealmixturereservoir.NewVolume(*document.VolumeCubicMetres)
	if err != nil {
		return idealmixturereservoir.State{}, model("/reservoir/volume_m3", "nonpositive_quantity")
	}
	temperature, err := idealmixturereservoir.NewTemperature(*document.TemperatureKelvin)
	if err != nil {
		return idealmixturereservoir.State{}, model("/reservoir/temperature_k", "nonpositive_quantity")
	}
	components := make([]idealmixturereservoir.Component, len(document.Components))
	seen := map[string]struct{}{}
	for index, component := range document.Components {
		path := "/reservoir/components/" + strconv.Itoa(index)
		if err := lawcatalog.ValidateMachineToken(component.ID); err != nil {
			return idealmixturereservoir.State{}, schema(path+"/id", "invalid_token")
		}
		if _, exists := seen[component.ID]; exists {
			return idealmixturereservoir.State{}, schema(path+"/id", "duplicate_component_id")
		}
		seen[component.ID] = struct{}{}
		if component.MassKilograms == nil || component.SpecificGasConstantJoulesPerKilogramK == nil ||
			component.IsochoricHeatCapacityJoulesPerKilogramK == nil {
			return idealmixturereservoir.State{}, schema(path, "missing_member")
		}
		mass, err := idealmixturereservoir.NewMass(*component.MassKilograms)
		if err != nil {
			return idealmixturereservoir.State{}, model(path+"/mass_kg", "nonpositive_quantity")
		}
		gas, err := idealmixturereservoir.NewSpecificGasConstant(*component.SpecificGasConstantJoulesPerKilogramK)
		if err != nil {
			return idealmixturereservoir.State{}, model(path+"/specific_gas_constant_j_per_kg_k", "nonpositive_quantity")
		}
		cv, err := idealmixturereservoir.NewIsochoricHeatCapacity(*component.IsochoricHeatCapacityJoulesPerKilogramK)
		if err != nil {
			return idealmixturereservoir.State{}, model(path+"/isochoric_heat_capacity_j_per_kg_k", "nonpositive_quantity")
		}
		built, err := idealmixturereservoir.NewComponent(mass, gas, cv)
		if err != nil {
			return idealmixturereservoir.State{}, model("/reservoir/components", "invalid_component")
		}
		components[index] = built
	}
	state, err := idealmixturereservoir.NewState(components, volume, temperature)
	if err != nil {
		return idealmixturereservoir.State{}, model("/reservoir", "invalid_state")
	}
	return state, nil
}

func parseArea(document requestArea) (restrictionflow.AreaLaw, *Diagnostic) {
	if document.PrescribedSquareMetres == nil {
		return restrictionflow.AreaLaw{}, schema("/restriction/area/prescribed_m2", "missing_member")
	}
	prescribed, err := restrictionflow.NewArea(*document.PrescribedSquareMetres)
	if err != nil {
		return restrictionflow.AreaLaw{}, model("/restriction/area/prescribed_m2", "negative_area")
	}
	switch document.Law {
	case "prescribed":
		if document.ComplianceSquareMetres != nil || document.MaximumSquareMetres != nil {
			return restrictionflow.AreaLaw{}, schema("/restriction/area", "unexpected_area_member")
		}
		law, err := restrictionflow.NewPrescribedArea(prescribed)
		if err != nil {
			return restrictionflow.AreaLaw{}, model("/restriction/area", "invalid_area_law")
		}
		return law, nil
	case "linear-compliance":
		if document.ComplianceSquareMetres == nil || document.MaximumSquareMetres == nil {
			return restrictionflow.AreaLaw{}, schema("/restriction/area", "missing_member")
		}
		compliance, err := restrictionflow.NewAreaCompliance(*document.ComplianceSquareMetres)
		if err != nil {
			return restrictionflow.AreaLaw{}, model("/restriction/area/compliance_m2_per_pa", "negative_compliance")
		}
		maximum, err := restrictionflow.NewArea(*document.MaximumSquareMetres)
		if err != nil {
			return restrictionflow.AreaLaw{}, model("/restriction/area/maximum_m2", "negative_area")
		}
		law, err := restrictionflow.NewLinearComplianceArea(prescribed, compliance, maximum)
		if err != nil {
			return restrictionflow.AreaLaw{}, model("/restriction/area", "invalid_area_law")
		}
		return law, nil
	default:
		return restrictionflow.AreaLaw{}, schema("/restriction/area/law", "unsupported_area_law")
	}
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

func schema(path, reason string) *Diagnostic {
	return &Diagnostic{Code: "FART-E-SCHEMA-0005", Stage: "schema", Path: path, ReasonCode: reason}
}

func model(path, reason string) *Diagnostic {
	return &Diagnostic{Code: "FART-E-MODEL-0006", Stage: "model", Path: path, ReasonCode: reason}
}
