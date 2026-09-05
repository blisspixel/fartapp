package floatmath

import "math"

// ProductOver evaluates the product of values divided by divisor without
// forming intermediate products or a potentially overflowing reciprocal.
// Nonfinite inputs and final overflow/underflow retain IEEE-754 semantics.
func ProductOver(divisor float64, values ...float64) float64 {
	fraction, power := math.Frexp(divisor)
	mantissa, exponent := 1/fraction, -power
	for _, value := range values {
		fraction, power = math.Frexp(value)
		mantissa *= fraction
		exponent += power
		mantissa, power = math.Frexp(mantissa)
		exponent += power
	}
	return math.Ldexp(mantissa, exponent)
}
