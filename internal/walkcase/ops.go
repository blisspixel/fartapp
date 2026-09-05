package walkcase

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"runtime"

	"github.com/blisspixel/fartapp/internal/coupledblowdown"
	"github.com/blisspixel/fartapp/internal/idealmixturereservoir"
	"github.com/blisspixel/fartapp/internal/restrictionflow"
)

func Run(data []byte, operation string) Report {
	if len(data) > MaxInputBytes {
		return failure("FART-E-INPUT-0005", "input", "/", "input_too_large")
	}
	parsed, diagnostic := parseCase(data)
	if diagnostic != nil {
		return failure(diagnostic.Code, diagnostic.Stage, diagnostic.Path, diagnostic.ReasonCode)
	}
	switch operation {
	case "predict":
		return runPredict(parsed)
	case "simulate", "inspect", "explain", "certify":
		return runSimulate(parsed, operation)
	case "branch":
		return runBranch(parsed)
	case "witness", "reconstruct":
		return runWitness(parsed, operation)
	default:
		return failure("FART-E-SCHEMA-0005", "schema", "/operation", "unsupported_operation")
	}
}

func InputFailure(reason string, consultedInputs ...string) Report {
	report := failure("FART-E-IO-0005", "input", "/", reason)
	report.ValidationEnvironment.ConsultedInputs = append([]string(nil), consultedInputs...)
	return report
}

func runPredict(parsed parsedCase) Report {
	flow, err := initialFlow(parsed)
	if err != nil {
		return failure("FART-E-MODEL-0006", "model", "/restriction", classify(err))
	}
	fraction, err := coupledblowdown.EqualizationFraction(parsed.state, parsed.back, parsed.closure)
	if err != nil {
		return failure("FART-E-MODEL-0006", "model", "/restriction/back_pressure_pa", classify(err))
	}
	report := baseReport(parsed, "predict")
	report.EndpointReachability = "finite-time-model-endpoint"
	if fraction == 0 {
		report.EndpointReachability = "already-equalized"
	} else if flow.Regime() == restrictionflow.RegimeNoFlow {
		fraction = 0
		report.EndpointReachability = "unreachable-closed-restriction"
	} else if parsed.area.Prescribed().SquareMetres() == 0 {
		report.EndpointReachability = "asymptotic-limit"
	}
	report.EqualizationFraction = &fraction
	report.InitialRestriction = snapshotFlow(flow)
	initial := endpointFromState(parsed.state)
	report.Initial = &initial
	if fraction > 0 {
		withdrawal, err := idealmixturereservoir.NewWithdrawalFraction(fraction)
		if err != nil {
			return failure("FART-E-MODEL-0006", "model", "/", classify(err))
		}
		transition, err := idealmixturereservoir.WithdrawFraction(parsed.state, withdrawal, parsed.closure)
		if err != nil {
			return failure("FART-E-MODEL-0006", "model", "/", classify(err))
		}
		final := endpointFromState(transition.After())
		report.Final = &final
		massOut := transition.TotalMassOut().Kilograms()
		report.MassOutKilograms = &massOut
	} else {
		report.Final = &initial
		zero := 0.0
		report.MassOutKilograms = &zero
	}
	report.Explanation = []string{
		"This is an analytical endpoint prediction; it does not predict an elapsed time.",
		"The reservoir closure determines pressure and temperature along the withdrawal path. The restriction determines how quickly that path is traversed.",
	}
	if report.EndpointReachability == "unreachable-closed-restriction" {
		report.Explanation = append(report.Explanation, "The restriction is closed, so this model predicts no withdrawal and preserves the initial state.")
		report.Claims = []Claim{newClaim("walk.predict-no-flow", "closed-restriction-identity", "kg", report.Final.MassKilograms-initial.MassKilograms, roundoffTolerance(initial.MassKilograms))}
	} else {
		report.Claims = []Claim{newClaim("walk.predict-pressure-endpoint", "reservoir-endpoint-at-back-pressure", "Pa", report.Final.PressurePascals-parsed.back.Pascals(), roundoffTolerance(report.Final.PressurePascals, parsed.back.Pascals()))}
		if report.EndpointReachability == "asymptotic-limit" {
			report.Explanation = append(report.Explanation, "This compliant area closes as the pressure difference vanishes. Equalization is an infinite-time model limit, not a finite-duration prediction.")
		}
	}
	return finish(report, nil)
}

