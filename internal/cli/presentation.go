package cli

import (
	"fmt"
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

// wrapProse wraps whitespace-separated words at limit columns with a hanging indent.
// Unbroken tokens exceeding limit are never wrapped mid-token.
func wrapProse(text string, indent string, limit int) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}
	var lines []string
	currentLine := indent + words[0]
	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= limit {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = indent + word
		}
	}
	lines = append(lines, currentLine)
	return strings.Join(lines, "\n")
}

// formatIdentifierList writes a labeled identifier list with one item per line,
// preceded by a blank line, as required by the design language.
func formatIdentifierList(builder *strings.Builder, label string, items []string) {
	if len(items) == 0 {
		return
	}
	fmt.Fprintf(builder, "\n%s:\n", label)
	for _, item := range items {
		fmt.Fprintf(builder, "  %s\n", item)
	}
}
