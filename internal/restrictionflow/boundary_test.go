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
