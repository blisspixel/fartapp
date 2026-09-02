package idealmixturereservoir

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

func TestSyntheticMixtureClosedForms(t *testing.T) {
	initial := syntheticState(t)
	assertNear(t, "total mass", initial.TotalMass().Kilograms(), 4, 0)
	assertNear(t, "mixture R", initial.MixtureGasConstant().JoulesPerKilogramKelvin(), 350, 0)
	assertNear(t, "mixture cv", initial.MixtureIsochoricHeatCapacity().JoulesPerKilogramKelvin(), 700, 0)
	assertNear(t, "mixture cp", initial.MixtureIsobaricHeatCapacity().JoulesPerKilogramKelvin(), 1050, 0)
	assertNear(t, "gamma", initial.HeatCapacityRatio(), 1.5, 0)
	assertNear(t, "pressure", initial.Pressure().Pascals(), 560000, 0)
	assertNear(t, "energy", initial.InternalEnergy().Joules(), 1120000, 0)

	withdrawal := mustWithdrawal(t, 0.75)
	tests := []struct {
		name        string
		closure     Closure
		temperature float64
		pressure    float64
		energy      float64
		enthalpyOut float64
		heatIn      float64
	}{
		{
			name:        "adiabatic",
			closure:     RigidAdiabatic,
			temperature: 200,
			pressure:    70000,
			energy:      140000,
			enthalpyOut: 980000,
		},
		{
			name:        "isothermal",
			closure:     RigidIsothermal,
			temperature: 400,
			pressure:    140000,
			energy:      280000,
			enthalpyOut: 1260000,
			heatIn:      420000,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			transition, err := WithdrawFraction(initial, withdrawal, test.closure)
			if err != nil {
				t.Fatalf("WithdrawFraction: %v", err)
			}
			after := transition.After()
			masses := after.Components()
			assertNear(t, "component 0 mass", masses[0].Mass().Kilograms(), 0.25, 0)
			assertNear(t, "component 1 mass", masses[1].Mass().Kilograms(), 0.75, 0)
			assertNear(t, "total mass", after.TotalMass().Kilograms(), 1, 0)
			assertNear(t, "temperature", after.Temperature().Kelvin(), test.temperature, 1e-13)
			assertNear(t, "pressure", after.Pressure().Pascals(), test.pressure, 1e-10)
			assertNear(t, "energy", after.InternalEnergy().Joules(), test.energy, 1e-10)
			assertNear(t, "mass out", transition.TotalMassOut().Kilograms(), 3, 0)
			assertNear(t, "enthalpy out", transition.IntegratedEnthalpyOut().Joules(), test.enthalpyOut, 1e-9)
			assertNear(t, "heat in", transition.HeatIntoReservoir().Joules(), test.heatIn, 1e-9)
			assertNear(t, "boundary work", transition.BoundaryWorkByReservoir().Joules(), 0, 0)
			assertNear(t, "mass residual", transition.MassBalanceResidual().Kilograms(), 0, 1e-15)
			assertNear(t, "energy residual", transition.EnergyBalanceResidual().Joules(), 0, 1e-9)
			if transition.Before().Pressure().Pascals() != initial.Pressure().Pascals() {
				t.Fatal("transition did not retain the initial state")
			}
		})
	}
}

