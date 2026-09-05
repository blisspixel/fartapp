package restrictionflow

import "math"

type Regime uint8

const (
	RegimeNoFlow Regime = iota + 1
	RegimeSubsonic
	RegimeChoked
)

func (regime Regime) String() string {
	switch regime {
	case RegimeNoFlow:
		return "no-flow"
	case RegimeSubsonic:
		return "subsonic"
	case RegimeChoked:
		return "choked"
	default:
		return "unsupported"
	}
}

type AreaLaw struct {
	prescribed Area
	compliance AreaCompliance
	maximum    Area
}

func NewPrescribedArea(area Area) (AreaLaw, error) {
	if err := validateArea(area); err != nil {
		return AreaLaw{}, err
	}
	return AreaLaw{prescribed: area, maximum: area}, nil
}

func NewLinearComplianceArea(prescribed Area, compliance AreaCompliance, maximum Area) (AreaLaw, error) {
	if err := validateArea(prescribed); err != nil {
		return AreaLaw{}, err
	}
	if err := validateArea(maximum); err != nil {
		return AreaLaw{}, err
	}
	if err := validateCompliance(compliance); err != nil {
		return AreaLaw{}, err
	}
	if maximum.squareMetres < prescribed.squareMetres {
		return AreaLaw{}, ErrInvalidAreaLaw
	}
	return AreaLaw{prescribed: prescribed, compliance: compliance, maximum: maximum}, nil
}

func (law AreaLaw) Prescribed() Area { return law.prescribed }

func (law AreaLaw) Compliance() AreaCompliance { return law.compliance }

func (law AreaLaw) Maximum() Area { return law.maximum }

func (law AreaLaw) LawName() string {
	if law.compliance.squareMetresPerPascal == 0 {
		return "prescribed"
	}
	return "linear-compliance"
}

func (law AreaLaw) Effective(overpressure Pressure) (Area, error) {
	if err := validateAreaLaw(law); err != nil {
		return Area{}, err
	}
	if !finite(overpressure.pascals) {
		return Area{}, ErrNonFiniteValue
	}
	opening := 0.0
	if overpressure.pascals > 0 {
		opening = law.compliance.squareMetresPerPascal * overpressure.pascals
		if opening == 0 && law.compliance.squareMetresPerPascal > 0 &&
			law.prescribed.squareMetres == 0 && law.maximum.squareMetres > 0 {
			return Area{}, ErrNoRepresentableFlow
		}
	}
	effective := law.prescribed.squareMetres + opening
	if effective > law.maximum.squareMetres {
		effective = law.maximum.squareMetres
	}
	return NewArea(effective)
}

type Stagnation struct {
	pressure    Pressure
	temperature Temperature
	gasConstant SpecificGasConstant
	gamma       HeatCapacityRatio
}

func NewStagnation(
	pressure Pressure,
	temperature Temperature,
	gasConstant SpecificGasConstant,
	gamma HeatCapacityRatio,
) (Stagnation, error) {
	state := Stagnation{
		pressure:    pressure,
		temperature: temperature,
		gasConstant: gasConstant,
		gamma:       gamma,
	}
	if err := validateStagnation(state); err != nil {
		return Stagnation{}, err
	}
	return state, nil
}

func (state Stagnation) Pressure() Pressure { return state.pressure }

func (state Stagnation) Temperature() Temperature { return state.temperature }

func (state Stagnation) SpecificGasConstant() SpecificGasConstant { return state.gasConstant }

func (state Stagnation) HeatCapacityRatio() HeatCapacityRatio { return state.gamma }

func (state Stagnation) CriticalPressureRatio() float64 {
	gamma := state.gamma.value
	return math.Exp(-gamma / (gamma - 1) * math.Log1p((gamma-1)/2))
}

func (state Stagnation) CriticalPressure() Pressure {
	return Pressure{pascals: state.pressure.pascals * state.CriticalPressureRatio()}
}

type Request struct {
	stagnation Stagnation
	back       Pressure
	area       AreaLaw
	cd         DischargeCoefficient
}

func NewRequest(
	stagnation Stagnation,
	back Pressure,
	area AreaLaw,
	cd DischargeCoefficient,
) (Request, error) {
	request := Request{stagnation: stagnation, back: back, area: area, cd: cd}
	if err := validateRequest(request); err != nil {
		return Request{}, err
	}
	return request, nil
}

func (request Request) Stagnation() Stagnation { return request.stagnation }

func (request Request) BackPressure() Pressure { return request.back }

func (request Request) AreaLaw() AreaLaw { return request.area }

func (request Request) DischargeCoefficient() DischargeCoefficient { return request.cd }

type Result struct {
	request               Request
	regime                Regime
	effectiveArea         Area
	criticalPressureRatio float64
	backPressureRatio     float64
	throatMach            MachNumber
	exitPressure          Pressure
	exitTemperature       Temperature
	exitSpeed             Speed
	massFlow              MassFlow
	sonicMassFlow         MassFlow
	thrust                Force
	recoil                Force
	massFlowResidual      MassFlowResidual
	thrustResidual        ForceResidual
	recoilResidual        ForceResidual
}

