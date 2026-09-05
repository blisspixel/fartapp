package coupledblowdown

import (
	"errors"
	"math"
	"testing"

	"github.com/blisspixel/fartapp/internal/idealmixturereservoir"
	"github.com/blisspixel/fartapp/internal/restrictionflow"
)

func TestIsothermalChokedMassFollowsExponential(t *testing.T) {
	config := mustConfig(t, idealmixturereservoir.RigidIsothermal, 0.01, 0.005, 1024, 0)
	mdot0 := 0.01 * math.Sqrt(96000)
	mass0 := config.Reservoir().TotalMass().Kilograms()
	tau := mass0 / mdot0
	config = mustConfig(t, idealmixturereservoir.RigidIsothermal, 0.01, 0.002, 4096, -tau*math.Log(0.9))
	result := mustSimulate(t, config)
	want := mass0 * math.Exp(-result.ElapsedSeconds()/tau)
	assertNear(t, "mass", result.Final().TotalMass().Kilograms(), want, 5e-3*mass0)
	if !result.Signature().ChokedOccurred() {
		t.Fatal("isothermal fixture did not remain choked")
	}
	if result.Stop() != StopMaxTime && result.Stop() != StopMaxSteps {
		t.Fatalf("stop = %s", result.Stop())
	}
}

func TestAdiabaticPathMatchesReservoirEndpoint(t *testing.T) {
	result := mustSimulate(t, mustConfig(t, idealmixturereservoir.RigidAdiabatic, 0.01, 0.01, 64, 0.002))
	fraction := result.MassOutKilograms() / result.Config().Reservoir().TotalMass().Kilograms()
	if fraction <= 0 {
		t.Fatalf("no mass left the reservoir: %#v", result)
	}
	withdrawal, err := idealmixturereservoir.NewWithdrawalFraction(fraction)
	if err != nil {
		t.Fatal(err)
	}
	transition, err := idealmixturereservoir.WithdrawFraction(
		result.Config().Reservoir(), withdrawal, idealmixturereservoir.RigidAdiabatic,
	)
	if err != nil {
		t.Fatal(err)
	}
	assertNear(t, "temperature", result.Final().Temperature().Kelvin(), transition.After().Temperature().Kelvin(), 1e-6)
	assertNear(t, "pressure", result.Final().Pressure().Pascals(), transition.After().Pressure().Pascals(), 1e-4)
	assertNear(t, "mass residual", result.Ledgers().MassResidualKilograms(), 0, 1e-12)
	assertNear(t, "energy residual", result.Ledgers().EnergyResidualJoules(), 0, 1e-6)
	assertNear(t, "impulse residual", result.Ledgers().ImpulseResidualNewtonSeconds(), 0, 1e-12)
}

func TestZeroAreaIsNoFlowIdentity(t *testing.T) {
	result := mustSimulate(t, mustConfig(t, idealmixturereservoir.RigidAdiabatic, 0, 0.01, 8, 1))
	if result.Stop() != StopNoFlow || result.MassOutKilograms() != 0 || result.ElapsedSeconds() != 0 {
		t.Fatalf("zero area = %#v", result)
	}
	if result.Final().Pressure().Pascals() != result.Config().Reservoir().Pressure().Pascals() {
		t.Fatal("zero area changed the reservoir")
	}
	if _, defined := result.Signature().FormationNumber(); defined {
		t.Fatal("zero area invented a formation number")
	}
}

func TestEqualizationFractionIsothermal(t *testing.T) {
	state := oneComponentState(t)
	back := mustPressure(t, 62500)
	fraction, err := EqualizationFraction(state, back, idealmixturereservoir.RigidIsothermal)
	if err != nil {
		t.Fatal(err)
	}
	assertNear(t, "equalization", fraction, 0.5, 1e-15)
	zero, err := EqualizationFraction(state, mustPressure(t, 125000), idealmixturereservoir.RigidIsothermal)
	if err != nil || zero != 0 {
		t.Fatalf("equal pressure = (%g, %v)", zero, err)
	}
}

func TestEqualPressureStopsWithoutFlow(t *testing.T) {
	state := oneComponentState(t)
	config, err := NewConfig(
		state, idealmixturereservoir.RigidIsothermal,
		mustPressure(t, state.Pressure().Pascals()),
		mustArea(t, 0.01), mustDischarge(t, 1),
		0.01, 8, 1,
	)
	if err != nil {
		t.Fatal(err)
	}
	result := mustSimulate(t, config)
	if result.Stop() != StopNoFlow || result.MassOutKilograms() != 0 {
		t.Fatalf("equal pressure = %#v", result)
	}
}

