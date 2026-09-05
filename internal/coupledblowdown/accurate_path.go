package coupledblowdown

import (
	"math"
	"sort"

	"github.com/blisspixel/fartapp/internal/floatmath"
	"github.com/blisspixel/fartapp/internal/idealmixturereservoir"
)

type accuratePath struct {
	config                           Config
	mass, temperature, gas, gamma, n float64
	equalMassRatio, excessMassRatio  float64
	logEqualMassRatio, criticalLog   float64
	back, restArea, compliance, cap  float64
	discharge, sonicFactor           float64
	breaks                           []float64
}

func newAccuratePath(config Config) (accuratePath, error) {
	state := config.reservoir
	p := accuratePath{
		config: config, mass: state.TotalMass().Kilograms(), temperature: state.Temperature().Kelvin(),
		gas: state.MixtureGasConstant().JoulesPerKilogramKelvin(), gamma: state.HeatCapacityRatio(), n: 1,
		back: config.back.Pascals(), restArea: config.area.Prescribed().SquareMetres(),
		compliance: config.area.Compliance().SquareMetresPerPascal(), cap: config.area.Maximum().SquareMetres(),
		discharge: config.cd.Value(),
	}
	if config.closure == idealmixturereservoir.RigidAdiabatic {
		p.n = p.gamma
	}
	pressure := state.Pressure().Pascals()
	logPressureRatio := math.Log(pressure) - math.Log(p.back)
	if p.back > pressure/2 {
		logPressureRatio = math.Log1p((pressure - p.back) / p.back)
	}
	p.logEqualMassRatio = -logPressureRatio / p.n
	p.equalMassRatio = math.Exp(p.logEqualMassRatio)
	p.excessMassRatio = -math.Expm1(p.logEqualMassRatio)
	p.criticalLog = p.gamma / (p.gamma - 1) * math.Log1p((p.gamma-1)/2)
	p.sonicFactor = math.Sqrt(p.gamma) * math.Exp(-(0.5+1/(p.gamma-1))*math.Log1p((p.gamma-1)/2))
	// Below this excess mass, endpoint representation dominates a useful
	// controlled clock calculation. Refuse instead of labelling it no-flow.
	if p.equalMassRatio <= 0 || !finite(p.excessMassRatio/p.equalMassRatio) ||
		p.excessMassRatio < 128*(math.Nextafter(1, 2)-1) || p.sonicFactor <= 0 || !finite(p.sonicFactor) {
		return accuratePath{}, ErrAccuracyNotAchieved
	}
	p.addBreak(p.criticalLog)
	if p.compliance > 0 {
		// Split the actual cap kink, and also the smooth crossover where
		// compliance and rest area are equal. The latter resolves a narrow
		// near-closure time contribution when the rest area is very small.
		for _, delta := range []float64{(p.cap - p.restArea) / p.compliance, p.restArea / p.compliance} {
			if delta > 0 && delta < pressure-p.back {
				logRatio := math.Log1p(delta / p.back)
				if !finite(logRatio) {
					logRatio = math.Log(p.back+delta) - math.Log(p.back)
				}
				p.addBreak(logRatio)
			}
		}
	}
	sort.Float64s(p.breaks)
	return p, nil
}

func (p *accuratePath) addBreak(logPressureRatio float64) {
	excess := math.Expm1(logPressureRatio / p.n)
	z := math.Sqrt(accurateRatio([]float64{p.equalMassRatio, excess}, []float64{p.excessMassRatio}))
	if z > 0 && z < 1 {
		p.breaks = append(p.breaks, z)
	}
}

