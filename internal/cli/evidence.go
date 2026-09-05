package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/blisspixel/fartapp/internal/walkcase"
	"github.com/blisspixel/fartapp/internal/walkevidence"
)

const evidenceHelp = `Retain and inspect an experimental walk evidence carrier.

Usage:
  fartapp evidence capture <case.json|-> --output <record.fartevidence> [--format text|json]
  fartapp evidence <inspect|verify|replay|reconstruct> <record.fartevidence|-> [--format text|json]

Operations:
  capture      Calculate one witness and publish its request and report atomically.
  inspect      Show member hashes, normalized input binding, and evidence limits.
  verify       Check integrity and account self-consistency without running a solver.
  replay       Emit the retained JSON report with a final newline. No solver runs.
  reconstruct  Calculate again against the retained witness and report a mismatch.

The output directory must exist. Capture never replaces an existing destination;
the filesystem must support hard links. A failed stdout write can occur after a
valid archive has been committed. Inspect that file before retrying.

Replay defaults to JSON and accepts only --format json. Other operations default
to English text. Input is bounded to 64 KiB for cases and 24 MiB for carriers.
Carriers contain exact authored request bytes and a compact complete report.
They are uncompressed JSON with two fixed base64 members, not extracted files.

Integrity does not establish authenticity, scientific validation, or a certified
case archive. Reconstruction requires a matching implementation/runtime profile.
This provisional .fartevidence format is separate from the planned .fart format.

Exit status:
  0  The requested operation completed.
  1  Usage, input, integrity, reconstruction, filesystem, or output failure.
`

func runEvidence(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		writeDiagnostic(stderr, "usage: fartapp evidence <capture|inspect|verify|replay|reconstruct>\n")
		return 1
	}
	if len(args) == 1 && (isHelpRequest(args) || args[0] == "help") {
		return writeText(stdout, stderr, evidenceHelp)
	}
	operation := args[0]
	if !evidenceOperation(operation) {
		writeDiagnostic(stderr, "unknown evidence command %s\n", quoteInput(operation))
		return 1
	}
	options, destination, err := evidenceOptions(args[1:], operation)
	if err != nil {
		writeDiagnostic(stderr, "invalid evidence %s: %v\n", operation, err)
		return 1
	}
	if options.help {
		return writeText(stdout, stderr, evidenceHelp)
	}
	var source io.Reader = stdin
	if options.positional[0] != "-" {
		file, err := os.Open(options.positional[0])
		if err != nil {
			return evidenceFailure(options.format, operation, err, stdout, stderr)
		}
		defer file.Close()
		source = file
	}
	var record walkevidence.Record
	if operation == "capture" {
		request, err := walkcase.ReadBounded(source)
		if err != nil {
			return evidenceFailure(options.format, operation, err, stdout, stderr)
		}
		record, err = walkevidence.Capture(request)
		if err == nil {
			err = walkevidence.WriteNew(context.Background(), destination, record)
		}
		if err != nil {
			return evidenceFailure(options.format, operation, err, stdout, stderr)
		}
	} else {
		record, err = walkevidence.Read(source)
		if err != nil {
			return evidenceFailure(options.format, operation, err, stdout, stderr)
		}
	}
	if operation == "replay" {
		return writeText(stdout, stderr, string(record.Replay())+"\n")
	}
	if operation == "reconstruct" {
		report := record.Reconstruct()
		code := writeValue(stdout, stderr, options.format, report, formatWalk)
		if !report.Predicted() {
			return 1
		}
		return code
	}
	return writeValue(stdout, stderr, options.format, record.Summary(), formatEvidence)
}

func evidenceOperation(operation string) bool {
	switch operation {
	case "capture", "inspect", "verify", "replay", "reconstruct":
		return true
	default:
		return false
	}
}

func evidenceOptions(args []string, operation string) (outputOptions, string, error) {
	options, values, err := parseValuedOptions(args, "--output")
	if err != nil {
		return outputOptions{}, "", err
	}
	destination, seenOutput := values["--output"]
	if len(options.positional) > 1 || (!options.help && len(options.positional) != 1) {
		return outputOptions{}, "", errors.New("exactly one input path or - is required")
	}
	if seenOutput && operation != "capture" {
		return outputOptions{}, "", errors.New("--output is only valid for capture")
	}
	if operation == "capture" && !options.help && destination == "" {
		return outputOptions{}, "", errors.New("capture requires --output with an explicit destination")
	}
	if operation == "replay" {
		if options.formatSpecified && options.format != outputJSON {
			return outputOptions{}, "", errors.New("replay supports only --format json")
		}
		options.format = outputJSON
	}
	return options, destination, nil
}

func evidenceFailure(format outputFormat, operation string, err error, stdout, stderr io.Writer) int {
	reason := "input_or_filesystem_failure"
	switch {
	case errors.Is(err, walkevidence.ErrInvalidArchive):
		reason = "invalid_archive"
	case errors.Is(err, walkevidence.ErrInvalidRequest):
		reason = "invalid_capture_request"
	case errors.Is(err, walkevidence.ErrDestinationExists):
		reason = "destination_exists"
	case errors.Is(err, walkevidence.ErrTooLarge), errors.Is(err, walkcase.ErrInputTooLarge):
		reason = "input_too_large"
	}
	if format == outputJSON {
		value := struct {
			Schema    string `json:"schema"`
			Status    string `json:"status"`
			Operation string `json:"operation"`
			Reason    string `json:"reason_code"`
		}{"fart.evidence-operation/v0alpha1", "invalid", operation, reason}
		_ = writeValue[any](stdout, stderr, outputJSON, value, func(any) string { return "" })
	} else {
		writeDiagnostic(stderr, "evidence %s failed: %s: %s\n", operation, reason, quoteInput(err.Error()))
	}
	return 1
}

func formatEvidence(value walkevidence.Summary) string {
	return fmt.Sprintf("Evidence: %s\nArtifact SHA-256: %s\nRequest: %d bytes; SHA-256 %s\nReport: %d bytes; SHA-256 %s\nWitness: %s\nNormalized input binding: %s\nAccount witness consistency: %s\nSolver reexecution: %s\nAuthenticity: %s\nScientific validation: %s\n",
		value.Status, value.ArtifactSHA256, value.RequestBytes, value.RequestSHA256, value.ReportBytes, value.ReportSHA256,
		value.Witness, value.InputBinding, value.AccountConsistency, value.SolverReexecution, value.Authenticity, value.ScientificValidation)
}
