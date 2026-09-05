package restrictionflow

import (
	"math"
	"sort"

	"github.com/blisspixel/fartapp/internal/floatmath"
)

const smallestNormal = 0x1p-1022

func normalPositive(values ...float64) bool {
	for _, value := range values {
		if !finite(value) || value < smallestNormal {
			return false
		}
	}
	return true
}

// Keep the ordinary arithmetic path unchanged. Scaled fallbacks avoid losing
// a representable reported value through an unrepresentable intermediate.
func throatTemperature(temperature, gamma float64) float64 {
	value := temperature * 2 / (gamma + 1)
	if finite(value) && value > 0 {
		return value
	}
	return floatmath.ProductOver(gamma+1, temperature, 2)
}

func flowSpeed(mach, gamma, gasConstant, temperature float64) float64 {
	thermalFactor := gamma * gasConstant
	squaredSoundSpeed := thermalFactor * temperature
	value := mach * math.Sqrt(squaredSoundSpeed)
	if finite(value) && value > 0 && normalPositive(thermalFactor, squaredSoundSpeed) {
		return value
	}
	return floatmath.Product(mach, math.Sqrt(gamma), math.Sqrt(gasConstant), math.Sqrt(temperature))
}

func scaledMassFlow(cd, area, mach, pressure, gamma, gasConstant, temperature float64) float64 {
	// Continuity and the ideal gas law give
	// mdot = Cd*A*M*p*sqrt(gamma)/(sqrt(R)*sqrt(T)). The square-root
	// reciprocals remain finite for every positive finite binary64 input.
	return floatmath.Product(cd, area, mach, pressure, math.Sqrt(gamma),
		1/math.Sqrt(gasConstant), 1/math.Sqrt(temperature))
}

func scaledContinuity(cd, area, pressure, speed, gasConstant, temperature float64) float64 {
	// Reconstruct Cd*rho*A*v from the reported exit speed. Factoring the
	// two reciprocals separately avoids materializing either density or R*T.
	rootR := 1 / math.Sqrt(gasConstant)
	rootT := 1 / math.Sqrt(temperature)
	return floatmath.Product(cd, area, pressure, speed, rootR, rootR, rootT, rootT)
}

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
