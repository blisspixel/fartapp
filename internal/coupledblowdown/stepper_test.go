package coupledblowdown

import (
	"errors"
	"math"
	"strconv"
	"testing"

	"github.com/blisspixel/fartapp/internal/idealmixturereservoir"
	"github.com/blisspixel/fartapp/internal/restrictionflow"
)

func TestTruncatedHistoryRetainsEveryCompletedStep(t *testing.T) {
	for _, steps := range []int{1, 3, MaxSteps} {
		t.Run(strconv.Itoa(steps), func(t *testing.T) {
			config := mustConfig(t, idealmixturereservoir.RigidIsothermal, 0.01, 1e-5, steps, 0)
			result := mustSimulate(t, config)
			if result.Stop() != StopMaxSteps || result.Steps() != steps || len(result.Samples()) != steps+1 {
				t.Fatalf("steps=%d: stop=%s steps=%d samples=%d", steps, result.Stop(), result.Steps(), len(result.Samples()))
			}
			last := result.Samples()[steps]
			if last.Time() != result.ElapsedSeconds() || last.Pressure() != result.Final().Pressure().Pascals() ||
				last.Mass() != result.Final().TotalMass().Kilograms() {
				t.Fatalf("last retained state differs from final result: %#v", last)
			}
			if len(result.Samples()) > MaxSamples {
				t.Fatal("sample budget exceeded")
			}
		})
	}
}

func TestTimeBudgetHasPriorityOverStepTruncation(t *testing.T) {
	config := mustConfig(t, idealmixturereservoir.RigidIsothermal, 0.01, 0.01, 1, 0.001)
	result := mustSimulate(t, config)
	if result.Stop() != StopMaxTime || result.Steps() != 1 || result.ElapsedSeconds() != 0.001 {
		t.Fatalf("time-limited result: stop=%s steps=%d time=%g", result.Stop(), result.Steps(), result.ElapsedSeconds())
	}
	assertNear(t, "first-order withdrawn mass", result.MassOutKilograms(), 0.001*0.01*math.Sqrt(96000), 1e-15)
	if config.MaxWithdrawalFractionPerStep() != 0.01 || config.MaxSteps() != 1 || config.MaxTimeSeconds() != 0.001 {
		t.Fatal("step policy lost authored input")
	}
}

func TestSimulationRevalidatesConfigAndRefusesReverseFlow(t *testing.T) {
	valid := mustConfig(t, idealmixturereservoir.RigidIsothermal, 0.01, 0.01, 1, 0)
	for _, mutate := range []func(*Config){
		func(c *Config) { *c = Config{} },
		func(c *Config) { c.maxSteps = MaxSteps + 1 },
		func(c *Config) { c.maxFraction = math.NaN() },
		func(c *Config) { c.maxTime = math.Inf(1) },
		func(c *Config) { c.closure = 0 },
		func(c *Config) { c.reservoir = idealmixturereservoir.State{} },
		func(c *Config) { c.back = restrictionflow.Pressure{} },
	} {
		forged := valid
		mutate(&forged)
		if result, err := Simulate(forged); err == nil || len(result.Samples()) != 0 {
			t.Fatalf("invalid configuration returned an account: %#v, %v", forged, err)
		}
	}
	back := mustPressure(t, 125001)
	if _, err := NewConfig(valid.reservoir, valid.closure, back, valid.area, valid.cd, 0.01, 1, 0); !errors.Is(err, restrictionflow.ErrAdversePressure) {
		t.Fatalf("initial reverse flow: %v", err)
	}
	if _, err := EqualizationFraction(valid.reservoir, back, valid.closure); !errors.Is(err, restrictionflow.ErrAdversePressure) {
		t.Fatalf("reverse-flow endpoint: %v", err)
	}
	if _, err := EqualizationFraction(idealmixturereservoir.State{}, valid.back, valid.closure); err == nil {
		t.Fatal("zero reservoir yielded an endpoint")
	}
	if _, err := EqualizationFraction(valid.reservoir, valid.back, 0); err == nil {
		t.Fatal("unknown closure yielded an endpoint")
	}
}

