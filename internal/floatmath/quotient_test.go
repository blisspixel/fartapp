package floatmath

import (
	"math"
	"testing"
)

func TestProductOverPreservesFiniteExtremeRatios(t *testing.T) {
	for _, test := range []struct {
		divisor float64
		factors []float64
		want    float64
	}{
		{math.Ldexp(1, -1060), []float64{math.Ldexp(1, -1000), math.Ldexp(1, -60)}, 1},
		{math.Ldexp(1, 1000), []float64{math.Ldexp(1, 1000), math.Ldexp(1, 1000)}, math.Ldexp(1, 1000)},
		{math.Ldexp(1, -60), []float64{math.Ldexp(1, -500), math.Ldexp(1, -560), math.Ldexp(1, 1000)}, 1},
		{2, nil, 0.5}, {-2, []float64{4}, -2},
	} {
		if got := ProductOver(test.divisor, test.factors...); got != test.want {
			t.Fatalf("got %g, want %g", got, test.want)
		}
	}
	if !math.IsNaN(ProductOver(0, 0)) || !math.IsInf(ProductOver(0, 1), 1) ||
		ProductOver(math.Inf(1), 1) != 0 || !math.IsNaN(ProductOver(math.Inf(1), math.Inf(1))) {
		t.Fatal("nonfinite ratio semantics changed")
	}
}
