package cli

import (
	"io"

	"github.com/blisspixel/fartapp/internal/assurance"
)

const assuranceHelp = `Inspect declared engineering invariants and their evidence references.

Usage:
  fartapp assurance list [--format text|json]
  fartapp assurance inspect <invariant-id> [--format text|json]

Examples:
  fartapp assurance list --format json
  fartapp assurance inspect PHY-001
  fartapp help assurance inspect

The embedded registry is metadata. Inspection runs no test, solver, shell,
network request, or filesystem evidence check. An executable candidate declares
checks; it does not report that those checks passed or that an invariant applies
to a supplied case. Design candidates remain planned obligations.

Invariant IDs and related benchmark IDs belong to separate namespaces.
Current text is an English presentation. JSON retains versioned metadata.

Exit status:
  0  The requested metadata or help was written.
  1  Usage, unknown invariant, registry, or output failure.
`

func runAssurance(args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && (args[0] == "help" || isHelpRequest(args)) {
		return writeText(stdout, stderr, assuranceHelp)
	}
	if len(args) == 0 || (args[0] != "list" && args[0] != "inspect") {
		writeDiagnostic(stderr, "usage: fartapp assurance <list|inspect> [--format text|json]\n")
		return 1
	}
	operation := args[0]
	options, err := parseOutputOptions(args[1:])
	if err != nil {
		writeDiagnostic(stderr, "invalid assurance %s: %v\n", operation, err)
		return 1
	}
	if operation == "list" && len(options.positional) != 0 ||
		operation == "inspect" && (len(options.positional) > 1 || !options.help && len(options.positional) != 1) {
		writeDiagnostic(stderr, "invalid assurance %s arguments; run 'fartapp help assurance'\n", operation)
		return 1
	}
	if options.help {
		return writeText(stdout, stderr, assuranceHelp)
	}
	registry, err := assurance.BuiltIn()
	if err != nil {
		writeDiagnostic(stderr, "invalid embedded assurance registry\n")
		return 1
	}
	var output string
	if operation == "list" {
		if options.format == outputJSON {
			var encoded []byte
			encoded, err = registry.ListJSON()
			output = string(encoded) + "\n"
		} else {
			output = registry.ListText()
		}
	} else {
		id := options.positional[0]
		if options.format == outputJSON {
			var encoded []byte
			encoded, err = registry.InspectJSON(id)
			output = string(encoded) + "\n"
		} else {
			output, err = registry.InspectText(id)
		}
	}
	if err != nil {
		writeDiagnostic(stderr, "assurance inspection failed; use 'fartapp assurance list' for known invariant IDs\n")
		return 1
	}
	return writeText(stdout, stderr, output)
}
