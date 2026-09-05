package coupledblowdown

import (
	"errors"
	"math"
	"testing"

	"github.com/blisspixel/fartapp/internal/idealmixturereservoir"
	"github.com/blisspixel/fartapp/internal/restrictionflow"
)

func TestAccurateEqualizationMatchesIndependentReferences(t *testing.T) {
	for _, closure := range []idealmixturereservoir.Closure{idealmixturereservoir.RigidIsothermal, idealmixturereservoir.RigidAdiabatic} {
		for _, gamma := range []float64{1.4, 1.5} {
			t.Run(closure.String()+"/"+referenceGammaName(gamma), func(t *testing.T) {
				config := referenceConfig(t, closure, gamma, 0.4, 0)
				result, evidence := mustSimulateAccurate(t, config, 1e-10)
				x := math.Pow(referenceBack/referencePressure, 1/referenceExponent(closure, gamma))
				want := referenceGamma15Time(x, closure)
				if gamma == 1.4 {
					want = referenceGamma14PublishedTime(x, closure, referenceBack/referencePressure)
				}
				referenceNear(t, "accurate complete time", result.ElapsedSeconds(), want, 2e-11*want)
				referenceEndpoint(t, result, x, referenceExponent(closure, gamma), gamma, 2e-13)
				if result.Stop() != StopEqualized || !evidence.DischargeComplete || !evidence.ToleranceSatisfied {
					t.Fatalf("completion/evidence = %s, %+v", result.Stop(), evidence)
				}
				if result.Steps() > 3 || len(result.Samples()) != result.Steps()+1 {
					t.Fatalf("coarse retained sampling changed: steps=%d samples=%d", result.Steps(), len(result.Samples()))
				}
				if evidence.Evaluations > 5000 || evidence.AcceptedIntervals < result.Steps() {
					t.Fatalf("work evidence = %+v", evidence)
				}
				gap := result.Final().Pressure().Pascals() - config.BackPressure().Pascals()
				if gap < 0 || gap > result.EqualizationPressureTolerancePascals() {
					t.Fatalf("invalid represented endpoint gap %.17g", gap)
				}
				t.Logf("time=%.16g reference=%.16g error=%.3g estimate=%.3g evaluations=%d steps=%d", result.ElapsedSeconds(), want,
					math.Abs(result.ElapsedSeconds()-want), evidence.EstimatedTimeErrorSeconds, evidence.Evaluations, result.Steps())
			})
		}
	}
}

func TestAccurateToleranceRefinementAgainstPublishedTime(t *testing.T) {
	for _, closure := range []idealmixturereservoir.Closure{idealmixturereservoir.RigidIsothermal, idealmixturereservoir.RigidAdiabatic} {
		t.Run(closure.String(), func(t *testing.T) {
			config := referenceConfig(t, closure, 1.4, 0.9999, 0)
			config.back = mustPressure(t, referencePressure*0.001)
			x := math.Pow(0.001, 1/referenceExponent(closure, 1.4))
			want := referenceGamma14PublishedTime(x, closure, 0.001)
			previousError, previousWork := math.Inf(1), 0
			for _, tolerance := range []float64{1e-2, 1e-4, 1e-7, 1e-10} {
				result, evidence := mustSimulateAccurate(t, config, tolerance)
				actualError := math.Abs(result.ElapsedSeconds()/want - 1)
				if actualError > previousError && actualError > 2e-13 {
					t.Fatalf("time error increased: tolerance=%g error=%g prior=%g", tolerance, actualError, previousError)
				}
				if evidence.Evaluations < previousWork || actualError > tolerance {
					t.Fatalf("refinement evidence failed: tolerance=%g error=%g evidence=%+v", tolerance, actualError, evidence)
				}
				previousError, previousWork = actualError, evidence.Evaluations
				t.Logf("tolerance=%g relative_error=%.3g evaluations=%d refinements=%d", tolerance, actualError, evidence.Evaluations, evidence.Refinements)
			}
		})
	}
}

