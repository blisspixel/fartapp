package restrictionhistory

import (
	"errors"
	"math"
	"testing"

	"github.com/blisspixel/fartapp/internal/restrictionflow"
)

func TestConstantChokedHistoryClosedForm(t *testing.T) {
	history := mustIntegrate(t, chokedSamples(t, 0, 0.01, 0.01, 0.01))
	mdot := 0.01 * math.Sqrt(96000)
	assertNear(t, "mass", history.MassOutKilograms(), mdot*0.01, 1e-12)
	assertNear(t, "impulse", history.ImpulseNewtonSeconds(), 11, 1e-12)
	assertNear(t, "recoil impulse", history.RecoilImpulseNewtonSeconds(), -11, 1e-12)
	assertNear(t, "recoil residual", history.RecoilResidualNewtonSeconds(), 0, 1e-15)
	cp := 1.5 * 200 / 0.5
	assertNear(t, "enthalpy", history.EnthalpyOutJoules(), mdot*cp*320*0.01, 1e-9)
	assertNear(t, "exit kinetic energy", history.KineticEnergyOutJoules(), mdot*48000*0.01, 1e-9)
	assertNear(t, "stagnation enthalpy", history.TotalEnthalpyOutJoules(), mdot*240000*0.01, 1e-9)
	assertNear(t, "transport energy", history.EnthalpyOutJoules()+history.KineticEnergyOutJoules(), history.TotalEnthalpyOutJoules(), 1e-9)
	if len(history.Samples()) != 2 || history.Samples()[0].Result().Regime() != restrictionflow.RegimeChoked {
		t.Fatalf("samples = %#v", history.Samples())
	}
	if history.Samples()[0].Result().MassFlow().KilogramsPerSecond() == 0 {
		t.Fatal("history hid the frozen stagnation flow")
	}
	if history.Stagnation().Pressure().Pascals() != 125000 ||
		history.BackPressure().Pascals() != 50000 ||
		history.DischargeCoefficient().Value() != 1 {
		t.Fatalf("history lost its inputs: %#v", history)
	}
}

func TestSingleSampleHasZeroIntegrals(t *testing.T) {
	history := mustIntegrate(t, chokedSamples(t, 0, 0.01))
	if history.MassOutKilograms() != 0 || history.EnthalpyOutJoules() != 0 ||
		history.ImpulseNewtonSeconds() != 0 || history.RecoilImpulseNewtonSeconds() != 0 {
		t.Fatalf("single sample leaked an integral: %#v", history)
	}
}

func TestNonFiniteIntegralsAreRefused(t *testing.T) {
	_, err := Integrate(mustStagnation(t), mustPressure(t, 50000), mustDischarge(t, 1),
		chokedSamples(t, 0, 0.01, math.MaxFloat64, 0.01))
	if !errors.Is(err, ErrNonFiniteIntegral) {
		t.Fatalf("overflowing frozen-source history: %v", err)
	}
}

func TestZeroFlowDoesNotEvaluateUnneededOverflowingHeatCapacity(t *testing.T) {
	stagnation := mustThermodynamicState(t, 1, 1e308, math.Nextafter(1, 2))
	for _, test := range []struct {
		name string
		back float64
		area float64
	}{
		{"closed", 50000, 0},
		{"closed adverse pressure", 200000, 0},
		{"equal pressure", 125000, 0.01},
	} {
		t.Run(test.name, func(t *testing.T) {
			history, err := Integrate(stagnation, mustPressure(t, test.back), mustDischarge(t, 1),
				chokedSamples(t, 0, test.area, 1, test.area))
			if err != nil {
				t.Fatalf("exact zero-flow interval: %v", err)
			}
			for _, value := range []float64{
				history.MassOutKilograms(), history.EnthalpyOutJoules(), history.KineticEnergyOutJoules(),
				history.TotalEnthalpyOutJoules(), history.ImpulseNewtonSeconds(), history.RecoilImpulseNewtonSeconds(),
			} {
				if value != 0 {
					t.Fatalf("no-flow integral = %g, want exactly zero", value)
				}
			}
			for _, sample := range history.Samples() {
				if sample.Result().Regime() != restrictionflow.RegimeNoFlow {
					t.Fatalf("no-flow sample regime = %s", sample.Result().Regime())
				}
			}
		})
	}
}

func TestPositiveMassCannotHideUnrepresentableTransportEnergy(t *testing.T) {
	_, err := Integrate(mustThermodynamicState(t, 1, 1e-20, 1.5),
		mustPressure(t, 50000), mustDischarge(t, 1),
		chokedSamples(t, 0, 0.01, math.SmallestNonzeroFloat64, 0.01))
	if !errors.Is(err, ErrNonFiniteIntegral) {
		t.Fatalf("positive mass with unrepresentable energy: %v", err)
	}
}

