package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/blisspixel/fartapp/internal/restrictionhistoryprediction"
	"github.com/blisspixel/fartapp/internal/restrictionprediction"
)

const restrictionHelp = `F.A.R.T. Lab analytical restriction-flow predictor

Usage:
  fartapp restriction predict <request.json|-> [--format text|json]
  fartapp restriction history <request.json|-> [--format text|json]

Commands:
  predict  Calculate one quasi-steady isentropic converging-restriction state.
  history  Integrate prescribed-area samples over a frozen stagnation state.

Help:
  fartapp help restriction predict
  fartapp help restriction history

The predictor is a standalone SI continuum model. It supplies no body, world,
coupled reservoir blowdown, plume, sound, observer, or case identity. It
performs no operation admission or certificate issuance.

Example:
  fartapp restriction predict testdata/restriction/gamma15-choked.json
  fartapp restriction history testdata/restriction/gamma15-choked-history.json
`

const restrictionPredictHelp = `Predict one quasi-steady isentropic converging-restriction state.

Usage:
  fartapp restriction predict <request.json|-> [--format text|json]

Arguments:
  request.json  Strict JSON using fart.restriction-prediction-request/v0alpha1.
  -             Read the request from standard input.

Options:
  --format text  Write a concise scientific report. This is the default.
  --format json  Write the complete typed prediction report.
  -h, --help     Show this help.

Limits and effects:
  Input is limited to 65,536 bytes. The model is quasi-steady, calorically
  perfect, isentropic to the controlling section, and converging only. Area is
  prescribed or a bounded linear compliance of pressure difference. Prediction
  is read-only and commits no case.

Exit status:
  0  A prediction was produced.
  1  Usage, input, syntax, schema, model, invariant, or output failure.
     A native Unix SIGPIPE retains the operating system's pipeline status.
`

const restrictionHistoryHelp = `Integrate prescribed-area restriction samples over frozen stagnation.

Usage:
  fartapp restriction history <request.json|-> [--format text|json]

Arguments:
  request.json  Strict JSON using fart.restriction-history-request/v0alpha1.
  -             Read the request from standard input.

Options:
  --format text  Write a concise scientific report. This is the default.
  --format json  Write the complete typed history report.
  -h, --help     Show this help.

Limits and effects:
  Input is limited to 65,536 bytes and 256 samples. Times must be nonnegative
  and strictly increasing. Stagnation is frozen; this is not a blowdown.

Exit status:
  0  A history was produced.
  1  Usage, input, syntax, schema, model, invariant, or output failure.
     A native Unix SIGPIPE retains the operating system's pipeline status.
`

func runRestriction(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		writeDiagnostic(stderr, "usage: fartapp restriction <predict|history> <request.json|-> [--format text|json]\n")
		return 1
	}
	if repeatedHelpRequest(args) {
		writeDiagnostic(stderr, "invalid restriction help: --help may be specified only once\n")
		return 1
	}
	if args[0] == "help" {
		if len(args) != 1 {
			writeDiagnostic(stderr, "usage: fartapp restriction help\n")
			return 1
		}
		return writeText(stdout, stderr, restrictionHelp)
	}
	if isHelpRequest(args) {
		return writeText(stdout, stderr, restrictionHelp)
	}
	if args[0] != "predict" && args[0] != "history" {
		writeDiagnostic(stderr, "unknown restriction command %s\n", quoteInput(args[0]))
		return 1
	}
	command := args[0]
	options, err := parseOutputOptions(args[1:])
	if err != nil {
		writeDiagnostic(stderr, "invalid restriction %s: %v\n", command, err)
		return 1
	}
	if len(options.positional) > 1 || (!options.help && len(options.positional) != 1) {
		writeDiagnostic(stderr, "usage: fartapp restriction %s <request.json|-> [--format text|json]\n", command)
		return 1
	}
	if options.help {
		if command == "history" {
			return writeText(stdout, stderr, restrictionHistoryHelp)
		}
		return writeText(stdout, stderr, restrictionPredictHelp)
	}
	if command == "history" {
		return writeRestrictionHistory(options.positional[0], options.format, stdin, stdout, stderr)
	}

	report := readRestrictionPrediction(options.positional[0], stdin)
	if report.Predicted() {
		return writeValue(stdout, stderr, options.format, report, formatRestrictionPrediction)
	}
	if options.format == outputJSON {
		if writeValue(stdout, stderr, options.format, report, formatRestrictionPrediction) != 0 {
			return 1
		}
		return 1
	}
	writeRestrictionDiagnostic(stderr, report.Diagnostics[0])
	return 1
}

func readRestrictionPrediction(source string, stdin io.Reader) restrictionprediction.Report {
	reader := stdin
	if source != "-" {
		opened, err := os.Open(source)
		if err != nil {
			return restrictionprediction.InputFailure(
				classifyRestrictionInputError(err),
				"input_source_reference",
			)
		}
		reader = opened
		defer func() { _ = opened.Close() }()
	}
	data, err := restrictionprediction.ReadBounded(reader)
	if err != nil {
		if errors.Is(err, restrictionprediction.ErrInputTooLarge) {
			return restrictionprediction.InputFailure("input_too_large", "input_stream")
		}
		return restrictionprediction.InputFailure("input_unavailable", "input_stream")
	}
	return restrictionprediction.Predict(data)
}

func classifyRestrictionInputError(err error) string {
	switch {
	case errors.Is(err, os.ErrNotExist):
		return "input_not_found"
	case errors.Is(err, os.ErrPermission):
		return "input_permission_denied"
	default:
		return "input_unavailable"
	}
}

func writeRestrictionDiagnostic(stderr io.Writer, diagnostic restrictionprediction.Diagnostic) {
	writeDiagnostic(
		stderr,
		"restriction prediction failed: %s %s at %s\n",
		diagnostic.Code,
		diagnostic.ReasonCode,
		strconv.QuoteToASCII(diagnostic.Path),
	)
}

func formatRestrictionPrediction(report restrictionprediction.Report) string {
	var output strings.Builder
	output.WriteString("RESTRICTION STATE PREDICTED\n\n")
	fmt.Fprintf(&output, "Model: %s@%s\n", report.Model.ID, report.Model.Version)
	fmt.Fprintf(&output, "Implementation: %s\n", report.ImplementationRevision)
	fmt.Fprintf(&output, "Quantity system: %s (explicit)\n", report.QuantitySystem)
	fmt.Fprintf(&output, "Regime: %s\n", report.Flow.Regime)
	output.WriteString("\nSTAGNATION\n")
	fmt.Fprintf(&output, "  Pressure:     %s Pa\n", formatScientificValue(report.Stagnation.PressurePascals))
	fmt.Fprintf(&output, "  Temperature:  %s K\n", formatScientificValue(report.Stagnation.TemperatureKelvin))
	fmt.Fprintf(&output, "  R:            %s J/(kg K)\n", formatScientificValue(report.Stagnation.SpecificGasConstantJoulesPerKilogramKelvin))
	fmt.Fprintf(&output, "  gamma:        %s\n", formatScientificValue(report.Stagnation.HeatCapacityRatio))
	output.WriteString("\nRESTRICTION\n")
	fmt.Fprintf(&output, "  Area law:     %s\n", report.Area.Law)
	fmt.Fprintf(&output, "  Prescribed:   %s m^2\n", formatScientificValue(report.Area.PrescribedSquareMetres))
	if report.Area.ComplianceSquareMetresPa != nil {
		fmt.Fprintf(&output, "  Compliance:   %s m^2/Pa\n", formatScientificValue(*report.Area.ComplianceSquareMetresPa))
	}
	if report.Area.MaximumSquareMetres != nil {
		fmt.Fprintf(&output, "  Maximum:      %s m^2\n", formatScientificValue(*report.Area.MaximumSquareMetres))
	}
	fmt.Fprintf(&output, "  Effective:    %s m^2\n", formatScientificValue(report.Area.EffectiveSquareMetres))
	fmt.Fprintf(&output, "  Cd:           %s\n", formatScientificValue(*report.DischargeCoefficient))
	fmt.Fprintf(&output, "  Back pressure: %s Pa\n", formatScientificValue(*report.BackPressurePascals))
	output.WriteString("\nCONTROL SECTION\n")
	fmt.Fprintf(&output, "  Critical ratio: %s\n", formatScientificValue(report.Flow.CriticalPressureRatio))
	fmt.Fprintf(&output, "  Back ratio:     %s\n", formatScientificValue(report.Flow.BackPressureRatio))
	fmt.Fprintf(&output, "  Throat Mach:    %s\n", formatScientificValue(report.Flow.ThroatMach))
	fmt.Fprintf(&output, "  Exit pressure:  %s Pa\n", formatScientificValue(report.Flow.ExitPressurePascals))
	fmt.Fprintf(&output, "  Exit temperature: %s K\n", formatScientificValue(report.Flow.ExitTemperatureKelvin))
	fmt.Fprintf(&output, "  Exit speed:     %s m/s\n", formatScientificValue(report.Flow.ExitSpeedMetresPerSecond))
	fmt.Fprintf(&output, "  Mass flow:      %s kg/s\n", formatScientificValue(report.Flow.MassFlowKilogramsPerSecond))
	fmt.Fprintf(&output, "  Sonic mass flow: %s kg/s\n", formatScientificValue(report.Flow.SonicMassFlowKilogramsPerS))
	fmt.Fprintf(&output, "  Thrust:         %s N\n", formatScientificValue(report.Flow.ThrustNewtons))
	fmt.Fprintf(&output, "  Recoil:         %s N\n", formatScientificValue(report.Flow.RecoilNewtons))
	output.WriteString("\nBALANCE CLAIMS\n")
	for _, claim := range report.Claims {
		fmt.Fprintf(
			&output,
			"  %s: %s; residual %s %s; tolerance %s %s\n",
			claim.ID,
			claim.Status,
			formatScientificValue(claim.Residual),
			claim.ResidualUnit,
			formatScientificValue(claim.Tolerance),
			claim.ResidualUnit,
		)
	}
	fmt.Fprintf(&output, "\nAssumptions: %s\n", strings.Join(report.Assumptions, ", "))
	fmt.Fprintf(&output, "Model nonclaims: %s\n", strings.Join(report.Nonclaims.Model, ", "))
	fmt.Fprintf(&output, "Operation nonclaims: %s\n", strings.Join(report.Nonclaims.Operation, ", "))
	fmt.Fprintf(&output, "Evidence nonclaims: %s\n", strings.Join(report.Nonclaims.Evidence, ", "))
	output.WriteString("Ambient inputs: none\n")
	return output.String()
}

