package coupledblowdown

import (
	"math"
	"testing"

	"github.com/blisspixel/fartapp/internal/idealmixturereservoir"
	"github.com/blisspixel/fartapp/internal/restrictionflow"
	"github.com/blisspixel/fartapp/internal/restrictionhistory"
)

// These references integrate the mass-flow law independently of the production
// restriction evaluator and reservoir transition. Constant area, discharge
// coefficient, volume, gas properties, and back pressure are required.
//
// With x=m/m0 and n=1 (isothermal) or n=gamma (adiabatic), P/P0=x^n,
// T/T0=x^(n-1), and the choked ODE is dx/dt=-k*x^((n+1)/2).
// NASA's mass-flow and isentropic equations supply the restriction relation:
// https://www.grc.nasa.gov/www/k-12/BGP/mflchk.html
// https://www.grc.nasa.gov/www/k-12/airplane/isentrop.html
// They do not validate a transient tank, a biological source, or a heat law.

const (
	referenceMass        = 1.5625
	referencePressure    = 125000.0
	referenceTemperature = 400.0
	referenceGasConstant = 200.0
	referenceVolume      = 1.0
	referenceArea        = 0.01
	referenceBack        = 50000.0
)

func TestChokedHistoryConvergesToIndependentTimeSolution(t *testing.T) {
	for _, closure := range []idealmixturereservoir.Closure{
		idealmixturereservoir.RigidIsothermal,
		idealmixturereservoir.RigidAdiabatic,
	} {
		for _, gamma := range []float64{1.4, 1.5, 5.0 / 3.0} {
			t.Run(closure.String()+"/"+referenceGammaName(gamma), func(t *testing.T) {
				n := referenceExponent(closure, gamma)
				xCritical := math.Pow(referenceBack/referencePressure/referenceCriticalRatio(gamma), 1/n)
				endTime := 0.75 * referenceChokedTime(xCritical, n, gamma)
				previousMassError := math.Inf(1)
				for _, fraction := range []float64{0.004, 0.002, 0.001} {
					result := mustSimulate(t, referenceConfig(t, closure, gamma, fraction, endTime))
					if result.Stop() != StopMaxTime {
						t.Fatalf("fraction %g stopped with %s before prescribed time", fraction, result.Stop())
					}
					referenceNear(t, "elapsed time", result.ElapsedSeconds(), endTime, 1e-12*endTime)
					x := referenceChokedMass(endTime, n, gamma)
					massError := math.Abs(result.Final().TotalMass().Kilograms()/referenceMass - x)
					if massError > 0.65*previousMassError {
						t.Fatalf("mass error did not converge when fraction halved: %g after %g", massError, previousMassError)
					}
					previousMassError = massError
					referenceEndpoint(t, result, x, n, gamma, fraction)
					for _, sample := range result.Samples() {
						if sample.Regime() != restrictionflow.RegimeChoked {
							t.Fatalf("pre-transition sample at %g s has regime %s", sample.Time(), sample.Regime())
						}
						expectedMass := referenceMass * referenceChokedMass(sample.Time(), n, gamma)
						referenceNear(t, "sample mass", sample.Mass(), expectedMass, referenceMass*fraction)
					}

					// Thrust includes pressure force. The integral of x^n is
					// 2*(1-x^((n+1)/2))/(k*(n+1)); momentum alone fails here.
					k := referenceChokedRateConstant(gamma)
					pressureCoefficient := referenceArea * referencePressure * referenceCriticalRatio(gamma) * (gamma + 1)
					integratedPower := 2 * (1 - math.Pow(x, (n+1)/2)) / (k * (n + 1))
					expectedImpulse := pressureCoefficient*integratedPower - referenceArea*referenceBack*endTime
					referenceNear(t, "thrust impulse", result.ImpulseNewtonSeconds(), expectedImpulse, 5*fraction*expectedImpulse)
					referenceNear(t, "recoil impulse", result.RecoilImpulseNewtonSeconds(), -expectedImpulse, 5*fraction*expectedImpulse)
					u0 := math.Sqrt(2 * gamma * referenceGasConstant * referenceTemperature / (gamma + 1))
					expectedStroke := -u0 / k * math.Log(x)
					referenceNear(t, "stroke", result.Signature().StrokeLengthMetres(), expectedStroke, 5*fraction*expectedStroke)
				}
			})
		}
	}
}