func TestZeroWithdrawalIsExactIdentity(t *testing.T) {
	initial := syntheticState(t)
	for _, closure := range []Closure{RigidAdiabatic, RigidIsothermal} {
		transition, err := WithdrawFraction(initial, mustWithdrawal(t, 0), closure)
		if err != nil {
			t.Fatalf("WithdrawFraction(%s): %v", closure, err)
		}
		if !reflect.DeepEqual(transition.Before(), initial) ||
			!reflect.DeepEqual(transition.After(), initial) {
			t.Fatalf("zero transition changed state for %s", closure)
		}
		if transition.TotalMassOut().Kilograms() != 0 ||
			transition.IntegratedEnthalpyOut().Joules() != 0 ||
			transition.HeatIntoReservoir().Joules() != 0 ||
			transition.MassBalanceResidual().Kilograms() != 0 ||
			transition.EnergyBalanceResidual().Joules() != 0 {
			t.Fatalf("zero transition has a nonzero ledger for %s: %#v", closure, transition)
		}
		if transition.Closure() != closure || transition.Withdrawal().Value() != 0 {
			t.Fatalf("zero transition lost its rule or input for %s", closure)
		}
		for _, mass := range transition.ComponentMassOut() {
			if mass.Kilograms() != 0 {
				t.Fatalf("zero transition exported component mass for %s", closure)
			}
		}
		for _, residual := range transition.ComponentMassResiduals() {
			if residual.Kilograms() != 0 {
				t.Fatalf("zero transition has a component residual for %s", closure)
			}
		}
	}
}

func TestTransitionCompositionAndSemigroup(t *testing.T) {
	initial := syntheticState(t)
	one, err := WithdrawFraction(initial, mustWithdrawal(t, 0.75), RigidAdiabatic)
	if err != nil {
		t.Fatal(err)
	}
	first, err := WithdrawFraction(initial, mustWithdrawal(t, 0.5), RigidAdiabatic)
	if err != nil {
		t.Fatal(err)
	}
	second, err := WithdrawFraction(first.After(), mustWithdrawal(t, 0.5), RigidAdiabatic)
	if err != nil {
		t.Fatal(err)
	}
	assertStateNear(t, second.After(), one.After(), 1e-9)

	beforeComponents := initial.Components()
	afterComponents := one.After().Components()
	for index := range beforeComponents {
		beforeFraction := beforeComponents[index].Mass().Kilograms() / initial.TotalMass().Kilograms()
		afterFraction := afterComponents[index].Mass().Kilograms() / one.After().TotalMass().Kilograms()
		assertNear(t, "composition", afterFraction, beforeFraction, 1e-15)
	}
}

func TestComponentPermutationDoesNotChangeAggregates(t *testing.T) {
	initial := syntheticState(t)
	components := initial.Components()
	reversed := []Component{components[1], components[0]}
	permuted, err := NewState(reversed, initial.Volume(), initial.Temperature())
	if err != nil {
		t.Fatal(err)
	}
	if initial.TotalMass() != permuted.TotalMass() ||
		initial.MixtureGasConstant() != permuted.MixtureGasConstant() ||
		initial.MixtureIsochoricHeatCapacity() != permuted.MixtureIsochoricHeatCapacity() ||
		initial.Pressure() != permuted.Pressure() ||
		initial.InternalEnergy() != permuted.InternalEnergy() {
		t.Fatalf("permutation changed aggregate: %#v %#v", initial, permuted)
	}
}

func TestInputsAndResultsAreImmutable(t *testing.T) {
	initial := syntheticState(t)
	components := initial.Components()
	components[0] = Component{}
	if initial.Components()[0] == (Component{}) {
		t.Fatal("Components exposed the state backing array")
	}
	transition, err := WithdrawFraction(initial, mustWithdrawal(t, 0.5), RigidIsothermal)
	if err != nil {
		t.Fatal(err)
	}
	before := transition.Before()
	exposed := before.Components()
	exposed[0] = Component{}
	if transition.Before().Components()[0] == (Component{}) {
		t.Fatal("Before exposed the transition backing array")
	}
	exported := transition.ComponentMassOut()
	exported[0] = Mass{}
	if transition.ComponentMassOut()[0] == (Mass{}) {
		t.Fatal("ComponentMassOut exposed the transition backing array")
	}
	residuals := transition.ComponentMassResiduals()
	residuals[0] = MassResidual{kilograms: 1}
	if transition.ComponentMassResiduals()[0].Kilograms() == 1 {
		t.Fatal("ComponentMassResiduals exposed the transition backing array")
	}
}

