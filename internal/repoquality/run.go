// Package repoquality implements the dependency-free repository policy checkers.
//
// Domain logic lives here so Windows, macOS, Linux, local hooks, and CI share
// one implementation. The thin command in tools/repoquality only forwards
// arguments and exit status.
package repoquality

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		writeDiagnostic(stderr, "usage: repoquality <repository|coverage|fuzz>\n")
		return 1
	}
	switch args[0] {
	case "repository":
		return runRepository(args[1:], stdout, stderr)
	case "coverage":
		return runCoverage(args[1:], stdout, stderr)
	case "fuzz":
		return runFuzz(args[1:], stdout, stderr)
	case "-h", "--help", "help":
		_, _ = io.WriteString(stdout, commandHelp)
		return 0
	default:
		writeDiagnostic(stderr, "unknown repoquality command %s\n", quote(args[0]))
		return 1
	}
}

const commandHelp = `F.A.R.T. Lab repository policy checker

Usage:
  repoquality repository
  repoquality coverage [--profile coverage.out] [--aggregate 90] [--package 80]
  repoquality fuzz [--time 5s]

Commands:
  repository  Check npm and Go dependency policy, local Markdown links, and media.
  coverage    Enforce aggregate and per-package Go statement coverage.
  fuzz        Run the declared Go fuzz targets for a bounded duration.
`

func writeDiagnostic(stderr io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(stderr, format, args...)
}

func quote(value string) string {
	if len(value) > 32 {
		value = value[:32] + "..."
	}
	return strconv.QuoteToASCII(value)
}

func parseDurationArgument(value string) (time.Duration, error) {
	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return 0, fmt.Errorf("fuzz time must be a positive Go duration")
	}
	return duration, nil
}

func joinFailures(failures []string) string {
	return strings.Join(failures, "; ")
}
