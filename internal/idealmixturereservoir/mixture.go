package idealmixturereservoir

import "math"

// weightedProperty scales mass-property products before adding them. Forming
// m*R or m*cv first can quantize subnormal products and change composition
// properties during an otherwise composition-preserving withdrawal.
func weightedProperty(components []Component, total float64, property func(Component) float64) float64 {
	exponents := make([]int, len(components))
	mantissas := make([]float64, len(components))
	maximum := math.MinInt
	for index, component := range components {
		mass, massExponent := math.Frexp(component.mass.kilograms)
		value, valueExponent := math.Frexp(property(component))
		mantissas[index], exponents[index] = math.Frexp(mass * value)
		exponents[index] += massExponent + valueExponent
		maximum = max(maximum, exponents[index])
	}
	for index := range mantissas {
		mantissas[index] = math.Ldexp(mantissas[index], exponents[index]-maximum)
	}
	mass, exponent := math.Frexp(total)
	return math.Ldexp(stableSum(mantissas)/mass, maximum-exponent)
}