func runSimulate(parsed parsedCase, operation string) Report {
	result, err := simulate(parsed, parsed.area)
	if err != nil {
		return failure("FART-E-MODEL-0006", "model", "/", classify(err))
	}
	report := simulateReport(parsed, operation, result)
	if operation == "explain" || operation == "certify" {
		report.Explanation = explain(parsed, result)
	}
	if operation == "certify" {
		report.Explanation = append(report.Explanation, "These are arithmetic balance checks for this numerical account. Passing them does not establish time-step accuracy, empirical validity, calibration, or approval by a certificate authority.")
	}
	return finish(report, &result)
}

func runBranch(parsed parsedCase) Report {
	if parsed.branchArea == nil {
		return failure("FART-E-SCHEMA-0005", "schema", "/branch", "missing_member")
	}
	baseline, err := simulate(parsed, parsed.area)
	if err != nil {
		return failure("FART-E-MODEL-0006", "model", "/", classify(err))
	}
	area, err := restrictionflow.NewArea(*parsed.branchArea)
	if err != nil {
		return failure("FART-E-MODEL-0006", "model", "/branch/prescribed_area_m2", classify(err))
	}
	law, err := restrictionflow.NewPrescribedArea(area)
	if err != nil {
		return failure("FART-E-MODEL-0006", "model", "/branch/prescribed_area_m2", classify(err))
	}
	variant, err := simulate(parsed, law)
	if err != nil {
		return failure("FART-E-MODEL-0006", "model", "/branch", classify(err))
	}
	report := simulateReport(parsed, "branch", baseline)
	variantParsed := parsed
	variantParsed.area = law
	variantParsed.document.Restriction.Area = requestArea{Law: "prescribed", PrescribedSquareMetres: parsed.branchArea}
	variantParsed.document.Branch = nil
	variantReport := finish(simulateReport(variantParsed, "simulate", variant), &variant)
	if !variantReport.Predicted() {
		return variantReport
	}
	tolerance := 256 * float64(1+baseline.Steps()+variant.Steps()) * (math.Nextafter(1, 2) - 1) * parsed.state.TotalMass().Kilograms()
	bothEqualized := baseline.Stop() == coupledblowdown.StopEqualized && variant.Stop() == coupledblowdown.StopEqualized
	same := bothEqualized && math.Abs(baseline.MassOutKilograms()-variant.MassOutKilograms()) <= tolerance
	report.Branch = &BranchComparison{
		PrescribedAreaSquareMetres: *parsed.branchArea,
		BaselineStop:               baseline.Stop().String(), VariantStop: variant.Stop().String(),
		BothEqualized: bothEqualized, MassComparisonToleranceKg: tolerance, Variant: &variantReport,
		BaselineElapsedSeconds: baseline.ElapsedSeconds(),
		VariantElapsedSeconds:  variant.ElapsedSeconds(),
		BaselineMassOutKg:      baseline.MassOutKilograms(),
		VariantMassOutKg:       variant.MassOutKilograms(),
		SameMassEndpoint:       same,
	}
	report.Explanation = []string{
		"This counterfactual changes the restriction to the declared prescribed area and keeps the initial reservoir, closure, back pressure, and stopping budgets.",
	}
	if bothEqualized {
		report.Explanation = append(report.Explanation, "Both calculations reached pressure equalization. Area changes the duration; the closure and back pressure determine the final mass.")
	} else {
		report.Explanation = append(report.Explanation, "At least one calculation stopped before pressure equalization. The reported masses and durations describe those stops, so this comparison makes no common-endpoint claim.")
	}
	return finish(report, &baseline)
}

func runWitness(parsed parsedCase, operation string) Report {
	if operation == "reconstruct" && parsed.expectedWitness == "" {
		return failure("FART-E-SCHEMA-0005", "schema", "/expected_witness", "missing_member")
	}
	result, err := simulate(parsed, parsed.area)
	if err != nil {
		return failure("FART-E-MODEL-0006", "model", "/", classify(err))
	}
	report := finish(simulateReport(parsed, "simulate", result), &result)
	if !report.Predicted() {
		return report
	}
	witness, inputDigest, err := witnessOf(report)
	if err != nil {
		return failure("FART-E-NUMERICAL-0004", "model", "/", "witness_encoding_failure")
	}
	report.Operation = operation
	report.Witness = witness
	report.WitnessSchema = WitnessSchema
	report.WitnessAlgorithm = "sha256"
	report.InputDigest = inputDigest
	report.InputDigestSchema = InputDigestSchema
	report.Explanation = []string{"This digest binds the normalized inputs and the full numerical account, including component identities, history, balances, model revision, and runtime profile. It is a software comparison, not an occurrence identity, signature, or empirical proof."}
	if operation == "reconstruct" {
		report.ExpectedWitness = parsed.expectedWitness
		report.ReconstructedWitness = witness
		match := witness == parsed.expectedWitness
		report.WitnessMatch = &match
		if !match {
			report.Status = "mismatch"
			report.Diagnostics = []Diagnostic{{Code: "FART-E-NUMERICAL-0004", Stage: "comparison", Path: "/expected_witness", ReasonCode: "witness_mismatch"}}
		}
	}
	return report
}

func simulate(parsed parsedCase, area restrictionflow.AreaLaw) (coupledblowdown.Result, error) {
	config, err := coupledblowdown.NewConfig(
		parsed.state, parsed.closure, parsed.back, area, parsed.cd,
		parsed.fraction, parsed.maxSteps, parsed.maxTime,
	)
	if err != nil {
		return coupledblowdown.Result{}, err
	}
	return coupledblowdown.Simulate(config)
}

func initialFlow(parsed parsedCase) (restrictionflow.Result, error) {
	pressure, err := restrictionflow.NewPressure(parsed.state.Pressure().Pascals())
	if err != nil {
		return restrictionflow.Result{}, err
	}
	temperature, err := restrictionflow.NewTemperature(parsed.state.Temperature().Kelvin())
	if err != nil {
		return restrictionflow.Result{}, err
	}
	gas, err := restrictionflow.NewSpecificGasConstant(parsed.state.MixtureGasConstant().JoulesPerKilogramKelvin())
	if err != nil {
		return restrictionflow.Result{}, err
	}
	gamma, err := restrictionflow.NewHeatCapacityRatio(parsed.state.HeatCapacityRatio())
	if err != nil {
		return restrictionflow.Result{}, err
	}
	stagnation, err := restrictionflow.NewStagnation(pressure, temperature, gas, gamma)
	if err != nil {
		return restrictionflow.Result{}, err
	}
	request, err := restrictionflow.NewRequest(stagnation, parsed.back, parsed.area, parsed.cd)
	if err != nil {
		return restrictionflow.Result{}, err
	}
	return restrictionflow.Evaluate(request)
}

