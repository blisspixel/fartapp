package coupledblowdown

import (
	"errors"
	"math"

	"github.com/blisspixel/fartapp/internal/floatmath"
	"github.com/blisspixel/fartapp/internal/idealmixturereservoir"
	"github.com/blisspixel/fartapp/internal/restrictionflow"
)

// MaxAccuracyEvaluations bounds work independently of retained history size.
const MaxAccuracyEvaluations = 1000000

var (
	ErrInvalidAccuracyOptions    = errors.New("accuracy options are outside the supported range")
	ErrUnsupportedAccuracyConfig = errors.New("accurate blowdown does not support this configuration")
	ErrAccuracyBudgetExhausted   = errors.New("accurate blowdown exhausted its evaluation budget")
	ErrAccuracyNotAchieved       = errors.New("accurate blowdown cannot establish the requested numerical accuracy")
)

// AccuracyOptions has no implicit defaults. RelativeTolerance must be in
// [1e-12, 0.1], AbsoluteTimeToleranceSeconds must be finite and nonnegative,
// and MaxEvaluations must be in [15, MaxAccuracyEvaluations]. The time request
// is absolute tolerance + relative tolerance * elapsed time. Impulse and
// stroke use the relative tolerance alone.
type AccuracyOptions struct {
	RelativeTolerance            float64
	AbsoluteTimeToleranceSeconds float64
	MaxEvaluations               int
}

// AccuracyEvidence describes numerical quadrature on the ideal model path.
// Estimates are not rigorous bounds and exclude input uncertainty, physical
// model error, and floating-point representation of the path and endpoint.
// Error estimates are meaningful only on a nil error return. Work counters
// remain available on failure. ToleranceSatisfied does not imply completed
// discharge: the retained-history step budget can truncate an accurate run.
type AccuracyEvidence struct {
	EstimatedTimeErrorSeconds          float64
	EstimatedImpulseErrorNewtonSeconds float64
	EstimatedStrokeErrorMetres         float64
	RequestedTimeToleranceSeconds      float64
	Evaluations                        int
	AcceptedIntervals                  int
	Refinements                        int
	ToleranceSatisfied                 bool
	DischargeComplete                  bool
}

// SimulateAccurate is an opt-in regularized mass-coordinate integrator. It
// leaves Simulate's left-rate method and existing Config semantics unchanged.
// Flowing cases require positive prescribed area and MaxTimeSeconds()==0.
// Both reservoir closures and capped linear compliance are supported. A
// zero-rest-area compliant opening has infinite equalization time and is
// refused. No-flow identities remain supported for every valid Config.
//
// MaxWithdrawalFractionPerStep selects retained mass samples; it does not
// control quadrature accuracy. Every completed mass step has a sample and
// MaxSteps still truncates the history. Completed time reaches the analytical
// equalization endpoint; its stored state is within the reported pressure
// tolerance and is never clamped below back pressure.
func SimulateAccurate(config Config, options AccuracyOptions) (Result, AccuracyEvidence, error) {
	evidence := AccuracyEvidence{}
	if err := validateAccuracyOptions(options); err != nil {
		return Result{}, evidence, err
	}
	if _, err := NewConfig(config.reservoir, config.closure, config.back, config.area,
		config.cd, config.maxFraction, config.maxSteps, config.maxTime); err != nil {
		return Result{}, evidence, err
	}
	flow, err := evaluateState(config.reservoir, config)
	if err != nil {
		return Result{}, evidence, err
	}
	if flow.MassFlow().KilogramsPerSecond() == 0 {
		result, err := Simulate(config)
		evidence.ToleranceSatisfied = err == nil
		evidence.RequestedTimeToleranceSeconds = options.AbsoluteTimeToleranceSeconds
		return result, evidence, err
	}
	if config.maxTime != 0 || config.area.Prescribed().SquareMetres() == 0 {
		return Result{}, evidence, ErrUnsupportedAccuracyConfig
	}
	path, err := newAccuratePath(config)
	if err != nil {
		return Result{}, evidence, err
	}
	quadrature := accurateQuadrature{path: path, options: options, evidence: &evidence}
	result, err := accurateHistory(config, path, &quadrature, flow)
	if err != nil {
		return Result{}, evidence, err
	}
	evidence.RequestedTimeToleranceSeconds = options.AbsoluteTimeToleranceSeconds + options.RelativeTolerance*result.elapsed
	evidence.ToleranceSatisfied = finite(evidence.RequestedTimeToleranceSeconds) &&
		evidence.EstimatedTimeErrorSeconds <= evidence.RequestedTimeToleranceSeconds &&
		evidence.EstimatedImpulseErrorNewtonSeconds <= options.RelativeTolerance*result.impulse &&
		evidence.EstimatedStrokeErrorMetres <= options.RelativeTolerance*result.signature.strokeLength
	if !evidence.ToleranceSatisfied {
		return Result{}, evidence, ErrAccuracyNotAchieved
	}
	evidence.DischargeComplete = result.stop == StopEqualized
	return result, evidence, nil
}

func validateAccuracyOptions(options AccuracyOptions) error {
	if !finite(options.RelativeTolerance) || options.RelativeTolerance < 1e-12 || options.RelativeTolerance > 0.1 ||
		!finite(options.AbsoluteTimeToleranceSeconds) || options.AbsoluteTimeToleranceSeconds < 0 ||
		options.MaxEvaluations < 15 || options.MaxEvaluations > MaxAccuracyEvaluations {
		return ErrInvalidAccuracyOptions
	}
	return nil
}

