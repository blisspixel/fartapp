// Package coupledblowdown couples a rigid ideal-mixture reservoir to a
// quasi-steady converging restriction. Restriction flow sets the rate;
// reservoir closure sets the thermodynamic path. The package does not model
// heat-transfer laws, stratification, a plume, or acoustics.
package coupledblowdown

import (
	"errors"
	"math"

	"github.com/blisspixel/fartapp/internal/idealmixturereservoir"
	"github.com/blisspixel/fartapp/internal/restrictionflow"
)

const (
	MaxSteps   = 4096
	MaxSamples = MaxSteps + 1
)

var (
	ErrInvalidConfig       = errors.New("blowdown configuration is outside the model domain")
	ErrInvalidStepPolicy   = errors.New("step policy is outside the model domain")
	ErrNoRepresentableStep = errors.New("restriction flow produces no representable reservoir progress")
)

type StopReason uint8

const (
	StopNoFlow StopReason = iota + 1
	StopEqualized
	StopMaxSteps
	StopMaxTime
	StopNoProgress
	StopPressureTolerance
)

func (reason StopReason) String() string {
	switch reason {
	case StopNoFlow:
		return "no-flow"
	case StopEqualized:
		return "equalized"
	case StopMaxSteps:
		return "max-steps"
	case StopMaxTime:
		return "max-time"
	case StopNoProgress:
		return "no-progress"
	case StopPressureTolerance:
		return "pressure-tolerance"
	default:
		return "unsupported"
	}
}

type Config struct {
	reservoir   idealmixturereservoir.State
	closure     idealmixturereservoir.Closure
	back        restrictionflow.Pressure
	area        restrictionflow.AreaLaw
	cd          restrictionflow.DischargeCoefficient
	maxFraction float64
	maxSteps    int
	maxTime     float64
}

func NewConfig(
	reservoir idealmixturereservoir.State,
	closure idealmixturereservoir.Closure,
	back restrictionflow.Pressure,
	area restrictionflow.AreaLaw,
	cd restrictionflow.DischargeCoefficient,
	maxFraction float64,
	maxSteps int,
	maxTime float64,
) (Config, error) {
	if err := validateBoundary(reservoir, back, closure); err != nil {
		return Config{}, err
	}
	if !finite(maxFraction) || maxFraction <= 0 || maxFraction >= 1 {
		return Config{}, ErrInvalidStepPolicy
	}
	if maxSteps <= 0 || maxSteps > MaxSteps {
		return Config{}, ErrInvalidStepPolicy
	}
	if !finite(maxTime) || maxTime < 0 {
		return Config{}, ErrInvalidStepPolicy
	}
	request, err := restrictionRequest(reservoir, back, area, cd)
	if err != nil {
		return Config{}, err
	}
	if _, err := restrictionflow.Evaluate(request); err != nil {
		return Config{}, err
	}
	return Config{
		reservoir:   reservoir,
		closure:     closure,
		back:        back,
		area:        area,
		cd:          cd,
		maxFraction: maxFraction,
		maxSteps:    maxSteps,
		maxTime:     maxTime,
	}, nil
}

func (config Config) Reservoir() idealmixturereservoir.State { return config.reservoir }

func (config Config) Closure() idealmixturereservoir.Closure { return config.closure }

func (config Config) BackPressure() restrictionflow.Pressure { return config.back }

func (config Config) AreaLaw() restrictionflow.AreaLaw { return config.area }

func (config Config) DischargeCoefficient() restrictionflow.DischargeCoefficient {
	return config.cd
}

func (config Config) MaxWithdrawalFractionPerStep() float64 { return config.maxFraction }

func (config Config) MaxSteps() int { return config.maxSteps }

func (config Config) MaxTimeSeconds() float64 { return config.maxTime }

type Sample struct {
	time             float64
	mass             float64
	pressure         float64
	temperature      float64
	massFlow         float64
	exitSpeed        float64
	thrust           float64
	recoil           float64
	regime           restrictionflow.Regime
	exitPressure     float64
	exitTemperature  float64
	effectiveArea    float64
	enthalpyFlow     float64
	componentMasses  []float64
	componentMassOut []float64
}

func (sample Sample) Time() float64 { return sample.time }

func (sample Sample) Mass() float64 { return sample.mass }

func (sample Sample) Pressure() float64 { return sample.pressure }

