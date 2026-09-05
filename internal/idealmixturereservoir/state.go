package idealmixturereservoir

import "github.com/blisspixel/fartapp/internal/floatmath"

type Component struct {
	mass        Mass
	gasConstant SpecificGasConstant
	heatCV      IsochoricHeatCapacity
}

func NewComponent(
	mass Mass,
	gasConstant SpecificGasConstant,
	heatCV IsochoricHeatCapacity,
) (Component, error) {
	component := Component{mass: mass, gasConstant: gasConstant, heatCV: heatCV}
	if err := validateComponent(component); err != nil {
		return Component{}, err
	}
	return component, nil
}

func (component Component) Mass() Mass { return component.mass }

func (component Component) SpecificGasConstant() SpecificGasConstant {
	return component.gasConstant
}

func (component Component) IsochoricHeatCapacity() IsochoricHeatCapacity {
	return component.heatCV
}

type State struct {
	components  []Component
	volume      Volume
	temperature Temperature
}

func NewState(components []Component, volume Volume, temperature Temperature) (State, error) {
	state := State{
		components:  append([]Component(nil), components...),
		volume:      volume,
		temperature: temperature,
	}
	if err := validateState(state); err != nil {
		return State{}, err
	}
	return state, nil
}

func (state State) Components() []Component {
	return append([]Component(nil), state.components...)
}

func (state State) Volume() Volume { return state.volume }

func (state State) Temperature() Temperature { return state.temperature }

func (state State) TotalMass() Mass {
	return Mass{kilograms: stableSum(componentMasses(state.components))}
}

func (state State) MixtureGasConstant() SpecificGasConstant {
	return SpecificGasConstant{joulesPerKilogramKelvin: weightedProperty(state.components, state.TotalMass().kilograms,
		func(component Component) float64 { return component.gasConstant.joulesPerKilogramKelvin })}
}

func (state State) MixtureIsochoricHeatCapacity() IsochoricHeatCapacity {
	return IsochoricHeatCapacity{joulesPerKilogramKelvin: weightedProperty(state.components, state.TotalMass().kilograms,
		func(component Component) float64 { return component.heatCV.joulesPerKilogramKelvin })}
}

func (state State) MixtureIsobaricHeatCapacity() IsobaricHeatCapacity {
	return IsobaricHeatCapacity{joulesPerKilogramKelvin: state.MixtureIsochoricHeatCapacity().joulesPerKilogramKelvin +
		state.MixtureGasConstant().joulesPerKilogramKelvin}
}

func (state State) HeatCapacityRatio() float64 {
	return state.MixtureIsobaricHeatCapacity().joulesPerKilogramKelvin /
		state.MixtureIsochoricHeatCapacity().joulesPerKilogramKelvin
}

func (state State) Pressure() Pressure {
	return Pressure{pascals: floatmath.ProductOver(state.volume.cubicMetres, state.TotalMass().kilograms,
		state.MixtureGasConstant().joulesPerKilogramKelvin, state.temperature.kelvin)}
}

func (state State) InternalEnergy() Energy {
	terms := make([]float64, len(state.components))
	for index, component := range state.components {
		terms[index] = floatmath.Product(component.mass.kilograms,
			component.heatCV.joulesPerKilogramKelvin, state.temperature.kelvin)
	}
	return Energy{joules: stableSum(terms)}
}

func copyState(state State) State {
	state.components = append([]Component(nil), state.components...)
	return state
}

func validateState(state State) error {
	if len(state.components) == 0 || len(state.components) > MaxComponents {
		return ErrInvalidComponentSet
	}
	if err := positiveFinite(state.volume.cubicMetres); err != nil {
		return err
	}
	if err := positiveFinite(state.temperature.kelvin); err != nil {
		return err
	}
	for _, component := range state.components {
		if err := validateComponent(component); err != nil {
			return err
		}
	}
	derived := []float64{
		state.TotalMass().kilograms,
		state.MixtureGasConstant().joulesPerKilogramKelvin,
		state.MixtureIsochoricHeatCapacity().joulesPerKilogramKelvin,
		state.MixtureIsobaricHeatCapacity().joulesPerKilogramKelvin,
		state.HeatCapacityRatio(),
		state.Pressure().pascals,
		state.InternalEnergy().joules,
	}
	for _, value := range derived {
		if err := positiveFinite(value); err != nil {
			return err
		}
	}
	return nil
}

func validateComponent(component Component) error {
	values := []float64{
		component.mass.kilograms,
		component.gasConstant.joulesPerKilogramKelvin,
		component.heatCV.joulesPerKilogramKelvin,
	}
	for _, value := range values {
		if err := positiveFinite(value); err != nil {
			return err
		}
	}
	return nil
}
