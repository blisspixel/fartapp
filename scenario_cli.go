package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/blisspixel/fartapp/internal/scenarioprobe"
)

const scenarioHelp = `F.A.R.T. Lab experimental scenario-document probe

Usage:
  fartapp scenario validate <scenario.json|-> [--format text|json]

Commands:
  validate  Validate a bounded document of law-scoped capability requests.

Help:
  fartapp help scenario validate

The provisional envelope supplies no implicit Earth or other world, source,
body, species, gas, geometry, time, observer, identity, mapping, seed, admission
policy, or solver.

Text output is a current English presentation. Stable IDs and JSON fields are
locale-invariant engineering tokens, not a claim of shared language or meaning.

Example:
  fartapp scenario validate testdata/scenarios/atemporal-probe.json
`

const scenarioValidateHelp = `Validate a bounded document of law-scoped capability requests without realization.

Usage:
  fartapp scenario validate <scenario.json|-> [--format text|json]

Arguments:
  scenario.json  Strict JSON document using fart.scenario-probe/v0alpha1.
  -              Read the document from standard input.

Options:
  --format text  Write a concise text report. This is the default.
  --format json  Write the complete typed report as JSON.
  -h, --help     Show this help.

Limits and effects:
  Input is limited to 65,536 bytes. Validation is read-only: it applies no
  defaults, creates no record or identity, consults no ambient inputs, and
  invokes no admission policy or solver.

Exit status:
  0  The document and all requested capability references are valid.
  1  Usage, input, syntax, schema, law, capability, or detected output failure.
     A native Unix SIGPIPE retains the operating system's pipeline status.

Examples:
  fartapp scenario validate testdata/scenarios/atemporal-probe.json
  fartapp scenario validate testdata/scenarios/atemporal-probe.json --format json

Recovery:
  Correct the first reported diagnostic and run validation again. JSON output
  carries stable code, stage, path, and reason_code fields for automation.
`

func runScenario(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		writeDiagnostic(stderr, "usage: fartapp scenario validate <scenario.json|-> [--format text|json]\n")
		return 1
	}
	if repeatedHelpRequest(args) {
		writeDiagnostic(stderr, "invalid scenario help: --help may be specified only once\n")
		return 1
	}
	if args[0] == "help" {
		if len(args) != 1 {
			writeDiagnostic(stderr, "usage: fartapp scenario help\n")
			return 1
		}
		return writeText(stdout, stderr, scenarioHelp)
	}
	if isHelpRequest(args) {
		return writeText(stdout, stderr, scenarioHelp)
	}
	if args[0] != "validate" {
		writeDiagnostic(stderr, "unknown scenario command %s\n", quoteInput(args[0]))
		return 1
	}
	options, err := parseOutputOptions(args[1:])
	if err != nil {
		writeDiagnostic(stderr, "invalid scenario validate: %v\n", err)
		return 1
	}
	if len(options.positional) > 1 || (!options.help && len(options.positional) != 1) {
		writeDiagnostic(stderr, "usage: fartapp scenario validate <scenario.json|-> [--format text|json]\n")
		return 1
	}
	if options.help {
		return writeText(stdout, stderr, scenarioValidateHelp)
	}

	report := readAndValidateScenario(options.positional[0], stdin)
	if report.Valid() {
		return writeValue(stdout, stderr, options.format, report, formatScenarioReport)
	}
	if options.format == outputJSON {
		if writeValue(stdout, stderr, options.format, report, formatScenarioReport) != 0 {
			return 1
		}
		return 1
	}
	writeScenarioDiagnostic(stderr, report.Diagnostics[0])
	return 1
}

func readAndValidateScenario(source string, stdin io.Reader) scenarioprobe.Report {
	reader := stdin
	consultedInput := "standard_input"
	if source != "-" {
		consultedInput = "named_input"
		opened, err := os.Open(source)
		if err != nil {
			return scenarioInputFailure(classifyScenarioInputError(err), consultedInput)
		}
		reader = opened
		defer func() { _ = opened.Close() }()
	}
	data, err := scenarioprobe.ReadBounded(reader)
	if err != nil {
		if errors.Is(err, scenarioprobe.ErrInputTooLarge) {
			return scenarioprobe.InputFailure(scenarioprobe.Diagnostic{
				Code:       "FART-E-INPUT-0001",
				Stage:      "input",
				Path:       "/",
				ReasonCode: "input_too_large",
			}, consultedInput)
		}
		return scenarioInputFailure(classifyScenarioInputError(err), consultedInput)
	}
	return scenarioprobe.Validate(data)
}

func classifyScenarioInputError(err error) string {
	switch {
	case errors.Is(err, os.ErrNotExist):
		return "input_not_found"
	case errors.Is(err, os.ErrPermission):
		return "input_permission_denied"
	default:
		return "input_unavailable"
	}
}

