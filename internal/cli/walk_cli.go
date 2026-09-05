package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/blisspixel/fartapp/internal/walkcase"
)

const walkHelp = `F.A.R.T. Lab experimental walking skeleton

Usage:
  fartapp walk <operation> <case.json|-> [options]
  fartapp help walk <operation>

Commands:
  predict      Reachable endpoint or asymptotic limit and initial restriction.
  simulate     Quasi-steady coupled blowdown with ledgers and signature.
  inspect      The simulated account with dimension diagnostics when declared.
  explain      Explain closure, stopping condition, choking, and numerical limits.
  branch       One prescribed-area counterfactual comparison.
  certify      Inspect arithmetic balances with residuals and tolerances.
  witness      Hash normalized inputs and the complete numerical account.
  reconstruct  Compare a new calculation with a retained expected_witness.
  refine       Estimate quadrature error under explicit tolerances and a work budget.

Help:
  fartapp help walk simulate

The walkthrough is an explicit SI model, not an ordinary pfft calibration or
Reference Pfft. An optional law_context must be earth.continuum.si@v0alpha1;
incompatible contexts are refused. No context is selected when it is omitted.
Use evidence help for a provisional retained software carrier. There is no
case commitment, certificate authority, or physical audio.

Example:
  fartapp walk explain testdata/walk/ordinary-low-pressure.json
  fartapp walk branch testdata/walk/ordinary-low-pressure.json
`

var walkOperations = map[string]struct{}{
	"predict": {}, "simulate": {}, "inspect": {}, "explain": {},
	"branch": {}, "certify": {}, "witness": {}, "reconstruct": {},
}

func runWalk(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		writeDiagnostic(stderr, "usage: fartapp walk <operation> <case.json|-> [options]; run 'fartapp help walk'\n")
		return 1
	}
	if repeatedHelpRequest(args) {
		writeDiagnostic(stderr, "invalid walk help: --help may be specified only once\n")
		return 1
	}
	if args[0] == "help" {
		if len(args) != 1 {
			writeDiagnostic(stderr, "usage: fartapp walk help\n")
			return 1
		}
		return writeText(stdout, stderr, walkHelp)
	}
	if isHelpRequest(args) {
		return writeText(stdout, stderr, walkHelp)
	}
	if args[0] == "refine" {
		return runRefine(args[1:], stdin, stdout, stderr)
	}
	if _, ok := walkOperations[args[0]]; !ok {
		writeDiagnostic(stderr, "unknown walk command %s; run 'fartapp help walk'\n", quoteInput(args[0]))
		return 1
	}
	operation := args[0]
	options, err := parseOutputOptions(args[1:])
	if err != nil {
		writeDiagnostic(stderr, "invalid walk %s: %v\n", operation, err)
		return 1
	}
	if len(options.positional) > 1 || (!options.help && len(options.positional) != 1) {
		writeDiagnostic(stderr, "usage: fartapp walk %s <case.json|-> [--format text|json]\n", operation)
		return 1
	}
	if options.help {
		return writeText(stdout, stderr, walkHelpForOperation(operation))
	}
	report := readWalk(options.positional[0], operation, stdin)
	return writeWalkReport(report, operation, options.format, stdout, stderr)
}

func writeWalkReport(report walkcase.Report, operation string, format outputFormat, stdout, stderr io.Writer) int {
	if report.Predicted() {
		return writeValue(stdout, stderr, format, report, formatWalk)
	}
	if format == outputJSON {
		if writeValue(stdout, stderr, format, report, formatWalk) != 0 {
			return 1
		}
		return 1
	}
	diagnostic := report.Diagnostics[0]
	writePredictionDiagnostic(stderr, "walk "+operation, diagnostic.Code, diagnostic.ReasonCode, diagnostic.Path)
	return 1
}

func readWalk(source, operation string, stdin io.Reader) walkcase.Report {
	return readWalkWith(source, stdin, func(data []byte) walkcase.Report { return walkcase.Run(data, operation) })
}

func readWalkWith(source string, stdin io.Reader, calculate func([]byte) walkcase.Report) walkcase.Report {
	reader := stdin
	if source != "-" {
		opened, err := os.Open(source)
		if err != nil {
			return walkcase.InputFailure(classifyWalkInputError(err), "input_source_reference")
		}
		reader = opened
		defer func() { _ = opened.Close() }()
	}
	data, err := walkcase.ReadBounded(reader)
	if err != nil {
		if errors.Is(err, walkcase.ErrInputTooLarge) {
			return walkcase.InputFailure("input_too_large", "input_stream")
		}
		return walkcase.InputFailure("input_unavailable", "input_stream")
	}
	return calculate(data)
}

func classifyWalkInputError(err error) string {
	switch {
	case errors.Is(err, os.ErrNotExist):
		return "input_not_found"
	case errors.Is(err, os.ErrPermission):
		return "input_permission_denied"
	default:
		return "input_unavailable"
	}
}

