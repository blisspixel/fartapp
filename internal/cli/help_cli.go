package cli

import (
	"io"
	"strings"
)

const rootHelp = `F.A.R.T. Lab

Experimental law inspection, scenario validation, and analytical prediction.

Usage:
  fartapp <command> [arguments]
  fartapp <intensity>
  fartapp help [command [operation]]

Available now:
  <intensity>  Run the permanent v0.6 toy output, levels 1 to 5.
  law          List contexts and inspect capability axes.
  scenario     Validate a document's law and capability references.
  reservoir    Predict a rigid ideal-mixture reservoir endpoint.
  restriction  Predict flow or integrate prescribed-area samples.
  walk         Explore coupled blowdown, compare an area, or refine its clock.
  evidence     Capture, inspect, verify, replay, and reconstruct an account.
  assurance    Inspect declared invariants and evidence references.
  help         Show contextual command help.

Predictors require explicit SI inputs and a named model. The scenario probe
checks references; it executes no mapping or realization. Help and reports
use English; stable IDs and JSON fields retain their exact spelling.

Examples:
  fartapp 3
  fartapp walk explain testdata/walk/ordinary-low-pressure.json
  fartapp help walk branch

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
	case len(args) == 1 && args[0] == "reservoir":
		return reservoirHelp, true
	case len(args) == 2 && args[0] == "reservoir" && args[1] == "predict":
		return reservoirPredictHelp, true
	case len(args) == 1 && args[0] == "restriction":
		return restrictionHelp, true
	case len(args) == 2 && args[0] == "restriction" && args[1] == "predict":
		return restrictionPredictHelp, true
	case len(args) == 2 && args[0] == "restriction" && args[1] == "history":
		return restrictionHistoryHelp, true
	case len(args) == 1 && args[0] == "walk":
		return walkHelp, true
	case len(args) == 1 && args[0] == "evidence":
		return evidenceHelp, true
	case len(args) == 1 && args[0] == "assurance":
		return assuranceHelp, true
	case len(args) == 2 && args[0] == "assurance" && (args[1] == "list" || args[1] == "inspect"):
		return assuranceHelp, true
	case len(args) == 2 && args[0] == "evidence" && evidenceOperation(args[1]):
		return evidenceHelp, true
	case len(args) == 2 && args[0] == "walk" && args[1] == "refine":
		return refineHelp, true
	case len(args) == 2 && args[0] == "walk":
		if _, ok := walkOperations[args[1]]; ok {
			return walkHelpForOperation(args[1]), true
		}
		return "", false
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
