package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/blisspixel/fartapp/internal/reservoirprediction"
)

const reservoirHelp = `F.A.R.T. Lab analytical reservoir predictor

Usage:
  fartapp reservoir predict <request.json|-> [--format text|json]

Commands:
  predict  Calculate one exact rigid ideal-mixture reservoir endpoint.

Help:
  fartapp help reservoir predict

The predictor is a standalone SI continuum model. It supplies no body, world,
exterior, aperture, duration, flow rate, plume, sound, observer, or case
identity. It performs no operation admission or certificate issuance.

Example:
  fartapp reservoir predict testdata/reservoir/synthetic-mixture-adiabatic.json
`

const reservoirPredictHelp = `Predict one rigid ideal-mixture reservoir endpoint.

Usage:
  fartapp reservoir predict <request.json|-> [--format text|json]

Arguments:
  request.json  Strict JSON using fart.reservoir-prediction-request/v0alpha1.
  -             Read the request from standard input.

Options:
  --format text  Write a concise scientific report. This is the default.
  --format json  Write the complete typed prediction report.
  -h, --help     Show this help.

Limits and effects:
  Input is limited to 65,536 bytes. The model is rigid, homogeneous,
  nonreacting, single-phase, calorically perfect, and perfectly mixed. The
  request explicitly supplies every component, SI quantity, closure, and
  withdrawal fraction. Prediction is read-only and commits no case.

Example:
  fartapp reservoir predict testdata/reservoir/synthetic-mixture-adiabatic.json

Exit status:
  0  A prediction was produced.
  1  Usage, input, syntax, schema, model, invariant, or output failure.
     A native Unix SIGPIPE retains the operating system's pipeline status.
`

func runReservoir(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		writeDiagnostic(stderr, "usage: fartapp reservoir predict <request.json|-> [--format text|json]\n")
		return 1
	}
	if repeatedHelpRequest(args) {
		writeDiagnostic(stderr, "invalid reservoir help: --help may be specified only once\n")
		return 1
	}
	if args[0] == "help" {
		if len(args) != 1 {
			writeDiagnostic(stderr, "usage: fartapp reservoir help\n")
			return 1
		}
		return writeText(stdout, stderr, reservoirHelp)
	}
	if isHelpRequest(args) {
		return writeText(stdout, stderr, reservoirHelp)
	}
	if args[0] != "predict" {
		writeDiagnostic(stderr, "unknown reservoir command %s; run 'fartapp help reservoir'\n", quoteInput(args[0]))
		return 1
	}
	options, err := parseOutputOptions(args[1:])
	if err != nil {
		writeDiagnostic(stderr, "invalid reservoir predict: %v\n", err)
		return 1
	}
	if len(options.positional) > 1 || (!options.help && len(options.positional) != 1) {
		writeDiagnostic(stderr, "usage: fartapp reservoir predict <request.json|-> [--format text|json]\n")
		return 1
	}
	if options.help {
		return writeText(stdout, stderr, reservoirPredictHelp)
	}

	report := readReservoirPrediction(options.positional[0], stdin)
	if report.Predicted() {
		return writeValue(stdout, stderr, options.format, report, formatReservoirPrediction)
	}
	if options.format == outputJSON {
		if writeValue(stdout, stderr, options.format, report, formatReservoirPrediction) != 0 {
			return 1
		}
		return 1
	}
	writeReservoirDiagnostic(stderr, report.Diagnostics[0])
	return 1
}

func readReservoirPrediction(source string, stdin io.Reader) reservoirprediction.Report {
	reader := stdin
	if source != "-" {
		opened, err := os.Open(source)
		if err != nil {
			return reservoirprediction.InputFailure(
				classifyReservoirInputError(err),
				"input_source_reference",
			)
		}
		reader = opened
		defer func() { _ = opened.Close() }()
	}
	data, err := reservoirprediction.ReadBounded(reader)
	if err != nil {
		if errors.Is(err, reservoirprediction.ErrInputTooLarge) {
			return reservoirprediction.InputFailure("input_too_large", "input_stream")
		}
		return reservoirprediction.InputFailure("input_unavailable", "input_stream")
	}
	return reservoirprediction.Predict(data)
}

