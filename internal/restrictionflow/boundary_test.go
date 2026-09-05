package restrictionflow

import (
	"errors"
	"math"
	"testing"
)

func TestSmallPositivePressureDifferenceProducesSubsonicFlow(t *testing.T) {
	p0 := 125000.0
	pb := math.Nextafter(p0, 0)
	result := mustEvaluate(t, mustRequest(t, mustStagnation(t, p0, 400, 200, 1.5),
		mustPressure(t, pb), mustPrescribed(t, 0.01), mustDischarge(t, 1)))
	if result.Regime() != RegimeSubsonic || result.MassFlow().KilogramsPerSecond() <= 0 || result.ThroatMach().Value() <= 0 {
		t.Fatalf("representable positive flow was erased: %#v", result)
	}
	// The linear pressure limit has corrections far below binary64 here.
	want := 0.01 * math.Sqrt(2*(p0/(200*400))*(p0-pb))
	assertNear(t, "small-gradient mass flow", result.MassFlow().KilogramsPerSecond(), want, want*1e-14)
}

func TestHeatCapacityRatioApproachesIsothermalLimit(t *testing.T) {
	state := mustStagnation(t, 125000, 400, 200, math.Nextafter(1, 2))
	assertNear(t, "critical ratio limit", state.CriticalPressureRatio(), math.Exp(-0.5), 2e-16)
	result := mustEvaluate(t, mustRequest(t, state, mustPressure(t, 100000), mustPrescribed(t, 0.01), mustDischarge(t, 1)))
	if result.Regime() != RegimeSubsonic || !finite(result.MassFlow().KilogramsPerSecond()) || result.MassFlow().KilogramsPerSecond() <= 0 {
		t.Fatalf("near-unit gamma lost a finite subsonic state: %#v", result)
	}
}

func TestUnderflowCannotMasqueradeAsNoFlow(t *testing.T) {
	request := mustRequest(t, mustStagnation(t, 125000, 400, 200, 1.5),
		mustPressure(t, 100000), mustPrescribed(t, math.SmallestNonzeroFloat64), mustDischarge(t, math.SmallestNonzeroFloat64))
	if _, err := Evaluate(request); !errors.Is(err, ErrNoRepresentableFlow) {
		t.Fatalf("unrepresentable positive mass flow: %v", err)
	}
}

func TestComplianceUnderflowIsNotAnAuthoredClosedArea(t *testing.T) {
	law, err := NewLinearComplianceArea(mustArea(t, 0),
		AreaCompliance{squareMetresPerPascal: math.SmallestNonzeroFloat64}, mustArea(t, 1))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := law.Effective(Pressure{pascals: 0.25}); !errors.Is(err, ErrNoRepresentableFlow) {
		t.Fatalf("positive opening underflow: %v", err)
	}
}

func TestChokedExtremeScalesRetainFiniteFlow(t *testing.T) {
	for _, test := range []struct {
		name        string
		temperature float64
		gasConstant float64
		exitTemp    float64
		exitSpeed   float64
	}{
		{
			name:        "throat temperature intermediate overflow",
			temperature: math.Ldexp(1, 1023), gasConstant: math.Ldexp(1, -1020),
			exitTemp: 0.8 * math.Ldexp(1, 1023), exitSpeed: math.Sqrt(9.6),
		},
		{
			name:        "unrepresentable density with finite flux",
			temperature: 400, gasConstant: math.Ldexp(1, -1020),
			exitTemp: 320, exitSpeed: math.Sqrt(480) * math.Ldexp(1, -510),
		},
		{
			name:        "squared speed overflow and density underflow",
			temperature: 400, gasConstant: math.Ldexp(1, 1023),
			exitTemp: 320, exitSpeed: math.Sqrt(960) * math.Ldexp(1, 511),
		},
		{
			name:        "subnormal thermal factor with normal squared speed",
			temperature: math.Ldexp(1, 1000), gasConstant: 17 * math.SmallestNonzeroFloat64,
			exitTemp: 0.8 * math.Ldexp(1, 1000), exitSpeed: math.Sqrt(20.4) * math.Ldexp(1, -37),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := mustEvaluate(t, mustRequest(t,
				mustStagnation(t, 125000, test.temperature, test.gasConstant, 1.5),
				mustPressure(t, 50000), mustPrescribed(t, 0.01), mustDischarge(t, 1)))
			if result.Regime() != RegimeChoked {
				t.Fatalf("regime = %s", result.Regime())
			}
			// At gamma=1.5, p*=64000 Pa and T*=0.8*T0. Consequently
			// mdot*v=gamma*p*area=960 N at every thermodynamic scale;
			// the independent pressure contribution is 140 N.
			massFlow := 960 / test.exitSpeed
			assertNear(t, "exit temperature", result.ExitTemperature().Kelvin(), test.exitTemp, test.exitTemp*4e-15)
			assertNear(t, "exit speed", result.ExitSpeed().MetresPerSecond(), test.exitSpeed, test.exitSpeed*4e-15)
			assertNear(t, "mass flow", result.MassFlow().KilogramsPerSecond(), massFlow, massFlow*4e-15)
			assertNear(t, "sonic mass flow", result.SonicMassFlow().KilogramsPerSecond(), massFlow, massFlow*4e-15)
			assertNear(t, "thrust", result.Thrust().Newtons(), 1100, 5e-12)
			assertNear(t, "mass residual", result.MassFlowResidual().KilogramsPerSecond(), 0, massFlow*4e-15)
			assertNear(t, "thrust residual", result.ThrustResidual().Newtons(), 0, 5e-12)
		})
	}
}