func TestMaxStepsStopAndSampleAccessors(t *testing.T) {
	result := mustSimulate(t, mustConfig(t, idealmixturereservoir.RigidAdiabatic, 0.01, 0.01, 3, 0))
	if result.Stop() != StopMaxSteps {
		t.Fatalf("stop = %s", result.Stop())
	}
	sample := result.Samples()[0]
	if sample.Pressure() <= 0 || sample.Temperature() <= 0 || sample.Regime() == 0 ||
		sample.MassFlow() < 0 || sample.ExitSpeed() < 0 {
		t.Fatalf("sample = %#v", sample)
	}
	_ = sample.Thrust()
	_ = sample.Recoil()
	if result.EnthalpyOutJoules() <= 0 || result.RecoilImpulseNewtonSeconds() >= 0 {
		t.Fatalf("transfers = %#v", result)
	}
	if result.Config().BackPressure().Pascals() != 50000 ||
		result.Config().AreaLaw().Prescribed().SquareMetres() != 0.01 ||
		result.Config().Reservoir().Volume().CubicMetres() != 1 {
		t.Fatal("config accessors lost inputs")
	}
}

func TestIsothermalSimulationEqualizes(t *testing.T) {
	result := mustSimulate(t, mustConfig(t, idealmixturereservoir.RigidIsothermal, 0.01, 0.2, 64, 0))
	if result.Stop() != StopEqualized && result.Stop() != StopNoFlow {
		t.Fatalf("stop = %s", result.Stop())
	}
	if result.Final().Pressure().Pascals() < 50000-1 {
		t.Fatalf("overshot back pressure: %g", result.Final().Pressure().Pascals())
	}
	if StopEqualized.String() != "equalized" || StopMaxSteps.String() != "max-steps" ||
		StopMaxTime.String() != "max-time" || StopNoProgress.String() != "no-progress" {
		t.Fatal("stop names changed")
	}
}

func TestAdiabaticEqualizationFraction(t *testing.T) {
	state := oneComponentState(t)
	back := mustPressure(t, 64000)
	fraction, err := EqualizationFraction(state, back, idealmixturereservoir.RigidAdiabatic)
	if err != nil {
		t.Fatal(err)
	}
	if fraction <= 0 || fraction >= 1 {
		t.Fatalf("adiabatic equalization = %g", fraction)
	}
	if _, err := EqualizationFraction(state, restrictionflow.Pressure{}, idealmixturereservoir.RigidAdiabatic); err == nil {
		t.Fatal("zero back pressure was accepted")
	}
}

func TestLedgersAndSignatureAccessors(t *testing.T) {
	result := mustSimulate(t, mustConfig(t, idealmixturereservoir.RigidIsothermal, 0.01, 0.02, 32, 0.001))
	if result.Steps() <= 0 || result.ElapsedSeconds() <= 0 || result.MassOutKilograms() <= 0 {
		t.Fatalf("no progress: %#v", result)
	}
	if result.HeatInJoules() <= 0 {
		t.Fatal("isothermal blowdown had no heat")
	}
	samples := result.Samples()
	samples[0] = Sample{}
	if result.Samples()[0].Time() != 0 || result.Samples()[0].Mass() <= 0 {
		t.Fatal("Samples exposed the backing array")
	}
	number, defined := result.Signature().FormationNumber()
	if !defined || number <= 0 || result.Signature().EquivalentDiameterMetres() <= 0 {
		t.Fatalf("signature = %#v", result.Signature())
	}
	if result.Config().Closure() != idealmixturereservoir.RigidIsothermal ||
		result.Config().DischargeCoefficient().Value() != 1 {
		t.Fatal("config accessors lost inputs")
	}
	if StopReason(0).String() != "unsupported" || StopNoFlow.String() != "no-flow" {
		t.Fatal("stop names changed")
	}
}

func TestConstructorsRejectInvalidPolicy(t *testing.T) {
	state := oneComponentState(t)
	back := mustPressure(t, 50000)
	area := mustArea(t, 0.01)
	cd := mustDischarge(t, 1)
	if _, err := NewConfig(state, 0, back, area, cd, 0.01, 8, 1); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("bad closure = %v", err)
	}
	if _, err := NewConfig(state, idealmixturereservoir.RigidAdiabatic, back, area, cd, 0, 8, 1); !errors.Is(err, ErrInvalidStepPolicy) {
		t.Fatalf("bad fraction = %v", err)
	}
	if _, err := NewConfig(state, idealmixturereservoir.RigidAdiabatic, back, area, cd, 0.01, 0, 1); !errors.Is(err, ErrInvalidStepPolicy) {
		t.Fatalf("bad steps = %v", err)
	}
	if _, err := NewConfig(state, idealmixturereservoir.RigidAdiabatic, back, area, cd, 0.01, 8, math.NaN()); !errors.Is(err, ErrInvalidStepPolicy) {
		t.Fatalf("bad time = %v", err)
	}
}

