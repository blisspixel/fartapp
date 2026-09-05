// Package restrictionhistory integrates quasi-steady restriction samples over a
// prescribed area history. Stagnation state is frozen. The package does not
// couple a reservoir, resolve an ODE blowdown, or invent species-resolved
// composition.
package restrictionhistory

import (
	"errors"
	"math"
	"sort"

	"github.com/blisspixel/fartapp/internal/floatmath"
	"github.com/blisspixel/fartapp/internal/restrictionflow"
)

const MaxSamples = 256

var (
	ErrInvalidSampleCount = errors.New("sample count must be within the model limit")
	ErrInvalidTime        = errors.New("sample times must be finite, nonnegative, and strictly increasing")
	ErrInvalidSample      = errors.New("history sample is outside the model domain")
	ErrNonFiniteIntegral  = errors.New("history integral is outside finite representable bounds")
)

type Seconds struct{ value float64 }

func NewSeconds(value float64) (Seconds, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
		return Seconds{}, ErrInvalidTime
	}
	return Seconds{value: value}, nil
}

func (value Seconds) Value() float64 { return value.value }

type Sample struct {
	time Seconds
	area restrictionflow.Area
}

func NewSample(time Seconds, area restrictionflow.Area) (Sample, error) {
	if time.value < 0 || math.IsNaN(time.value) || math.IsInf(time.value, 0) {
		return Sample{}, ErrInvalidTime
	}
	if area.SquareMetres() < 0 || math.IsNaN(area.SquareMetres()) || math.IsInf(area.SquareMetres(), 0) {
		return Sample{}, ErrInvalidSample
	}
	return Sample{time: time, area: area}, nil
}

func (sample Sample) Time() Seconds { return sample.time }

func (sample Sample) Area() restrictionflow.Area { return sample.area }

type Instant struct {
	time   Seconds
	result restrictionflow.Result
}

func (instant Instant) Time() Seconds { return instant.time }

func (instant Instant) Result() restrictionflow.Result { return instant.result }

type History struct {
	stagnation       restrictionflow.Stagnation
	back             restrictionflow.Pressure
	cd               restrictionflow.DischargeCoefficient
	samples          []Instant
	massOut          float64
	enthalpyOut      float64
	kineticEnergyOut float64
	totalEnthalpyOut float64
	impulse          float64
	recoilImpulse    float64
	recoilResidual   float64
}

func (history History) Stagnation() restrictionflow.Stagnation { return history.stagnation }

func (history History) BackPressure() restrictionflow.Pressure { return history.back }

func (history History) DischargeCoefficient() restrictionflow.DischargeCoefficient {
	return history.cd
}

func (history History) Samples() []Instant {
	return append([]Instant(nil), history.samples...)
}

func (history History) MassOutKilograms() float64 { return history.massOut }

func (history History) EnthalpyOutJoules() float64 { return history.enthalpyOut }

func (history History) KineticEnergyOutJoules() float64 { return history.kineticEnergyOut }

func (history History) TotalEnthalpyOutJoules() float64 { return history.totalEnthalpyOut }

func (history History) ImpulseNewtonSeconds() float64 { return history.impulse }

func (history History) RecoilImpulseNewtonSeconds() float64 { return history.recoilImpulse }

func (history History) RecoilResidualNewtonSeconds() float64 { return history.recoilResidual }

