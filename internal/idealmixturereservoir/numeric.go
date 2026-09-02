package idealmixturereservoir

import (
	"math"
	"sort"
)

func positiveFinite(value float64) error {
	if !finite(value) {
		return ErrNonFiniteValue
	}
	if value <= 0 {
		return ErrNonPositiveValue
	}
	return nil
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func componentMasses(components []Component) []float64 {
	values := make([]float64, len(components))
	for index, component := range components {
		values[index] = component.mass.kilograms
	}
	return values
}

func massValues(masses []Mass) []float64 {
	values := make([]float64, len(masses))
	for index, mass := range masses {
		values[index] = mass.kilograms
	}
	return values
}

func stableSum(values []float64) float64 {
	ordered := append([]float64(nil), values...)
	sort.Float64s(ordered)
	var sum float64
	var compensation float64
	for _, value := range ordered {
		adjusted := value - compensation
		next := sum + adjusted
		compensation = (next - sum) - adjusted
		sum = next
	}
	return sum
}

func stableSignedSum(values []float64) float64 {
	ordered := append([]float64(nil), values...)
	sort.Slice(ordered, func(left, right int) bool {
		return math.Abs(ordered[left]) < math.Abs(ordered[right])
	})
	var sum float64
	var compensation float64
	for _, value := range ordered {
		next := sum + value
		if math.Abs(sum) >= math.Abs(value) {
			compensation += (sum - next) + value
		} else {
			compensation += (value - next) + sum
		}
		sum = next
	}
	return sum + compensation
}
