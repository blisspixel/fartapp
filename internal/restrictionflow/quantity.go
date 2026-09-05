// Package restrictionflow implements the first analytical restriction-flow oracle.
//
// The model is a quasi-steady, adiabatic, isentropic, calorically perfect gas
// through a converging restriction with an empirical discharge coefficient.
// Area is either prescribed or a bounded linear quasi-static compliance of
// pressure difference. The package does not model a reservoir history, a
// diverging or shock-containing nozzle, a plume, acoustics, reverse flow, or a
// source ontology.
package restrictionflow

import "errors"

var (
	ErrNonFiniteValue              = errors.New("value must be finite")
	ErrNonPositiveValue            = errors.New("value must be positive")
	ErrNegativeArea                = errors.New("area must be nonnegative")
	ErrNegativeCompliance          = errors.New("area compliance must be nonnegative")
	ErrInvalidDischargeCoefficient = errors.New("discharge coefficient must be in (0, 1]")
	ErrInvalidHeatCapacityRatio    = errors.New("heat-capacity ratio must be greater than one")
	ErrInvalidAreaLaw              = errors.New("area law is outside the model domain")
	ErrInvalidStagnation           = errors.New("stagnation state is outside the model domain")
	ErrAdversePressure             = errors.New("back pressure exceeds stagnation pressure; reverse flow is unsupported")
	ErrNoRepresentableFlow         = errors.New("inputs produce no representable restriction state")
)

type Pressure struct{ pascals float64 }

func NewPressure(pascals float64) (Pressure, error) {
	if err := positiveFinite(pascals); err != nil {
		return Pressure{}, err
	}
	return Pressure{pascals: pascals}, nil
}

func (value Pressure) Pascals() float64 { return value.pascals }

type Temperature struct{ kelvin float64 }

func NewTemperature(kelvin float64) (Temperature, error) {
	if err := positiveFinite(kelvin); err != nil {
		return Temperature{}, err
	}
	return Temperature{kelvin: kelvin}, nil
}

func (value Temperature) Kelvin() float64 { return value.kelvin }

type Area struct{ squareMetres float64 }

func NewArea(squareMetres float64) (Area, error) {
	if !finite(squareMetres) {
		return Area{}, ErrNonFiniteValue
	}
	if squareMetres < 0 {
		return Area{}, ErrNegativeArea
	}
	return Area{squareMetres: squareMetres}, nil
}

func (value Area) SquareMetres() float64 { return value.squareMetres }

type SpecificGasConstant struct{ joulesPerKilogramKelvin float64 }

func NewSpecificGasConstant(value float64) (SpecificGasConstant, error) {
	if err := positiveFinite(value); err != nil {
		return SpecificGasConstant{}, err
	}
	return SpecificGasConstant{joulesPerKilogramKelvin: value}, nil
}

func (value SpecificGasConstant) JoulesPerKilogramKelvin() float64 {
	return value.joulesPerKilogramKelvin
}

type HeatCapacityRatio struct{ value float64 }

func NewHeatCapacityRatio(value float64) (HeatCapacityRatio, error) {
	if !finite(value) {
		return HeatCapacityRatio{}, ErrNonFiniteValue
	}
	if value <= 1 {
		return HeatCapacityRatio{}, ErrInvalidHeatCapacityRatio
	}
	return HeatCapacityRatio{value: value}, nil
}

func (value HeatCapacityRatio) Value() float64 { return value.value }

type DischargeCoefficient struct{ value float64 }

func NewDischargeCoefficient(value float64) (DischargeCoefficient, error) {
	if !finite(value) {
		return DischargeCoefficient{}, ErrNonFiniteValue
	}
	if value <= 0 || value > 1 {
		return DischargeCoefficient{}, ErrInvalidDischargeCoefficient
	}
	return DischargeCoefficient{value: value}, nil
}

func (value DischargeCoefficient) Value() float64 { return value.value }

type AreaCompliance struct{ squareMetresPerPascal float64 }

func NewAreaCompliance(squareMetresPerPascal float64) (AreaCompliance, error) {
	if !finite(squareMetresPerPascal) {
		return AreaCompliance{}, ErrNonFiniteValue
	}
	if squareMetresPerPascal < 0 {
		return AreaCompliance{}, ErrNegativeCompliance
	}
	return AreaCompliance{squareMetresPerPascal: squareMetresPerPascal}, nil
}

func (value AreaCompliance) SquareMetresPerPascal() float64 {
	return value.squareMetresPerPascal
}

type MassFlow struct{ kilogramsPerSecond float64 }

func (value MassFlow) KilogramsPerSecond() float64 { return value.kilogramsPerSecond }

type Speed struct{ metresPerSecond float64 }

func (value Speed) MetresPerSecond() float64 { return value.metresPerSecond }

type Force struct{ newtons float64 }

func (value Force) Newtons() float64 { return value.newtons }

type MachNumber struct{ value float64 }

func (value MachNumber) Value() float64 { return value.value }

type MassFlowResidual struct{ kilogramsPerSecond float64 }

func (value MassFlowResidual) KilogramsPerSecond() float64 { return value.kilogramsPerSecond }

type ForceResidual struct{ newtons float64 }

func (value ForceResidual) Newtons() float64 { return value.newtons }
