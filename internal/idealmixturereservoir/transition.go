package idealmixturereservoir

import (
	"github.com/blisspixel/fartapp/internal/floatmath"
	"math"
)

type Closure uint8

const (
	RigidAdiabatic Closure = iota + 1
	RigidIsothermal
)

func (closure Closure) String() string {
	switch closure {
	case RigidAdiabatic:
		return "rigid-adiabatic"
	case RigidIsothermal:
		return "rigid-isothermal"
	default:
		return "unsupported"
	}
}

type WithdrawalFraction struct{ value float64 }

func NewWithdrawalFraction(value float64) (WithdrawalFraction, error) {
	if !finite(value) || value < 0 {
		return WithdrawalFraction{}, ErrInvalidWithdrawal
	}
	if value >= 1 {
		return WithdrawalFraction{}, ErrReservoirExhausted
	}
	return WithdrawalFraction{value: value}, nil
}

func (fraction WithdrawalFraction) Value() float64 { return fraction.value }

type Transition struct {
	before                  State
	after                   State
	closure                 Closure
	withdrawal              WithdrawalFraction
	componentMassOut        []Mass
	componentMassResiduals  []MassResidual
	totalMassOut            Mass
	integratedEnthalpyOut   Energy
	heatIntoReservoir       Energy
	boundaryWorkByReservoir Energy
	massBalanceResidual     MassResidual
	energyBalanceResidual   EnergyResidual
}

func (transition Transition) Before() State { return copyState(transition.before) }

func (transition Transition) After() State { return copyState(transition.after) }

func (transition Transition) Closure() Closure { return transition.closure }

func (transition Transition) Withdrawal() WithdrawalFraction { return transition.withdrawal }

func (transition Transition) ComponentMassOut() []Mass {
	return append([]Mass(nil), transition.componentMassOut...)
}

func (transition Transition) ComponentMassResiduals() []MassResidual {
	return append([]MassResidual(nil), transition.componentMassResiduals...)
}

func (transition Transition) TotalMassOut() Mass { return transition.totalMassOut }

func (transition Transition) IntegratedEnthalpyOut() Energy {
	return transition.integratedEnthalpyOut
}

func (transition Transition) HeatIntoReservoir() Energy { return transition.heatIntoReservoir }

func (transition Transition) BoundaryWorkByReservoir() Energy {
	return transition.boundaryWorkByReservoir
}

func (transition Transition) MassBalanceResidual() MassResidual {
	return transition.massBalanceResidual
}

func (transition Transition) EnergyBalanceResidual() EnergyResidual {
	return transition.energyBalanceResidual
}