func scenarioInputFailure(reason, consultedInput string) scenarioprobe.Report {
	return scenarioprobe.InputFailure(scenarioprobe.Diagnostic{
		Code:       "FART-E-IO-0001",
		Stage:      "input",
		Path:       "/",
		ReasonCode: reason,
	}, consultedInput)
}

func writeScenarioDiagnostic(stderr io.Writer, diagnostic scenarioprobe.Diagnostic) {
	if diagnostic.ByteOffset != 0 {
		writeDiagnostic(
			stderr,
			"scenario validation failed: %s %s at %s (byte %d)\n",
			diagnostic.Code,
			diagnostic.ReasonCode,
			strconv.QuoteToASCII(diagnostic.Path),
			diagnostic.ByteOffset,
		)
		return
	}
	writeDiagnostic(
		stderr,
		"scenario validation failed: %s %s at %s\n",
		diagnostic.Code,
		diagnostic.ReasonCode,
		strconv.QuoteToASCII(diagnostic.Path),
	)
}

func formatScenarioReport(report scenarioprobe.Report) string {
	var output strings.Builder
	output.WriteString("SCENARIO PROBE DOCUMENT VALID\n\n")
	fmt.Fprintf(&output, "Document schema: %s\n", report.DocumentSchema)
	fmt.Fprintf(&output, "Law context: %s@%s\n", report.LawContext.ID, report.LawContext.Version)
	fmt.Fprintf(&output, "Scope: %s\n", report.Scope.ID)
	fmt.Fprintf(
		&output,
		"Requested case operation: %s (%s)\n",
		report.RequestedCaseOperation.Selection.Status,
		report.RequestedCaseOperation.Selection.ReasonCode,
	)
	fmt.Fprintf(
		&output,
		"Operation admission: %s (%s)\n",
		report.RequestedCaseOperation.Admission.Status,
		report.RequestedCaseOperation.Admission.ReasonCode,
	)
	fmt.Fprintf(
		&output,
		"Operation execution: %s (%s)\n",
		report.RequestedCaseOperation.Execution.Status,
		report.RequestedCaseOperation.Execution.ReasonCode,
	)
	output.WriteString("Validation stages:\n")
	writeScenarioStage(&output, "syntax", report.ValidationStages.Syntax)
	writeScenarioStage(&output, "schema", report.ValidationStages.Schema)
	writeScenarioStage(&output, "law resolution", report.ValidationStages.LawResolution)
	writeScenarioStage(
		&output,
		"capability resolution",
		report.ValidationStages.CapabilityResolution,
	)
	fmt.Fprintf(
		&output,
		"Validator inputs: %s\n",
		strings.Join(report.Environment.ConsultedInputs, ", "),
	)
	fmt.Fprintf(&output, "Ambient inputs: %s\n", formatStringList(report.Environment.AmbientInputs))
	output.WriteString("\nCAPABILITY REQUESTS\n")
	for _, capability := range report.Capabilities {
		fmt.Fprintf(&output, "\n%s [%s]\n", capability.ID, capability.Resolution)
		writeAssessment(&output, "law definition", capability.LawDefinition)
		writeAssessment(&output, "implementation", capability.Implementation)
		writeAssessment(&output, "closure", capability.Closure)
		writeAssessment(&output, "applicability", capability.Applicability)
		writeAssessment(&output, "evidence", capability.Evidence)
		if len(capability.EvidenceReferences) != 0 {
			fmt.Fprintf(
				&output,
				"  %-20s %s\n",
				"evidence references:",
				strings.Join(capability.EvidenceReferences, ", "),
			)
		}
		writeAssessment(&output, "trust", capability.Trust)
		writeAssessment(&output, "backend feasibility", capability.BackendFeasibility)
		writeAssessment(&output, "resource feasibility", capability.ResourceFeasibility)
	}
	if len(report.EvidenceRegistry) != 0 {
		output.WriteString("\nCAPABILITY EVIDENCE REGISTRY\n")
		for _, evidence := range report.EvidenceRegistry {
			fmt.Fprintf(&output, "%s [%s/%s]\n", evidence.ID, evidence.Scope, evidence.Kind)
			fmt.Fprintf(
				&output,
				"  go test %s -run ^%s$\n",
				evidence.GoPackage,
				evidence.GoTest,
			)
		}
	}
	return output.String()
}

func writeScenarioStage(output *strings.Builder, label string, assessment scenarioprobe.StageAssessment) {
	fmt.Fprintf(output, "  %-23s %s", label+":", assessment.Status)
	if assessment.ReasonCode != "" {
		fmt.Fprintf(output, " (%s)", assessment.ReasonCode)
	}
	output.WriteByte('\n')
}

func formatStringList(values []string) string {
	if len(values) == 0 {
		return "none"
	}
	return strings.Join(values, ", ")
}