func TestFiniteEnthalpyRecoversFromOverflowingSpecificHeat(t *testing.T) {
	gamma := math.Nextafter(1, 2)
	history, err := Integrate(mustThermodynamicState(t, 1e-300, 1e308, gamma),
		mustPressure(t, 50000), mustDischarge(t, 1),
		chokedSamples(t, 0, 1e-10, 1, 1e-10))
	if err != nil {
		t.Fatalf("finite transported enthalpy despite unrepresentable cp: %v", err)
	}
	// At gamma=1+epsilon the sonic mass-rate limit differs only at roundoff.
	// R*T0=1e8, so this independent enthalpy factorization stays finite
	// without constructing cp or sharing production's exponent arithmetic.
	wantMass := 1e-10 * 125000 * math.Exp(-0.5) / math.Sqrt(1e8)
	wantTotal := wantMass * 1e8 * (gamma / (gamma - 1))
	wantKinetic := wantMass * (0.5 * 1e8)
	assertNear(t, "mass", history.MassOutKilograms(), wantMass, wantMass*4e-15)
	assertNear(t, "total enthalpy", history.TotalEnthalpyOutJoules(), wantTotal, wantTotal*4e-15)
	assertNear(t, "kinetic energy", history.KineticEnergyOutJoules(), wantKinetic, wantKinetic*4e-15)
	assertNear(t, "static enthalpy", history.EnthalpyOutJoules(), wantTotal-wantKinetic, wantTotal*4e-15)
	assertNear(t, "transport closure", history.EnthalpyOutJoules()+history.KineticEnergyOutJoules(),
		history.TotalEnthalpyOutJoules(), wantTotal*4e-15)
}

func TestClosedOpenClosedPulse(t *testing.T) {
	history := mustIntegrate(t, chokedSamples(t, 0, 0, 0.01, 0.01, 0.02, 0))
	if history.Samples()[0].Result().Regime() != restrictionflow.RegimeNoFlow ||
		history.Samples()[1].Result().Regime() != restrictionflow.RegimeChoked ||
		history.Samples()[2].Result().Regime() != restrictionflow.RegimeNoFlow {
		t.Fatalf("pulse regimes = %#v", history.Samples())
	}
	if history.MassOutKilograms() <= 0 || history.ImpulseNewtonSeconds() <= 0 {
		t.Fatalf("pulse integrals = %#v", history)
	}
}

func TestConstructorsAndIntegrateRejectInvalidInput(t *testing.T) {
	if _, err := NewSeconds(-1); !errors.Is(err, ErrInvalidTime) {
		t.Fatalf("negative time = %v", err)
	}
	if _, err := NewSeconds(math.NaN()); !errors.Is(err, ErrInvalidTime) {
		t.Fatalf("NaN time = %v", err)
	}
	if _, err := NewSample(Seconds{value: -1}, mustArea(t, 0)); !errors.Is(err, ErrInvalidTime) {
		t.Fatalf("forged sample time = %v", err)
	}
	if _, err := Integrate(restrictionflow.Stagnation{}, restrictionflow.Pressure{}, restrictionflow.DischargeCoefficient{}, nil); !errors.Is(err, ErrInvalidSampleCount) {
		t.Fatalf("empty samples = %v", err)
	}
	tooMany := make([]Sample, MaxSamples+1)
	for index := range tooMany {
		tooMany[index] = mustSample(t, float64(index), 0)
	}
	if _, err := Integrate(
		mustStagnation(t), mustPressure(t, 50000), mustDischarge(t, 1), tooMany,
	); !errors.Is(err, ErrInvalidSampleCount) {
		t.Fatalf("too many samples = %v", err)
	}
	if _, err := Integrate(
		mustStagnation(t), mustPressure(t, 50000), mustDischarge(t, 1),
		[]Sample{mustSample(t, 0.1, 0.01), mustSample(t, 0.1, 0.01)},
	); !errors.Is(err, ErrInvalidTime) {
		t.Fatalf("repeated time = %v", err)
	}
	if _, err := Integrate(
		mustStagnation(t), mustPressure(t, 200000), mustDischarge(t, 1),
		[]Sample{mustSample(t, 0, 0.01)},
	); !errors.Is(err, restrictionflow.ErrAdversePressure) {
		t.Fatalf("adverse history = %v", err)
	}
}