func TestAccurateTruncationMatchesChokedTimeImpulseAndStroke(t *testing.T) {
	for _, closure := range []idealmixturereservoir.Closure{idealmixturereservoir.RigidIsothermal, idealmixturereservoir.RigidAdiabatic} {
		config := referenceConfig(t, closure, 1.5, 0.1, 0)
		config.maxSteps = 1
		result, evidence := mustSimulateAccurate(t, config, 1e-10)
		x, n := 0.9, referenceExponent(closure, 1.5)
		wantTime := referenceChokedTime(x, n, 1.5)
		k := referenceChokedRateConstant(1.5)
		pressureCoefficient := referenceArea * referencePressure * referenceCriticalRatio(1.5) * 2.5
		wantImpulse := pressureCoefficient*2*(1-math.Pow(x, (n+1)/2))/(k*(n+1)) - referenceArea*referenceBack*wantTime
		wantStroke := -math.Sqrt(2*1.5*referenceGasConstant*referenceTemperature/2.5) / k * math.Log(x)
		referenceNear(t, "choked exact clock", result.ElapsedSeconds(), wantTime, 1e-11*wantTime)
		referenceNear(t, "choked exact impulse", result.ImpulseNewtonSeconds(), wantImpulse, 1e-11*wantImpulse)
		referenceNear(t, "choked exact recoil", result.RecoilImpulseNewtonSeconds(), -wantImpulse, 1e-11*wantImpulse)
		referenceNear(t, "choked exact stroke", result.Signature().StrokeLengthMetres(), wantStroke, 1e-11*wantStroke)
		if result.Stop() != StopMaxSteps || evidence.DischargeComplete || !evidence.ToleranceSatisfied || result.Steps() != 1 || len(result.Samples()) != 2 {
			t.Fatalf("truncated result/evidence = %s, %+v", result.Stop(), evidence)
		}
	}
}

func TestAccurateComplianceCapsMatchIndependentPrimitive(t *testing.T) {
	for _, closure := range []idealmixturereservoir.Closure{idealmixturereservoir.RigidIsothermal, idealmixturereservoir.RigidAdiabatic} {
		for _, capRatio := range []float64{1, 1.5, 2.2, 3} {
			config := referenceConfig(t, closure, 1.5, 0.6, 0)
			const restArea = referenceArea / 10
			config.area = accurateCompliance(t, restArea, restArea/referenceBack, capRatio*restArea)
			result, evidence := mustSimulateAccurate(t, config, 1e-10)
			want := accurateComplianceReferenceTime(closure, restArea, capRatio)
			referenceNear(t, "capped compliant exact time", result.ElapsedSeconds(), want, 3e-11*want)
			if !evidence.DischargeComplete || result.Samples()[0].EffectiveArea() != math.Min(capRatio, 2.5)*restArea {
				t.Fatalf("compliance cap evidence: %+v", evidence)
			}
		}
	}
}

// Independent gamma=3/2 reference with A0=C*Pb. Then A=A0*P/Pb
// until capped. In the isothermal subsonic coordinate
// s=sqrt(1-(r0/x)^(1/3)), dt/ds=B0*sqrt(6) is constant.
// In the adiabatic coordinate w=sqrt(sqrt(x)-b), b=r0^(1/3),
// dt/dw=4*B0*r0^(1/3)/(sqrt(6)*(w*w+b)^2), whose primitive
// is elementary. Choked uncapped flow has dx/dt=-k0*x^((3*n+1)/2)/r0.
func accurateComplianceReferenceTime(closure idealmixturereservoir.Closure, restArea, capRatio float64) float64 {
	const r0 = referenceBack / referencePressure
	n := referenceExponent(closure, 1.5)
	xCap := math.Min(1, math.Pow(capRatio*r0, 1/n))
	xCritical := math.Pow(r0/referenceCriticalRatio(1.5), 1/n)
	time := referenceGamma15Time(xCap, closure) * referenceArea / (restArea * capRatio)
	if xCap >= 1 {
		time = 0
	}
	if xCap > xCritical {
		k := referenceChokedRateConstant(1.5) * restArea / referenceArea
		exponent := (3*n - 1) / 2
		time += r0 / (k * exponent) * (math.Pow(xCritical, -exponent) - math.Pow(xCap, -exponent))
	}
	xStart := math.Min(xCap, xCritical)
	b0 := referenceVolume / (restArea * math.Sqrt(referenceGasConstant*referenceTemperature))
	if closure == idealmixturereservoir.RigidIsothermal {
		return time + b0*math.Sqrt(6)*math.Sqrt(1-math.Pow(r0/xStart, 1.0/3))
	}
	b := math.Pow(r0, 1.0/3)
	w := math.Sqrt(math.Sqrt(xStart) - b)
	primitive := w/(2*b*(w*w+b)) + math.Atan(w/math.Sqrt(b))/(2*math.Pow(b, 1.5))
	return time + 4*b0*b/math.Sqrt(6)*primitive
}