func simulateReport(parsed parsedCase, operation string, result coupledblowdown.Result) Report {
	report := baseReport(parsed, operation)
	initial := endpointFromState(parsed.state)
	final := endpointFromState(result.Final())
	elapsed := result.ElapsedSeconds()
	steps := result.Steps()
	massOut := result.MassOutKilograms()
	enthalpy := result.EnthalpyOutJoules()
	heat := result.HeatInJoules()
	impulse := result.ImpulseNewtonSeconds()
	recoilImpulse := result.RecoilImpulseNewtonSeconds()
	pressureTolerance := result.EqualizationPressureTolerancePascals()
	report.Stop = result.Stop().String()
	report.ElapsedSeconds = &elapsed
	report.Steps = &steps
	report.Initial = &initial
	report.Final = &final
	report.MassOutKilograms = &massOut
	report.EnthalpyOutJoules = &enthalpy
	report.HeatInJoules = &heat
	report.ImpulseNewtonSeconds = &impulse
	report.RecoilImpulseNewtonSeconds = &recoilImpulse
	report.EqualizationPressureTolerancePascals = &pressureTolerance
	report.History = historySamples(parsed, result)
	flow, err := initialFlow(parsed)
	if err == nil {
		report.InitialRestriction = snapshotFlow(flow)
	}
	sig := result.Signature()
	signature := Signature{
		EquivalentDiameterMetres: sig.EquivalentDiameterMetres(),
		StrokeLengthMetres:       sig.StrokeLengthMetres(),
		ChokedOccurred:           sig.ChokedOccurred(),
	}
	if number, defined := sig.FormationNumber(); defined {
		signature.FormationNumber = &number
	}
	report.Signature = &signature
	return report
}

func baseReport(parsed parsedCase, operation string) Report {
	inputs := parsed.document
	return Report{
		Schema:                 ReportSchema,
		Status:                 "predicted",
		Operation:              operation,
		RequestSchema:          RequestSchema,
		ImplementationRevision: ImplementationRevision,
		Model:                  &ModelReference{ID: ModelID, Version: ModelVersion},
		NumericalPolicy: &NumericalPolicy{
			Method:    "first-order-explicit-mass-step-with-exact-reservoir-withdrawal",
			Precision: "ieee754-binary64", GoVersion: runtime.Version(),
			OperatingSystem: runtime.GOOS, Architecture: runtime.GOARCH,
			MaximumSamples:   coupledblowdown.MaxSteps + 1,
			HistorySemantics: "initial-and-each-completed-step-including-final-state",
		},
		Inputs:         &inputs,
		QuantitySystem: QuantitySystem,
		LawContext:     parsed.lawContext,
		Closure:        parsed.closure.String(),
		Dimensions:     dimensionDiagnostics(parsed.lawContext),
		Assumptions: []string{
			"rigid-calorically-perfect-ideal-mixture",
			"quasi-steady-isentropic-converging-restriction",
			"composition-preserving-outflow",
			"restriction-rate-reservoir-path",
			"circular-equivalent-diameter-from-prescribed-area",
		},
		Nonclaims: &Nonclaims{
			Model: []string{
				"viscous-resolved-aperture",
				"plume-and-acoustics",
				"heat-transfer-law",
				"vortex-ring-prediction",
				"complete-dry-flow-similarity",
				"empirical-validation",
			},
			Operation: []string{"case-commitment", "certificate-authority", "archive", "occurrence-identity", "reference-pfft-ratification"},
			Evidence:  []string{"empirical-validation", "solution-error-bound", "cross-platform-bitwise-reconstruction", "immutable-build-provenance", "cryptographic-signature"},
		},
		ValidationEnvironment: ValidationEnvironment{
			ConsultedInputs: []string{"document_bytes", "built_in_model", "implementation_runtime_profile"}, AmbientInputs: []string{},
		},
	}
}

func finish(report Report, result *coupledblowdown.Result) Report {
	if report.Claims == nil {
		report.Claims = claims(result)
	}
	for _, claim := range report.Claims {
		if claim.Status != "satisfied-within-roundoff" && claim.Status != "not-applicable" {
			report.Status = "invalid"
			report.Diagnostics = []Diagnostic{{Code: "FART-E-NUMERICAL-0004", Stage: "model", Path: "/claims", ReasonCode: "invariant_violation"}}
			return report
		}
	}
	return report
}