func WithdrawFraction(
	before State,
	withdrawal WithdrawalFraction,
	closure Closure,
) (Transition, error) {
	if err := validateState(before); err != nil {
		return Transition{}, ErrInvalidState
	}
	if !finite(withdrawal.value) || withdrawal.value < 0 {
		return Transition{}, ErrInvalidWithdrawal
	}
	if withdrawal.value >= 1 {
		return Transition{}, ErrReservoirExhausted
	}
	if closure != RigidAdiabatic && closure != RigidIsothermal {
		return Transition{}, ErrInvalidClosure
	}
	if withdrawal.value == 0 {
		return zeroTransition(before, withdrawal, closure), nil
	}

	retained := 1 - withdrawal.value
	if retained == 1 {
		return Transition{}, ErrNoRepresentableProgress
	}
	afterComponents := make([]Component, len(before.components))
	massOut := make([]Mass, len(before.components))
	componentResiduals := make([]MassResidual, len(before.components))
	for index, component := range before.components {
		afterMass := component.mass.kilograms * retained
		if afterMass <= 0 || afterMass >= component.mass.kilograms {
			return Transition{}, ErrNoRepresentableProgress
		}
		afterComponents[index] = component
		afterComponents[index].mass = Mass{kilograms: afterMass}
		massOut[index] = Mass{kilograms: component.mass.kilograms - afterMass}
		if massOut[index].kilograms <= 0 {
			return Transition{}, ErrNoRepresentableProgress
		}
		componentResiduals[index] = MassResidual{kilograms: stableSignedSum([]float64{
			component.mass.kilograms,
			-afterMass,
			-massOut[index].kilograms,
		})}
	}

	afterTemperature := before.temperature.kelvin
	logRetained := math.Log1p(-withdrawal.value)
	if closure == RigidAdiabatic {
		exponent := before.MixtureGasConstant().joulesPerKilogramKelvin /
			before.MixtureIsochoricHeatCapacity().joulesPerKilogramKelvin
		decay := exponent * logRetained
		factor := math.Exp(decay)
		if factor < math.Ldexp(1, -1022) {
			afterTemperature = math.Exp(math.Log(before.temperature.kelvin) + decay)
		} else {
			afterTemperature *= factor
		}
	}
	after := State{
		components:  afterComponents,
		volume:      before.volume,
		temperature: Temperature{kelvin: afterTemperature},
	}
	if err := validateState(after); err != nil {
		return Transition{}, ErrInvalidState
	}

	totalMassOut := stableSum(massValues(massOut))
	beforeEnergy := before.InternalEnergy().joules
	afterEnergy := after.InternalEnergy().joules
	var enthalpyOut float64
	var heatIn float64
	if closure == RigidAdiabatic {
		gamma := before.HeatCapacityRatio()
		oneMinusEnergyRatio := -math.Expm1(gamma * logRetained)
		enthalpyOut = floatmath.Product(beforeEnergy, oneMinusEnergyRatio)
	} else {
		enthalpyTerms := make([]float64, len(before.components))
		heatTerms := make([]float64, len(before.components))
		for index, component := range before.components {
			cp := component.heatCV.joulesPerKilogramKelvin +
				component.gasConstant.joulesPerKilogramKelvin
			enthalpyTerms[index] = floatmath.Product(massOut[index].kilograms, cp, before.temperature.kelvin)
			heatTerms[index] = floatmath.Product(massOut[index].kilograms,
				component.gasConstant.joulesPerKilogramKelvin, before.temperature.kelvin)
		}
		enthalpyOut = stableSum(enthalpyTerms)
		heatIn = stableSum(heatTerms)
	}
	values := []float64{totalMassOut, beforeEnergy, afterEnergy, enthalpyOut, heatIn}
	if enthalpyOut <= 0 || (closure == RigidIsothermal && heatIn <= 0) {
		return Transition{}, ErrNoRepresentableProgress
	}
	for _, value := range values {
		if !finite(value) || value < 0 {
			return Transition{}, ErrInvalidState
		}
	}

	afterMass := after.TotalMass().kilograms
	beforeMass := before.TotalMass().kilograms
	return Transition{
		before:                  copyState(before),
		after:                   after,
		closure:                 closure,
		withdrawal:              withdrawal,
		componentMassOut:        massOut,
		componentMassResiduals:  componentResiduals,
		totalMassOut:            Mass{kilograms: totalMassOut},
		integratedEnthalpyOut:   Energy{joules: enthalpyOut},
		heatIntoReservoir:       Energy{joules: heatIn},
		boundaryWorkByReservoir: Energy{},
		massBalanceResidual: MassResidual{
			kilograms: stableSignedSum([]float64{afterMass, totalMassOut, -beforeMass}),
		},
		energyBalanceResidual: EnergyResidual{
			joules: stableSignedSum([]float64{afterEnergy, enthalpyOut, -heatIn, -beforeEnergy}),
		},
	}, nil
}

func zeroTransition(state State, withdrawal WithdrawalFraction, closure Closure) Transition {
	componentMassOut := make([]Mass, len(state.components))
	componentResiduals := make([]MassResidual, len(state.components))
	return Transition{
		before:                 copyState(state),
		after:                  copyState(state),
		closure:                closure,
		withdrawal:             withdrawal,
		componentMassOut:       componentMassOut,
		componentMassResiduals: componentResiduals,
	}
}
