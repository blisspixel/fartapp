package reservoirprediction

import (
	"math"

	"github.com/blisspixel/fartapp/internal/idealmixturereservoir"
)

func buildReport(request parsedRequest, transition idealmixturereservoir.Transition) Report {
	withdrawalFraction := request.withdrawal.Value()
	initial := stateReport(request.ids, transition.Before())
	final := stateReport(request.ids, transition.After())
	massOut := transition.ComponentMassOut()
	componentResiduals := transition.ComponentMassResiduals()
	componentTransfers := make([]ComponentTransfer, len(request.ids))
	componentBalances := make([]ComponentMassBalance, len(request.ids))
	for index, id := range request.ids {
		componentTransfers[index] = ComponentTransfer{
			ID: id, MassOutKilograms: massOut[index].Kilograms(),
		}
		componentBalances[index] = ComponentMassBalance{
			ID: id, ResidualKilograms: componentResiduals[index].Kilograms(),
		}
	}
	transfers := Transfers{
		Components:                    componentTransfers,
		TotalMassOutKilograms:         transition.TotalMassOut().Kilograms(),
		IntegratedEnthalpyOutJoules:   transition.IntegratedEnthalpyOut().Joules(),
		HeatIntoReservoirJoules:       transition.HeatIntoReservoir().Joules(),
		BoundaryWorkByReservoirJoules: transition.BoundaryWorkByReservoir().Joules(),
	}
	balances := Balances{
		Components:                 componentBalances,
		TotalMassResidualKilograms: transition.MassBalanceResidual().Kilograms(),
		EnergyResidualJoules:       transition.EnergyBalanceResidual().Joules(),
		InitialEOSResidualJoules:   equationOfStateResidual(initial),
		FinalEOSResidualJoules:     equationOfStateResidual(final),
	}
	claims := buildClaims(initial, final, transfers, balances)
	for _, claim := range claims {
		if claim.Status != "satisfied-within-roundoff" {
			return failure("FART-E-NUMERICAL-0001", "model", "/", "invariant_violation")
		}
	}
	assumptions := []string{
		"calorically-perfect-components",
		"homogeneous-equilibrium-state",
		"nonreacting-mixture",
		"single-gas-phase",
		"perfectly-mixed-nonselective-outflow",
		"rigid-volume",
		"sensible-energy-datum-cv-times-temperature",
	}
	if request.closure == idealmixturereservoir.RigidAdiabatic {
		assumptions = append(assumptions, "adiabatic-no-heat-transfer")
	} else {
		assumptions = append(assumptions, "prescribed-isothermal-ideal-thermostat")
	}
	return Report{
		Schema:                 ReportSchema,
		Status:                 "predicted",
		RequestSchema:          RequestSchema,
		Model:                  &ModelReference{ID: ModelID, Version: ModelVersion},
		ImplementationRevision: ImplementationRevision,
		QuantitySystem:         QuantitySystem,
		Closure:                request.closure.String(),
		WithdrawalFraction:     &withdrawalFraction,
		Initial:                &initial,
		Final:                  &final,
		Transfers:              &transfers,
		Balances:               &balances,
		Assumptions:            assumptions,
		Nonclaims: &Nonclaims{
			Model: []string{
				"aperture-and-restriction-flow", "elapsed-time-history", "exterior-state",
				"momentum-and-recoil", "phase-change-and-reaction", "plume-and-acoustics",
			},
			Operation: []string{"case-commitment", "certificate-issuance"},
			Evidence:  []string{"empirical-validation"},
		},
		Claims: claims,
		ValidationEnvironment: ValidationEnvironment{
			ConsultedInputs: []string{"document_bytes"}, AmbientInputs: []string{},
		},
	}
}