func TestConstructorsRejectInvalidValues(t *testing.T) {
	invalidPositive := []struct {
		name string
		call func() error
		want error
	}{
		{name: "mass zero", call: func() error { _, err := NewMass(0); return err }, want: ErrNonPositiveValue},
		{name: "mass negative", call: func() error { _, err := NewMass(-1); return err }, want: ErrNonPositiveValue},
		{name: "volume NaN", call: func() error { _, err := NewVolume(math.NaN()); return err }, want: ErrNonFiniteValue},
		{name: "temperature infinity", call: func() error { _, err := NewTemperature(math.Inf(1)); return err }, want: ErrNonFiniteValue},
		{name: "gas constant zero", call: func() error { _, err := NewSpecificGasConstant(0); return err }, want: ErrNonPositiveValue},
		{name: "heat capacity negative infinity", call: func() error { _, err := NewIsochoricHeatCapacity(math.Inf(-1)); return err }, want: ErrNonFiniteValue},
	}
	for _, test := range invalidPositive {
		t.Run(test.name, func(t *testing.T) {
			if err := test.call(); !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}

	for _, value := range []float64{-1, math.NaN(), math.Inf(1)} {
		if _, err := NewWithdrawalFraction(value); !errors.Is(err, ErrInvalidWithdrawal) {
			t.Fatalf("NewWithdrawalFraction(%v) = %v", value, err)
		}
	}
	for _, value := range []float64{1, 2, math.Inf(1)} {
		_, err := NewWithdrawalFraction(value)
		if value == math.Inf(1) {
			if !errors.Is(err, ErrInvalidWithdrawal) {
				t.Fatalf("NewWithdrawalFraction(+Inf) = %v", err)
			}
		} else if !errors.Is(err, ErrReservoirExhausted) {
			t.Fatalf("NewWithdrawalFraction(%v) = %v", value, err)
		}
	}
}

func TestStateAndTransitionDefendTheirDomains(t *testing.T) {
	validComponent := mustComponent(t, 1, 200, 400)
	volume := mustVolume(t, 1)
	temperature := mustTemperature(t, 300)
	for _, components := range [][]Component{nil, make([]Component, MaxComponents+1), {{}}} {
		if _, err := NewState(components, volume, temperature); err == nil {
			t.Fatalf("NewState accepted %d invalid components", len(components))
		}
	}
	tooMany := make([]Component, MaxComponents+1)
	for index := range tooMany {
		tooMany[index] = validComponent
	}
	if _, err := NewState(tooMany, volume, temperature); !errors.Is(err, ErrInvalidComponentSet) {
		t.Fatalf("NewState too many = %v", err)
	}
	if _, err := NewState([]Component{validComponent}, Volume{}, temperature); err == nil {
		t.Fatal("NewState accepted a forged volume")
	}
	if _, err := NewComponent(Mass{}, mustGasConstant(t, 1), mustHeatCV(t, 1)); err == nil {
		t.Fatal("NewComponent accepted a forged mass")
	}

	state := syntheticState(t)
	if _, err := WithdrawFraction(State{}, mustWithdrawal(t, 0), RigidAdiabatic); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("forged state = %v", err)
	}
	if _, err := WithdrawFraction(state, WithdrawalFraction{value: math.NaN()}, RigidAdiabatic); !errors.Is(err, ErrInvalidWithdrawal) {
		t.Fatalf("forged withdrawal = %v", err)
	}
	if _, err := WithdrawFraction(state, WithdrawalFraction{value: 1}, RigidAdiabatic); !errors.Is(err, ErrReservoirExhausted) {
		t.Fatalf("dry boundary = %v", err)
	}
	if _, err := WithdrawFraction(state, mustWithdrawal(t, 0.5), Closure(99)); !errors.Is(err, ErrInvalidClosure) {
		t.Fatalf("unsupported closure = %v", err)
	}
	if Closure(99).String() != "unsupported" || RigidAdiabatic.String() != "rigid-adiabatic" ||
		RigidIsothermal.String() != "rigid-isothermal" {
		t.Fatal("closure names changed")
	}
}

func TestRepresentabilityFailureIsRejectedWithoutClamping(t *testing.T) {
	component := mustComponent(t, math.SmallestNonzeroFloat64, 1, math.SmallestNonzeroFloat64)
	state, err := NewState([]Component{component}, mustVolume(t, 1), mustTemperature(t, 1))
	if err == nil {
		t.Fatalf("underflowing initial state was accepted: %#v", state)
	}

	state = oneComponentState(t, 1e-200, 1e200, 1, 1, 1)
	if _, err := WithdrawFraction(state, mustWithdrawal(t, 0.5), RigidAdiabatic); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("underflowing transition = %v", err)
	}
	if _, err := WithdrawFraction(state, WithdrawalFraction{value: math.SmallestNonzeroFloat64}, RigidIsothermal); !errors.Is(err, ErrNoRepresentableProgress) {
		t.Fatalf("unrepresentable progress = %v", err)
	}
	state = oneComponentState(t, math.SmallestNonzeroFloat64, 1e307, 1e307, 1, 1)
	if _, err := WithdrawFraction(state, WithdrawalFraction{value: math.Ldexp(1, -53)}, RigidIsothermal); !errors.Is(err, ErrNoRepresentableProgress) {
		t.Fatalf("subnormal component progress = %v", err)
	}
}

