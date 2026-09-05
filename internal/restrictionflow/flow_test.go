package restrictionflow

import (
	"errors"
	"math"
	"testing"
)

func TestChokedGammaFifteenClosedForm(t *testing.T) {
	result := mustEvaluate(t, chokedGammaFifteenRequest(t))
	if result.Regime() != RegimeChoked {
		t.Fatalf("regime = %s", result.Regime())
	}
	assertNear(t, "critical ratio", result.CriticalPressureRatio(), 64.0/125.0, 1e-15)
	assertNear(t, "back ratio", result.BackPressureRatio(), 0.4, 0)
	assertNear(t, "mach", result.ThroatMach().Value(), 1, 0)
	assertNear(t, "exit pressure", result.ExitPressure().Pascals(), 64000, 1e-10)
	assertNear(t, "exit temperature", result.ExitTemperature().Kelvin(), 320, 1e-12)
	expectedSpeed := math.Sqrt(96000)
	assertNear(t, "exit speed", result.ExitSpeed().MetresPerSecond(), expectedSpeed, 1e-12)
	assertNear(t, "mass flow", result.MassFlow().KilogramsPerSecond(), 0.01*expectedSpeed, 1e-12)
	assertNear(t, "sonic mass flow", result.SonicMassFlow().KilogramsPerSecond(), 0.01*expectedSpeed, 1e-12)
	assertNear(t, "thrust", result.Thrust().Newtons(), 1100, 1e-12)
	assertNear(t, "recoil", result.Recoil().Newtons(), -1100, 1e-12)
	assertNear(t, "mass residual", result.MassFlowResidual().KilogramsPerSecond(), 0, 1e-15)
	assertNear(t, "thrust residual", result.ThrustResidual().Newtons(), 0, 1e-12)
	assertNear(t, "recoil residual", result.RecoilResidual().Newtons(), 0, 1e-12)
	if result.EffectiveArea().SquareMetres() != 0.01 {
		t.Fatalf("area = %g", result.EffectiveArea().SquareMetres())
	}
}

func TestSubsonicTwoThirdsMachClosedForm(t *testing.T) {
	stagnation := mustStagnation(t, 1_000_000, 400, 200, 1.5)
	request := mustRequest(
		t,
		stagnation,
		mustPressure(t, 729_000),
		mustPrescribed(t, 0.01),
		mustDischarge(t, 1),
	)
	result := mustEvaluate(t, request)
	if result.Regime() != RegimeSubsonic {
		t.Fatalf("regime = %s", result.Regime())
	}
	assertNear(t, "mach", result.ThroatMach().Value(), 2.0/3.0, 1e-15)
	assertNear(t, "exit pressure", result.ExitPressure().Pascals(), 729_000, 0)
	assertNear(t, "exit temperature", result.ExitTemperature().Kelvin(), 360, 1e-12)
	expectedSpeed := (2.0 / 3.0) * math.Sqrt(108000)
	assertNear(t, "exit speed", result.ExitSpeed().MetresPerSecond(), expectedSpeed, 1e-12)
	expectedMassFlow := 10.125 * 0.01 * expectedSpeed
	assertNear(t, "mass flow", result.MassFlow().KilogramsPerSecond(), expectedMassFlow, 1e-12)
	assertNear(t, "thrust", result.Thrust().Newtons(), expectedMassFlow*expectedSpeed, 1e-9)
	if result.MassFlow().KilogramsPerSecond() >= result.SonicMassFlow().KilogramsPerSecond() {
		t.Fatalf("subsonic mass flow %g was not below sonic %g",
			result.MassFlow().KilogramsPerSecond(), result.SonicMassFlow().KilogramsPerSecond())
	}
}