func TestAccurateAreaScalingAndComponentWitness(t *testing.T) {
	for _, closure := range []idealmixturereservoir.Closure{idealmixturereservoir.RigidIsothermal, idealmixturereservoir.RigidAdiabatic} {
		config := referenceConfig(t, closure, 1.5, 0.1, 0)
		var components []idealmixturereservoir.Component
		for _, r := range []float64{100, 300} {
			mass, _ := idealmixturereservoir.NewMass(referenceMass / 2)
			gas, _ := idealmixturereservoir.NewSpecificGasConstant(r)
			cv, _ := idealmixturereservoir.NewIsochoricHeatCapacity(2 * r)
			component, err := idealmixturereservoir.NewComponent(mass, gas, cv)
			if err != nil {
				t.Fatal(err)
			}
			components = append(components, component)
		}
		state, err := idealmixturereservoir.NewState(components, config.reservoir.Volume(), config.reservoir.Temperature())
		if err != nil {
			t.Fatal(err)
		}
		config.reservoir = state
		config.area = accurateCompliance(t, 0.001, 2e-8, 0.0015)
		baseline, _ := mustSimulateAccurate(t, config, 1e-10)
		config.area = accurateCompliance(t, 0.002, 4e-8, 0.003)
		branch, _ := mustSimulateAccurate(t, config, 1e-10)
		referenceNear(t, "doubled area clock", branch.ElapsedSeconds(), baseline.ElapsedSeconds()/2, 2e-13*baseline.ElapsedSeconds())
		referenceNear(t, "doubled area impulse", branch.ImpulseNewtonSeconds(), baseline.ImpulseNewtonSeconds(), 2e-13*baseline.ImpulseNewtonSeconds())
		referenceNear(t, "doubled area stroke", branch.Signature().StrokeLengthMetres(), baseline.Signature().StrokeLengthMetres()/2, 2e-13*baseline.Signature().StrokeLengthMetres())
		for _, result := range []Result{baseline, branch} {
			for index, sample := range result.Samples() {
				masses, out := sample.ComponentMassesKilograms(), sample.ComponentMassOutKilograms()
				for component := range masses {
					referenceNear(t, "component closure", masses[component]+out[component], referenceMass/2, 2e-15*referenceMass)
					referenceNear(t, "constant mass fraction", masses[component]/sample.Mass(), 0.5, 2e-15)
				}
				if index > 0 && (sample.Time() <= result.Samples()[index-1].Time() || sample.Mass() >= result.Samples()[index-1].Mass()) {
					t.Fatal("history is not strictly advancing")
				}
			}
			referenceNear(t, "mixture energy ledger", result.Ledgers().EnergyResidualJoules(), 0, 2e-10)
		}
	}
}

func TestAccurateNoFlowAndUnsupportedScope(t *testing.T) {
	base := referenceConfig(t, idealmixturereservoir.RigidIsothermal, 1.5, 0.1, 0)
	for _, name := range []string{"zero-area", "equal-pressure"} {
		config := base
		config.maxTime = 1
		if name == "zero-area" {
			config.area = mustArea(t, 0)
		} else {
			config.back = mustPressure(t, referencePressure)
		}
		result, evidence := mustSimulateAccurate(t, config, 1e-8)
		if result.Stop() != StopNoFlow || result.ElapsedSeconds() != 0 || result.MassOutKilograms() != 0 || result.Steps() != 0 ||
			evidence.Evaluations != 0 || !evidence.ToleranceSatisfied || evidence.DischargeComplete {
			t.Fatalf("%s is not an exact no-flow identity: %+v", name, evidence)
		}
	}
	for _, config := range []Config{
		func() Config { c := base; c.maxTime = 1; return c }(),
		func() Config { c := base; c.area = accurateCompliance(t, 0, 1e-8, referenceArea); return c }(),
	} {
		_, evidence, err := SimulateAccurate(config, accurateTestOptions(1e-8))
		if !errors.Is(err, ErrUnsupportedAccuracyConfig) || evidence.ToleranceSatisfied || evidence.Evaluations != 0 {
			t.Fatalf("unsupported scope = %+v, %v", evidence, err)
		}
	}
	base.back = mustPressure(t, 2*referencePressure)
	if _, evidence, err := SimulateAccurate(base, accurateTestOptions(1e-8)); !errors.Is(err, restrictionflow.ErrAdversePressure) || evidence.ToleranceSatisfied {
		t.Fatalf("config was not revalidated: %+v, %v", evidence, err)
	}
}

