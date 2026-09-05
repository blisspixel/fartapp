package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/blisspixel/fartapp/internal/walkcase"
)

func TestWalkCLISimulateAndFailures(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run(
		[]string{"fartapp", "walk", "simulate", filepath.FromSlash("testdata/walk/isothermal-choked.json")},
		strings.NewReader(""),
		&stdout,
		&stderr,
	)
	if code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), "WALK SIMULATE") ||
		!strings.Contains(stdout.String(), "Choked occurred:") {
		t.Fatalf("simulate = (%d, %q, %q)", code, stdout.String(), stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run(
		[]string{"fartapp", "walk", "reconstruct", filepath.FromSlash("testdata/walk/isothermal-choked.json")},
		strings.NewReader(""),
		&stdout,
		&stderr,
	)
	if code != 1 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "missing_member at \"/expected_witness\"") {
		t.Fatalf("reconstruct = (%d, %q, %q)", code, stdout.String(), stderr.String())
	}

	tests := []struct {
		args []string
		want string
	}{
		{args: []string{"fartapp", "walk"}, want: "usage: fartapp walk <predict|simulate|inspect|explain|branch|certify|witness|reconstruct> <case.json|-> [--format text|json]\n"},
		{args: []string{"fartapp", "walk", "dance"}, want: "unknown walk command \"dance\"\n"},
		{args: []string{"fartapp", "walk", "simulate"}, want: "usage: fartapp walk simulate <case.json|-> [--format text|json]\n"},
		{args: []string{"fartapp", "walk", "simulate", "missing.json"}, want: "walk simulate failed: FART-E-IO-0005 input_not_found at \"/\"\n"},
	}
	for _, test := range tests {
		stdout.Reset()
		stderr.Reset()
		if code := Run(test.args, strings.NewReader(""), &stdout, &stderr); code != 1 ||
			stdout.Len() != 0 || stderr.String() != test.want {
			t.Fatalf("run(%q) = (%d, %q, %q), want %q", test.args, code, stdout.String(), stderr.String(), test.want)
		}
	}
	if got := classifyWalkInputError(os.ErrPermission); got != "input_permission_denied" {
		t.Fatalf("permission = %q", got)
	}
	if got := classifyWalkInputError(errors.New("other")); got != "input_unavailable" {
		t.Fatalf("other = %q", got)
	}
}

func TestWalkCLIAllOperationsAndRetainedReconstruction(t *testing.T) {
	input, err := os.ReadFile(filepath.FromSlash("testdata/walk/isothermal-choked.json"))
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"fartapp", "walk", "witness", "-", "--format", "json"}, bytes.NewReader(input), &stdout, &stderr); code != 0 || stderr.Len() != 0 {
		t.Fatalf("witness = (%d, %q)", code, stderr.String())
	}
	var witness walkcase.Report
	if err := json.Unmarshal(stdout.Bytes(), &witness); err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(input, &document); err != nil {
		t.Fatal(err)
	}
	document["expected_witness"] = witness.Witness
	retained, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	for _, operation := range []string{"predict", "simulate", "inspect", "explain", "branch", "certify", "witness", "reconstruct"} {
		for _, format := range []string{"text", "json"} {
			t.Run(operation+"/"+format, func(t *testing.T) {
				stdout.Reset()
				stderr.Reset()
				code := Run([]string{"fartapp", "walk", operation, "-", "--format", format}, bytes.NewReader(retained), &stdout, &stderr)
				if code != 0 || stderr.Len() != 0 {
					t.Fatalf("operation = (%d, %q)", code, stderr.String())
				}
				if format == "json" {
					var report walkcase.Report
					if err := json.Unmarshal(stdout.Bytes(), &report); err != nil || !report.Predicted() || report.Operation != operation {
						t.Fatalf("report = %s; %v", stdout.String(), err)
					}
					if operation == "reconstruct" && (report.WitnessMatch == nil || !*report.WitnessMatch || report.ExpectedWitness != witness.Witness) {
						t.Fatal("retained CLI witness was not compared")
					}
				} else if !strings.Contains(stdout.String(), "residual ") || !strings.Contains(stdout.String(), "Evidence nonclaims:") {
					t.Fatal("text omitted balance evidence or its limits")
				}
			})
		}
	}
	document["expected_witness"] = strings.Repeat("0", 64)
	mismatch, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	code := Run([]string{"fartapp", "walk", "reconstruct", "-", "--format", "json"}, bytes.NewReader(mismatch), &stdout, &stderr)
	var report walkcase.Report
	if err := json.Unmarshal(stdout.Bytes(), &report); code != 1 || err != nil || report.Status != "mismatch" || report.ExpectedWitness != strings.Repeat("0", 64) || report.ReconstructedWitness != witness.Witness || stderr.Len() != 0 {
		t.Fatalf("mismatch = (%d,%s,%q,%v)", code, stdout.String(), stderr.String(), err)
	}
}