// With x=m/m0, xe=(Pb/P0)^(1/n), and d=1-xe, use x=xe+d*z^2.
// Integrating z from zero to one traverses the path backwards. The positive
// Jacobian is dm/dz=2*m0*d*z. For positive rest area, q is proportional to z
// at equalization, so (dm/dz)/q has a finite limit.
//
// The restriction relation is the NASA isentropic mass-flow function:
// https://www.grc.nasa.gov/www/k-12/BGP/mflchk.html
// Its subsonic speed satisfies u^2=2*R*T*(1-exp(-a*L))/a, where
// a=(gamma-1)/gamma and L=log(P/Pb)=n*log1p(d*z^2/xe). Factoring out z
// before division avoids pressure subtraction and a numerical 0/0 tail.
func (p accuratePath) integrand(z float64) ([3]float64, error) {
	v := floatmath.Product(p.excessMassRatio/p.equalMassRatio, z, z)
	h := math.Log1p(v)
	logPressureRatio := p.n * h
	logMassRatio := math.Min(0, p.logEqualMassRatio+h)
	temperature := p.temperature * math.Exp((p.n-1)*logMassRatio)
	delta := p.back * math.Expm1(logPressureRatio)
	if logPressureRatio > 1 {
		// A finite pressure need not have a representable P/Pb ratio.
		// Subtraction is well conditioned away from the equalization tail.
		delta = p.config.reservoir.Pressure().Pascals()*math.Exp(p.n*logMassRatio) - p.back
	}
	area := math.Min(p.cap, p.restArea+p.compliance*delta)
	var timeJacobian, speed, pressureImpulse float64
	if logPressureRatio <= p.criticalLog {
		a := (p.gamma - 1) / p.gamma
		w := a * logPressureRatio
		logFactor, expFactor := 1.0, 1.0
		if v > 0 {
			logFactor = h / v
		}
		if w > 0 {
			expFactor = -math.Expm1(-w) / w
		}
		speedOverZ := accurateRatio([]float64{math.Sqrt(2), math.Sqrt(p.gas), math.Sqrt(temperature),
			math.Sqrt(p.n), math.Sqrt(p.excessMassRatio), math.Sqrt(logFactor), math.Sqrt(expFactor)},
			[]float64{math.Sqrt(p.equalMassRatio)})
		exitTemperature := temperature * math.Exp(-w)
		timeJacobian = accurateRatio([]float64{2, p.mass, p.excessMassRatio, p.gas, exitTemperature},
			[]float64{p.discharge, area, p.back, speedOverZ})
		speed = floatmath.Product(z, speedOverZ)
	} else {
		pressure := p.config.reservoir.Pressure().Pascals() * math.Exp(p.n*logMassRatio)
		timeJacobian = accurateRatio([]float64{2, p.mass, p.excessMassRatio, z, math.Sqrt(p.gas), math.Sqrt(temperature)},
			[]float64{p.discharge, area, pressure, p.sonicFactor})
		speed = floatmath.Product(math.Sqrt(2*(p.gamma/(p.gamma+1))), math.Sqrt(p.gas), math.Sqrt(temperature))
		exitPressure := pressure * math.Exp(-p.criticalLog)
		pressureImpulse = floatmath.Product(math.Max(0, exitPressure-p.back), area, timeJacobian)
	}
	value := [3]float64{
		timeJacobian,
		floatmath.Product(2, p.mass, p.excessMassRatio, z, speed) + pressureImpulse,
		floatmath.Product(speed, timeJacobian),
	}
	if !finite(temperature) || temperature <= 0 || !finite(area) || area <= 0 ||
		!finite(speed) || speed < 0 || !finite(value[0]) || value[0] <= 0 ||
		!finite(value[1]) || value[1] < 0 || !finite(value[2]) || value[2] < 0 {
		return [3]float64{}, ErrAccuracyNotAchieved
	}
	return value, nil
}

func (p accuratePath) state(x float64, endpoint bool) (idealmixturereservoir.State, error) {
	for attempt := 0; attempt < 16; attempt++ {
		components := p.config.reservoir.Components()
		for index, component := range components {
			mass, err := idealmixturereservoir.NewMass(floatmath.Product(component.Mass().Kilograms(), x))
			if err != nil {
				return idealmixturereservoir.State{}, ErrAccuracyNotAchieved
			}
			components[index], err = idealmixturereservoir.NewComponent(mass, component.SpecificGasConstant(), component.IsochoricHeatCapacity())
			if err != nil {
				return idealmixturereservoir.State{}, err
			}
		}
		temperature, err := idealmixturereservoir.NewTemperature(p.temperature * math.Exp((p.n-1)*math.Log(x)))
		if err != nil {
			return idealmixturereservoir.State{}, ErrAccuracyNotAchieved
		}
		state, err := idealmixturereservoir.NewState(components, p.config.reservoir.Volume(), temperature)
		if err != nil {
			return idealmixturereservoir.State{}, err
		}
		if state.Pressure().Pascals() >= p.back {
			if endpoint && state.Pressure().Pascals()-p.back > pressureTolerance(p.back) {
				return idealmixturereservoir.State{}, ErrAccuracyNotAchieved
			}
			return state, nil
		}
		if !endpoint {
			return idealmixturereservoir.State{}, ErrAccuracyNotAchieved
		}
		x = math.Nextafter(x, 1)
	}
	return idealmixturereservoir.State{}, ErrAccuracyNotAchieved
}

// Factoring both products avoids intermediate reciprocal overflow, including
// small-area/long-duration configurations whose final integral is finite.
func accurateRatio(numerators, denominators []float64) float64 {
	mantissa, exponent := 1.0, 0
	for _, value := range numerators {
		fraction, power := math.Frexp(value)
		mantissa *= fraction
		exponent += power
		mantissa, power = math.Frexp(mantissa)
		exponent += power
	}
	for _, value := range denominators {
		fraction, power := math.Frexp(value)
		mantissa /= fraction
		exponent -= power
		mantissa, power = math.Frexp(mantissa)
		exponent += power
	}
	return math.Ldexp(mantissa, exponent)
}
