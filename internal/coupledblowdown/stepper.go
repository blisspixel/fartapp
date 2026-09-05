package coupledblowdown

import (
	"errors"
	"math"

	"github.com/blisspixel/fartapp/internal/floatmath"
	"github.com/blisspixel/fartapp/internal/idealmixturereservoir"
	"github.com/blisspixel/fartapp/internal/restrictionflow"
)

// Simulate uses left-endpoint restriction rates and exact finite reservoir
// withdrawals. Every completed withdrawal has a retained endpoint sample.
// Conservation of these discrete transfers does not establish time accuracy.
func Simulate(config Config) (Result, error) {
	if _, err := NewConfig(config.reservoir, config.closure, config.back, config.area,
		config.cd, config.maxFraction, config.maxSteps, config.maxTime); err != nil {
		return Result{}, err
	}
	state := config.reservoir
	flow, err := evaluateState(state, config)
	if err != nil {
		return Result{}, err
	}
	components := state.Components()
	componentOut := make([]accumulator, len(components))
	samples := make([]Sample, 0, min(config.maxSteps+1, 64))
	initial, err := retainedSample(0, state, flow, componentOut)
	if err != nil {
		return Result{}, err
	}
	samples = append(samples, initial)
	var massOut, enthalpyOut, heatIn, impulse, recoil, stroke accumulator
	elapsed := 0.0
	steps := 0
	stop := StopMaxSteps
	choked := flow.Regime() == restrictionflow.RegimeChoked
	asymptotic := config.area.Prescribed().SquareMetres() == 0 &&
		config.area.Compliance().SquareMetresPerPascal() > 0

	for steps < config.maxSteps {
		if flow.MassFlow().KilogramsPerSecond() == 0 {
			stop = StopNoFlow
			break
		}
		equalization := equalizationFraction(state.Pressure().Pascals(), config.back.Pascals(),
			state.HeatCapacityRatio(), config.closure)
		fraction := math.Min(config.maxFraction, equalization)
		if asymptotic {
			// A zero-rest-area compliant opening closes as pressure falls. Its
			// exact equalization time is infinite; retain a positive gap.
			fraction = math.Min(fraction, equalization/2)
		}
		timeLimited := false
		if config.maxTime > 0 {
			remaining := config.maxTime - elapsed
			if remaining <= 0 {
				stop = StopMaxTime
				break
			}
			limit := flow.MassFlow().KilogramsPerSecond() * remaining / state.TotalMass().Kilograms()
			if limit <= fraction {
				fraction = limit
				timeLimited = true
			}
		}
		transition, err := boundedWithdrawal(state, fraction, config)
		if err != nil {
			if errors.Is(err, idealmixturereservoir.ErrNoRepresentableProgress) {
				stop = StopNoProgress
				if steps > 0 && state.Pressure().Pascals()-config.back.Pascals() <= pressureTolerance(config.back.Pascals()) {
					stop = StopEqualized
					if asymptotic {
						stop = StopPressureTolerance
					}
				}
				break
			}
			return Result{}, err
		}
		dt := transition.TotalMassOut().Kilograms() / flow.MassFlow().KilogramsPerSecond()
		nextElapsed := elapsed + dt
		if config.maxTime > 0 && nextElapsed >= config.maxTime {
			timeLimited = true
		}
		if timeLimited {
			// The represented mass withdrawal differs from q*dt by roundoff.
			// Preserve the authored time boundary and the actual mass ledger.
			dt = config.maxTime - elapsed
			nextElapsed = config.maxTime
		}
		if !finite(dt) || !finite(nextElapsed) || dt <= 0 || nextElapsed <= elapsed {
			stop = StopNoProgress
			break
		}
		massOut.add(transition.TotalMassOut().Kilograms())
		enthalpyOut.add(transition.IntegratedEnthalpyOut().Joules())
		heatIn.add(transition.HeatIntoReservoir().Joules())
		momentumImpulse := floatmath.Product(transition.TotalMassOut().Kilograms(), flow.ExitSpeed().MetresPerSecond())
		pressureImpulse := floatmath.Product(flow.ExitPressure().Pascals()-config.back.Pascals(),
			flow.EffectiveArea().SquareMetres(), dt)
		stepImpulse := momentumImpulse + pressureImpulse
		impulse.add(stepImpulse)
		recoil.add(-stepImpulse)
		stroke.add(floatmath.Product(flow.ExitSpeed().MetresPerSecond(), dt))
		for index, out := range transition.ComponentMassOut() {
			componentOut[index].add(out.Kilograms())
		}
		for _, total := range []accumulator{massOut, enthalpyOut, heatIn, impulse, recoil, stroke} {
			if !finite(total.value()) {
				return Result{}, ErrNoRepresentableStep
			}
		}
		state = transition.After()
		elapsed = nextElapsed
		steps++
		flow, err = evaluateState(state, config)
		if err != nil {
			return Result{}, err
		}
		last, err := retainedSample(elapsed, state, flow, componentOut)
		if err != nil {
			return Result{}, err
		}
		samples = append(samples, last)
		choked = choked || flow.Regime() == restrictionflow.RegimeChoked
		gap := state.Pressure().Pascals() - config.back.Pascals()
		if asymptotic && gap <= pressureTolerance(config.back.Pascals()) {
			stop = StopPressureTolerance
			break
		}
		if !asymptotic && gap <= pressureTolerance(config.back.Pascals()) {
			stop = StopEqualized
			break
		}
		if timeLimited {
			stop = StopMaxTime
			break
		}
	}

	componentMassOut := make([]float64, len(components))
	componentResiduals := make([]float64, len(components))
	for index, final := range state.Components() {
		componentMassOut[index] = componentOut[index].value()
		componentResiduals[index] = signedSum(final.Mass().Kilograms(),
			componentMassOut[index], -components[index].Mass().Kilograms())
	}
	sig := signature(config.area.Prescribed().SquareMetres(), stroke.value(), choked)
	if !finite(sig.equivalentDiameter) || !finite(sig.formationNumber) {
		return Result{}, ErrNoRepresentableStep
	}
	ledgers := Ledgers{
		massResidual:           signedSum(state.TotalMass().Kilograms(), massOut.value(), -config.reservoir.TotalMass().Kilograms()),
		energyResidual:         signedSum(state.InternalEnergy().Joules(), enthalpyOut.value(), -heatIn.value(), -config.reservoir.InternalEnergy().Joules()),
		impulseResidual:        signedSum(recoil.value(), impulse.value()),
		componentMassResiduals: componentResiduals,
	}
	for _, residual := range append(componentResiduals, ledgers.massResidual, ledgers.energyResidual, ledgers.impulseResidual) {
		if !finite(residual) {
			return Result{}, ErrNoRepresentableStep
		}
	}
	return Result{
		config: config, final: state, samples: samples,
		massOut: massOut.value(), enthalpyOut: enthalpyOut.value(), heatIn: heatIn.value(),
		impulse: impulse.value(), recoilImpulse: recoil.value(), elapsed: elapsed,
		steps: steps, stop: stop, signature: sig, componentMassOut: componentMassOut,
		ledgers: ledgers,
	}, nil
}