func (sample Sample) Temperature() float64 { return sample.temperature }

func (sample Sample) MassFlow() float64 { return sample.massFlow }

func (sample Sample) ExitSpeed() float64 { return sample.exitSpeed }

func (sample Sample) Thrust() float64 { return sample.thrust }

func (sample Sample) Recoil() float64 { return sample.recoil }

func (sample Sample) Regime() restrictionflow.Regime { return sample.regime }

func (sample Sample) ExitPressure() float64 { return sample.exitPressure }

func (sample Sample) ExitTemperature() float64 { return sample.exitTemperature }

func (sample Sample) EffectiveArea() float64 { return sample.effectiveArea }

// EnthalpyFlow is the source stagnation enthalpy rate, including the exit
// kinetic-energy contribution, in joules per second.
func (sample Sample) EnthalpyFlow() float64 { return sample.enthalpyFlow }

func (sample Sample) ComponentMassesKilograms() []float64 {
	return append([]float64(nil), sample.componentMasses...)
}

func (sample Sample) ComponentMassOutKilograms() []float64 {
	return append([]float64(nil), sample.componentMassOut...)
}

type Ledgers struct {
	massResidual           float64
	energyResidual         float64
	impulseResidual        float64
	componentMassResiduals []float64
}

func (ledgers Ledgers) MassResidualKilograms() float64 { return ledgers.massResidual }

func (ledgers Ledgers) EnergyResidualJoules() float64 { return ledgers.energyResidual }

func (ledgers Ledgers) ImpulseResidualNewtonSeconds() float64 { return ledgers.impulseResidual }

func (ledgers Ledgers) ComponentMassResidualsKilograms() []float64 {
	return append([]float64(nil), ledgers.componentMassResiduals...)
}

type Signature struct {
	equivalentDiameter float64
	strokeLength       float64
	formationNumber    float64
	formationDefined   bool
	chokedOccurred     bool
}

func (signature Signature) EquivalentDiameterMetres() float64 { return signature.equivalentDiameter }

func (signature Signature) StrokeLengthMetres() float64 { return signature.strokeLength }

func (signature Signature) FormationNumber() (float64, bool) {
	return signature.formationNumber, signature.formationDefined
}

func (signature Signature) ChokedOccurred() bool { return signature.chokedOccurred }

type Result struct {
	config           Config
	final            idealmixturereservoir.State
	samples          []Sample
	massOut          float64
	enthalpyOut      float64
	heatIn           float64
	impulse          float64
	recoilImpulse    float64
	elapsed          float64
	steps            int
	stop             StopReason
	ledgers          Ledgers
	signature        Signature
	componentMassOut []float64
}

func (result Result) Config() Config { return result.config }

func (result Result) Final() idealmixturereservoir.State { return result.final }

func (result Result) Samples() []Sample { return append([]Sample(nil), result.samples...) }

func (result Result) MassOutKilograms() float64 { return result.massOut }

func (result Result) ComponentMassOutKilograms() []float64 {
	return append([]float64(nil), result.componentMassOut...)
}

func (result Result) EnthalpyOutJoules() float64 { return result.enthalpyOut }

func (result Result) HeatInJoules() float64 { return result.heatIn }

func (result Result) ImpulseNewtonSeconds() float64 { return result.impulse }

func (result Result) RecoilImpulseNewtonSeconds() float64 { return result.recoilImpulse }

func (result Result) ElapsedSeconds() float64 { return result.elapsed }

func (result Result) Steps() int { return result.steps }

func (result Result) Stop() StopReason { return result.stop }

func (result Result) Ledgers() Ledgers { return result.ledgers }

func (result Result) Signature() Signature { return result.signature }

func (result Result) EqualizationPressureTolerancePascals() float64 {
	return pressureTolerance(result.config.back.Pascals())
}

func EqualizationFraction(state idealmixturereservoir.State, back restrictionflow.Pressure, closure idealmixturereservoir.Closure) (float64, error) {
	if err := validateBoundary(state, back, closure); err != nil {
		return 0, err
	}
	pressure := state.Pressure().Pascals()
	if pressure <= back.Pascals() {
		return 0, nil
	}
	fraction := equalizationFraction(pressure, back.Pascals(), state.HeatCapacityRatio(), closure)
	if fraction >= 1 {
		return 0, idealmixturereservoir.ErrReservoirExhausted
	}
	return fraction, nil
}