func TestSubsonicHistoryConvergesToIndependentTimeSolution(t *testing.T) {
	const gamma = 1.5
	for _, closure := range []idealmixturereservoir.Closure{
		idealmixturereservoir.RigidIsothermal,
		idealmixturereservoir.RigidAdiabatic,
	} {
		t.Run(closure.String(), func(t *testing.T) {
			n := referenceExponent(closure, gamma)
			xCritical := math.Pow(referenceBack/referencePressure/referenceCriticalRatio(gamma), 1/n)
			xEqualized := math.Pow(referenceBack/referencePressure, 1/n)
			x := (xCritical + xEqualized) / 2
			endTime := referenceGamma15Time(x, closure)
			previousMassError := math.Inf(1)
			for _, fraction := range []float64{0.008, 0.004, 0.002, 0.001} {
				result := mustSimulate(t, referenceConfig(t, closure, gamma, fraction, endTime))
				if result.Stop() != StopMaxTime {
					t.Fatalf("fraction %g stopped with %s before subsonic target", fraction, result.Stop())
				}
				referenceNear(t, "elapsed time", result.ElapsedSeconds(), endTime, 1e-12*endTime)
				referenceEndpoint(t, result, x, n, gamma, 2*fraction)
				massError := math.Abs(result.Final().TotalMass().Kilograms()/referenceMass - x)
				if massError > 0.65*previousMassError {
					t.Fatalf("subsonic mass error did not converge: %g after %g", massError, previousMassError)
				}
				previousMassError = massError
				referenceCheckTransition(t, result, n, gamma, fraction)
			}
		})
	}
}

func TestEqualizationTimeConvergesToIndependentCompleteSolution(t *testing.T) {
	const gamma = 1.5
	for _, closure := range []idealmixturereservoir.Closure{
		idealmixturereservoir.RigidIsothermal,
		idealmixturereservoir.RigidAdiabatic,
	} {
		t.Run(closure.String(), func(t *testing.T) {
			n := referenceExponent(closure, gamma)
			x := math.Pow(referenceBack/referencePressure, 1/n)
			expectedTime := referenceGamma15Time(x, closure)
			previousError := math.Inf(1)
			for _, intervals := range []int{128, 256, 512, 1024, 2048} {
				// Align the mass endpoint across refinements. The time integrand
				// behaves as 1/sqrt(m-meq), so left rectangles have O(sqrt(f))
				// equalization-time error, despite O(f) convergence away from it.
				fraction := -math.Expm1(math.Log(x) / float64(intervals))
				result := mustSimulate(t, referenceConfig(t, closure, gamma, fraction, 0))
				if result.Stop() != StopEqualized && result.Stop() != StopNoFlow {
					t.Fatalf("%d intervals stopped with %s: steps=%d time=%.17g pressure=%.17g gap=%.17g tolerance=%.17g",
						intervals, result.Stop(), result.Steps(), result.ElapsedSeconds(), result.Final().Pressure().Pascals(),
						result.Final().Pressure().Pascals()-referenceBack, result.EqualizationPressureTolerancePascals())
				}
				referenceEndpoint(t, result, x, n, gamma, 2e-10)
				referenceNear(t, "equalized pressure", result.Final().Pressure().Pascals(), referenceBack, 2e-10*referencePressure)
				relativeError := math.Abs(result.ElapsedSeconds()/expectedTime - 1)
				if relativeError > 0.8*previousError {
					t.Fatalf("equalization-time error did not converge: %g after %g", relativeError, previousError)
				}
				if relativeError > math.Sqrt(fraction) {
					t.Fatalf("equalization-time error %g exceeds sqrt(fraction)=%g", relativeError, math.Sqrt(fraction))
				}
				previousError = relativeError
				t.Logf("intervals=%d fraction=%.9g equalization_seconds=%.12g reference_seconds=%.12g relative_error=%.6g",
					intervals, fraction, result.ElapsedSeconds(), expectedTime, relativeError)
			}
			if previousError > 0.015 {
				t.Fatalf("finest equalization-time relative error %g exceeds 1.5 percent", previousError)
			}
		})
	}
}