func writeRestrictionHistory(
	source string,
	format outputFormat,
	stdin io.Reader,
	stdout, stderr io.Writer,
) int {
	report := readRestrictionHistory(source, stdin)
	if report.Predicted() {
		return writeValue(stdout, stderr, format, report, formatRestrictionHistory)
	}
	if format == outputJSON {
		if writeValue(stdout, stderr, format, report, formatRestrictionHistory) != 0 {
			return 1
		}
		return 1
	}
	writeDiagnostic(
		stderr,
		"restriction history failed: %s %s at %s\n",
		report.Diagnostics[0].Code,
		report.Diagnostics[0].ReasonCode,
		strconv.QuoteToASCII(report.Diagnostics[0].Path),
	)
	return 1
}

func readRestrictionHistory(source string, stdin io.Reader) restrictionhistoryprediction.Report {
	reader := stdin
	if source != "-" {
		opened, err := os.Open(source)
		if err != nil {
			return restrictionhistoryprediction.InputFailure(
				classifyRestrictionInputError(err),
				"input_source_reference",
			)
		}
		reader = opened
		defer func() { _ = opened.Close() }()
	}
	data, err := restrictionhistoryprediction.ReadBounded(reader)
	if err != nil {
		if errors.Is(err, restrictionhistoryprediction.ErrInputTooLarge) {
			return restrictionhistoryprediction.InputFailure("input_too_large", "input_stream")
		}
		return restrictionhistoryprediction.InputFailure("input_unavailable", "input_stream")
	}
	return restrictionhistoryprediction.Predict(data)
}

func formatRestrictionHistory(report restrictionhistoryprediction.Report) string {
	var output strings.Builder
	output.WriteString("RESTRICTION HISTORY PREDICTED\n\n")
	fmt.Fprintf(&output, "Model: %s@%s\n", report.Model.ID, report.Model.Version)
	fmt.Fprintf(&output, "Implementation: %s\n", report.ImplementationRevision)
	fmt.Fprintf(&output, "Samples: %d\n", len(report.Samples))
	output.WriteString("\nSAMPLES\n")
	for _, sample := range report.Samples {
		fmt.Fprintf(
			&output,
			"  t %s s; A %s m^2; %s; mdot %s kg/s; thrust %s N; recoil %s N\n",
			formatScientificValue(sample.TimeSeconds),
			formatScientificValue(sample.EffectiveSquareMetres),
			sample.Regime,
			formatScientificValue(sample.MassFlowKilogramsPerS),
			formatScientificValue(sample.ThrustNewtons),
			formatScientificValue(sample.RecoilNewtons),
		)
	}
	output.WriteString("\nINTEGRALS\n")
	fmt.Fprintf(&output, "  Mass out:       %s kg\n", formatScientificValue(report.Totals.MassOutKilograms))
	fmt.Fprintf(&output, "  Static exit enthalpy: %s J\n", formatScientificValue(report.Totals.EnthalpyOutJoules))
	fmt.Fprintf(&output, "  Exit kinetic energy: %s J\n", formatScientificValue(report.Totals.KineticEnergyOutJoules))
	fmt.Fprintf(&output, "  Total enthalpy out:   %s J\n", formatScientificValue(report.Totals.TotalEnthalpyOutJoules))
	fmt.Fprintf(&output, "  Impulse:        %s N s\n", formatScientificValue(report.Totals.ImpulseNewtonSeconds))
	fmt.Fprintf(&output, "  Recoil impulse: %s N s\n", formatScientificValue(report.Totals.RecoilImpulseNewtonSeconds))
	output.WriteString("\nBALANCE CLAIMS\n")
	for _, claim := range report.Claims {
		fmt.Fprintf(
			&output,
			"  %s: %s; residual %s %s; tolerance %s %s\n",
			claim.ID,
			claim.Status,
			formatScientificValue(claim.Residual),
			claim.ResidualUnit,
			formatScientificValue(claim.Tolerance),
			claim.ResidualUnit,
		)
	}
	fmt.Fprintf(&output, "\nAssumptions: %s\n", strings.Join(report.Assumptions, ", "))
	fmt.Fprintf(&output, "Model nonclaims: %s\n", strings.Join(report.Nonclaims.Model, ", "))
	output.WriteString("Ambient inputs: none\n")
	return output.String()
}