func validateBoundary(state idealmixturereservoir.State, back restrictionflow.Pressure, closure idealmixturereservoir.Closure) error {
	if closure != idealmixturereservoir.RigidAdiabatic && closure != idealmixturereservoir.RigidIsothermal {
		return ErrInvalidConfig
	}
	if _, err := idealmixturereservoir.NewState(state.Components(), state.Volume(), state.Temperature()); err != nil {
		return ErrInvalidConfig
	}
	if !finite(back.Pascals()) || back.Pascals() <= 0 {
		return ErrInvalidConfig
	}
	if state.Pressure().Pascals() < back.Pascals() {
		return restrictionflow.ErrAdversePressure
	}
	return nil
}

func equalizationFraction(pressure, back, gamma float64, closure idealmixturereservoir.Closure) float64 {
	if pressure <= back {
		return 0
	}
	logRatio := math.Log(back) - math.Log(pressure)
	if back > pressure/2 {
		logRatio = math.Log1p(-(pressure - back) / pressure)
	}
	if closure == idealmixturereservoir.RigidAdiabatic {
		logRatio /= gamma
	}
	return -math.Expm1(logRatio)
}

func evaluateState(
	state idealmixturereservoir.State,
	config Config,
) (restrictionflow.Result, error) {
	request, err := restrictionRequest(state, config.back, config.area, config.cd)
	if err != nil {
		return restrictionflow.Result{}, err
	}
	return restrictionflow.Evaluate(request)
}

func restrictionRequest(
	state idealmixturereservoir.State,
	back restrictionflow.Pressure,
	area restrictionflow.AreaLaw,
	cd restrictionflow.DischargeCoefficient,
) (restrictionflow.Request, error) {
	pressure, err := restrictionflow.NewPressure(state.Pressure().Pascals())
	if err != nil {
		return restrictionflow.Request{}, err
	}
	temperature, err := restrictionflow.NewTemperature(state.Temperature().Kelvin())
	if err != nil {
		return restrictionflow.Request{}, err
	}
	gas, err := restrictionflow.NewSpecificGasConstant(state.MixtureGasConstant().JoulesPerKilogramKelvin())
	if err != nil {
		return restrictionflow.Request{}, err
	}
	gamma, err := restrictionflow.NewHeatCapacityRatio(state.HeatCapacityRatio())
	if err != nil {
		return restrictionflow.Request{}, err
	}
	stagnation, err := restrictionflow.NewStagnation(pressure, temperature, gas, gamma)
	if err != nil {
		return restrictionflow.Request{}, err
	}
	return restrictionflow.NewRequest(stagnation, back, area, cd)
}

func snapshot(time float64, state idealmixturereservoir.State, flow restrictionflow.Result) Sample {
	sample := Sample{
		time:        time,
		mass:        state.TotalMass().Kilograms(),
		pressure:    state.Pressure().Pascals(),
		temperature: state.Temperature().Kelvin(),
	}
	for _, component := range state.Components() {
		sample.componentMasses = append(sample.componentMasses, component.Mass().Kilograms())
	}
	if flow.Regime() != 0 {
		sample.massFlow = flow.MassFlow().KilogramsPerSecond()
		sample.exitSpeed = flow.ExitSpeed().MetresPerSecond()
		sample.thrust = flow.Thrust().Newtons()
		sample.recoil = flow.Recoil().Newtons()
		sample.regime = flow.Regime()
		sample.exitPressure = flow.ExitPressure().Pascals()
		sample.exitTemperature = flow.ExitTemperature().Kelvin()
		sample.effectiveArea = flow.EffectiveArea().SquareMetres()
		sample.enthalpyFlow = flow.MassFlow().KilogramsPerSecond() *
			state.MixtureIsobaricHeatCapacity().JoulesPerKilogramKelvin() * state.Temperature().Kelvin()
	}
	return sample
}

func signature(prescribedArea, stroke float64, choked bool) Signature {
	result := Signature{strokeLength: stroke, chokedOccurred: choked}
	if prescribedArea > 0 {
		result.equivalentDiameter = 2 / math.Sqrt(math.Pi) * math.Sqrt(prescribedArea)
		result.formationNumber = stroke / result.equivalentDiameter
		result.formationDefined = true
	}
	return result
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