func TestIsothermalHeatIsIndependentOfLargeInternalEnergy(t *testing.T) {
	state := oneComponentState(t, 1, 1, 1e20, 1, 1)
	transition, err := WithdrawFraction(state, mustWithdrawal(t, 0.5), RigidIsothermal)
	if err != nil {
		t.Fatal(err)
	}
	assertNear(t, "heat in", transition.HeatIntoReservoir().Joules(), 0.5, 0)
	assertNear(t, "enthalpy out", transition.IntegratedEnthalpyOut().Joules(), 5e19, 8192)
	energyScale := math.Abs(transition.After().InternalEnergy().Joules()) +
		math.Abs(transition.IntegratedEnthalpyOut().Joules()) +
		math.Abs(transition.HeatIntoReservoir().Joules()) +
		math.Abs(state.InternalEnergy().Joules())
	if math.Abs(transition.EnergyBalanceResidual().Joules()) > 16*(math.Nextafter(1, 2)-1)*energyScale {
		t.Fatalf("energy residual = %g for scale %g", transition.EnergyBalanceResidual().Joules(), energyScale)
	}
}

func FuzzWithdrawFraction(f *testing.F) {
	f.Add(1.0, 3.0, 200.0, 400.0, 400.0, 800.0, 1.0, 400.0, 0.75, uint8(1))
	f.Add(0.1, 0.2, 287.0, 188.9, 718.0, 659.0, 0.01, 250.0, 0.0, uint8(2))
	f.Fuzz(func(
		t *testing.T,
		massA, massB, gasA, gasB, cvA, cvB, volume, temperature, withdrawal float64,
		closureByte uint8,
	) {
		values := []float64{massA, massB, gasA, gasB, cvA, cvB, volume, temperature}
		for _, value := range values {
			if !finite(value) || value <= 1e-100 || value >= 1e100 {
				t.Skip()
			}
		}
		if !finite(withdrawal) || withdrawal < 0 || withdrawal > 0.99 {
			t.Skip()
		}
		closure := RigidAdiabatic
		if closureByte%2 == 0 {
			closure = RigidIsothermal
		}
		components := []Component{
			mustComponent(t, massA, gasA, cvA),
			mustComponent(t, massB, gasB, cvB),
		}
		state, err := NewState(components, mustVolume(t, volume), mustTemperature(t, temperature))
		if err != nil {
			t.Skip()
		}
		transition, err := WithdrawFraction(state, mustWithdrawal(t, withdrawal), closure)
		if err != nil {
			t.Skip()
		}
		after := transition.After()
		for _, value := range []float64{
			after.TotalMass().Kilograms(), after.Temperature().Kelvin(),
			after.Pressure().Pascals(), after.InternalEnergy().Joules(),
		} {
			if !finite(value) || value <= 0 {
				t.Fatalf("nonpositive or nonfinite result %g", value)
			}
		}
		massScale := state.TotalMass().Kilograms()
		energyScale := math.Abs(state.InternalEnergy().Joules()) +
			math.Abs(after.InternalEnergy().Joules()) +
			math.Abs(transition.IntegratedEnthalpyOut().Joules()) +
			math.Abs(transition.HeatIntoReservoir().Joules())
		if math.Abs(transition.MassBalanceResidual().Kilograms()) > 1e-12*massScale ||
			math.Abs(transition.EnergyBalanceResidual().Joules()) > 1e-12*energyScale {
			t.Fatalf("ledger did not close: %#v", transition)
		}
	})
}