func TestSamplesAreCopied(t *testing.T) {
	history := mustIntegrate(t, chokedSamples(t, 0, 0.01, 0.01, 0.01))
	exposed := history.Samples()
	exposed[0] = Instant{}
	if history.Samples()[0].Time().Value() != 0 {
		t.Fatal("Samples exposed the backing array")
	}
}

func FuzzIntegrate(f *testing.F) {
	f.Add(0.0, 0.01, 0.01, 0.01)
	f.Add(0.0, 0.0, 0.02, 0.01)
	f.Fuzz(func(t *testing.T, t0, a0, t1, a1 float64) {
		if !finite(t0) || !finite(t1) || t0 < 0 || t1 <= t0 || t1 > 10 {
			t.Skip()
		}
		if !finite(a0) || !finite(a1) || a0 < 0 || a1 < 0 || a0 > 1 || a1 > 1 {
			t.Skip()
		}
		history, err := Integrate(
			mustStagnation(t), mustPressure(t, 50000), mustDischarge(t, 1),
			[]Sample{mustSample(t, t0, a0), mustSample(t, t1, a1)},
		)
		if err != nil {
			t.Skip()
		}
		if history.MassOutKilograms() < 0 || history.EnthalpyOutJoules() < 0 {
			t.Fatalf("negative integral: %#v", history)
		}
		if math.Abs(history.RecoilImpulseNewtonSeconds()+history.ImpulseNewtonSeconds()) > 1e-9*(1+math.Abs(history.ImpulseNewtonSeconds())) {
			t.Fatalf("recoil did not cancel: %#v", history)
		}
	})
}

func chokedSamples(t *testing.T, pairs ...float64) []Sample {
	t.Helper()
	if len(pairs)%2 != 0 {
		t.Fatal("expected time, area pairs")
	}
	samples := make([]Sample, 0, len(pairs)/2)
	for index := 0; index < len(pairs); index += 2 {
		samples = append(samples, mustSample(t, pairs[index], pairs[index+1]))
	}
	return samples
}

func mustIntegrate(t *testing.T, samples []Sample) History {
	t.Helper()
	history, err := Integrate(mustStagnation(t), mustPressure(t, 50000), mustDischarge(t, 1), samples)
	if err != nil {
		t.Fatalf("Integrate: %v", err)
	}
	return history
}

func mustStagnation(t *testing.T) restrictionflow.Stagnation {
	t.Helper()
	return mustThermodynamicState(t, 400, 200, 1.5)
}

func mustThermodynamicState(t *testing.T, temperatureK, gasConstant, heatCapacityRatio float64) restrictionflow.Stagnation {
	t.Helper()
	pressure, err := restrictionflow.NewPressure(125000)
	if err != nil {
		t.Fatal(err)
	}
	temperature, err := restrictionflow.NewTemperature(temperatureK)
	if err != nil {
		t.Fatal(err)
	}
	gas, err := restrictionflow.NewSpecificGasConstant(gasConstant)
	if err != nil {
		t.Fatal(err)
	}
	gamma, err := restrictionflow.NewHeatCapacityRatio(heatCapacityRatio)
	if err != nil {
		t.Fatal(err)
	}
	stagnation, err := restrictionflow.NewStagnation(pressure, temperature, gas, gamma)
	if err != nil {
		t.Fatal(err)
	}
	return stagnation
}

func mustPressure(t *testing.T, pascals float64) restrictionflow.Pressure {
	t.Helper()
	value, err := restrictionflow.NewPressure(pascals)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func mustDischarge(t *testing.T, value float64) restrictionflow.DischargeCoefficient {
	t.Helper()
	result, err := restrictionflow.NewDischargeCoefficient(value)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func mustSample(t *testing.T, time, area float64) Sample {
	t.Helper()
	sample, err := NewSample(mustSeconds(t, time), mustArea(t, area))
	if err != nil {
		t.Fatal(err)
	}
	return sample
}

func mustSeconds(t *testing.T, value float64) Seconds {
	t.Helper()
	seconds, err := NewSeconds(value)
	if err != nil {
		t.Fatal(err)
	}
	return seconds
}

func mustArea(t *testing.T, value float64) restrictionflow.Area {
	t.Helper()
	area, err := restrictionflow.NewArea(value)
	if err != nil {
		t.Fatal(err)
	}
	return area
}

func assertNear(t *testing.T, name string, got, want, tolerance float64) {
	t.Helper()
	if math.IsNaN(got) || math.IsInf(got, 0) || math.Abs(got-want) > tolerance {
		t.Fatalf("%s = %g, want %g ± %g", name, got, want, tolerance)
	}
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
