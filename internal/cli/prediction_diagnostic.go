package cli

import (
	"io"
	"strconv"
	"strings"
)

// These are presentation hints, never input repairs or replacements for a
// report's diagnostic code. The structured report remains unchanged.
var predictionRecovery = map[string]string{
	"input_not_found":                    "Check the input path, or use '-' to read standard input.",
	"input_permission_denied":            "Check that the input file is readable by the current user.",
	"input_unavailable":                  "Check the input file or stream, then retry with a readable source.",
	"input_too_large":                    "Reduce the input document to at most 65,536 bytes.",
	"empty_input":                        "Supply one JSON document through a file or standard input.",
	"malformed_json":                     "Correct the JSON syntax at the reported path, then retry.",
	"trailing_json_value":                "Supply exactly one JSON document; remove the trailing value.",
	"duplicate_member":                   "Keep only one occurrence of the named JSON member.",
	"maximum_depth_exceeded":             "Reduce the JSON nesting depth; see the command's request contract.",
	"member_name_too_long":               "Use the field names defined by the command's request contract.",
	"unsupported_quantity_system":        "Supply explicit SI quantities and the quantity_system value 'si'.",
	"nonfinite_quantity":                 "Supply a finite number for this quantity.",
	"nonpositive_quantity":               "Supply a finite, strictly positive quantity in the stated SI unit.",
	"negative_area":                      "Supply a finite, nonnegative area in square metres.",
	"negative_compliance":                "Supply a finite, nonnegative compliance in m^2/Pa.",
	"invalid_discharge_coefficient":      "Supply a finite discharge coefficient greater than 0 and at most 1.",
	"invalid_heat_capacity_ratio":        "Supply a finite heat-capacity ratio greater than 1.",
	"invalid_withdrawal":                 "Supply a withdrawal fraction at least 0 and less than 1.",
	"reservoir_depletion":                "Complete depletion is outside this model's admissible state domain.",
	"adverse_pressure":                   "Back pressure exceeds source pressure; this model admits no reverse flow.",
	"duplicate_component_id":             "Give each declared component a distinct ID.",
	"invalid_time":                       "Supply nonnegative sample times in strictly increasing order.",
	"invalid_sample_count":               "Supply between 1 and 256 history samples.",
	"incompatible_law_context":           "This model supports earth.continuum.si@v0alpha1; use it only when intended.",
	"exact_law_revision_required":        "Declare the intended law-context revision explicitly in its version field.",
	"invalid_witness_digest":             "Use the retained witness's exact 64-character lowercase SHA-256 value.",
	"witness_mismatch":                   "Inspect both digests with --format json and check the retained input/profile.",
	"unsupported_accuracy_configuration": "Refine requires positive resting area and step.max_time_s = 0 for flowing cases.",
	"accuracy_budget_exhausted":          "Review --format json work evidence, then raise --max-evaluations if intended.",
	"accuracy_not_achieved":              "Review --format json estimates and numerical limits before changing tolerances.",
	"no_representable_progress":          "The requested change is lost at floating-point precision; review its scale.",
	"no_representable_flow":              "Flow cannot be represented at this scale; review the quantities and units.",
	"numerical_domain_error":             "Review input scales and model limits; use --format json for the full refusal.",
	"invariant_violation":                "Retain the input and JSON refusal for investigation; balance checks failed.",
}

func writePredictionDiagnostic(stderr io.Writer, command, code, reason, path string) {
	label := command
	switch command {
	case "reservoir predict":
		label = "reservoir prediction"
	case "restriction predict":
		label = "restriction prediction"
	case "restriction history":
	default:
		operation, isWalk := strings.CutPrefix(command, "walk ")
		_, supported := walkOperations[operation]
		if !isWalk || (!supported && operation != "refine") {
			label, command = "prediction", ""
		}
	}
	hint := predictionRecovery[reason]
	if command == "walk reconstruct" && path == "/expected_witness" && reason == "missing_member" {
		hint = "Retain a 'walk witness' value, then add it as expected_witness to the same case."
	}
	if hint == "" {
		hint = "Check the request contract with 'fartapp help"
		if command != "" {
			hint += " " + command
		}
		hint += "'."
	}
	writeDiagnostic(stderr, "%s failed: %s %s at %s\n  %s\n", label,
		boundedDiagnosticToken(code), boundedDiagnosticToken(reason), quoteDiagnosticPath(path), hint)
}

func boundedDiagnosticToken(value string) string {
	const maximumTokenBytes = 64
	if len(value) > maximumTokenBytes {
		value = value[:maximumTokenBytes] + "..."
	}
	for _, character := range value {
		if character != '-' && character != '_' && (character < 'a' || character > 'z') &&
			(character < 'A' || character > 'Z') && (character < '0' || character > '9') {
			return strconv.QuoteToASCII(value)
		}
	}
	if value == "" {
		return `""`
	}
	return value
}

func quoteDiagnosticPath(path string) string {
	const maximumPathBytes = 256
	if len(path) > maximumPathBytes {
		path = path[:maximumPathBytes] + "..."
	}
	return strconv.QuoteToASCII(path)
}