func (result Result) Request() Request { return result.request }

func (result Result) Regime() Regime { return result.regime }

func (result Result) EffectiveArea() Area { return result.effectiveArea }

func (result Result) CriticalPressureRatio() float64 { return result.criticalPressureRatio }

func (result Result) BackPressureRatio() float64 { return result.backPressureRatio }

func (result Result) ThroatMach() MachNumber { return result.throatMach }

func (result Result) ExitPressure() Pressure { return result.exitPressure }

func (result Result) ExitTemperature() Temperature { return result.exitTemperature }

func (result Result) ExitSpeed() Speed { return result.exitSpeed }

func (result Result) MassFlow() MassFlow { return result.massFlow }

func (result Result) SonicMassFlow() MassFlow { return result.sonicMassFlow }

func (result Result) Thrust() Force { return result.thrust }

func (result Result) Recoil() Force { return result.recoil }

func (result Result) MassFlowResidual() MassFlowResidual { return result.massFlowResidual }

func (result Result) ThrustResidual() ForceResidual { return result.thrustResidual }

func (result Result) RecoilResidual() ForceResidual { return result.recoilResidual }

func Evaluate(request Request) (Result, error) {
	if err := validateRequest(request); err != nil {
		return Result{}, err
	}
	delta := request.stagnation.pressure.pascals - request.back.pascals
	effective, err := request.area.Effective(Pressure{pascals: delta})
	if err != nil {
		return Result{}, err
	}
	criticalRatio := request.stagnation.CriticalPressureRatio()
	backRatio := request.back.pascals / request.stagnation.pressure.pascals
	if !finite(criticalRatio) || criticalRatio <= 0 || criticalRatio >= 1 || !finite(backRatio) || backRatio < 0 {
		return Result{}, ErrNoRepresentableFlow
	}

	if effective.squareMetres == 0 {
		return zeroFlow(request, effective, criticalRatio, backRatio), nil
	}
	if request.back.pascals > request.stagnation.pressure.pascals {
		return Result{}, ErrAdversePressure
	}
	if request.back.pascals == request.stagnation.pressure.pascals {
		return zeroFlow(request, effective, criticalRatio, backRatio), nil
	}

	gamma := request.stagnation.gamma.value
	p0 := request.stagnation.pressure.pascals
	t0 := request.stagnation.temperature.kelvin
	gasR := request.stagnation.gasConstant.joulesPerKilogramKelvin
	area := effective.squareMetres
	cd := request.cd.value
	pb := request.back.pascals

	sonicMassFlow, err := sonicMassFlowRate(cd, area, p0, t0, gasR, gamma, criticalRatio)
	if err != nil {
		return Result{}, err
	}

	regime := RegimeSubsonic
	mach := 0.0
	exitPressure := pb
	exitTemperature := t0
	if backRatio <= criticalRatio {
		regime = RegimeChoked
		mach = 1
		exitPressure = p0 * criticalRatio
		exitTemperature = t0 * 2 / (gamma + 1)
	} else {
		logPressureRatio := math.Log1p((p0 - pb) / pb)
		machSquared := 2 / (gamma - 1) * math.Expm1((gamma-1)/gamma*logPressureRatio)
		mach = math.Sqrt(machSquared)
		if mach >= 1 {
			if mach > 1+8*(math.Nextafter(1, 2)-1) {
				return Result{}, ErrNoRepresentableFlow
			}
			// A subsonic pressure ratio can round to unit Mach at the
			// continuous sonic boundary. Retain the subsonic regime with
			// the closest representable Mach below one, within eight ulps.
			mach = math.Nextafter(1, 0)
		}
		exitTemperature = t0 / (1 + (gamma-1)/2*mach*mach)
	}
	if err := positiveFinite(exitPressure); err != nil {
		return Result{}, ErrNoRepresentableFlow
	}
	if err := positiveFinite(exitTemperature); err != nil {
		return Result{}, ErrNoRepresentableFlow
	}

	exitSpeed := mach * math.Sqrt(gamma*gasR*exitTemperature)
	density := exitPressure / (gasR * exitTemperature)
	kinematicMassFlow := density * area * exitSpeed
	massFlow := cd * kinematicMassFlow
	pressureThrust := (exitPressure - pb) * area
	thrust := massFlow*exitSpeed + pressureThrust
	recoil := -thrust
	values := []float64{exitSpeed, density, kinematicMassFlow, massFlow, pressureThrust, thrust, recoil, sonicMassFlow}
	for _, value := range values {
		if !finite(value) {
			return Result{}, ErrNoRepresentableFlow
		}
	}
	if massFlow <= 0 || exitSpeed <= 0 || sonicMassFlow <= 0 {
		return Result{}, ErrNoRepresentableFlow
	}

	return Result{
		request:               request,
		regime:                regime,
		effectiveArea:         effective,
		criticalPressureRatio: criticalRatio,
		backPressureRatio:     backRatio,
		throatMach:            MachNumber{value: mach},
		exitPressure:          Pressure{pascals: exitPressure},
		exitTemperature:       Temperature{kelvin: exitTemperature},
		exitSpeed:             Speed{metresPerSecond: exitSpeed},
		massFlow:              MassFlow{kilogramsPerSecond: massFlow},
		sonicMassFlow:         MassFlow{kilogramsPerSecond: sonicMassFlow},
		thrust:                Force{newtons: thrust},
		recoil:                Force{newtons: recoil},
		massFlowResidual: MassFlowResidual{
			kilogramsPerSecond: stableSignedSum([]float64{massFlow, -cd * kinematicMassFlow}),
		},
		thrustResidual: ForceResidual{
			newtons: stableSignedSum([]float64{thrust, -massFlow * exitSpeed, -pressureThrust}),
		},
		recoilResidual: ForceResidual{
			newtons: stableSignedSum([]float64{recoil, thrust}),
		},
	}, nil
}

