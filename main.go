package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const maximumDisplayedInputBytes = 32

func run(args []string, stdout, stderr io.Writer) int {
	return runWithInput(args, strings.NewReader(""), stdout, stderr)
}

func runWithInput(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) >= 2 && repeatedHelpRequest(args[1:]) {
		writeDiagnostic(stderr, "invalid help: --help may be specified only once\n")
		return 1
	}
	if len(args) == 2 && isHelpRequest(args[1:]) {
		return writeText(stdout, stderr, rootHelp)
	}
	if len(args) >= 2 && args[1] == "help" {
		return runHelp(args[2:], stdout, stderr)
	}
	if len(args) >= 2 && args[1] == "law" {
		return runLaw(args[2:], stdout, stderr)
	}
	if len(args) >= 2 && args[1] == "scenario" {
		return runScenario(args[2:], stdin, stdout, stderr)
	}

	if len(args) != 2 {
		writeDiagnostic(stderr, "usage: fartapp <intensity>\n")
		return 1
	}

	rawIntensity := args[1]
	value, err := strconv.Atoi(rawIntensity)
	if err != nil {
		writeDiagnostic(
			stderr,
			"invalid intensity %s: must be a canonical integer from %d to %d\n",
			quoteInput(rawIntensity),
			minimumIntensity,
			maximumIntensity,
		)
		return 1
	}
	if strconv.Itoa(value) != rawIntensity {
		writeDiagnostic(
			stderr,
			"invalid intensity %s: must be a canonical integer from %d to %d\n",
			quoteInput(rawIntensity),
			minimumIntensity,
			maximumIntensity,
		)
		return 1
	}

	level, err := newIntensity(value)
	if err != nil {
		writeDiagnostic(stderr, "invalid %v\n", err)
		return 1
	}

	event := level.emission()
	if _, err := fmt.Fprintf(stdout, "%s (%s)\n", event.sound, event.rating); err != nil {
		writeDiagnostic(stderr, "write output: %v\n", err)
		return 1
	}
	return 0
}

func quoteInput(value string) string {
	if len(value) > maximumDisplayedInputBytes {
		value = value[:maximumDisplayedInputBytes] + "..."
	}
	return strconv.QuoteToASCII(value)
}

func writeDiagnostic(stderr io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(stderr, format, args...)
}

func main() {
	os.Exit(runWithInput(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