func syntheticState(t *testing.T) State {
	t.Helper()
	components := []Component{
		mustComponent(t, 1, 200, 400),
		mustComponent(t, 3, 400, 800),
	}
	state, err := NewState(components, mustVolume(t, 1), mustTemperature(t, 400))
	if err != nil {
		t.Fatalf("NewState: %v", err)
	}
	return state
}

func oneComponentState(t *testing.T, mass, gasConstant, cv, volume, temperature float64) State {
	t.Helper()
	state, err := NewState(
		[]Component{mustComponent(t, mass, gasConstant, cv)},
		mustVolume(t, volume),
		mustTemperature(t, temperature),
	)
	if err != nil {
		t.Fatalf("NewState: %v", err)
	}
	return state
}

func mustComponent(t *testing.T, mass, gasConstant, cv float64) Component {
	t.Helper()
	component, err := NewComponent(mustMass(t, mass), mustGasConstant(t, gasConstant), mustHeatCV(t, cv))
	if err != nil {
		t.Fatalf("NewComponent: %v", err)
	}
	return component
}

func mustMass(t *testing.T, value float64) Mass {
	t.Helper()
	result, err := NewMass(value)
	if err != nil {
		t.Fatalf("NewMass: %v", err)
	}
	return result
}

func mustVolume(t *testing.T, value float64) Volume {
	t.Helper()
	result, err := NewVolume(value)
	if err != nil {
		t.Fatalf("NewVolume: %v", err)
	}
	return result
}

func mustTemperature(t *testing.T, value float64) Temperature {
	t.Helper()
	result, err := NewTemperature(value)
	if err != nil {
		t.Fatalf("NewTemperature: %v", err)
	}
	return result
}

func mustGasConstant(t *testing.T, value float64) SpecificGasConstant {
	t.Helper()
	result, err := NewSpecificGasConstant(value)
	if err != nil {
		t.Fatalf("NewSpecificGasConstant: %v", err)
	}
	return result
}

func mustHeatCV(t *testing.T, value float64) IsochoricHeatCapacity {
	t.Helper()
	result, err := NewIsochoricHeatCapacity(value)
	if err != nil {
		t.Fatalf("NewIsochoricHeatCapacity: %v", err)
	}
	return result
}

func mustWithdrawal(t *testing.T, value float64) WithdrawalFraction {
	t.Helper()
	result, err := NewWithdrawalFraction(value)
	if err != nil {
		t.Fatalf("NewWithdrawalFraction: %v", err)
	}
	return result
}

func assertStateNear(t *testing.T, got, want State, tolerance float64) {
	t.Helper()
	assertNear(t, "mass", got.TotalMass().Kilograms(), want.TotalMass().Kilograms(), tolerance)
	assertNear(t, "temperature", got.Temperature().Kelvin(), want.Temperature().Kelvin(), tolerance)
	assertNear(t, "pressure", got.Pressure().Pascals(), want.Pressure().Pascals(), tolerance)
	assertNear(t, "energy", got.InternalEnergy().Joules(), want.InternalEnergy().Joules(), tolerance)
}

func assertNear(t *testing.T, name string, got, want, tolerance float64) {
	t.Helper()
	if math.Abs(got-want) > tolerance {
		t.Fatalf("%s = %.17g, want %.17g within %.3g", name, got, want, tolerance)
	}
}
