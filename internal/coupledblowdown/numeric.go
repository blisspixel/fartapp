package coupledblowdown

import "math"

// Neumaier accumulation limits roundoff in a long series of finite transfers.
type accumulator struct {
	sum, compensation float64
}

func (a *accumulator) add(value float64) {
	next := a.sum + value
	if math.Abs(a.sum) >= math.Abs(value) {
		a.compensation += (a.sum - next) + value
	} else {
		a.compensation += (value - next) + a.sum
	}
	a.sum = next
}

func (a accumulator) value() float64 { return a.sum + a.compensation }

func signedSum(values ...float64) float64 {
	scale := 0.0
	for _, value := range values {
		scale = math.Max(scale, math.Abs(value))
	}
	if scale == 0 {
		return 0
	}
	var sum accumulator
	for _, value := range values {
		// Scaling keeps individually finite debit and credit terms from
		// overflowing before they cancel in a residual calculation.
		sum.add(value / scale)
	}
	return sum.value() * scale
}