func accurateHistory(config Config, path accuratePath, quadrature *accurateQuadrature, initialFlow restrictionflow.Result) (Result, error) {
	result := Result{config: config, final: config.reservoir, stop: StopMaxSteps}
	components := config.reservoir.Components()
	componentOut := make([]accumulator, len(components))
	first, err := retainedSample(0, result.final, initialFlow, componentOut)
	if err != nil {
		return Result{}, err
	}
	result.samples = append(make([]Sample, 0, min(config.maxSteps+1, 64)), first)
	choked := initialFlow.Regime() == restrictionflow.RegimeChoked
	x, z := 1.0, 1.0
	var totals, estimated [3]accumulator
	for result.steps < config.maxSteps {
		nextX := math.Max(path.equalMassRatio, x*(1-config.maxFraction))
		if nextX >= x {
			return Result{}, ErrAccuracyNotAchieved
		}
		nextZ := math.Sqrt((nextX - path.equalMassRatio) / path.excessMassRatio)
		if nextZ >= z || !finite(nextZ) {
			return Result{}, ErrAccuracyNotAchieved
		}
		integral, estimate, err := quadrature.integrate(nextZ, z)
		if err != nil {
			return Result{}, err
		}
		for index := range totals {
			totals[index].add(integral[index])
			estimated[index].add(estimate[index])
			if !finite(totals[index].value()) || !finite(estimated[index].value()) {
				return Result{}, ErrAccuracyNotAchieved
			}
		}
		nextState, err := path.state(nextX, nextZ == 0)
		if err != nil {
			return Result{}, err
		}
		if nextState.TotalMass().Kilograms() >= result.final.TotalMass().Kilograms() || totals[0].value() <= result.elapsed {
			return Result{}, ErrAccuracyNotAchieved
		}
		for index, component := range nextState.Components() {
			out := components[index].Mass().Kilograms() - component.Mass().Kilograms()
			if out <= componentOut[index].value() {
				return Result{}, ErrAccuracyNotAchieved
			}
			componentOut[index] = accumulator{sum: out}
		}
		flow, err := evaluateState(nextState, config)
		if err != nil {
			return Result{}, err
		}
		last, err := retainedSample(totals[0].value(), nextState, flow, componentOut)
		if err != nil {
			return Result{}, err
		}
		result.samples = append(result.samples, last)
		result.final, result.elapsed = nextState, totals[0].value()
		result.steps++
		x, z = nextX, nextZ
		if z == 0 {
			result.stop = StopEqualized
			break
		}
	}
	result.impulse, result.recoilImpulse = totals[1].value(), -totals[1].value()
	result.signature = signature(config.area.Prescribed().SquareMetres(), totals[2].value(), choked)
	if result.impulse <= 0 || result.signature.strokeLength <= 0 {
		return Result{}, ErrAccuracyNotAchieved
	}
	if err := accurateLedgers(&result, x, componentOut); err != nil {
		return Result{}, err
	}
	quadrature.evidence.EstimatedTimeErrorSeconds = estimated[0].value()
	quadrature.evidence.EstimatedImpulseErrorNewtonSeconds = estimated[1].value()
	quadrature.evidence.EstimatedStrokeErrorMetres = estimated[2].value()
	return result, nil
}

// Transfers are cumulative differences along the initial constant-composition
// path. They are independent of the retained sample count and quadrature work.
func accurateLedgers(result *Result, massRatio float64, componentOut []accumulator) error {
	initial := result.config.reservoir
	components := initial.Components()
	result.componentMassOut = make([]float64, len(components))
	result.ledgers.componentMassResiduals = make([]float64, len(components))
	var mass, enthalpy, heat accumulator
	for index, component := range result.final.Components() {
		out := componentOut[index].value()
		result.componentMassOut[index] = out
		result.ledgers.componentMassResiduals[index] = signedSum(component.Mass().Kilograms(), out, -components[index].Mass().Kilograms())
		mass.add(out)
		if result.config.closure == idealmixturereservoir.RigidIsothermal {
			r := components[index].SpecificGasConstant().JoulesPerKilogramKelvin()
			cv := components[index].IsochoricHeatCapacity().JoulesPerKilogramKelvin()
			enthalpy.add(floatmath.Product(out, r+cv, initial.Temperature().Kelvin()))
			heat.add(floatmath.Product(out, r, initial.Temperature().Kelvin()))
		}
	}
	result.massOut, result.enthalpyOut, result.heatIn = mass.value(), enthalpy.value(), heat.value()
	if result.config.closure == idealmixturereservoir.RigidAdiabatic {
		result.enthalpyOut = floatmath.Product(initial.InternalEnergy().Joules(), -math.Expm1(initial.HeatCapacityRatio()*math.Log(massRatio)))
	}
	if result.enthalpyOut <= 0 || (result.config.closure == idealmixturereservoir.RigidIsothermal && result.heatIn <= 0) {
		return ErrAccuracyNotAchieved
	}
	result.ledgers.massResidual = signedSum(result.final.TotalMass().Kilograms(), result.massOut, -initial.TotalMass().Kilograms())
	result.ledgers.energyResidual = signedSum(result.final.InternalEnergy().Joules(), result.enthalpyOut, -result.heatIn, -initial.InternalEnergy().Joules())
	result.ledgers.impulseResidual = signedSum(result.impulse, result.recoilImpulse)
	values := []float64{result.massOut, result.enthalpyOut, result.heatIn, result.signature.equivalentDiameter,
		result.signature.formationNumber, result.ledgers.massResidual, result.ledgers.energyResidual, result.ledgers.impulseResidual}
	for _, value := range append(values, result.ledgers.componentMassResiduals...) {
		if !finite(value) {
			return ErrAccuracyNotAchieved
		}
	}
	return nil
}