func TestPositiveUnrepresentableWithdrawalNeverClaimsEqualization(t *testing.T) {
	result := mustSimulate(t, mustConfig(t, idealmixturereservoir.RigidIsothermal, 0.01, 1e-20, 8, 0))
	if result.Stop() != StopNoProgress || result.Steps() != 0 || len(result.Samples()) != 1 || result.MassOutKilograms() != 0 {
		t.Fatalf("unrepresentable step: %#v", result)
	}
}

func TestMixtureHistoryClosesEveryComponent(t *testing.T) {
	base := oneComponentState(t)
	first := base.Components()[0]
	secondMass, _ := idealmixturereservoir.NewMass(0.5)
	secondR, _ := idealmixturereservoir.NewSpecificGasConstant(100)
	secondCV, _ := idealmixturereservoir.NewIsochoricHeatCapacity(300)
	second, _ := idealmixturereservoir.NewComponent(secondMass, secondR, secondCV)
	state, err := idealmixturereservoir.NewState([]idealmixturereservoir.Component{first, second}, base.Volume(), base.Temperature())
	if err != nil {
		t.Fatal(err)
	}
	config, err := NewConfig(state, idealmixturereservoir.RigidAdiabatic, mustPressure(t, 50000), mustArea(t, 0.01), mustDischarge(t, 1), 0.005, 32, 0)
	if err != nil {
		t.Fatal(err)
	}
	result := mustSimulate(t, config)
	previousTime := -1.0
	for _, sample := range result.Samples() {
		if sample.Time() <= previousTime {
			t.Fatal("history time did not strictly advance")
		}
		previousTime = sample.Time()
		for i, mass := range sample.ComponentMassesKilograms() {
			out := sample.ComponentMassOutKilograms()[i]
			assertNear(t, "component account", mass+out, state.Components()[i].Mass().Kilograms(), 1e-14)
		}
		if sample.ExitPressure() <= 0 || sample.ExitTemperature() <= 0 || sample.EffectiveArea() != 0.01 || sample.EnthalpyFlow() <= 0 {
			t.Fatalf("incomplete source history: %#v", sample)
		}
		// Stagnation enthalpy is static exit enthalpy plus kinetic energy.
		cp := state.MixtureIsobaricHeatCapacity().JoulesPerKilogramKelvin()
		wantRate := sample.MassFlow() * (cp*sample.ExitTemperature() + sample.ExitSpeed()*sample.ExitSpeed()/2)
		assertNear(t, "total enthalpy rate", sample.EnthalpyFlow(), wantRate, 1e-9)
	}
	for _, residual := range result.Ledgers().ComponentMassResidualsKilograms() {
		assertNear(t, "component residual", residual, 0, 1e-14)
	}
	last := result.Samples()[result.Steps()]
	for i, out := range result.ComponentMassOutKilograms() {
		if out != last.ComponentMassOutKilograms()[i] {
			t.Fatal("component total differs from retained history")
		}
	}
	exposed := last.ComponentMassesKilograms()
	exposed[0] = 0
	if result.Samples()[result.Steps()].ComponentMassesKilograms()[0] == 0 {
		t.Fatal("component accessor exposed mutable state")
	}
}

func TestEqualizationNearAmbientPreservesSmallPositiveDifference(t *testing.T) {
	state := oneComponentState(t)
	backValue := math.Nextafter(state.Pressure().Pascals(), 0)
	back := mustPressure(t, backValue)
	fraction, err := EqualizationFraction(state, back, idealmixturereservoir.RigidIsothermal)
	if err != nil || fraction <= 0 {
		t.Fatalf("positive pressure difference lost: fraction=%g err=%v", fraction, err)
	}
	want := (125000 - backValue) / 125000
	assertNear(t, "small fraction", fraction, want, 1e-30)
}