func claims(result *coupledblowdown.Result) []Claim {
	if result == nil {
		return nil
	}
	ledgers := result.Ledgers()
	tolerance := func(terms ...float64) float64 {
		return float64(1+result.Steps()) * roundoffTolerance(terms...)
	}
	return []Claim{
		newClaim("walk.mass-ledger", "double-entry-balance", "kg", ledgers.MassResidualKilograms(), tolerance(result.MassOutKilograms(), result.Final().TotalMass().Kilograms(), result.Config().Reservoir().TotalMass().Kilograms())),
		newClaim("walk.energy-ledger", "double-entry-balance", "J", ledgers.EnergyResidualJoules(), tolerance(result.EnthalpyOutJoules(), result.HeatInJoules(), result.Final().InternalEnergy().Joules(), result.Config().Reservoir().InternalEnergy().Joules())),
		newClaim("walk.impulse-ledger", "equal-and-opposite-force-accounting", "N s", ledgers.ImpulseResidualNewtonSeconds(), tolerance(result.ImpulseNewtonSeconds(), result.RecoilImpulseNewtonSeconds())),
	}
}

func newClaim(id, method, unit string, residual, tolerance float64) Claim {
	status := "failed"
	if finite(residual) && finite(tolerance) && math.Abs(residual) <= tolerance {
		status = "satisfied-within-roundoff"
	}
	return Claim{
		ID: id, Status: status, Method: method, EquationRevision: ModelID + "@" + ModelVersion,
		Residual: residual, Tolerance: tolerance, ResidualUnit: unit,
	}
}

func explain(parsed parsedCase, result coupledblowdown.Result) []string {
	lines := []string{
		"This is one explicit synthetic SI reservoir and restriction. It contains no biological default, exterior plume, sound, or Reference Pfft calibration.",
	}
	if parsed.closure == idealmixturereservoir.RigidIsothermal {
		lines = append(lines, "Temperature is prescribed constant. The reported heat input is the heat this closure requires; it is not a simulated wall heat-transfer law.")
	} else {
		lines = append(lines, "The rigid reservoir receives no heat. Its remaining gas cools as escaping gas carries source total enthalpy out.")
	}
	if result.Signature().ChokedOccurred() {
		lines = append(lines, "The critical pressure ratio was reached, so this model reached sonic flow at the restriction. That does not imply a resolved supersonic exterior plume.")
	} else {
		lines = append(lines, "The restriction remained subsonic or had no flow. The pressure ratio never reached this gas model's choking boundary.")
	}
	switch result.Stop().String() {
	case "equalized":
		lines = append(lines, "The calculation reached back pressure within the reported roundoff tolerance. This stopping check is not a bound on elapsed-time error.")
	case "pressure-tolerance":
		lines = append(lines, "The pressure difference reached the numerical tolerance. A compliant restriction with zero resting area approaches equalization asymptotically; this finite stop is not exact physical equalization.")
	case "no-flow":
		lines = append(lines, "The configured restriction permits no flow. The initial reservoir state is preserved.")
	case "max-time":
		lines = append(lines, "The declared time budget stopped the calculation before equalization. The final state is a budget-limited sample.")
	case "max-steps":
		lines = append(lines, "The declared step budget stopped the calculation before equalization. The final state is a budget-limited sample.")
	default:
		lines = append(lines, "The calculation could make no representable floating-point progress. Its final state does not establish physical equilibrium.")
	}
	lines = append(lines, "The time integrator is first order. Refine the withdrawal fraction and compare the same observable before interpreting numerical accuracy.")
	return lines
}

func dimensionDiagnostics(lawContext string) []DimensionDiagnostic {
	if lawContext == "" {
		return nil
	}
	return []DimensionDiagnostic{
		{Quantity: "mass", Unit: "kg", Dimension: "M", Status: "declared"},
		{Quantity: "length", Unit: "m", Dimension: "L", Status: "declared"},
		{Quantity: "time", Unit: "s", Dimension: "T", Status: "declared"},
		{Quantity: "temperature", Unit: "K", Dimension: "Theta", Status: "declared"},
		{Quantity: "pressure", Unit: "Pa", Dimension: "M L^-1 T^-2", Status: "declared"},
		{Quantity: "energy", Unit: "J", Dimension: "M L^2 T^-2", Status: "declared"},
	}
}