func classifyReservoirInputError(err error) string {
	switch {
	case errors.Is(err, os.ErrNotExist):
		return "input_not_found"
	case errors.Is(err, os.ErrPermission):
		return "input_permission_denied"
	default:
		return "input_unavailable"
	}
}

func writeReservoirDiagnostic(stderr io.Writer, diagnostic reservoirprediction.Diagnostic) {
	writePredictionDiagnostic(stderr, "reservoir predict", diagnostic.Code, diagnostic.ReasonCode, diagnostic.Path)
}

func formatReservoirPrediction(report reservoirprediction.Report) string {
	var output strings.Builder
	output.WriteString("RESERVOIR ENDPOINT PREDICTED\n\n")
	fmt.Fprintf(&output, "Model: %s@%s\n", report.Model.ID, report.Model.Version)
	fmt.Fprintf(&output, "Implementation: %s\n", report.ImplementationRevision)
	fmt.Fprintf(&output, "Quantity system: %s (explicit)\n", report.QuantitySystem)
	fmt.Fprintf(&output, "Closure: %s\n", report.Closure)
	fmt.Fprintf(&output, "Prescribed withdrawal fraction: %s\n", formatScientificValue(*report.WithdrawalFraction))
	writeReservoirState(&output, "INITIAL STATE", *report.Initial)
	writeReservoirState(&output, "FINAL STATE", *report.Final)
	output.WriteString("\nTRANSFERS\n")
	for _, component := range report.Transfers.Components {
		fmt.Fprintf(&output, "  %-20s mass out %s kg\n", component.ID, formatScientificValue(component.MassOutKilograms))
	}
	fmt.Fprintf(&output, "  Total mass out:          %s kg\n", formatScientificValue(report.Transfers.TotalMassOutKilograms))
	fmt.Fprintf(&output, "  Integrated enthalpy out: %s J\n", formatScientificValue(report.Transfers.IntegratedEnthalpyOutJoules))
	fmt.Fprintf(&output, "  Heat into reservoir:     %s J\n", formatScientificValue(report.Transfers.HeatIntoReservoirJoules))
	fmt.Fprintf(&output, "  Boundary work:           %s J\n", formatScientificValue(report.Transfers.BoundaryWorkByReservoirJoules))
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
	output.WriteString(numericPresentationNote)
	return output.String()
}

func writeReservoirState(
	output *strings.Builder,
	title string,
	state reservoirprediction.ReservoirState,
) {
	fmt.Fprintf(output, "\n%s\n", title)
	for _, component := range state.Components {
		fmt.Fprintf(
			output,
			"  %-20s %s kg; R %s J/(kg K); cv %s J/(kg K)\n",
			component.ID,
			formatScientificValue(component.MassKilograms),
			formatScientificValue(component.SpecificGasConstantJoulesPerKilogramKelvin),
			formatScientificValue(component.SpecificIsochoricHeatCapacityJoulesPerKilogramKelvin),
		)
	}
	fmt.Fprintf(output, "  Total mass:      %s kg\n", formatScientificValue(state.TotalMassKilograms))
	fmt.Fprintf(output, "  Volume:          %s m^3\n", formatScientificValue(state.VolumeCubicMetres))
	fmt.Fprintf(output, "  Temperature:     %s K\n", formatScientificValue(state.TemperatureKelvin))
	fmt.Fprintf(output, "  Pressure:        %s Pa\n", formatScientificValue(state.PressurePascals))
	fmt.Fprintf(output, "  Internal energy: %s J\n", formatScientificValue(state.InternalEnergyJoules))
	fmt.Fprintf(output, "  R_mix:           %s J/(kg K)\n", formatScientificValue(state.MixtureGasConstantJoulesPerKilogramKelvin))
	fmt.Fprintf(output, "  cv_mix:          %s J/(kg K)\n", formatScientificValue(state.MixtureSpecificIsochoricHeatJoulesPerKilogramKelvin))
	fmt.Fprintf(output, "  cp_mix:          %s J/(kg K)\n", formatScientificValue(state.MixtureSpecificIsobaricHeatJoulesPerKilogramKelvin))
	fmt.Fprintf(output, "  gamma:           %s\n", formatScientificValue(state.HeatCapacityRatio))
}
