package restrictionprediction

import (
	"math"

	"github.com/blisspixel/fartapp/internal/restrictionflow"
)

func buildReport(result restrictionflow.Result) Report {
	request := result.Request()
	stagnation := request.Stagnation()
	back := request.BackPressure().Pascals()
	cd := request.DischargeCoefficient().Value()
	areaLaw := request.AreaLaw()
	area := AreaState{
		Law:                    areaLaw.LawName(),
		PrescribedSquareMetres: areaLaw.Prescribed().SquareMetres(),
		EffectiveSquareMetres:  result.EffectiveArea().SquareMetres(),
	}
	if areaLaw.LawName() == "linear-compliance" {
		compliance := areaLaw.Compliance().SquareMetresPerPascal()
		maximum := areaLaw.Maximum().SquareMetres()
		area.ComplianceSquareMetresPa = &compliance
		area.MaximumSquareMetres = &maximum
	}
	flow := FlowState{
		Regime:                     result.Regime().String(),
		CriticalPressureRatio:      result.CriticalPressureRatio(),
		BackPressureRatio:          result.BackPressureRatio(),
		ThroatMach:                 result.ThroatMach().Value(),
		ExitPressurePascals:        result.ExitPressure().Pascals(),
		ExitTemperatureKelvin:      result.ExitTemperature().Kelvin(),
		ExitSpeedMetresPerSecond:   result.ExitSpeed().MetresPerSecond(),
		MassFlowKilogramsPerSecond: result.MassFlow().KilogramsPerSecond(),
		SonicMassFlowKilogramsPerS: result.SonicMassFlow().KilogramsPerSecond(),
		ThrustNewtons:              result.Thrust().Newtons(),
		RecoilNewtons:              result.Recoil().Newtons(),
	}
	balances := Balances{
		MassFlowResidualKilogramsPerSecond: result.MassFlowResidual().KilogramsPerSecond(),
		ThrustResidualNewtons:              result.ThrustResidual().Newtons(),
		RecoilResidualNewtons:              result.RecoilResidual().Newtons(),
	}
	stagnationState := StagnationState{
		PressurePascals:   stagnation.Pressure().Pascals(),
		TemperatureKelvin: stagnation.Temperature().Kelvin(),
		SpecificGasConstantJoulesPerKilogramKelvin: stagnation.SpecificGasConstant().JoulesPerKilogramKelvin(),
		HeatCapacityRatio:                          stagnation.HeatCapacityRatio().Value(),
	}
	claims := buildClaims(flow, balances)
	for _, claim := range claims {
		if claim.Status != "satisfied-within-roundoff" {
			return failure("FART-E-NUMERICAL-0002", "model", "/", "invariant_violation")
		}
	}
	return Report{
		Schema:                 ReportSchema,
		Status:                 "predicted",
		RequestSchema:          RequestSchema,
		Model:                  &ModelReference{ID: ModelID, Version: ModelVersion},
		ImplementationRevision: ImplementationRevision,
		QuantitySystem:         QuantitySystem,
		Stagnation:             &stagnationState,
		BackPressurePascals:    &back,
		DischargeCoefficient:   &cd,
		Area:                   &area,
		Flow:                   &flow,
		Balances:               &balances,
		Assumptions: []string{
			"calorically-perfect-gas",
			"quasi-steady-flow",
			"isentropic-control-section",
			"converging-restriction-only",
			"discharge-coefficient-scales-mass-flow-only",
			"no-reverse-flow",
			"no-shock-inside-restriction",
			"prescribed-or-linear-quasi-static-area",
		},
		Nonclaims: &Nonclaims{
			Model: []string{
				"elapsed-time-history",
				"reservoir-mass-energy-coupling",
				"shock-containing-or-underexpanded-plume",
				"viscous-resolved-vena-contracta",
				"phase-change-and-reaction",
				"acoustics",
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

func buildClaims(flow FlowState, balances Balances) []Claim {
	return []Claim{
		newClaim(
			"restriction.mass-flow-definition",
			"cd-scaled-exit-mass-flux",
			"kg/s",
			balances.MassFlowResidualKilogramsPerSecond,
			roundoffTolerance(flow.MassFlowKilogramsPerSecond, flow.SonicMassFlowKilogramsPerS),
		),
		newClaim(
			"restriction.thrust-control-surface",
			"momentum-and-pressure-thrust",
			"N",
			balances.ThrustResidualNewtons,
			roundoffTolerance(flow.ThrustNewtons, flow.MassFlowKilogramsPerSecond*flow.ExitSpeedMetresPerSecond),
		),
		newClaim(
			"restriction.recoil-action-reaction",
			"equal-and-opposite-force",
			"N",
			balances.RecoilResidualNewtons,
			roundoffTolerance(flow.RecoilNewtons, flow.ThrustNewtons),
		),
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