func sonicMassFlowRate(cd, area, p0, t0, gasR, gamma, criticalRatio float64) (float64, error) {
	throatTemperature := t0 * 2 / (gamma + 1)
	throatPressure := p0 * criticalRatio
	if err := positiveFinite(throatTemperature); err != nil {
		return 0, ErrNoRepresentableFlow
	}
	if err := positiveFinite(throatPressure); err != nil {
		return 0, ErrNoRepresentableFlow
	}
	density := throatPressure / (gasR * throatTemperature)
	speed := math.Sqrt(gamma * gasR * throatTemperature)
	massFlow := cd * density * area * speed
	if !finite(density) || !finite(speed) || !finite(massFlow) || massFlow < 0 {
		return 0, ErrNoRepresentableFlow
	}
	return massFlow, nil
}

func zeroFlow(request Request, effective Area, criticalRatio, backRatio float64) Result {
	sonic := 0.0
	if effective.squareMetres > 0 {
		gamma := request.stagnation.gamma.value
		computed, err := sonicMassFlowRate(
			request.cd.value,
			effective.squareMetres,
			request.stagnation.pressure.pascals,
			request.stagnation.temperature.kelvin,
			request.stagnation.gasConstant.joulesPerKilogramKelvin,
			gamma,
			criticalRatio,
		)
		if err == nil {
			sonic = computed
		}
	}
	return Result{
		request:               request,
		regime:                RegimeNoFlow,
		effectiveArea:         effective,
		criticalPressureRatio: criticalRatio,
		backPressureRatio:     backRatio,
		exitPressure:          request.back,
		exitTemperature:       request.stagnation.temperature,
		sonicMassFlow:         MassFlow{kilogramsPerSecond: sonic},
	}
}

func validateRequest(request Request) error {
	if err := validateStagnation(request.stagnation); err != nil {
		return err
	}
	if err := positiveFinite(request.back.pascals); err != nil {
		return err
	}
	if err := validateAreaLaw(request.area); err != nil {
		return err
	}
	if err := validateDischargeCoefficient(request.cd); err != nil {
		return err
	}
	return nil
}

func validateStagnation(state Stagnation) error {
	if err := positiveFinite(state.pressure.pascals); err != nil {
		return ErrInvalidStagnation
	}
	if err := positiveFinite(state.temperature.kelvin); err != nil {
		return ErrInvalidStagnation
	}
	if err := positiveFinite(state.gasConstant.joulesPerKilogramKelvin); err != nil {
		return ErrInvalidStagnation
	}
	if !finite(state.gamma.value) || state.gamma.value <= 1 {
		return ErrInvalidStagnation
	}
	critical := state.CriticalPressureRatio()
	if !finite(critical) || critical <= 0 || critical >= 1 {
		return ErrInvalidStagnation
	}
	return nil
}

func validateAreaLaw(law AreaLaw) error {
	if err := validateArea(law.prescribed); err != nil {
		return ErrInvalidAreaLaw
	}
	if err := validateArea(law.maximum); err != nil {
		return ErrInvalidAreaLaw
	}
	if err := validateCompliance(law.compliance); err != nil {
		return ErrInvalidAreaLaw
	}
	if law.maximum.squareMetres < law.prescribed.squareMetres {
		return ErrInvalidAreaLaw
	}
	return nil
}

func validateArea(area Area) error {
	if !finite(area.squareMetres) {
		return ErrNonFiniteValue
	}
	if area.squareMetres < 0 {
		return ErrNegativeArea
	}
	return nil
}

func validateCompliance(compliance AreaCompliance) error {
	if !finite(compliance.squareMetresPerPascal) {
		return ErrNonFiniteValue
	}
	if compliance.squareMetresPerPascal < 0 {
		return ErrNegativeCompliance
	}
	return nil
}

func validateDischargeCoefficient(cd DischargeCoefficient) error {
	if !finite(cd.value) {
		return ErrNonFiniteValue
	}
	if cd.value <= 0 || cd.value > 1 {
		return ErrInvalidDischargeCoefficient
	}
	return nil
}