func TestSonicBoundaryIsChokedWithMatchingExitPressure(t *testing.T) {
	stagnation := mustStagnation(t, 125000, 400, 200, 1.5)
	critical := stagnation.CriticalPressure()
	request := mustRequest(t, stagnation, critical, mustPrescribed(t, 0.01), mustDischarge(t, 1))
	result := mustEvaluate(t, request)
	if result.Regime() != RegimeChoked {
		t.Fatalf("regime = %s", result.Regime())
	}
	assertNear(t, "mach", result.ThroatMach().Value(), 1, 0)
	assertNear(t, "exit pressure", result.ExitPressure().Pascals(), critical.Pascals(), 1e-12)
	assertNear(t, "pressure thrust", result.Thrust().Newtons()-result.MassFlow().KilogramsPerSecond()*result.ExitSpeed().MetresPerSecond(), 0, 1e-9)
}

func TestZeroAreaAndEqualPressureAreExactNoFlow(t *testing.T) {
	stagnation := mustStagnation(t, 125000, 400, 200, 1.5)
	closed := mustEvaluate(t, mustRequest(
		t, stagnation, mustPressure(t, 50000), mustPrescribed(t, 0), mustDischarge(t, 1),
	))
	if closed.Regime() != RegimeNoFlow || closed.MassFlow().KilogramsPerSecond() != 0 ||
		closed.Thrust().Newtons() != 0 || closed.Recoil().Newtons() != 0 ||
		closed.ThroatMach().Value() != 0 || closed.SonicMassFlow().KilogramsPerSecond() != 0 {
		t.Fatalf("closed restriction = %#v", closed)
	}

	equal := mustEvaluate(t, mustRequest(
		t, stagnation, stagnation.Pressure(), mustPrescribed(t, 0.01), mustDischarge(t, 1),
	))
	if equal.Regime() != RegimeNoFlow || equal.MassFlow().KilogramsPerSecond() != 0 ||
		equal.Thrust().Newtons() != 0 || equal.ExitPressure().Pascals() != stagnation.Pressure().Pascals() {
		t.Fatalf("equal-pressure restriction = %#v", equal)
	}
	if equal.SonicMassFlow().KilogramsPerSecond() <= 0 {
		t.Fatal("equal-pressure case hid the unused sonic mass-flow comparison")
	}
}

func TestClosedRestrictionIgnoresAdversePressure(t *testing.T) {
	stagnation := mustStagnation(t, 1000, 300, 200, 1.4)
	result := mustEvaluate(t, mustRequest(
		t, stagnation, mustPressure(t, 2000), mustPrescribed(t, 0), mustDischarge(t, 1),
	))
	if result.Regime() != RegimeNoFlow || result.MassFlow().KilogramsPerSecond() != 0 {
		t.Fatalf("closed adverse = %#v", result)
	}
}

func TestAdversePressureIsRejectedWhenOpen(t *testing.T) {
	stagnation := mustStagnation(t, 1000, 300, 200, 1.4)
	_, err := Evaluate(mustRequest(
		t, stagnation, mustPressure(t, 2000), mustPrescribed(t, 0.01), mustDischarge(t, 1),
	))
	if !errors.Is(err, ErrAdversePressure) {
		t.Fatalf("adverse error = %v", err)
	}
}

func TestDischargeCoefficientScalesMassFlowOnly(t *testing.T) {
	full := mustEvaluate(t, chokedGammaFifteenRequest(t))
	halfRequest := chokedGammaFifteenRequest(t)
	halfRequest.cd = mustDischarge(t, 0.5)
	half := mustEvaluate(t, halfRequest)
	assertNear(t, "mass flow", half.MassFlow().KilogramsPerSecond(), 0.5*full.MassFlow().KilogramsPerSecond(), 1e-15)
	assertNear(t, "exit speed", half.ExitSpeed().MetresPerSecond(), full.ExitSpeed().MetresPerSecond(), 0)
	assertNear(t, "exit pressure", half.ExitPressure().Pascals(), full.ExitPressure().Pascals(), 0)
	momentum := half.MassFlow().KilogramsPerSecond() * half.ExitSpeed().MetresPerSecond()
	pressure := (half.ExitPressure().Pascals() - half.Request().BackPressure().Pascals()) *
		half.EffectiveArea().SquareMetres()
	assertNear(t, "thrust", half.Thrust().Newtons(), momentum+pressure, 1e-12)
}

