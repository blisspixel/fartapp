package floatmath

import (
	"math"
	"testing"
)

func TestProductPreservesRepresentablePowersOfTwo(t *testing.T) {
	for _, values := range [][]float64{
		{math.Ldexp(1, 900), math.Ldexp(1, 900), math.Ldexp(1, -900)},
		{math.Ldexp(1, -900), math.Ldexp(1, -900), math.Ldexp(1, 900), math.Ldexp(1, 900), math.Ldexp(1, 900)},
		{math.SmallestNonzeroFloat64, math.Ldexp(1, 1000), math.Ldexp(1, 974)},
	} {
		if got, want := Product(values...), math.Ldexp(1, 900); got != want {
			t.Fatalf("Product(%v)=%g, want %g", values, got, want)
		}
	}
}

func TestProductFinalBoundaryAndSigns(t *testing.T) {
	if Product() != 1 || Product(0, math.MaxFloat64, math.MaxFloat64) != 0 || Product(-2, 3, 0.5) != -3 {
		t.Fatal("identity, zero, or sign changed")
	}
	if !math.IsInf(Product(math.MaxFloat64, 2), 1) || Product(math.SmallestNonzeroFloat64, 0.25) != 0 {
		t.Fatal("unrepresentable final product was hidden")
	}
	if !math.IsNaN(Product(0, math.Inf(1))) || !math.IsNaN(Product(math.NaN(), 0)) {
		t.Fatal("nonfinite input was hidden by a zero factor")
	}
}