func TestWalkCLIInputOutputAndArgumentBoundaries(t *testing.T) {
	for _, args := range [][]string{
		{"fartapp", "walk", "simulate", "--format", "xml"},
		{"fartapp", "walk", "simulate", "a", "b"},
		{"fartapp", "walk", "simulate", "--unknown"},
	} {
		var stdout, stderr bytes.Buffer
		if code := Run(args, strings.NewReader(""), &stdout, &stderr); code != 1 || stdout.Len() != 0 || stderr.Len() == 0 {
			t.Fatalf("invalid args=%v code=%d", args, code)
		}
	}
	for _, input := range []io.Reader{nil, walkCLIErrorReader{}, strings.NewReader(strings.Repeat("x", walkcase.MaxInputBytes+1)), strings.NewReader("{}")} {
		var stdout, stderr bytes.Buffer
		code := Run([]string{"fartapp", "walk", "simulate", "-", "--format", "json"}, input, &stdout, &stderr)
		var report walkcase.Report
		if err := json.Unmarshal(stdout.Bytes(), &report); code != 1 || err != nil || report.Predicted() || stderr.Len() != 0 {
			t.Fatalf("invalid input=(%d,%s,%q,%v)", code, stdout.String(), stderr.String(), err)
		}
	}
	for _, source := range []string{"testdata/walk/isothermal-choked.json", "-"} {
		for _, format := range []string{"text", "json"} {
			var stderr bytes.Buffer
			code := Run([]string{"fartapp", "walk", "simulate", source, "--format", format}, strings.NewReader("{}"), failingWriter{}, &stderr)
			if code != 1 || stderr.Len() == 0 {
				t.Fatal("output failure was hidden")
			}
		}
	}
	var stderr bytes.Buffer
	if code := Run([]string{"fartapp", "walk", "--help"}, nil, failingWriter{}, &stderr); code != 1 {
		t.Fatal("help output failure was hidden")
	}
}

func TestWalkCLIFileAndStdinReportsAreIdentical(t *testing.T) {
	path := filepath.FromSlash("testdata/walk/ordinary-low-pressure.json")
	input, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, operation := range []string{"simulate", "branch", "witness"} {
		var file, stdin, stderr bytes.Buffer
		if code := Run([]string{"fartapp", "walk", operation, path, "--format", "json"}, nil, &file, &stderr); code != 0 {
			t.Fatalf("file status %d: %s", code, stderr.String())
		}
		if code := Run([]string{"fartapp", "walk", operation, "-", "--format", "json"}, bytes.NewReader(input), &stdin, &stderr); code != 0 {
			t.Fatalf("stdin status %d: %s", code, stderr.String())
		}
		if !bytes.Equal(file.Bytes(), stdin.Bytes()) {
			t.Fatal("file and stdin reports differed")
		}
	}
}

type walkCLIErrorReader struct{}

func (walkCLIErrorReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }
