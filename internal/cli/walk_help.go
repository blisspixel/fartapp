package cli

import "fmt"

// Each leaf describes its own inputs and evidence. In particular, a plain case
// is not advertised as a successful reconstruction without a retained witness.
func walkHelpForOperation(operation string) string {
	var purpose, details string
	example := "  fartapp walk " + operation + " testdata/walk/ordinary-low-pressure.json\n"
	switch operation {
	case "predict":
		purpose = "Predict the reachable endpoint or asymptotic limit of a walk case."
		details = "The endpoint and initial restriction follow the declared thermal closure.\nA zero-rest-area compliant opening has an asymptotic equalization limit;\nthe endpoint prediction does not assign it a finite discharge time.\n"
	case "simulate":
		purpose = "Simulate coupled reservoir blowdown and retain its numerical account."
		details = "The report separates stopping reason from arithmetic balance checks. Time\nand step budgets may truncate the history. The first-order clock requires\nstep refinement to assess accuracy; see 'fartapp help walk refine' for the\nopt-in quadrature method and its explicit error-estimate limits.\n"
	case "inspect":
		purpose = "Inspect the simulated account and any declared dimension diagnostics."
		details = "Dimension diagnostics are evaluated only when a compatible law_context is\ndeclared. No law context is inferred when that field is omitted.\n"
	case "explain":
		purpose = "Explain the closure, stopping condition, choking, and numerical limits."
		details = "The explanation accompanies a calculated account under the case's explicit\nmodel and step policy. It does not infer empirical calibration.\n"
	case "branch":
		purpose = "Compare one prescribed-area counterfactual with the base walk case."
		details = "Set branch.prescribed_area_m2 to the counterfactual outlet area in square\nmetres. Both complete accounts and their stopping conditions are retained.\nThe ordinary-low-pressure fixture compares 1e-6 with 2e-6 m^2.\n"
	case "certify":
		purpose = "Inspect arithmetic balance claims, residuals, and declared tolerances."
		details = "Satisfied balance claims establish arithmetic consistency within the\nreported tolerances. This operation issues no certificate and makes no\nempirical-validity or discharge-time-accuracy claim.\n"
	case "witness":
		purpose = "Hash normalized inputs and the complete numerical account."
		details = "Retain the witness value from the JSON report for later reconstruction.\nThe digest uses versioned Go JSON; it is not canonical scientific identity\nor a signature. Matching requires the same model, implementation, runtime\nprofile, and normalized inputs. See 'fartapp help walk reconstruct'.\n"
		example = "  fartapp walk witness testdata/walk/ordinary-low-pressure.json --format json\n"
	case "reconstruct":
		purpose = "Compare a new calculation with a retained expected_witness."
		details = "First run witness and retain its witness value. In a copy of the same case,\nset expected_witness to that exact lowercase SHA-256 value. Run reconstruct\non that copy. Status 1 with witness_mismatch preserves both digests in JSON;\ninspect the input and implementation profile before changing the target.\nThe comparison target is retained evidence, not a second newly generated run.\n"
		example = "  fartapp walk witness testdata/walk/ordinary-low-pressure.json --format json\n  # After adding the retained expected_witness to a copy named retained-case.json:\n  fartapp walk reconstruct retained-case.json --format json\n"
	default:
		return ""
	}
	return fmt.Sprintf(`%s

Usage:
  fartapp walk %s <case.json|-> [--format text|json]

Arguments:
  case.json  Strict JSON using fart.walk-case/v0alpha1, at most 65,536 bytes.
  -          Read the case from standard input.

Options:
  --format text  Write a concise scientific report. This is the default.
  --format json  Write the complete typed account and structured refusals.
  -h, --help     Show this help without reading input.

%s
Scope:
  Explicit SI model. An optional law_context must be
  earth.continuum.si@v0alpha1; incompatible contexts are refused.
  Calculation is read-only and commits no case or physical measurement.

Example:
%s
Exit status:
  0  An account was produced; inspect its stopping condition separately.
  1  Usage, input, model, invariant, comparison, or output failure.
     A native Unix SIGPIPE retains the operating system's pipeline status.
`, purpose, operation, details, example)
}