// boundedWithdrawal protects the one-way boundary against floating-point
// overshoot without changing the reported pressure or reservoir closure.
func boundedWithdrawal(state idealmixturereservoir.State, fraction float64, config Config) (idealmixturereservoir.Transition, error) {
	if fraction <= 0 {
		return idealmixturereservoir.Transition{}, idealmixturereservoir.ErrNoRepresentableProgress
	}
	for attempt := 0; attempt < 16; attempt++ {
		withdrawal, err := idealmixturereservoir.NewWithdrawalFraction(fraction)
		if err != nil {
			return idealmixturereservoir.Transition{}, err
		}
		transition, err := idealmixturereservoir.WithdrawFraction(state, withdrawal, config.closure)
		if err != nil {
			return idealmixturereservoir.Transition{}, err
		}
		if transition.After().Pressure().Pascals() >= config.back.Pascals() {
			return transition, nil
		}
		// The endpoint pressure rounds at the scale of the retained mass,
		// not at the much smaller ulp of a nearly exhausted withdrawal.
		fraction -= math.Min(fraction/2, 4*(math.Nextafter(1, 2)-1))
	}
	return idealmixturereservoir.Transition{}, idealmixturereservoir.ErrNoRepresentableProgress
}

func retainedSample(time float64, state idealmixturereservoir.State, flow restrictionflow.Result, componentOut []accumulator) (Sample, error) {
	sample := snapshot(time, state, flow)
	sample.componentMassOut = make([]float64, len(componentOut))
	for index, total := range componentOut {
		sample.componentMassOut[index] = total.value()
	}
	if !finite(sample.enthalpyFlow) {
		return Sample{}, ErrNoRepresentableStep
	}
	return sample, nil
}

func pressureTolerance(back float64) float64 {
	return 16 * (math.Nextafter(1, 2) - 1) * back
}