func TestAccurateOptionsAndWorkBudgetRefusals(t *testing.T) {
	config := referenceConfig(t, idealmixturereservoir.RigidAdiabatic, 1.5, 0.4, 0)
	valid := accurateTestOptions(1e-8)
	for _, invalid := range []AccuracyOptions{
		{}, {RelativeTolerance: math.NaN(), MaxEvaluations: 100}, {RelativeTolerance: math.Inf(1), MaxEvaluations: 100},
		{RelativeTolerance: 1e-13, MaxEvaluations: 100}, {RelativeTolerance: 0.11, MaxEvaluations: 100},
		{RelativeTolerance: 1e-8, AbsoluteTimeToleranceSeconds: -1, MaxEvaluations: 100},
		{RelativeTolerance: 1e-8, AbsoluteTimeToleranceSeconds: math.NaN(), MaxEvaluations: 100},
		{RelativeTolerance: 1e-8, AbsoluteTimeToleranceSeconds: math.Inf(1), MaxEvaluations: 100},
		{RelativeTolerance: 1e-8, MaxEvaluations: 14}, {RelativeTolerance: 1e-8, MaxEvaluations: MaxAccuracyEvaluations + 1},
	} {
		if _, evidence, err := SimulateAccurate(config, invalid); !errors.Is(err, ErrInvalidAccuracyOptions) || evidence.ToleranceSatisfied || evidence.Evaluations != 0 {
			t.Fatalf("invalid options accepted: %+v, %+v, %v", invalid, evidence, err)
		}
	}
	for _, budget := range []int{15, 29, 30, 44} {
		options := valid
		options.MaxEvaluations = budget
		_, evidence, err := SimulateAccurate(config, options)
		if !errors.Is(err, ErrAccuracyBudgetExhausted) || evidence.ToleranceSatisfied || evidence.DischargeComplete || evidence.Evaluations > budget {
			t.Fatalf("budget %d: %+v, %v", budget, evidence, err)
		}
	}
	valid.AbsoluteTimeToleranceSeconds = 1e-6
	result, evidence, err := SimulateAccurate(config, valid)
	if err != nil {
		t.Fatal(err)
	}
	referenceNear(t, "combined request", evidence.RequestedTimeToleranceSeconds, 1e-6+valid.RelativeTolerance*result.ElapsedSeconds(), 1e-20)
}

func TestAccurateExtremeScalesRemainFiniteOrRefuse(t *testing.T) {
	base := referenceConfig(t, idealmixturereservoir.RigidIsothermal, 1.5, 0.4, 0)
	wantTime := referenceGamma15Time(referenceBack/referencePressure, base.closure)
	for _, scales := range [][2]float64{{1e-100, 1e-200}, {1e100, 1e100}, {1e-100, 1e-100}, {1e100, 1e50}} {
		config := base
		mass, _ := idealmixturereservoir.NewMass(referenceMass * scales[0])
		component := base.reservoir.Components()[0]
		component, err := idealmixturereservoir.NewComponent(mass, component.SpecificGasConstant(), component.IsochoricHeatCapacity())
		if err != nil {
			t.Fatal(err)
		}
		volume, _ := idealmixturereservoir.NewVolume(referenceVolume * scales[0])
		config.reservoir, err = idealmixturereservoir.NewState([]idealmixturereservoir.Component{component}, volume, base.reservoir.Temperature())
		if err != nil {
			t.Fatal(err)
		}
		config.area = mustArea(t, referenceArea*scales[1])
		result, _ := mustSimulateAccurate(t, config, 1e-9)
		expected := wantTime * (scales[0] / scales[1])
		referenceNear(t, "scaled exact clock", result.ElapsedSeconds(), expected, 1e-10*expected)
	}
	near := base
	near.back = mustPressure(t, math.Nextafter(referencePressure, 0))
	if _, evidence, err := SimulateAccurate(near, accurateTestOptions(1e-8)); !errors.Is(err, ErrAccuracyNotAchieved) || evidence.ToleranceSatisfied {
		t.Fatalf("roundoff-dominated endpoint must refuse: %+v, %v", evidence, err)
	}
	base.maxFraction = math.SmallestNonzeroFloat64
	if _, evidence, err := SimulateAccurate(base, accurateTestOptions(1e-8)); !errors.Is(err, ErrAccuracyNotAchieved) || evidence.ToleranceSatisfied {
		t.Fatalf("unrepresentable sampling step must refuse: %+v, %v", evidence, err)
	}
}