func FuzzSimulate(f *testing.F) {
	f.Add(0.01, 0.01, 16, 0.001)
	f.Add(0.0, 0.02, 4, 1.0)
	f.Fuzz(func(t *testing.T, area, fraction float64, steps int, maxTime float64) {
		if !finite(area) || area < 0 || area > 0.05 {
			t.Skip()
		}
		if !finite(fraction) || fraction < 0.001 || fraction > 0.2 {
			t.Skip()
		}
		if steps < 1 || steps > 64 {
			t.Skip()
		}
		if !finite(maxTime) || maxTime < 0 || maxTime > 1 {
			t.Skip()
		}
		config, err := NewConfig(
			oneComponentState(t), idealmixturereservoir.RigidAdiabatic,
			mustPressure(t, 50000), mustArea(t, area), mustDischarge(t, 1),
			fraction, steps, maxTime,
		)
		if err != nil {
			t.Skip()
		}
		result, err := Simulate(config)
		if err != nil {
			t.Skip()
		}
		if result.MassOutKilograms() < 0 || result.ElapsedSeconds() < 0 {
			t.Fatalf("negative progress: %#v", result)
		}
		if math.Abs(result.Ledgers().ImpulseResidualNewtonSeconds()) > 1e-9*(1+math.Abs(result.ImpulseNewtonSeconds())) {
			t.Fatalf("impulse ledger: %#v", result.Ledgers())
		}
	})
}

func mustSimulate(t *testing.T, config Config) Result {
	t.Helper()
	result, err := Simulate(config)
	if err != nil {
		t.Fatalf("Simulate: %v", err)
	}
	return result
}

func mustConfig(
	t *testing.T,
	closure idealmixturereservoir.Closure,
	area, fraction float64,
	steps int,
	maxTime float64,
) Config {
	t.Helper()
	config, err := NewConfig(
		oneComponentState(t), closure, mustPressure(t, 50000),
		mustArea(t, area), mustDischarge(t, 1), fraction, steps, maxTime,
	)
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}
	return config
}

func oneComponentState(t *testing.T) idealmixturereservoir.State {
	t.Helper()
	mass, err := idealmixturereservoir.NewMass(1.5625)
	if err != nil {
		t.Fatal(err)
	}
	gas, err := idealmixturereservoir.NewSpecificGasConstant(200)
	if err != nil {
		t.Fatal(err)
	}
	cv, err := idealmixturereservoir.NewIsochoricHeatCapacity(400)
	if err != nil {
		t.Fatal(err)
	}
	component, err := idealmixturereservoir.NewComponent(mass, gas, cv)
	if err != nil {
		t.Fatal(err)
	}
	volume, err := idealmixturereservoir.NewVolume(1)
	if err != nil {
		t.Fatal(err)
	}
	temperature, err := idealmixturereservoir.NewTemperature(400)
	if err != nil {
		t.Fatal(err)
	}
	state, err := idealmixturereservoir.NewState([]idealmixturereservoir.Component{component}, volume, temperature)
	if err != nil {
		t.Fatal(err)
	}
	return state
}

func mustPressure(t *testing.T, pascals float64) restrictionflow.Pressure {
	t.Helper()
	value, err := restrictionflow.NewPressure(pascals)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func mustArea(t *testing.T, squareMetres float64) restrictionflow.AreaLaw {
	t.Helper()
	area, err := restrictionflow.NewArea(squareMetres)
	if err != nil {
		t.Fatal(err)
	}
	law, err := restrictionflow.NewPrescribedArea(area)
	if err != nil {
		t.Fatal(err)
	}
	return law
}

func mustDischarge(t *testing.T, value float64) restrictionflow.DischargeCoefficient {
	t.Helper()
	result, err := restrictionflow.NewDischargeCoefficient(value)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func assertNear(t *testing.T, name string, got, want, tolerance float64) {
	t.Helper()
	if math.IsNaN(got) || math.IsInf(got, 0) || math.Abs(got-want) > tolerance {
		t.Fatalf("%s = %g, want %g ± %g", name, got, want, tolerance)
	}
}