// witnessOf hashes the full simulation report under a separate versioned
// envelope. Go struct field order and encoding/json bytes define this bounded
// software format. This is deliberately not a canonical scientific identity.
func witnessOf(report Report) (string, string, error) {
	input, err := json.Marshal(struct {
		Schema string           `json:"schema"`
		Inputs *requestDocument `json:"inputs"`
	}{Schema: InputDigestSchema, Inputs: report.Inputs})
	if err != nil {
		return "", "", err
	}
	inputSum := sha256.Sum256(input)
	payload, err := json.Marshal(struct {
		Schema  string `json:"schema"`
		Account Report `json:"account"`
	}{Schema: WitnessSchema, Account: report})
	if err != nil {
		return "", "", err
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), hex.EncodeToString(inputSum[:]), nil
}

func historySamples(parsed parsedCase, result coupledblowdown.Result) []HistorySample {
	samples := result.Samples()
	history := make([]HistorySample, len(samples))
	for index, sample := range samples {
		masses := sample.ComponentMassesKilograms()
		massOut := sample.ComponentMassOutKilograms()
		components := make([]ComponentMass, len(masses))
		for componentIndex, mass := range masses {
			components[componentIndex] = ComponentMass{
				ID:            parsed.document.Reservoir.Components[componentIndex].ID,
				MassKilograms: mass, MassOutKilograms: massOut[componentIndex],
			}
		}
		history[index] = HistorySample{
			TimeSeconds: sample.Time(), MassKilograms: sample.Mass(),
			PressurePascals: sample.Pressure(), TemperatureKelvin: sample.Temperature(),
			MassFlowKilogramsPerSecond: sample.MassFlow(), SourceTotalEnthalpyWatts: sample.EnthalpyFlow(),
			ExitSpeedMetresPerSecond: sample.ExitSpeed(), ExitPressurePascals: sample.ExitPressure(),
			ExitTemperatureKelvin: sample.ExitTemperature(), EffectiveAreaSquareMetres: sample.EffectiveArea(),
			ThrustNewtons: sample.Thrust(), RecoilNewtons: sample.Recoil(), Regime: sample.Regime().String(),
			Components: components,
		}
	}
	return history
}

func snapshotFlow(flow restrictionflow.Result) *RestrictionSnapshot {
	return &RestrictionSnapshot{
		Regime:                flow.Regime().String(),
		MassFlowKilogramsPerS: flow.MassFlow().KilogramsPerSecond(),
		CriticalPressureRatio: flow.CriticalPressureRatio(),
		BackPressureRatio:     flow.BackPressureRatio(),
		ThrustNewtons:         flow.Thrust().Newtons(),
	}
}

func endpointFromState(state idealmixturereservoir.State) Endpoint {
	return Endpoint{
		MassKilograms:        state.TotalMass().Kilograms(),
		PressurePascals:      state.Pressure().Pascals(),
		TemperatureKelvin:    state.Temperature().Kelvin(),
		InternalEnergyJoules: state.InternalEnergy().Joules(),
	}
}

func classify(err error) string {
	switch {
	case errors.Is(err, restrictionflow.ErrAdversePressure):
		return "adverse_pressure"
	case errors.Is(err, idealmixturereservoir.ErrReservoirExhausted):
		return "reservoir_depletion"
	case errors.Is(err, coupledblowdown.ErrInvalidStepPolicy):
		return "invalid_step_policy"
	default:
		return "numerical_domain_error"
	}
}

func failure(code, stage, path, reason string) Report {
	return Report{
		Schema:                 ReportSchema,
		Status:                 "invalid",
		ImplementationRevision: ImplementationRevision,
		ValidationEnvironment: ValidationEnvironment{
			ConsultedInputs: []string{"document_bytes"}, AmbientInputs: []string{},
		},
		Diagnostics: []Diagnostic{{Code: code, Stage: stage, Path: path, ReasonCode: reason}},
	}
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