func formatWalk(report walkcase.Report) string {
	var output strings.Builder
	fmt.Fprintf(&output, "WALK %s\n\n", strings.ToUpper(report.Operation))
	fmt.Fprintf(&output, "Implementation: %s\n", report.ImplementationRevision)
	if report.LawContext != "" {
		fmt.Fprintf(&output, "Law context: %s\n", report.LawContext)
	} else {
		output.WriteString("Law context: none\n")
	}
	if report.Closure != "" {
		fmt.Fprintf(&output, "Closure: %s\n", report.Closure)
	}
	if report.Model != nil {
		fmt.Fprintf(&output, "Model: %s@%s\n", report.Model.ID, report.Model.Version)
	}
	if report.Initial != nil {
		fmt.Fprintf(&output, "Initial pressure: %s Pa\n", formatScientificValue(report.Initial.PressurePascals))
		fmt.Fprintf(&output, "Initial mass: %s kg\n", formatScientificValue(report.Initial.MassKilograms))
	}
	if report.InitialRestriction != nil {
		fmt.Fprintf(&output, "Initial regime: %s\n", report.InitialRestriction.Regime)
	}
	if report.EqualizationFraction != nil {
		fmt.Fprintf(&output, "Equalization fraction: %s\n", formatScientificValue(*report.EqualizationFraction))
	}
	if report.EndpointReachability != "" {
		fmt.Fprintf(&output, "Endpoint reachability: %s\n", report.EndpointReachability)
	}
	if report.Stop != "" {
		fmt.Fprintf(&output, "Stop: %s\n", report.Stop)
	}
	if report.ElapsedSeconds != nil {
		fmt.Fprintf(&output, "Elapsed: %s s\n", formatScientificValue(*report.ElapsedSeconds))
	}
	if report.Final != nil {
		fmt.Fprintf(&output, "Final pressure: %s Pa\n", formatScientificValue(report.Final.PressurePascals))
		fmt.Fprintf(&output, "Final temperature: %s K\n", formatScientificValue(report.Final.TemperatureKelvin))
	}
	if report.MassOutKilograms != nil {
		fmt.Fprintf(&output, "Mass out: %s kg\n", formatScientificValue(*report.MassOutKilograms))
	}
	if len(report.History) > 0 {
		fmt.Fprintf(&output, "History: %d samples including initial and final state (full values in JSON)\n", len(report.History))
	}
	if report.Signature != nil {
		fmt.Fprintf(&output, "Choked occurred: %t\n", report.Signature.ChokedOccurred)
		if report.Signature.FormationNumber != nil {
			fmt.Fprintf(&output, "L/D: %s\n", formatScientificValue(*report.Signature.FormationNumber))
		}
	}
	if report.Accuracy != nil {
		accuracy := report.Accuracy
		fmt.Fprintf(&output, "Accuracy tolerance satisfied: %t\nDischarge complete: %t\nEvaluations: %d / %d\n",
			accuracy.ToleranceSatisfied, accuracy.DischargeComplete, accuracy.Evaluations, accuracy.MaxEvaluations)
		if accuracy.Estimates != nil {
			fmt.Fprintf(&output, "Estimated time error: %s s (requested %s s)\n",
				formatScientificValue(accuracy.Estimates.TimeErrorSeconds),
				formatScientificValue(accuracy.Estimates.RequestedTimeToleranceSeconds))
		}
	}
	if report.Branch != nil {
		fmt.Fprintf(&output, "Branch stops baseline/variant: %s / %s\n", report.Branch.BaselineStop, report.Branch.VariantStop)
		fmt.Fprintf(&output, "Branch both equalized: %t\n", report.Branch.BothEqualized)
		if report.Branch.BothEqualized {
			fmt.Fprintf(&output, "Branch same mass endpoint: %t\n", report.Branch.SameMassEndpoint)
		}
		fmt.Fprintf(&output, "Branch elapsed baseline/variant: %s / %s s\n",
			formatScientificValue(report.Branch.BaselineElapsedSeconds),
			formatScientificValue(report.Branch.VariantElapsedSeconds))
		fmt.Fprintf(&output, "Branch mass out baseline/variant: %s / %s kg\n",
			formatScientificValue(report.Branch.BaselineMassOutKg), formatScientificValue(report.Branch.VariantMassOutKg))
	}
	if report.Witness != "" {
		fmt.Fprintf(&output, "Witness: %s\n", report.Witness)
	}
	if report.WitnessMatch != nil {
		fmt.Fprintf(&output, "Witness match: %t\n", *report.WitnessMatch)
	}
	if len(report.Explanation) > 0 {
		output.WriteString("\nEXPLANATION\n")
		for _, explanation := range report.Explanation {
			fmt.Fprintf(&output, "  %s\n", explanation)
		}
	}
	if len(report.Claims) > 0 {
		output.WriteString("\nCLAIMS\n")
		for _, claim := range report.Claims {
			fmt.Fprintf(&output, "  %s: %s (residual %s %s; tolerance %s %s)\n", claim.ID, claim.Status,
				formatScientificValue(claim.Residual), claim.ResidualUnit, formatScientificValue(claim.Tolerance), claim.ResidualUnit)
		}
	}
	if len(report.Dimensions) > 0 {
		output.WriteString("\nDIMENSIONS\n")
		for _, dimension := range report.Dimensions {
			fmt.Fprintf(&output, "  %s %s [%s]: %s\n", dimension.Quantity, dimension.Unit, dimension.Dimension, dimension.Status)
		}
	}
	fmt.Fprintf(&output, "\nAssumptions: %s\n", strings.Join(report.Assumptions, ", "))
	if report.Nonclaims != nil {
		fmt.Fprintf(&output, "Model nonclaims: %s\n", strings.Join(report.Nonclaims.Model, ", "))
		fmt.Fprintf(&output, "Operation nonclaims: %s\n", strings.Join(report.Nonclaims.Operation, ", "))
		fmt.Fprintf(&output, "Evidence nonclaims: %s\n", strings.Join(report.Nonclaims.Evidence, ", "))
	}
	output.WriteString("Ambient inputs: none\n")
	output.WriteString(numericPresentationNote)
	return output.String()
}
