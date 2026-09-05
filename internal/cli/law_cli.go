package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/blisspixel/fartapp/internal/lawcatalog"
)

const lawHelp = `F.A.R.T. Lab candidate law catalog

Usage:
  fartapp law list [--format text|json]
  fartapp law inspect <law-context-id[@version]> [--format text|json]

Commands:
  list     List built-in candidate law contexts.
  inspect  Show one context and all eight capability axes.

Help:
  fartapp help law list
  fartapp help law inspect

Inspection is read-only and invokes no solver or realization. Text output is a
current English presentation; stable IDs and JSON fields are locale-invariant
engineering tokens, not a claim of shared language or meaning.

Examples:
  fartapp law list
  fartapp law inspect conformance.relation.atemporal@v0alpha1
`

const lawListHelp = `List built-in candidate law contexts.

Usage:
  fartapp law list [--format text|json]

Options:
  --format text  Write concise text with optional English presentation metadata.
  --format json  Write the complete locale-invariant typed catalog listing.
  -h, --help     Show this help.

Examples:
  fartapp law list
  fartapp law list --format json

Exit status:
  0  The built-in catalog was written successfully.
  1  Usage, option, encoding, or detected output failure.

Next:
  Run 'fartapp law inspect <id@version>' to inspect every capability axis.
`

const lawInspectHelp = `Inspect one candidate law-context revision and all eight capability axes.

Usage:
  fartapp law inspect <law-context-id[@version]> [--format text|json]

Arguments:
  law-context-id  Exact locale-invariant ID. Add @version for an exact revision.

Options:
  --format text  Write concise text with optional English presentation metadata.
  --format json  Write the complete locale-invariant typed inspection.
  -h, --help     Show this help.

Resolution:
  An unversioned ID resolves only when the built-in catalog contains one
  matching version. Inspection does not invoke a solver or realization.

Examples:
  fartapp law inspect conformance.relation.atemporal@v0alpha1
  fartapp law inspect earth.continuum.si --format json

Exit status:
  0  The requested catalog entry was written successfully.
  1  Usage, option, resolution, encoding, or detected output failure.

Recovery:
  Run 'fartapp law list' to discover exact IDs and versions.
`

func runLaw(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		writeDiagnostic(stderr, "usage: fartapp law <list|inspect> [--format text|json]\n")
		return 1
	}
	if repeatedHelpRequest(args) {
		writeDiagnostic(stderr, "invalid law help: --help may be specified only once\n")
		return 1
	}
	if args[0] == "help" {
		if len(args) != 1 {
			writeDiagnostic(stderr, "usage: fartapp law help\n")
			return 1
		}
		return writeText(stdout, stderr, lawHelp)
	}
	if isHelpRequest(args) {
		return writeText(stdout, stderr, lawHelp)
	}

	switch args[0] {
	case "list":
		options, err := parseOutputOptions(args[1:])
		if err != nil {
			writeDiagnostic(stderr, "invalid law list: %v\n", err)
			return 1
		}
		if len(options.positional) != 0 {
			writeDiagnostic(stderr, "usage: fartapp law list [--format text|json]\n")
			return 1
		}
		if options.help {
			return writeText(stdout, stderr, lawListHelp)
		}
		return writeValue(stdout, stderr, options.format, lawcatalog.List(), formatLawList)
	case "inspect":
		options, err := parseOutputOptions(args[1:])
		if err != nil {
			writeDiagnostic(stderr, "invalid law inspect: %v\n", err)
			return 1
		}
		if len(options.positional) > 1 || (!options.help && len(options.positional) != 1) {
			writeDiagnostic(stderr, "usage: fartapp law inspect <law-context-id[@version]> [--format text|json]\n")
			return 1
		}
		if options.help {
			return writeText(stdout, stderr, lawInspectHelp)
		}
		inspection, resolution := lawcatalog.Resolve(options.positional[0])
		switch resolution {
		case lawcatalog.ResolutionAmbiguous:
			writeDiagnostic(
				stderr,
				"ambiguous law context %s: specify id@version\n",
				quoteInput(options.positional[0]),
			)
			return 1
		case lawcatalog.ResolutionNotFound:
			writeDiagnostic(stderr, "unknown law context %s\n", quoteInput(options.positional[0]))
			return 1
		case lawcatalog.ResolutionFound:
		default:
			writeDiagnostic(stderr, "resolve law context: internal resolution error\n")
			return 1
		}
		return writeValue(stdout, stderr, options.format, inspection, formatLawInspection)
	default:
		writeDiagnostic(stderr, "unknown law command %s\n", quoteInput(args[0]))
		return 1
	}
}

