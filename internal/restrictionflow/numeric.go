package restrictionflow

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