func TestSubnormalSonicIntermediateDoesNotChangeChokedMassFlow(t *testing.T) {
	stagnation := mustStagnation(t, 140625, 400, 200, 1.5)
	area := 1e300
	cd := math.SmallestNonzeroFloat64
	result := mustEvaluate(t, mustRequest(t, stagnation, stagnation.CriticalPressure(),
		mustPrescribed(t, area), mustDischarge(t, cd)))
	// p*=72000 Pa, T*=320 K, and rho*=9/8 kg/m^3. Multiplying
	// Cd*rho first loses 1/9 of the answer before the large area rescales it.
	want := (cd * area) * (9.0 / 8.0) * math.Sqrt(96000)
	assertNear(t, "mass flow", result.MassFlow().KilogramsPerSecond(), want, want*4e-15)
	assertNear(t, "sonic mass flow", result.SonicMassFlow().KilogramsPerSecond(), want, want*4e-15)
}

func TestSubnormalDensityDoesNotContaminateRepresentableFlux(t *testing.T) {
	result := mustEvaluate(t, mustRequest(t, mustStagnation(t, 1.25e-15, 1, 1e308, 1.5),
		mustPressure(t, 5e-16), mustPrescribed(t, 1e300), mustDischarge(t, 1)))
	// At the same gamma=1.5 choked ratios, momentum is 9.6e284 N
	// and pressure thrust is 1.4e284 N. The density itself is subnormal.
	wantSpeed := math.Sqrt(1.2e308)
	wantMass := 9.6e284 / wantSpeed
	assertNear(t, "exit speed", result.ExitSpeed().MetresPerSecond(), wantSpeed, wantSpeed*4e-15)
	assertNear(t, "mass flow", result.MassFlow().KilogramsPerSecond(), wantMass, wantMass*4e-15)
	assertNear(t, "thrust", result.Thrust().Newtons(), 1.1e285, 1.1e285*4e-15)
}

func TestRoundedPressureRatioCannotCreateAdverseExitPressure(t *testing.T) {
	stagnation := mustStagnation(t, 125000, 400, 200, 1.5)
	back := mustPressure(t, math.Nextafter(64000, math.Inf(1)))
	result := mustEvaluate(t, mustRequest(t, stagnation, back,
		mustPrescribed(t, 1), mustDischarge(t, 1e-30)))
	if result.BackPressureRatio() != result.CriticalPressureRatio() {
		t.Fatal("fixture no longer has indistinguishable rounded pressure ratios")
	}
	if result.Regime() != RegimeSubsonic || result.ExitPressure().Pascals() != back.Pascals() {
		t.Fatalf("rounded sonic boundary invented an adverse exit state: %#v", result)
	}
	// The physical pressure term is exactly zero on this subsonic branch.
	// A spurious one-ulp pressure difference would dominate the tiny but
	// representable forward momentum force and reverse its sign.
	wantMass := 1e-30 * math.Sqrt(96000)
	assertNear(t, "mass flow", result.MassFlow().KilogramsPerSecond(), wantMass, wantMass*4e-15)
	assertNear(t, "forward thrust", result.Thrust().Newtons(), 96000e-30, 96000e-30*4e-15)
	if result.Thrust().Newtons() <= 0 || result.Recoil().Newtons() >= 0 {
		t.Fatal("a rounded sonic boundary reversed the pressure-driven force")
	}
}

func TestRepresentableMachSurvivesUnrepresentableSquare(t *testing.T) {
	p0 := 125000.0
	pb := math.Nextafter(p0, 0)
	for _, test := range []struct {
		name  string
		gamma float64
	}{
		{"subnormal square", 1e307},
		{"lost square", 1e308},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := mustEvaluate(t, mustRequest(t, mustStagnation(t, p0, 400, 200, test.gamma),
				mustPressure(t, pb), mustPrescribed(t, 0.01), mustDischarge(t, 1)))
			if result.Regime() != RegimeSubsonic || result.ThroatMach().Value() <= 0 {
				t.Fatalf("small positive Mach was erased: %#v", result)
			}
			square := result.ThroatMach().Value() * result.ThroatMach().Value()
			if square >= math.Ldexp(1, -1022) || test.gamma == 1e308 && square != 0 {
				t.Fatal("fixture no longer exercises lost precision in the Mach square")
			}
			// The small-pressure-gap limit has higher-order corrections below
			// binary64 roundoff; it does not require computing gamma*M*M.
			wantSpeed := math.Sqrt(2 * (200 * 400) / p0 * (p0 - pb))
			wantMass := 0.01 * math.Sqrt(2*(p0/(200*400))*(p0-pb))
			assertNear(t, "exit speed", result.ExitSpeed().MetresPerSecond(), wantSpeed, wantSpeed*2e-14)
			assertNear(t, "mass flow", result.MassFlow().KilogramsPerSecond(), wantMass, wantMass*2e-14)
			assertNear(t, "thrust", result.Thrust().Newtons(), 0.02*(p0-pb), 0.02*(p0-pb)*4e-14)
		})
	}
}