func TestLinearComplianceOpensTowardTheBound(t *testing.T) {
	stagnation := mustStagnation(t, 125000, 400, 200, 1.5)
	back := mustPressure(t, 50000)
	prescribed := mustArea(t, 0.001)
	compliance := mustCompliance(t, 1e-7)
	maximum := mustArea(t, 0.01)
	law, err := NewLinearComplianceArea(prescribed, compliance, maximum)
	if err != nil {
		t.Fatal(err)
	}
	if law.LawName() != "linear-compliance" {
		t.Fatalf("law name = %s", law.LawName())
	}
	result := mustEvaluate(t, mustRequest(t, stagnation, back, law, mustDischarge(t, 1)))
	assertNear(t, "effective area", result.EffectiveArea().SquareMetres(), 0.0085, 1e-15)
	if result.Regime() != RegimeChoked {
		t.Fatalf("regime = %s", result.Regime())
	}

	cappedLaw, err := NewLinearComplianceArea(prescribed, mustCompliance(t, 1), maximum)
	if err != nil {
		t.Fatal(err)
	}
	capped := mustEvaluate(t, mustRequest(t, stagnation, back, cappedLaw, mustDischarge(t, 1)))
	assertNear(t, "capped area", capped.EffectiveArea().SquareMetres(), 0.01, 0)
}

func TestOrdinaryEarthBiologicalPressureCannotChokeAtSeaLevel(t *testing.T) {
	const ambient = 101325.0
	const ordinaryGauge = 930.0
	stagnation := mustStagnation(t, ambient+ordinaryGauge, 310, 287.0, 1.4)
	request := mustRequest(
		t,
		stagnation,
		mustPressure(t, ambient),
		mustPrescribed(t, 0.0001),
		mustDischarge(t, 1),
	)
	result := mustEvaluate(t, request)
	if result.Regime() != RegimeSubsonic {
		t.Fatalf("ordinary pressure choked: %#v", result)
	}
	if result.CriticalPressureRatio() >= result.BackPressureRatio() {
		t.Fatalf("ordinary back ratio %g reached critical %g",
			result.BackPressureRatio(), result.CriticalPressureRatio())
	}
	assertNear(t, "gamma 1.4 critical ratio", result.CriticalPressureRatio(), 0.52828, 5e-6)
}