func TestGamma14HistoryMatchesPublishedDischargeRelations(t *testing.T) {
	const gamma = 1.4
	for _, closure := range []idealmixturereservoir.Closure{
		idealmixturereservoir.RigidIsothermal,
		idealmixturereservoir.RigidAdiabatic,
	} {
		for _, initial := range []struct {
			name          string
			backRatio     float64
			endPressure   float64
			initialRegime restrictionflow.Regime
		}{
			{name: "initially-choked", backRatio: 0.4, endPressure: 0.55, initialRegime: restrictionflow.RegimeChoked},
			{name: "initially-subsonic", backRatio: 0.8, endPressure: 0.9, initialRegime: restrictionflow.RegimeSubsonic},
		} {
			t.Run(closure.String()+"/"+initial.name, func(t *testing.T) {
				n := referenceExponent(closure, gamma)
				x := math.Pow(initial.endPressure, 1/n)
				endTime := referenceGamma14PublishedTime(x, closure, initial.backRatio)
				previousError := math.Inf(1)
				for _, fraction := range []float64{0.004, 0.002, 0.001} {
					base := referenceConfig(t, closure, gamma, fraction, endTime)
					config, err := NewConfig(base.Reservoir(), closure, mustPressure(t, initial.backRatio*referencePressure),
						base.AreaLaw(), base.DischargeCoefficient(), fraction, MaxSteps, endTime)
					if err != nil {
						t.Fatal(err)
					}
					result := mustSimulate(t, config)
					if result.Stop() != StopMaxTime {
						t.Fatalf("published-reference history stopped with %s", result.Stop())
					}
					samples := result.Samples()
					if samples[0].Regime() != initial.initialRegime || samples[len(samples)-1].Regime() != restrictionflow.RegimeSubsonic {
						t.Fatalf("published-reference history has wrong initial or final regime: %s, %s",
							samples[0].Regime(), samples[len(samples)-1].Regime())
					}
					referenceNear(t, "elapsed time", result.ElapsedSeconds(), endTime, 1e-12*endTime)
					referenceEndpoint(t, result, x, n, gamma, 2*fraction)
					massError := math.Abs(result.Final().TotalMass().Kilograms()/referenceMass - x)
					if massError > 0.65*previousError {
						t.Fatalf("published-reference mass error did not converge: %g after %g", massError, previousError)
					}
					previousError = massError
				}
			})
		}
	}
}

func TestZeroRestAreaComplianceDoesNotEqualizeInFiniteTime(t *testing.T) {
	const gamma = 1.5
	base := referenceConfig(t, idealmixturereservoir.RigidIsothermal, gamma, 0.2, 0)
	zero, err := restrictionflow.NewArea(0)
	if err != nil {
		t.Fatal(err)
	}
	maximum, err := restrictionflow.NewArea(referenceArea)
	if err != nil {
		t.Fatal(err)
	}
	compliance, err := restrictionflow.NewAreaCompliance(referenceArea / (referencePressure - referenceBack))
	if err != nil {
		t.Fatal(err)
	}
	law, err := restrictionflow.NewLinearComplianceArea(zero, compliance, maximum)
	if err != nil {
		t.Fatal(err)
	}
	config, err := NewConfig(base.Reservoir(), base.Closure(), base.BackPressure(), law, base.DischargeCoefficient(), 0.2, MaxSteps, 0)
	if err != nil {
		t.Fatal(err)
	}
	result := mustSimulate(t, config)
	// Near equality, q is proportional to (P-Pb)^(3/2), since both
	// A=C*(P-Pb) and the subsonic speed tend to zero. Its time integral
	// diverges. A finite numerical tolerance stop must retain a positive gap.
	if result.Stop() == StopEqualized || result.Stop() == StopNoFlow {
		t.Fatalf("zero-rest-area compliance falsely completed at finite time: %s at %g s", result.Stop(), result.ElapsedSeconds())
	}
	if result.Final().Pressure().Pascals() <= referenceBack || result.ElapsedSeconds() <= 0 || !finite(result.ElapsedSeconds()) {
		t.Fatalf("asymptotic tail lost positive pressure gap or finite progress: P=%g time=%g", result.Final().Pressure().Pascals(), result.ElapsedSeconds())
	}
}

