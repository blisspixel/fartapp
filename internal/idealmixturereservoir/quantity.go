// Package idealmixturereservoir implements the first analytical physics oracle.
//
// The model is deliberately narrow: a rigid reservoir contains a homogeneous,
// nonreacting, calorically perfect ideal mixture. A transition removes a
// prescribed fraction of every component under either an adiabatic or an
// isothermal closure. Sensible internal energy uses u_i = cv_i*T, with zero at
// zero kelvin inside this model coordinate. The package does not model an
// opening, an exterior, elapsed time, a flow rate, choking, a plume, acoustics,
// or a source ontology.
package idealmixturereservoir

import "errors"

const MaxComponents = 64

var (
	ErrNonFiniteValue          = errors.New("value must be finite")
	ErrNonPositiveValue        = errors.New("value must be positive")
	ErrInvalidComponentSet     = errors.New("component count must be within the model limit")
	ErrInvalidState            = errors.New("state is outside the model domain")
	ErrInvalidClosure          = errors.New("closure is not supported")
	ErrInvalidWithdrawal       = errors.New("withdrawal fraction must be finite and nonnegative")
	ErrReservoirExhausted      = errors.New("withdrawal must leave positive retained mass")
	ErrNoRepresentableProgress = errors.New("positive withdrawal produces no representable state change")
)

type Mass struct{ kilograms float64 }

func NewMass(kilograms float64) (Mass, error) {
	if err := positiveFinite(kilograms); err != nil {
		return Mass{}, err
	}
	return Mass{kilograms: kilograms}, nil
}

func (value Mass) Kilograms() float64 { return value.kilograms }

type Volume struct{ cubicMetres float64 }

func NewVolume(cubicMetres float64) (Volume, error) {
	if err := positiveFinite(cubicMetres); err != nil {
		return Volume{}, err
	}
	return Volume{cubicMetres: cubicMetres}, nil
}

func (value Volume) CubicMetres() float64 { return value.cubicMetres }

type Temperature struct{ kelvin float64 }

func NewTemperature(kelvin float64) (Temperature, error) {
	if err := positiveFinite(kelvin); err != nil {
		return Temperature{}, err
	}
	return Temperature{kelvin: kelvin}, nil
}

func (value Temperature) Kelvin() float64 { return value.kelvin }

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

type IsochoricHeatCapacity struct{ joulesPerKilogramKelvin float64 }

func NewIsochoricHeatCapacity(value float64) (IsochoricHeatCapacity, error) {
	if err := positiveFinite(value); err != nil {
		return IsochoricHeatCapacity{}, err
	}
	return IsochoricHeatCapacity{joulesPerKilogramKelvin: value}, nil
}

func (value IsochoricHeatCapacity) JoulesPerKilogramKelvin() float64 {
	return value.joulesPerKilogramKelvin
}

type IsobaricHeatCapacity struct{ joulesPerKilogramKelvin float64 }

func (value IsobaricHeatCapacity) JoulesPerKilogramKelvin() float64 {
	return value.joulesPerKilogramKelvin
}

type Pressure struct{ pascals float64 }

func (value Pressure) Pascals() float64 { return value.pascals }

type Energy struct{ joules float64 }

func (value Energy) Joules() float64 { return value.joules }

type MassResidual struct{ kilograms float64 }

func (value MassResidual) Kilograms() float64 { return value.kilograms }

type EnergyResidual struct{ joules float64 }

func (value EnergyResidual) Joules() float64 { return value.joules }