func TestAccurateOrdinaryLowPressureClockAndSampling(t *testing.T) {
	// The ordinary walkthrough case, checked directly against the published
	// initially-subsonic isothermal primitive with a stable pressure coordinate.
	mass, _ := idealmixturereservoir.NewMass(0.000121538188165196)
	gas, _ := idealmixturereservoir.NewSpecificGasConstant(287)
	cv, _ := idealmixturereservoir.NewIsochoricHeatCapacity(717.5)
	component, _ := idealmixturereservoir.NewComponent(mass, gas, cv)
	volume, _ := idealmixturereservoir.NewVolume(0.0001)
	temperature, _ := idealmixturereservoir.NewTemperature(293.15)
	state, err := idealmixturereservoir.NewState([]idealmixturereservoir.Component{component}, volume, temperature)
	if err != nil {
		t.Fatal(err)
	}
	config, err := NewConfig(state, idealmixturereservoir.RigidIsothermal, mustPressure(t, 101325),
		mustArea(t, 1e-6), mustDischarge(t, 0.8), 5e-5, MaxSteps, 0)
	if err != nil {
		t.Fatal(err)
	}
	chi := math.Sqrt(math.Expm1(2.0 / 7 * math.Log1p((state.Pressure().Pascals()-101325)/101325)))
	tchar := 0.0001 / (0.8 * 1e-6 * math.Sqrt(1.4*287*293.15))
	want := tchar * math.Sqrt(2*1.4*1.4/0.4) * (math.Pow(chi, 5)/5 + 2*chi*chi*chi/3 + chi)
	var previous Result
	for _, fraction := range []float64{5e-5, 0.01} {
		config.maxFraction = fraction
		result, evidence := mustSimulateAccurate(t, config, 1e-10)
		referenceNear(t, "ordinary exact clock", result.ElapsedSeconds(), want, 2e-12*want)
		if fraction == 5e-5 && result.Steps() != 183 {
			t.Fatalf("ordinary retained history has %d steps", result.Steps())
		}
		if previous.Steps() > 0 {
			referenceNear(t, "sampling-independent clock", result.ElapsedSeconds(), previous.ElapsedSeconds(), 2e-12*want)
			referenceNear(t, "sampling-independent enthalpy", result.EnthalpyOutJoules(), previous.EnthalpyOutJoules(), 1e-14)
		}
		previous = result
		t.Logf("fraction=%g time=%.16g steps=%d evaluations=%d error_estimate=%.3g", fraction, result.ElapsedSeconds(), result.Steps(), evidence.Evaluations, evidence.EstimatedTimeErrorSeconds)
	}
}

func TestAccurateSubnormalSourceRateHasRepresentableIntegral(t *testing.T) {
	mass, _ := idealmixturereservoir.NewMass(1e-171)
	gas, _ := idealmixturereservoir.NewSpecificGasConstant(2e-9)
	cv, _ := idealmixturereservoir.NewIsochoricHeatCapacity(4e-9)
	component, _ := idealmixturereservoir.NewComponent(mass, gas, cv)
	volume, _ := idealmixturereservoir.NewVolume(2e-171)
	temperature, _ := idealmixturereservoir.NewTemperature(1)
	state, err := idealmixturereservoir.NewState([]idealmixturereservoir.Component{component}, volume, temperature)
	if err != nil {
		t.Fatal(err)
	}
	config, err := NewConfig(state, idealmixturereservoir.RigidIsothermal, mustPressure(t, 4e-10),
		mustArea(t, 3.2e-319), mustDischarge(t, 1), 0.4, MaxSteps, 0)
	if err != nil {
		t.Fatal(err)
	}
	result, _ := mustSimulateAccurate(t, config, 1e-10)
	timeScale := (2e-171 / config.area.Prescribed().SquareMetres()) * referenceArea * math.Sqrt(referenceGasConstant*referenceTemperature/2e-9)
	want := referenceGamma15Time(0.4, config.closure) * timeScale
	referenceNear(t, "subnormal-rate analytical clock", result.ElapsedSeconds(), want, 2e-11*want)
	if result.ImpulseNewtonSeconds() <= 0 || result.EnthalpyOutJoules() <= 0 {
		t.Fatal("representable integrated transfers vanished")
	}
}