func TestConstructorsRejectInvalidValues(t *testing.T) {
	tests := []struct {
		name string
		call func() error
		want error
	}{
		{name: "pressure zero", call: func() error { _, err := NewPressure(0); return err }, want: ErrNonPositiveValue},
		{name: "temperature NaN", call: func() error { _, err := NewTemperature(math.NaN()); return err }, want: ErrNonFiniteValue},
		{name: "area negative", call: func() error { _, err := NewArea(-1); return err }, want: ErrNegativeArea},
		{name: "area infinity", call: func() error { _, err := NewArea(math.Inf(1)); return err }, want: ErrNonFiniteValue},
		{name: "gas constant zero", call: func() error { _, err := NewSpecificGasConstant(0); return err }, want: ErrNonPositiveValue},
		{name: "gamma one", call: func() error { _, err := NewHeatCapacityRatio(1); return err }, want: ErrInvalidHeatCapacityRatio},
		{name: "gamma below one", call: func() error { _, err := NewHeatCapacityRatio(0.5); return err }, want: ErrInvalidHeatCapacityRatio},
		{name: "Cd zero", call: func() error { _, err := NewDischargeCoefficient(0); return err }, want: ErrInvalidDischargeCoefficient},
		{name: "Cd above one", call: func() error { _, err := NewDischargeCoefficient(1.1); return err }, want: ErrInvalidDischargeCoefficient},
		{name: "compliance negative", call: func() error { _, err := NewAreaCompliance(-1); return err }, want: ErrNegativeCompliance},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.call(); !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
	if _, err := NewPrescribedArea(Area{squareMetres: -1}); !errors.Is(err, ErrNegativeArea) {
		t.Fatalf("forged prescribed area = %v", err)
	}
	if _, err := NewLinearComplianceArea(mustArea(t, 0.2), mustCompliance(t, 0), mustArea(t, 0.1)); !errors.Is(err, ErrInvalidAreaLaw) {
		t.Fatalf("maximum below prescribed = %v", err)
	}
	if _, err := NewStagnation(Pressure{}, mustTemperature(t, 1), mustGasConstant(t, 1), mustGamma(t, 1.4)); !errors.Is(err, ErrInvalidStagnation) {
		t.Fatalf("forged stagnation = %v", err)
	}
	if _, err := NewRequest(Stagnation{}, mustPressure(t, 1), mustPrescribed(t, 0), mustDischarge(t, 1)); !errors.Is(err, ErrInvalidStagnation) {
		t.Fatalf("forged request = %v", err)
	}
	if _, err := Evaluate(Request{}); !errors.Is(err, ErrInvalidStagnation) {
		t.Fatalf("forged evaluate = %v", err)
	}
}

func TestRegimeAndLawNames(t *testing.T) {
	if Regime(0).String() != "unsupported" || RegimeNoFlow.String() != "no-flow" ||
		RegimeSubsonic.String() != "subsonic" || RegimeChoked.String() != "choked" {
		t.Fatal("regime names changed")
	}
	if mustPrescribed(t, 0.01).LawName() != "prescribed" {
		t.Fatal("prescribed law name changed")
	}
}

func TestForgedLawsAndUnrepresentableStates(t *testing.T) {
	if _, err := NewAreaCompliance(math.NaN()); !errors.Is(err, ErrNonFiniteValue) {
		t.Fatalf("NaN compliance = %v", err)
	}
	if _, err := NewDischargeCoefficient(math.NaN()); !errors.Is(err, ErrNonFiniteValue) {
		t.Fatalf("NaN Cd = %v", err)
	}
	if _, err := NewHeatCapacityRatio(math.Inf(1)); !errors.Is(err, ErrNonFiniteValue) {
		t.Fatalf("Inf gamma = %v", err)
	}
	if _, err := NewLinearComplianceArea(mustArea(t, 0.1), AreaCompliance{squareMetresPerPascal: math.NaN()}, mustArea(t, 0.2)); err == nil {
		t.Fatal("NaN compliance law was accepted")
	}
	law := mustPrescribed(t, 0.01)
	if _, err := law.Effective(Pressure{pascals: math.NaN()}); !errors.Is(err, ErrNonFiniteValue) {
		t.Fatalf("NaN overpressure = %v", err)
	}
	if _, err := NewRequest(
		mustStagnation(t, 1000, 300, 200, 1.4),
		mustPressure(t, 500),
		AreaLaw{prescribed: Area{squareMetres: -1}, maximum: Area{squareMetres: -1}},
		mustDischarge(t, 1),
	); !errors.Is(err, ErrInvalidAreaLaw) {
		t.Fatalf("forged area law = %v", err)
	}
	if _, err := NewRequest(
		mustStagnation(t, 1000, 300, 200, 1.4),
		mustPressure(t, 500),
		mustPrescribed(t, 0.01),
		DischargeCoefficient{},
	); !errors.Is(err, ErrInvalidDischargeCoefficient) {
		t.Fatalf("forged Cd = %v", err)
	}
	if _, err := NewRequest(
		mustStagnation(t, 1000, 300, 200, 1.4),
		Pressure{},
		mustPrescribed(t, 0.01),
		mustDischarge(t, 1),
	); err == nil {
		t.Fatal("forged back pressure was accepted")
	}
}

func TestAccessorsRetainInputs(t *testing.T) {
	request := chokedGammaFifteenRequest(t)
	result := mustEvaluate(t, request)
	if result.Request().Stagnation().Pressure() != request.Stagnation().Pressure() ||
		result.Request().BackPressure() != request.BackPressure() ||
		result.Request().DischargeCoefficient() != request.DischargeCoefficient() ||
		result.Request().AreaLaw().Prescribed() != request.AreaLaw().Prescribed() ||
		result.Request().AreaLaw().Maximum() != request.AreaLaw().Maximum() ||
		result.Request().Stagnation().Temperature() != request.Stagnation().Temperature() ||
		result.Request().Stagnation().SpecificGasConstant() != request.Stagnation().SpecificGasConstant() ||
		result.Request().Stagnation().HeatCapacityRatio() != request.Stagnation().HeatCapacityRatio() ||
		result.Request().AreaLaw().Compliance().SquareMetresPerPascal() != 0 {
		t.Fatalf("accessors lost inputs: %#v", result)
	}
}

func FuzzEvaluate(f *testing.F) {
	f.Add(125000.0, 400.0, 200.0, 1.5, 50000.0, 0.01, 1.0, 0.0, 0.01)
	f.Add(1_000_000.0, 400.0, 200.0, 1.5, 729000.0, 0.01, 1.0, 0.0, 0.01)
	f.Add(102255.0, 310.0, 287.0, 1.4, 101325.0, 0.0001, 0.8, 1e-10, 0.001)
	f.Fuzz(func(
		t *testing.T,
		p0, t0, gasR, gamma, pb, area, cd, compliance, maximum float64,
	) {
		values := []float64{p0, t0, gasR, gamma, pb, area, cd, maximum}
		for _, value := range values {
			if !finite(value) || value <= 1e-12 || value >= 1e8 {
				t.Skip()
			}
		}
		if gamma <= 1.01 || gamma >= 2.5 || cd > 1 || area > maximum || compliance < 0 || compliance >= 1e-3 {
			t.Skip()
		}
		if !finite(compliance) {
			t.Skip()
		}
		stagnation, err := NewStagnation(
			mustPressure(t, p0), mustTemperature(t, t0), mustGasConstant(t, gasR), mustGamma(t, gamma),
		)
		if err != nil {
			t.Skip()
		}
		var law AreaLaw
		if compliance == 0 {
			law = mustPrescribed(t, area)
		} else {
			law, err = NewLinearComplianceArea(mustArea(t, area), mustCompliance(t, compliance), mustArea(t, maximum))
			if err != nil {
				t.Skip()
			}
		}
		request, err := NewRequest(stagnation, mustPressure(t, pb), law, mustDischarge(t, cd))
		if err != nil {
			t.Skip()
		}
		result, err := Evaluate(request)
		if err != nil {
			if errors.Is(err, ErrAdversePressure) || errors.Is(err, ErrNoRepresentableFlow) {
				return
			}
			t.Fatalf("unexpected error: %v", err)
		}
		if result.MassFlow().KilogramsPerSecond() < 0 || result.ExitSpeed().MetresPerSecond() < 0 ||
			result.SonicMassFlow().KilogramsPerSecond() < 0 {
			t.Fatalf("negative flow: %#v", result)
		}
		if math.Abs(result.Recoil().Newtons()+result.Thrust().Newtons()) > 1e-9*(1+math.Abs(result.Thrust().Newtons())) {
			t.Fatalf("recoil did not cancel thrust: %#v", result)
		}
		switch result.Regime() {
		case RegimeNoFlow:
			if result.MassFlow().KilogramsPerSecond() != 0 || result.Thrust().Newtons() != 0 {
				t.Fatalf("no-flow leaked a transfer: %#v", result)
			}
		case RegimeSubsonic:
			if result.ThroatMach().Value() >= 1 ||
				math.Abs(result.ExitPressure().Pascals()-pb) > 1e-9*(1+math.Abs(pb)) {
				t.Fatalf("subsonic invariants failed: %#v", result)
			}
		case RegimeChoked:
			if math.Abs(result.ThroatMach().Value()-1) > 1e-12 {
				t.Fatalf("choked Mach = %g", result.ThroatMach().Value())
			}
		default:
			t.Fatalf("unsupported regime: %#v", result)
		}
	})
}

func chokedGammaFifteenRequest(t *testing.T) Request {
	t.Helper()
	return mustRequest(
		t,
		mustStagnation(t, 125000, 400, 200, 1.5),
		mustPressure(t, 50000),
		mustPrescribed(t, 0.01),
		mustDischarge(t, 1),
	)
}

func mustEvaluate(t *testing.T, request Request) Result {
	t.Helper()
	result, err := Evaluate(request)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	return result
}

func mustRequest(
	t *testing.T,
	stagnation Stagnation,
	back Pressure,
	area AreaLaw,
	cd DischargeCoefficient,
) Request {
	t.Helper()
	request, err := NewRequest(stagnation, back, area, cd)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	return request
}

func mustStagnation(t *testing.T, pressure, temperature, gasConstant, gamma float64) Stagnation {
	t.Helper()
	state, err := NewStagnation(
		mustPressure(t, pressure),
		mustTemperature(t, temperature),
		mustGasConstant(t, gasConstant),
		mustGamma(t, gamma),
	)
	if err != nil {
		t.Fatalf("NewStagnation: %v", err)
	}
	return state
}

func mustPrescribed(t *testing.T, area float64) AreaLaw {
	t.Helper()
	law, err := NewPrescribedArea(mustArea(t, area))
	if err != nil {
		t.Fatalf("NewPrescribedArea: %v", err)
	}
	return law
}

func mustPressure(t *testing.T, pascals float64) Pressure {
	t.Helper()
	value, err := NewPressure(pascals)
	if err != nil {
		t.Fatalf("NewPressure: %v", err)
	}
	return value
}

func mustTemperature(t *testing.T, kelvin float64) Temperature {
	t.Helper()
	value, err := NewTemperature(kelvin)
	if err != nil {
		t.Fatalf("NewTemperature: %v", err)
	}
	return value
}

func mustArea(t *testing.T, squareMetres float64) Area {
	t.Helper()
	value, err := NewArea(squareMetres)
	if err != nil {
		t.Fatalf("NewArea: %v", err)
	}
	return value
}

func mustGasConstant(t *testing.T, value float64) SpecificGasConstant {
	t.Helper()
	result, err := NewSpecificGasConstant(value)
	if err != nil {
		t.Fatalf("NewSpecificGasConstant: %v", err)
	}
	return result
}

func mustGamma(t *testing.T, value float64) HeatCapacityRatio {
	t.Helper()
	result, err := NewHeatCapacityRatio(value)
	if err != nil {
		t.Fatalf("NewHeatCapacityRatio: %v", err)
	}
	return result
}

func mustDischarge(t *testing.T, value float64) DischargeCoefficient {
	t.Helper()
	result, err := NewDischargeCoefficient(value)
	if err != nil {
		t.Fatalf("NewDischargeCoefficient: %v", err)
	}
	return result
}

func mustCompliance(t *testing.T, value float64) AreaCompliance {
	t.Helper()
	result, err := NewAreaCompliance(value)
	if err != nil {
		t.Fatalf("NewAreaCompliance: %v", err)
	}
	return result
}

func assertNear(t *testing.T, name string, got, want, tolerance float64) {
	t.Helper()
	if math.IsNaN(got) || math.IsInf(got, 0) || math.Abs(got-want) > tolerance {
		t.Fatalf("%s = %g, want %g ± %g", name, got, want, tolerance)
	}
}
