package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
)

const maximumDisplayedInputBytes = 32

func run(args []string, stdout, stderr io.Writer) int {
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
	os.Exit(run(os.Args, os.Stdout, os.Stderr))
}