func TestAccurateHistoryAndAdaptiveWorkRemainBounded(t *testing.T) {
	config := referenceConfig(t, idealmixturereservoir.RigidIsothermal, 1.5, 1e-6, 0)
	result, evidence := mustSimulateAccurate(t, config, 1e-8)
	if result.Steps() != MaxSteps || len(result.Samples()) != MaxSamples || result.Stop() != StopMaxSteps || evidence.DischargeComplete {
		t.Fatalf("retained history escaped its budget: steps=%d samples=%d stop=%s evidence=%+v", result.Steps(), len(result.Samples()), result.Stop(), evidence)
	}
	// Extremely small positive rest area creates a very narrow tail. The
	// bounded method must refuse when it cannot resolve it, never call the
	// omitted time contribution a converged finite equalization result.
	config.maxFraction = 0.99
	config.area = accurateCompliance(t, 1e-200, 1e-8, referenceArea)
	options := accurateTestOptions(1e-12)
	options.MaxEvaluations = 1000
	_, evidence, err := SimulateAccurate(config, options)
	if (!errors.Is(err, ErrAccuracyBudgetExhausted) && !errors.Is(err, ErrAccuracyNotAchieved)) || evidence.ToleranceSatisfied ||
		evidence.DischargeComplete || evidence.Evaluations > options.MaxEvaluations {
		t.Fatalf("unresolved narrow tail was claimed complete: evidence=%+v err=%v", evidence, err)
	}
}

func TestAccurateLargeGammaChokedTruncation(t *testing.T) {
	// The typed model permits a very large finite cp/cv. The algebraic
	// factors (gamma+1)/(2*(gamma-1)) and 2*gamma/(gamma+1) must not first
	// overflow even though their limiting values are 1/2 and 2.
	mass, _ := idealmixturereservoir.NewMass(1)
	gas, _ := idealmixturereservoir.NewSpecificGasConstant(1e-15)
	cv, _ := idealmixturereservoir.NewIsochoricHeatCapacity(1e-323)
	component, _ := idealmixturereservoir.NewComponent(mass, gas, cv)
	volume, _ := idealmixturereservoir.NewVolume(1)
	temperature, _ := idealmixturereservoir.NewTemperature(1e307)
	state, err := idealmixturereservoir.NewState([]idealmixturereservoir.Component{component}, volume, temperature)
	if err != nil {
		t.Fatal(err)
	}
	config, err := NewConfig(state, idealmixturereservoir.RigidIsothermal, mustPressure(t, 1.5e-16),
		mustArea(t, 1e-146), mustDischarge(t, 1), 0.1, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	result, _ := mustSimulateAccurate(t, config, 1e-10)
	initial := result.Samples()[0]
	if initial.Regime() != restrictionflow.RegimeChoked || result.Stop() != StopMaxSteps {
		t.Fatal("large-gamma fixture was not a choked truncation")
	}
	// For prescribed isothermal choked discharge, the mass ODE is exactly
	// dm/dt=-q0*m/m0, independently of the value of gamma.
	want := -math.Log(0.9) / initial.MassFlow()
	referenceNear(t, "large-gamma exponential clock", result.ElapsedSeconds(), want, 2e-11*want)
}

func accurateTestOptions(tolerance float64) AccuracyOptions {
	return AccuracyOptions{RelativeTolerance: tolerance, MaxEvaluations: 100000}
}

func mustSimulateAccurate(t *testing.T, config Config, tolerance float64) (Result, AccuracyEvidence) {
	t.Helper()
	result, evidence, err := SimulateAccurate(config, accurateTestOptions(tolerance))
	if err != nil {
		t.Fatalf("accurate simulation failed: %v, evidence=%+v", err, evidence)
	}
	return result, evidence
}

func accurateCompliance(t *testing.T, rest, compliance, cap float64) restrictionflow.AreaLaw {
	t.Helper()
	base, _ := restrictionflow.NewArea(rest)
	coefficient, _ := restrictionflow.NewAreaCompliance(compliance)
	maximum, _ := restrictionflow.NewArea(cap)
	law, err := restrictionflow.NewLinearComplianceArea(base, coefficient, maximum)
	if err != nil {
		t.Fatal(err)
	}
	return law
}