func TestRepresentableEnergyTransfersProduceFiniteLedger(t *testing.T) {
	// Each state and transfer is finite, but Ufinal+Hout exceeds MaxFloat64
	// before the opposing Qin and Uinitial terms are included. An ordered
	// positive-first residual must not turn this valid account into NaN.
	mass, _ := idealmixturereservoir.NewMass(2)
	gas, _ := idealmixturereservoir.NewSpecificGasConstant(4e307)
	cv, _ := idealmixturereservoir.NewIsochoricHeatCapacity(5e307)
	component, err := idealmixturereservoir.NewComponent(mass, gas, cv)
	if err != nil {
		t.Fatal(err)
	}
	volume, _ := idealmixturereservoir.NewVolume(8e302)
	temperature, _ := idealmixturereservoir.NewTemperature(1)
	state, err := idealmixturereservoir.NewState([]idealmixturereservoir.Component{component}, volume, temperature)
	if err != nil {
		t.Fatal(err)
	}
	config, err := NewConfig(state, idealmixturereservoir.RigidIsothermal, mustPressure(t, 100),
		mustArea(t, 1e144), mustDischarge(t, 1), 0.998, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	result := mustSimulate(t, config)
	if result.Steps() != 1 {
		t.Fatalf("finite large-energy fixture made %d steps", result.Steps())
	}
	referenceNear(t, "scaled energy residual", result.Ledgers().EnergyResidualJoules()/1e308, 0, 1e-13)
}

func TestAuthoredTimeBoundaryNeverRoundsPastBudget(t *testing.T) {
	initial := mustSimulate(t, mustConfig(t, idealmixturereservoir.RigidIsothermal, referenceArea, 0.01, 1, 0))
	q := initial.Samples()[0].MassFlow()
	for _, fraction := range []float64{0.01, 0.03, 0.1, 0.2, 0.001, 0.0001, 0.00001, 0.123456789} {
		nominalTime := referenceMass * fraction / q
		for _, budget := range []float64{nominalTime, math.Nextafter(nominalTime, 0), math.Nextafter(nominalTime, math.Inf(1))} {
			config := mustConfig(t, idealmixturereservoir.RigidIsothermal, referenceArea, fraction, 1, budget)
			result := mustSimulate(t, config)
			if result.ElapsedSeconds() > budget {
				t.Fatalf("fraction %.17g exceeded authored time: elapsed=%.17g budget=%.17g excess=%.17g stop=%s",
					fraction, result.ElapsedSeconds(), budget, result.ElapsedSeconds()-budget, result.Stop())
			}
		}
	}
}

func TestChokingBoundaryAdmitsAdjacentPressureValues(t *testing.T) {
	for _, gamma := range []float64{1.01, 1.1, 1.4, 1.5, 5.0 / 3.0, 2, 3, 10} {
		base := referenceConfig(t, idealmixturereservoir.RigidIsothermal, gamma, 0.001, 0.001)
		actualGamma := base.Reservoir().HeatCapacityRatio()
		critical := referencePressure * referenceCriticalRatio(actualGamma)
		for _, direction := range []float64{0, math.Inf(1)} {
			pressure := critical
			for offset := 0; offset < 8; offset++ {
				pressure = math.Nextafter(pressure, direction)
				_, err := NewConfig(base.Reservoir(), base.Closure(), mustPressure(t, pressure),
					base.AreaLaw(), base.DischargeCoefficient(), 0.001, 1, 0.001)
				if err != nil {
					t.Fatalf("adjacent choking boundary was refused: gamma=%.17g critical=%.17g back=%.17g err=%v",
						actualGamma, critical, pressure, err)
				}
			}
		}
	}
}

func TestFrozenSourceConstantRateIntegratesAcrossScale(t *testing.T) {
	// A positive constant subnormal rate integrated over a long interval has
	// a representable total. Averaging the rates must not first erase them.
	temperature, _ := restrictionflow.NewTemperature(1)
	gas, _ := restrictionflow.NewSpecificGasConstant(2e-9)
	gamma, _ := restrictionflow.NewHeatCapacityRatio(1.5)
	stagnation, err := restrictionflow.NewStagnation(mustPressure(t, 1e-9), temperature, gas, gamma)
	if err != nil {
		t.Fatal(err)
	}
	area, _ := restrictionflow.NewArea(3.2e-319)
	start, _ := restrictionhistory.NewSeconds(0)
	const duration = 1e300
	end, _ := restrictionhistory.NewSeconds(duration)
	first, err := restrictionhistory.NewSample(start, area)
	if err != nil {
		t.Fatal(err)
	}
	last, err := restrictionhistory.NewSample(end, area)
	if err != nil {
		t.Fatal(err)
	}
	history, err := restrictionhistory.Integrate(stagnation, mustPressure(t, 4e-10), mustDischarge(t, 1),
		[]restrictionhistory.Sample{first, last})
	if err != nil {
		t.Fatal(err)
	}
	q := history.Samples()[0].Result().MassFlow().KilogramsPerSecond()
	expectedMass := q * duration
	if expectedMass <= 0 {
		t.Fatal("fixture failed to produce a positive representable mass integral")
	}
	referenceNear(t, "constant-rate integrated mass", history.MassOutKilograms(), expectedMass, 1e-14*expectedMass)
	// The source enthalpy of this frozen gas is exactly 6e-9 J/kg.
	expectedEnergy := expectedMass * 6e-9
	referenceNear(t, "constant-rate total enthalpy", history.TotalEnthalpyOutJoules(), expectedEnergy, 1e-14*expectedEnergy)
	flow := history.Samples()[0].Result()
	expectedImpulse := expectedMass*flow.ExitSpeed().MetresPerSecond() +
		(area.SquareMetres()*duration)*(flow.ExitPressure().Pascals()-4e-10)
	referenceNear(t, "constant-rate impulse", history.ImpulseNewtonSeconds(), expectedImpulse, 1e-14*expectedImpulse)
}

func TestCoupledImpulseRetainsRepresentableIntegratedMomentum(t *testing.T) {
	mass, _ := idealmixturereservoir.NewMass(1e-171)
	gas, _ := idealmixturereservoir.NewSpecificGasConstant(2e-9)
	cv, _ := idealmixturereservoir.NewIsochoricHeatCapacity(4e-9)
	component, err := idealmixturereservoir.NewComponent(mass, gas, cv)
	if err != nil {
		t.Fatal(err)
	}
	volume, _ := idealmixturereservoir.NewVolume(2e-171)
	temperature, _ := idealmixturereservoir.NewTemperature(1)
	state, err := idealmixturereservoir.NewState([]idealmixturereservoir.Component{component}, volume, temperature)
	if err != nil {
		t.Fatal(err)
	}
	config, err := NewConfig(state, idealmixturereservoir.RigidIsothermal, mustPressure(t, 4e-10),
		mustArea(t, 3.2e-319), mustDischarge(t, 1), 0.1, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	result := mustSimulate(t, config)
	if result.Steps() != 1 {
		t.Fatalf("representable integrated momentum fixture made %d steps", result.Steps())
	}
	initial := result.Samples()[0]
	expectedImpulse := result.MassOutKilograms()*initial.ExitSpeed() +
		(initial.EffectiveArea()*result.ElapsedSeconds())*(initial.ExitPressure()-4e-10)
	if expectedImpulse <= 0 || !finite(expectedImpulse) {
		t.Fatal("fixture failed to retain a positive finite impulse")
	}
	referenceNear(t, "integrated source momentum", result.ImpulseNewtonSeconds(), expectedImpulse, 1e-14*expectedImpulse)
	referenceNear(t, "integrated source recoil", result.RecoilImpulseNewtonSeconds(), -expectedImpulse, 1e-14*expectedImpulse)
}

func referenceConfig(t *testing.T, closure idealmixturereservoir.Closure, gamma, fraction, endTime float64) Config {
	t.Helper()
	mass, err := idealmixturereservoir.NewMass(referenceMass)
	if err != nil {
		t.Fatal(err)
	}
	gas, err := idealmixturereservoir.NewSpecificGasConstant(referenceGasConstant)
	if err != nil {
		t.Fatal(err)
	}
	cv, err := idealmixturereservoir.NewIsochoricHeatCapacity(referenceGasConstant / (gamma - 1))
	if err != nil {
		t.Fatal(err)
	}
	component, err := idealmixturereservoir.NewComponent(mass, gas, cv)
	if err != nil {
		t.Fatal(err)
	}
	volume, err := idealmixturereservoir.NewVolume(referenceVolume)
	if err != nil {
		t.Fatal(err)
	}
	temperature, err := idealmixturereservoir.NewTemperature(referenceTemperature)
	if err != nil {
		t.Fatal(err)
	}
	state, err := idealmixturereservoir.NewState([]idealmixturereservoir.Component{component}, volume, temperature)
	if err != nil {
		t.Fatal(err)
	}
	config, err := NewConfig(state, closure, mustPressure(t, referenceBack), mustArea(t, referenceArea), mustDischarge(t, 1), fraction, MaxSteps, endTime)
	if err != nil {
		t.Fatal(err)
	}
	return config
}

func referenceExponent(closure idealmixturereservoir.Closure, gamma float64) float64 {
	if closure == idealmixturereservoir.RigidIsothermal {
		return 1
	}
	return gamma
}

func referenceCriticalRatio(gamma float64) float64 {
	return math.Pow(2/(gamma+1), gamma/(gamma-1))
}

func referenceChokedRateConstant(gamma float64) float64 {
	return referenceArea / referenceVolume * math.Sqrt(gamma*referenceGasConstant*referenceTemperature) *
		math.Pow(2/(gamma+1), (gamma+1)/(2*(gamma-1)))
}

func referenceChokedMass(elapsed, n, gamma float64) float64 {
	k := referenceChokedRateConstant(gamma)
	if n == 1 {
		return math.Exp(-k * elapsed)
	}
	return math.Pow(1+(n-1)*k*elapsed/2, -2/(n-1))
}

func referenceChokedTime(x, n, gamma float64) float64 {
	k := referenceChokedRateConstant(gamma)
	if n == 1 {
		return -math.Log(x) / k
	}
	return 2 * (math.Pow(x, -(n-1)/2) - 1) / ((n - 1) * k)
}

// For gamma=3/2 the complete subsonic tail has elementary primitives.
// Write B=V/(Cd*A*sqrt(R*T0)), r=Pb/P0 and x=m/m0.
//
// Isothermal: s=sqrt(1-(r/x)^(1/3)) gives
// dt/ds=B*sqrt(6)/(1-s^2)^3. Its primitive is B*sqrt(6)*
// [s/(4*(1-s^2)^2)+3*s/(8*(1-s^2))+3*atanh(s)/8].
//
// Adiabatic: b=r^(1/3), w=sqrt(sqrt(x)-b) gives
// dt/dw=4*B*(w^2+b)/(sqrt(6)*r^(2/3)), with primitive
// 4*B*(w^3/3+b*w)/(sqrt(6)*r^(2/3)). Both primitives vanish
// at equality. Time is the choked time plus the difference of primitives.
func referenceGamma15Time(x float64, closure idealmixturereservoir.Closure) float64 {
	const gamma = 1.5
	n := referenceExponent(closure, gamma)
	xCritical := math.Pow(referenceBack/referencePressure/referenceCriticalRatio(gamma), 1/n)
	if x >= xCritical {
		return referenceChokedTime(x, n, gamma)
	}
	return referenceChokedTime(xCritical, n, gamma) +
		referenceGamma15Tail(xCritical, closure) - referenceGamma15Tail(x, closure)
}

func referenceGamma15Tail(x float64, closure idealmixturereservoir.Closure) float64 {
	r := referenceBack / referencePressure
	b := referenceVolume / (referenceArea * math.Sqrt(referenceGasConstant*referenceTemperature))
	if closure == idealmixturereservoir.RigidIsothermal {
		s := math.Sqrt(math.Max(0, 1-math.Cbrt(r/x)))
		d := 1 - s*s
		return b * math.Sqrt(6) * (s/(4*d*d) + 3*s/(8*d) + 3*math.Atanh(s)/8)
	}
	cubeRoot := math.Cbrt(r)
	w := math.Sqrt(math.Max(0, math.Sqrt(x)-cubeRoot))
	return 4 * b / (math.Sqrt(6) * cubeRoot * cubeRoot) * (w*w*w/3 + cubeRoot*w)
}

// Dutton and Coverdill (1997), equations (3)-(5), give these gamma=7/5
// unchoked time primitives. The published area ratio At/Ae is one here.
// Their chi=sqrt((P/Pb)^(2/7)-1) and tchar=V/(A*sqrt(gamma*R*T0)).
// Initially unchoked cases start at x=1 and t=0, without a choked segment.
// https://www.ijee.ie/articles/Vol13-2/ijee924.pdf
func referenceGamma14PublishedTime(x float64, closure idealmixturereservoir.Closure, backRatio float64) float64 {
	const gamma = 1.4
	n := referenceExponent(closure, gamma)
	xCritical := math.Pow(backRatio/referenceCriticalRatio(gamma), 1/n)
	startMass := math.Min(1, xCritical)
	startTime := 0.0
	if xCritical < 1 {
		startTime = referenceChokedTime(xCritical, n, gamma)
	}
	primitive := func(massFraction float64) float64 {
		pressureRatio := math.Pow(massFraction, n) / backRatio
		chi := math.Sqrt(math.Max(0, math.Pow(pressureRatio, 2.0/7.0)-1))
		tchar := referenceVolume / (referenceArea * math.Sqrt(gamma*referenceGasConstant*referenceTemperature))
		if closure == idealmixturereservoir.RigidIsothermal {
			polynomial := math.Pow(chi, 5)/5 + 2*math.Pow(chi, 3)/3 + chi
			return tchar * math.Sqrt(2*gamma*gamma/(gamma-1)) * polynomial
		}
		integral := (chi*chi*chi/4+5*chi/8)*math.Sqrt(chi*chi+1) + 3*math.Asinh(chi)/8
		return tchar * math.Sqrt(2/(gamma-1)) * math.Pow(backRatio, -(gamma-1)/(2*gamma)) * integral
	}
	return startTime + primitive(startMass) - primitive(x)
}

func referenceEndpoint(t *testing.T, result Result, x, n, gamma, relativeTolerance float64) {
	t.Helper()
	referenceNear(t, "mass", result.Final().TotalMass().Kilograms(), referenceMass*x, referenceMass*relativeTolerance)
	referenceNear(t, "pressure", result.Final().Pressure().Pascals(), referencePressure*math.Pow(x, n), n*referencePressure*relativeTolerance)
	referenceNear(t, "temperature", result.Final().Temperature().Kelvin(), referenceTemperature*math.Pow(x, n-1), referenceTemperature*relativeTolerance)
	cv := referenceGasConstant / (gamma - 1)
	cp := cv + referenceGasConstant
	energyScale := referenceMass * cp * referenceTemperature
	enthalpy := referenceMass * cv * referenceTemperature * (1 - math.Pow(x, gamma))
	heat := 0.0
	if n == 1 {
		enthalpy = referenceMass * cp * referenceTemperature * (1 - x)
		heat = referenceMass * referenceGasConstant * referenceTemperature * (1 - x)
	}
	referenceNear(t, "enthalpy out", result.EnthalpyOutJoules(), enthalpy, energyScale*relativeTolerance)
	referenceNear(t, "heat in", result.HeatInJoules(), heat, energyScale*relativeTolerance)
}

func referenceCheckTransition(t *testing.T, result Result, n, gamma, fraction float64) {
	t.Helper()
	samples := result.Samples()
	firstSubsonic := -1
	for index, sample := range samples {
		if index > 0 && sample.Time() <= samples[index-1].Time() {
			t.Fatalf("time did not increase at sample %d", index)
		}
		if sample.Regime() == restrictionflow.RegimeSubsonic && firstSubsonic < 0 {
			firstSubsonic = index
		}
		if firstSubsonic >= 0 && sample.Regime() != restrictionflow.RegimeSubsonic {
			t.Fatalf("flow returned from subsonic to %s", sample.Regime())
		}
	}
	if firstSubsonic < 1 {
		t.Fatal("history did not contain a choked-to-subsonic transition")
	}
	xCritical := math.Pow(referenceBack/referencePressure/referenceCriticalRatio(gamma), 1/n)
	expectedTime := referenceChokedTime(xCritical, n, gamma)
	transitionSample := samples[firstSubsonic]
	// A sampled transition is localized to one mass step, in addition to
	// the accumulated first-order timing error. This is not an exact event.
	transitionTolerance := 3 * fraction / referenceChokedRateConstant(gamma)
	referenceNear(t, "transition sample time", transitionSample.Time(), expectedTime, transitionTolerance)
	previous := samples[firstSubsonic-1]
	if previous.Pressure() < referenceBack/referenceCriticalRatio(gamma) ||
		transitionSample.Pressure() >= referenceBack/referenceCriticalRatio(gamma) {
		t.Fatal("transition samples did not straddle the analytical choking pressure")
	}
}

func referenceNear(t *testing.T, name string, actual, expected, tolerance float64) {
	t.Helper()
	if math.IsNaN(actual) || math.IsInf(actual, 0) || math.Abs(actual-expected) > tolerance {
		t.Fatalf("%s = %.17g, want %.17g within %.3g", name, actual, expected, tolerance)
	}
}

func referenceGammaName(gamma float64) string {
	switch gamma {
	case 1.4:
		return "gamma-1.4"
	case 1.5:
		return "gamma-1.5"
	default:
		return "gamma-5over3"
	}
}
