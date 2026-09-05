package idealmixturereservoir

import (
	"errors"
	"math"
	"testing"
)

func TestFiniteTransferSurvivesSubnormalIntermediateProducts(t *testing.T) {
	state := oneComponentState(t, math.Ldexp(1, -500), math.Ldexp(1, -560), math.Ldexp(1, -560),
		math.Ldexp(1, -60), math.Ldexp(1, 1000))
	fraction := math.Ldexp(1, -20)
	result, err := WithdrawFraction(state, mustWithdrawal(t, fraction), RigidIsothermal)
	if err != nil {
		t.Fatal(err)
	}
	if state.Pressure().Pascals() != 1 || result.After().Pressure().Pascals() != 1-fraction ||
		result.After().MixtureGasConstant() != state.MixtureGasConstant() ||
		result.After().MixtureIsochoricHeatCapacity() != state.MixtureIsochoricHeatCapacity() ||
		result.IntegratedEnthalpyOut().Joules() != math.Ldexp(1, -79) ||
		result.HeatIntoReservoir().Joules() != math.Ldexp(1, -80) ||
		result.EnergyBalanceResidual().Joules() != 0 {
		t.Fatal("finite thermodynamic transfer was erased or mixture properties changed")
	}
}

func TestMixtureScaleAndSplittingInvariance(t *testing.T) {
	for _, property := range []float64{math.SmallestNonzeroFloat64, math.Ldexp(1, -560), math.Ldexp(1, 900)} {
		for _, mass := range []float64{math.Ldexp(1, -500), 1, math.Ldexp(1, 900)} {
			components := make([]Component, 64)
			for i := range components {
				components[i] = mustComponent(t, mass/64, property, property)
			}
			for _, componentProperty := range []func(Component) float64{
				func(c Component) float64 { return c.gasConstant.joulesPerKilogramKelvin },
				func(c Component) float64 { return c.heatCV.joulesPerKilogramKelvin },
			} {
				if got := weightedProperty(components, mass, componentProperty); got != property {
					t.Fatalf("splitting changed property: got %g want %g", got, property)
				}
			}
		}
	}
}

func TestPositiveUnrepresentableHeatIsRefused(t *testing.T) {
	state := oneComponentState(t, 1, math.SmallestNonzeroFloat64, 1, 1, 1)
	if _, err := WithdrawFraction(state, mustWithdrawal(t, 0.25), RigidIsothermal); !errors.Is(err, ErrNoRepresentableProgress) {
		t.Fatalf("unrepresentable positive heat = %v", err)
	}
}

func TestAdiabaticCoolingRetainsRepresentableFinalTemperature(t *testing.T) {
	state := oneComponentState(t, math.Ldexp(1, -500), math.Ldexp(1, 533), math.Ldexp(1, 523),
		math.Ldexp(1, 10), math.Ldexp(1, 1000))
	result, err := WithdrawFraction(state, mustWithdrawal(t, 0.75), RigidAdiabatic)
	if err != nil {
		t.Fatal(err)
	}
	if state.Pressure().Pascals() != math.Ldexp(1, 1023) ||
		result.After().Temperature().Kelvin() != math.Ldexp(1, -1048) ||
		result.After().Pressure().Pascals() != math.Ldexp(1, -1027) ||
		result.After().InternalEnergy().Joules() != math.Ldexp(1, -1027) {
		t.Fatal("a vanishing intermediate cooling factor erased a finite endpoint")
	}
}