func Integrate(
	stagnation restrictionflow.Stagnation,
	back restrictionflow.Pressure,
	cd restrictionflow.DischargeCoefficient,
	samples []Sample,
) (History, error) {
	if len(samples) == 0 || len(samples) > MaxSamples {
		return History{}, ErrInvalidSampleCount
	}
	instants := make([]Instant, len(samples))
	for index, sample := range samples {
		if index > 0 && sample.time.value <= samples[index-1].time.value {
			return History{}, ErrInvalidTime
		}
		law, err := restrictionflow.NewPrescribedArea(sample.area)
		if err != nil {
			return History{}, err
		}
		request, err := restrictionflow.NewRequest(stagnation, back, law, cd)
		if err != nil {
			return History{}, err
		}
		result, err := restrictionflow.Evaluate(request)
		if err != nil {
			return History{}, err
		}
		instants[index] = Instant{time: sample.time, result: result}
	}

	massOut := 0.0
	enthalpyOut := 0.0
	kineticEnergyOut := 0.0
	totalEnthalpyOut := 0.0
	impulse := 0.0
	for index := 1; index < len(instants); index++ {
		dt := instants[index].time.value - instants[index-1].time.value
		left := instants[index-1].result
		right := instants[index].result
		leftRate := left.MassFlow().KilogramsPerSecond()
		rightRate := right.MassFlow().KilogramsPerSecond()
		if leftRate == 0 && rightRate == 0 {
			// An authored closed or equal-pressure interval has exactly no
			// transport. Unneeded heat-capacity products may overflow even
			// though every requested integral is identically zero.
			continue
		}
		mass := trapezoid(leftRate, rightRate, dt, 1)
		if mass == 0 && (leftRate > 0 || rightRate > 0) {
			return History{}, ErrNonFiniteIntegral
		}
		massOut += mass
		// Frozen stagnation and back pressure make the open-section
		// kinematics constant. A closed sample contributes no mass.
		open := left
		if leftRate == 0 {
			open = right
		}
		gasR := stagnation.SpecificGasConstant().JoulesPerKilogramKelvin()
		gamma := stagnation.HeatCapacityRatio().Value()
		cp := gasR + gasR/(gamma-1)
		enthalpy := floatmath.Product(mass, cp, open.ExitTemperature().Kelvin())
		kineticEnergy := floatmath.Product(mass, 0.5, open.ExitSpeed().MetresPerSecond(), open.ExitSpeed().MetresPerSecond())
		totalEnthalpy := floatmath.Product(mass, cp, stagnation.Temperature().Kelvin())
		if math.IsInf(cp, 0) || math.IsNaN(cp) {
			// Specific heat itself need not be representable when the
			// requested transported energy is. Keep its defining factors
			// separate until the complete mass*cp*T product is evaluated.
			enthalpy = floatmath.ProductOver(gamma-1, mass, gasR, gamma, open.ExitTemperature().Kelvin())
			totalEnthalpy = floatmath.ProductOver(gamma-1, mass, gasR, gamma, stagnation.Temperature().Kelvin())
		}
		momentumImpulse := floatmath.Product(mass, open.ExitSpeed().MetresPerSecond())
		pressureImpulse := trapezoid(left.EffectiveArea().SquareMetres(), right.EffectiveArea().SquareMetres(),
			dt, open.ExitPressure().Pascals()-back.Pascals())
		intervalImpulse := momentumImpulse + pressureImpulse
		for _, value := range []float64{enthalpy, kineticEnergy, totalEnthalpy, intervalImpulse} {
			if math.IsNaN(value) || math.IsInf(value, 0) || value <= 0 {
				return History{}, ErrNonFiniteIntegral
			}
		}
		enthalpyOut += enthalpy
		kineticEnergyOut += kineticEnergy
		totalEnthalpyOut += totalEnthalpy
		impulse += intervalImpulse
		for _, value := range []float64{massOut, enthalpyOut, kineticEnergyOut, totalEnthalpyOut, impulse} {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return History{}, ErrNonFiniteIntegral
			}
		}
	}
	recoilImpulse := -impulse
	return History{
		stagnation:       stagnation,
		back:             back,
		cd:               cd,
		samples:          instants,
		massOut:          massOut,
		enthalpyOut:      enthalpyOut,
		kineticEnergyOut: kineticEnergyOut,
		totalEnthalpyOut: totalEnthalpyOut,
		impulse:          impulse,
		recoilImpulse:    recoilImpulse,
		recoilResidual:   stableSignedSum([]float64{recoilImpulse, impulse}),
	}, nil
}

func trapezoid(left, right, dt, factor float64) float64 {
	scale := math.Max(left, right)
	if scale == 0 {
		return 0
	}
	return floatmath.Product(scale, (left/scale+right/scale)/2, dt, factor)
}

func stableSignedSum(values []float64) float64 {
	ordered := append([]float64(nil), values...)
	sort.Slice(ordered, func(left, right int) bool {
		return math.Abs(ordered[left]) < math.Abs(ordered[right])
	})
	var sum float64
	var compensation float64
	for _, value := range ordered {
		next := sum + value
		if math.Abs(sum) >= math.Abs(value) {
			compensation += (sum - next) + value
		} else {
			compensation += (value - next) + sum
		}
		sum = next
	}
	return sum + compensation
}
