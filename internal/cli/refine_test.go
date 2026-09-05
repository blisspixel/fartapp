package cli

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/blisspixel/fartapp/internal/walkcase"
)

func TestRefineCLIFileStdinAndEvidence(t *testing.T) {
	path := "testdata/walk/ordinary-low-pressure.json"
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	flags := []string{"--relative-tolerance", "1e-8", "--max-evaluations=100000", "--absolute-time-tolerance=0", "--format=json"}
	var previous []byte
	for _, source := range []string{path, "-"} {
		var stdout, stderr bytes.Buffer
		args := append([]string{"fartapp", "walk", "refine", source}, flags...)
		if code := Run(args, bytes.NewReader(data), &stdout, &stderr); code != 0 || stderr.Len() != 0 {
			t.Fatalf("refine (%d): %s", code, stderr.String())
		}
		var report walkcase.Report
		if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
			t.Fatal(err)
		}
		if !report.Predicted() || report.Accuracy == nil || !report.Accuracy.DischargeComplete ||
			math.Abs(*report.ElapsedSeconds-0.05839420446440555) > 2e-12 {
			t.Fatal("CLI lost reference result or evidence")
		}
		if previous != nil && !bytes.Equal(previous, stdout.Bytes()) {
			t.Fatal("file/stdin differed")
		}
		previous = append([]byte(nil), stdout.Bytes()...)
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"fartapp", "walk", "refine", path, "--relative-tolerance=1e-8", "--max-evaluations=100000"},
		nil, &stdout, &stderr)
	if code != 0 || !strings.Contains(stdout.String(), "Estimated time error:") ||
		!strings.Contains(stdout.String(), "Discharge complete: true") {
		t.Fatalf("text = %d %s %s", code, stdout.String(), stderr.String())
	}
}

func TestRefineCLIRefusalsAndHelpNeverRead(t *testing.T) {
	for _, args := range [][]string{
		{}, {"-"}, {"-", "--relative-tolerance=1e-8"}, {"-", "--max-evaluations=20"},
		{"-", "--relative-tolerance"}, {"-", "--relative-tolerance="},
		{"-", "--relative-tolerance=NaN", "--max-evaluations=20"},
		{"-", "--relative-tolerance=Inf", "--max-evaluations=20"},
		{"-", "--relative-tolerance=0", "--max-evaluations=20"},
		{"-", "--relative-tolerance=1e-8", "--max-evaluations=1e3"},
		{"-", "--relative-tolerance=1e-8", "--max-evaluations=14"},
		{"-", "--relative-tolerance=1e-8", "--max-evaluations=1000001"},
		{"-", "--relative-tolerance=1e-8", "--max-evaluations=20", "--absolute-time-tolerance=-1"},
		{"-", "--relative-tolerance=1e-8", "--max-evaluations=20", "--absolute-time-tolerance=NaN"},
		{"-", "--relative-tolerance=1e-8", "--max-evaluations=20", "--absolute-time-tolerance=Inf"},
		{"--format", "--relative-tolerance", "1e-8", "-", "--max-evaluations=20"},
		{"--help", "--relative-tolerance=1e-8", "--relative-tolerance=1e-8"},
		{"--help", "--max-evaluations", "--help"}, {"--help", "a", "b"},
		{"--", "-"}, {"--help", "--nope"}, {"--help", "--format"},
	} {
		var stdout, stderr bytes.Buffer
		if code := Run(append([]string{"fartapp", "walk", "refine"}, args...), tripwireReader{}, &stdout, &stderr); code != 1 || stdout.Len() != 0 || stderr.Len() == 0 {
			t.Fatalf("%q = %d %q %q", args, code, stdout.String(), stderr.String())
		}
	}
	for _, args := range [][]string{
		{"fartapp", "help", "walk", "refine"},
		{"fartapp", "walk", "refine", "--help"},
		{"fartapp", "walk", "refine", "unopened.json", "--help"},
	} {
		var stdout, stderr bytes.Buffer
		if code := Run(args, tripwireReader{}, &stdout, &stderr); code != 0 || stdout.String() != refineHelp || stderr.Len() != 0 {
			t.Fatalf("help = %d %q %q", code, stdout.String(), stderr.String())
		}
	}
	for _, source := range []string{"testdata/walk/isothermal-choked.json", "does-not-exist.json", "-"} {
		var stdout, stderr bytes.Buffer
		args := []string{"fartapp", "walk", "refine", source, "--relative-tolerance=1e-8", "--max-evaluations=100000", "--format=json"}
		if code := Run(args, nil, &stdout, &stderr); code != 1 {
			t.Fatal("expected refusal")
		}
		var report walkcase.Report
		if err := json.Unmarshal(stdout.Bytes(), &report); err != nil || report.Predicted() || report.ImplementationRevision != walkcase.RefinementRevision {
			t.Fatalf("refusal = %s", stdout.String())
		}
	}
}