func stateReport(ids []string, state idealmixturereservoir.State) ReservoirState {
	components := state.Components()
	reportComponents := make([]ComponentState, len(components))
	for index, component := range components {
		reportComponents[index] = ComponentState{
			ID:            ids[index],
			MassKilograms: component.Mass().Kilograms(),
			SpecificGasConstantJoulesPerKilogramKelvin:           component.SpecificGasConstant().JoulesPerKilogramKelvin(),
			SpecificIsochoricHeatCapacityJoulesPerKilogramKelvin: component.IsochoricHeatCapacity().JoulesPerKilogramKelvin(),
		}
	}
	return ReservoirState{
		Components:         reportComponents,
		TotalMassKilograms: state.TotalMass().Kilograms(),
		VolumeCubicMetres:  state.Volume().CubicMetres(),
		TemperatureKelvin:  state.Temperature().Kelvin(),
		MixtureGasConstantJoulesPerKilogramKelvin:           state.MixtureGasConstant().JoulesPerKilogramKelvin(),
		MixtureSpecificIsochoricHeatJoulesPerKilogramKelvin: state.MixtureIsochoricHeatCapacity().JoulesPerKilogramKelvin(),
		MixtureSpecificIsobaricHeatJoulesPerKilogramKelvin:  state.MixtureIsobaricHeatCapacity().JoulesPerKilogramKelvin(),
		HeatCapacityRatio:    state.HeatCapacityRatio(),
		PressurePascals:      state.Pressure().Pascals(),
		InternalEnergyJoules: state.InternalEnergy().Joules(),
	}
}

func equationOfStateResidual(state ReservoirState) float64 {
	return state.PressurePascals*state.VolumeCubicMetres -
		state.TotalMassKilograms*state.MixtureGasConstantJoulesPerKilogramKelvin*state.TemperatureKelvin
}

func buildClaims(initial, final ReservoirState, transfers Transfers, balances Balances) []Claim {
	initialPressureVolume := initial.PressurePascals * initial.VolumeCubicMetres
	initialMassGasTemperature := initial.TotalMassKilograms *
		initial.MixtureGasConstantJoulesPerKilogramKelvin * initial.TemperatureKelvin
	finalPressureVolume := final.PressurePascals * final.VolumeCubicMetres
	finalMassGasTemperature := final.TotalMassKilograms *
		final.MixtureGasConstantJoulesPerKilogramKelvin * final.TemperatureKelvin
	return []Claim{
		newClaim("reservoir.total-mass-balance", "double-entry-balance", "kg", balances.TotalMassResidualKilograms, roundoffTolerance(
			initial.TotalMassKilograms, final.TotalMassKilograms, transfers.TotalMassOutKilograms,
		)),
		newClaim("reservoir.energy-balance", "double-entry-balance", "J", balances.EnergyResidualJoules, roundoffTolerance(
			initial.InternalEnergyJoules, final.InternalEnergyJoules,
			transfers.IntegratedEnthalpyOutJoules, transfers.HeatIntoReservoirJoules,
			transfers.BoundaryWorkByReservoirJoules,
		)),
		newClaim("reservoir.initial-equation-of-state", "derived-state-consistency-residual", "J", balances.InitialEOSResidualJoules, roundoffTolerance(
			initialPressureVolume, initialMassGasTemperature,
		)),
		newClaim("reservoir.final-equation-of-state", "derived-state-consistency-residual", "J", balances.FinalEOSResidualJoules, roundoffTolerance(
			finalPressureVolume, finalMassGasTemperature,
		)),
	}
}

func newClaim(id, method, unit string, residual, tolerance float64) Claim {
	status := "failed"
	if finite(residual) && finite(tolerance) && tolerance >= 0 && math.Abs(residual) <= tolerance {
		status = "satisfied-within-roundoff"
	}
	return Claim{
		ID: id, Status: status, Method: method,
		EquationRevision: ModelID + "@" + ModelVersion,
		Residual:         residual, Tolerance: tolerance, ResidualUnit: unit,
	}
}

func roundoffTolerance(terms ...float64) float64 {
	const operationsAllowance = 64
	epsilon := math.Nextafter(1, 2) - 1
	maximumMagnitude := 0.0
	for _, term := range terms {
		magnitude := math.Abs(term)
		if !finite(magnitude) {
			return math.NaN()
		}
		maximumMagnitude = math.Max(maximumMagnitude, magnitude)
	}
	return (operationsAllowance*epsilon)*maximumMagnitude + math.SmallestNonzeroFloat64
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
