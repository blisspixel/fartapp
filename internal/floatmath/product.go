// Package floatmath contains shared binary64 arithmetic used by the scalar
// oracles. Domain admissibility and physical units remain caller-owned.
package floatmath

import "math"

// Product multiplies significands separately from exponents so intermediate
// overflow or underflow cannot erase an otherwise representable product.
// Final overflow, underflow, and nonfinite inputs retain IEEE-754 results.
func Product(values ...float64) float64 {
	mantissa := 1.0
	exponent := 0
	for _, value := range values {
		fraction, power := math.Frexp(value)
		mantissa *= fraction
		exponent += power
		mantissa, power = math.Frexp(mantissa)
		exponent += power
	}
	return math.Ldexp(mantissa, exponent)
}