func formatLawList(document lawcatalog.ListDocument) string {
	var output strings.Builder
	output.WriteString("LAW CONTEXTS\n\n")
	for _, context := range document.LawContexts {
		fmt.Fprintf(&output, "%s@%s [%s]\n", context.ID, context.Version, context.Maturity)
		presentation, found := selectPresentation(context.Presentations)
		if found && presentation.Name != "" {
			fmt.Fprintf(&output, "  %s\n", presentation.Name)
		}
		if found && presentation.Description != "" {
			fmt.Fprintf(&output, "  %s\n", presentation.Description)
		}
	}
	return output.String()
}

func formatLawInspection(inspection lawcatalog.Inspection) string {
	context := inspection.LawContext
	var output strings.Builder
	output.WriteString("LAW CONTEXT\n\n")
	fmt.Fprintf(&output, "ID: %s\n", context.ID)
	fmt.Fprintf(&output, "Version: %s\n", context.Version)
	fmt.Fprintf(&output, "Maturity: %s\n", context.Maturity)
	if presentation, found := selectPresentation(context.Presentations); found {
		if presentation.Name != "" {
			fmt.Fprintf(&output, "Name: %s\n", presentation.Name)
		}
		if presentation.Description != "" {
			fmt.Fprintf(&output, "Description: %s\n", presentation.Description)
		}
		fmt.Fprintf(&output, "Presentation locale: %s\n", presentation.Locale)
	}
	if len(context.StructuralModules) != 0 {
		fmt.Fprintf(&output, "Structural modules: %s\n", joinModuleIDs(context.StructuralModules))
	}
	if len(context.ExtensionRoles) != 0 {
		fmt.Fprintf(&output, "Context extension roles: %s\n", strings.Join(context.ExtensionRoles, ", "))
	}
	output.WriteString("\nCAPABILITY REPORT\n")
	fmt.Fprintf(
		&output,
		"Law context: %s@%s\n",
		inspection.CapabilityReport.LawContext.ID,
		inspection.CapabilityReport.LawContext.Version,
	)
	for _, capability := range inspection.CapabilityReport.Capabilities {
		fmt.Fprintf(&output, "\n%s\n", capability.ID)
		if presentation, found := selectPresentation(capability.Presentations); found &&
			presentation.Description != "" {
			fmt.Fprintf(&output, "  %s\n", presentation.Description)
		}
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
	if len(inspection.CapabilityReport.EvidenceRegistry) != 0 {
		output.WriteString("\nEVIDENCE REGISTRY\n")
		for _, evidence := range inspection.CapabilityReport.EvidenceRegistry {
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

func selectPresentation(
	presentations []lawcatalog.LocalizedPresentation,
) (lawcatalog.LocalizedPresentation, bool) {
	for _, presentation := range presentations {
		if presentation.Locale == "en" {
			return presentation, true
		}
	}
	if len(presentations) == 0 {
		return lawcatalog.LocalizedPresentation{}, false
	}
	return presentations[0], true
}

func joinModuleIDs(modules []lawcatalog.StructuralModule) string {
	ids := make([]string, len(modules))
	for index, module := range modules {
		ids[index] = module.ID
	}
	return strings.Join(ids, ", ")
}

func writeAssessment(output *strings.Builder, label string, assessment lawcatalog.Assessment) {
	fmt.Fprintf(output, "  %-20s %s", label+":", assessment.Status)
	if assessment.ReasonCode != "" {
		fmt.Fprintf(output, " (%s)", assessment.ReasonCode)
	}
	output.WriteByte('\n')
}
