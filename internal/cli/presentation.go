package cli

import (
	"math"
	"strconv"
	"strings"
)

const numericPresentationNote = "Human values: six significant digits; full precision in JSON.\n"

// formatScientificValue rounds display values only. Working from the rounded
// decimal digits avoids overflow, underflow, and a second floating-point rounding.
func formatScientificValue(value float64) string {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return "unavailable"
	}
	if value == 0 {
		return "0"
	}
	rounded := strconv.FormatFloat(value, 'e', 5, 64)
	mantissa, exponentText, _ := strings.Cut(rounded, "e")
	exponent, _ := strconv.Atoi(exponentText)
	if exponent < -3 || exponent >= 6 {
		return strings.TrimRight(strings.TrimRight(mantissa, "0"), ".") + "e" + strconv.Itoa(exponent)
	}
	sign := ""
	if strings.HasPrefix(mantissa, "-") {
		sign = "-"
		mantissa = mantissa[1:]
	}
	digits := strings.ReplaceAll(mantissa, ".", "")
	point := exponent + 1
	var decimal string
	if point <= 0 {
		decimal = "0." + strings.Repeat("0", -point) + digits
	} else {
		decimal = digits[:point] + "." + digits[point:]
	}
	return sign + strings.TrimRight(strings.TrimRight(decimal, "0"), ".")
}
