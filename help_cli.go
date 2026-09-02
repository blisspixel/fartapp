package main

import (
	"io"
	"strings"
)

const rootHelp = `F.A.R.T. Lab

Experimental law-context inspection and scenario-document validation.

Usage:
  fartapp <intensity>
  fartapp law <list|inspect>
  fartapp scenario validate <scenario.json|->
  fartapp help [law [list|inspect]|scenario [validate]]

Available now:
  <intensity>         Run the permanent v0.6 legacy string oracle from 1 to 5.
  law list           List built-in candidate law contexts.
  law inspect        Inspect one law-context revision and its capability axes.
  scenario validate  Validate a bounded document of law-scoped capability requests.
  help               Show root or command help.

This build supplies no implicit Earth or other world, source, body, species,
gas, geometry, time, observer, identity, mapping, admission policy, or solver.
It has no realization engine, TUI, native application, updater, or
natural-language generator. "Fart" is the comic umbrella, not the scientific
ontology. The current probe resolves declared law contexts and capability
references only; it executes no mapping or realization.

Text help and reports are currently English presentations. Stable IDs and JSON
fields are locale-invariant engineering tokens; this does not assert shared
language or meaning.

Examples:
  fartapp 3
  fartapp law list
  fartapp law inspect conformance.relation.atemporal@v0alpha1
  fartapp scenario validate testdata/scenarios/atemporal-probe.json
  fartapp help scenario validate

Exit status:
  0  Help or a successful command.
  1  A controlled command, input, validation, or detected output failure.
     A native Unix SIGPIPE retains the operating system's pipeline status.
`

func runHelp(args []string, stdout, stderr io.Writer) int {
	if output, found := helpOutput(args); found {
		return writeText(stdout, stderr, output)
	}
	writeDiagnostic(
		stderr,
		"unknown help topic %s; run 'fartapp help' for available topics\n",
		quoteInput(joinBounded(args)),
	)
	return 1
}

func helpOutput(args []string) (string, bool) {
	switch {
	case len(args) == 0:
		return rootHelp, true
	case len(args) == 1 && args[0] == "law":
		return lawHelp, true
	case len(args) == 2 && args[0] == "law" && args[1] == "list":
		return lawListHelp, true
	case len(args) == 2 && args[0] == "law" && args[1] == "inspect":
		return lawInspectHelp, true
	case len(args) == 1 && args[0] == "scenario":
		return scenarioHelp, true
	case len(args) == 2 && args[0] == "scenario" && args[1] == "validate":
		return scenarioValidateHelp, true
	default:
		return "", false
	}
}

func joinBounded(args []string) string {
	var joined strings.Builder
	const retainedBytes = maximumDisplayedInputBytes + 1
	for index, argument := range args {
		if index != 0 && joined.Len() < retainedBytes {
			joined.WriteByte(' ')
		}
		remaining := retainedBytes - joined.Len()
		if remaining <= 0 {
			return joined.String()
		}
		if len(argument) > remaining {
			joined.WriteString(argument[:remaining])
			return joined.String()
		}
		joined.WriteString(argument)
	}
	return joined.String()
}

func isHelpRequest(args []string) bool {
	return len(args) == 1 && isHelpToken(args[0])
}

func repeatedHelpRequest(args []string) bool {
	if len(args) < 2 {
		return false
	}
	for _, argument := range args {
		if !isHelpToken(argument) {
			return false
		}
	}
	return true
}

func isHelpToken(argument string) bool {
	return argument == "-h" || argument == "--help"
}
