package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/blisspixel/fartapp/internal/lawcatalog"
)

type lawOutputFormat uint8

const (
	lawOutputText lawOutputFormat = iota
	lawOutputJSON
)

const lawHelp = `F.A.R.T. Lab candidate law catalog

Usage:
  fartapp law list [--format text|json]
  fartapp law inspect <law-context-id[@version]> [--format text|json]

Commands:
  list     List built-in candidate law contexts.
  inspect  Show one context and all eight capability axes.
`

const lawListHelp = `Usage: fartapp law list [--format text|json]

List built-in candidate law contexts. Localized prose is optional metadata.
`

const lawInspectHelp = `Usage: fartapp law inspect <law-context-id[@version]> [--format text|json]

Inspect one exact law-context revision. An unversioned ID resolves only when the catalog contains one version.
`

func runLaw(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		writeDiagnostic(stderr, "usage: fartapp law <list|inspect> [--format text|json]\n")
		return 1
	}
	if len(args) == 1 && (args[0] == "help" || hasLawHelp(args)) {
		return writeLawText(stdout, stderr, lawHelp)
	}

	switch args[0] {
	case "list":
		if hasLawHelp(args[1:]) {
			return writeLawText(stdout, stderr, lawListHelp)
		}
		positional, format, err := parseLawFormat(args[1:])
		if err != nil {
			writeDiagnostic(stderr, "invalid law list: %v\n", err)
			return 1
		}
		if len(positional) != 0 {
			writeDiagnostic(stderr, "usage: fartapp law list [--format text|json]\n")
			return 1
		}
		return writeLawValue(stdout, stderr, format, lawcatalog.List(), formatLawList)
	case "inspect":
		if hasLawHelp(args[1:]) {
			return writeLawText(stdout, stderr, lawInspectHelp)
		}
		positional, format, err := parseLawFormat(args[1:])
		if err != nil {
			writeDiagnostic(stderr, "invalid law inspect: %v\n", err)
			return 1
		}
		if len(positional) != 1 {
			writeDiagnostic(stderr, "usage: fartapp law inspect <law-context-id[@version]> [--format text|json]\n")
			return 1
		}
		inspection, resolution := lawcatalog.Resolve(positional[0])
		switch resolution {
		case lawcatalog.ResolutionAmbiguous:
			writeDiagnostic(
				stderr,
				"ambiguous law context %s: specify id@version\n",
				quoteInput(positional[0]),
			)
			return 1
		case lawcatalog.ResolutionNotFound:
			writeDiagnostic(stderr, "unknown law context %s\n", quoteInput(positional[0]))
			return 1
		case lawcatalog.ResolutionFound:
		default:
			writeDiagnostic(stderr, "resolve law context: internal resolution error\n")
			return 1
		}
		return writeLawValue(stdout, stderr, format, inspection, formatLawInspection)
	default:
		if hasLawHelp(args[1:]) {
			return writeLawText(stdout, stderr, lawHelp)
		}
		writeDiagnostic(stderr, "unknown law command %s\n", quoteInput(args[0]))
		return 1
	}
}

func hasLawHelp(args []string) bool {
	for _, argument := range args {
		if argument == "-h" || argument == "--help" {
			return true
		}
	}
	return false
}

func writeLawText(stdout, stderr io.Writer, output string) int {
	return writeLawValue(
		stdout,
		stderr,
		lawOutputText,
		output,
		func(value string) string { return value },
	)
}

func parseLawFormat(args []string) ([]string, lawOutputFormat, error) {
	format := lawOutputText
	seenFormat := false
	positional := make([]string, 0, len(args))
	for index := 0; index < len(args); index++ {
		if args[index] != "--format" {
			positional = append(positional, args[index])
			continue
		}
		if seenFormat {
			return nil, 0, fmt.Errorf("--format may be specified only once")
		}
		seenFormat = true
		index++
		if index == len(args) {
			return nil, 0, fmt.Errorf("--format requires text or json")
		}
		switch args[index] {
		case "text":
			format = lawOutputText
		case "json":
			format = lawOutputJSON
		default:
			return nil, 0, fmt.Errorf(
				"unsupported format %s; expected text or json",
				quoteInput(args[index]),
			)
		}
	}
	return positional, format, nil
}

func writeLawValue[T any](
	stdout, stderr io.Writer,
	format lawOutputFormat,
	value T,
	humanFormatter func(T) string,
) int {
	var output []byte
	if format == lawOutputJSON {
		encoded, err := json.Marshal(value)
		if err != nil {
			writeDiagnostic(stderr, "encode output: %v\n", err)
			return 1
		}
		output = append(encoded, '\n')
	} else {
		output = []byte(humanFormatter(value))
	}
	written, err := stdout.Write(output)
	if err == nil && written != len(output) {
		err = io.ErrShortWrite
	}
	if err != nil {
		writeDiagnostic(stderr, "write output: %v\n", err)
		return 1
	}
	return 0
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
