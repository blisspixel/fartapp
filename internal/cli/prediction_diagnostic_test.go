package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/blisspixel/fartapp/internal/reservoirprediction"
	"github.com/blisspixel/fartapp/internal/restrictionhistoryprediction"
	"github.com/blisspixel/fartapp/internal/restrictionprediction"
	"github.com/blisspixel/fartapp/internal/walkcase"
)

func TestPredictionDiagnosticHintsDoNotAlterStructuredRefusals(t *testing.T) {
	input := []byte(`{"schema":"unsupported"}`)
	for _, test := range []struct {
		command []string
		report  any
	}{
		{[]string{"reservoir", "predict"}, reservoirprediction.Predict(input)},
		{[]string{"restriction", "predict"}, restrictionprediction.Predict(input)},
		{[]string{"restriction", "history"}, restrictionhistoryprediction.Predict(input)},
		{[]string{"walk", "simulate"}, walkcase.Run(input, "simulate")},
	} {
		t.Run(strings.Join(test.command, " "), func(t *testing.T) {
			args := append([]string{"fartapp"}, test.command...)
			args = append(args, "-")
			var stdout, stderr bytes.Buffer
			if code := Run(args, bytes.NewReader(input), &stdout, &stderr); code != 1 || stdout.Len() != 0 ||
				!strings.Contains(stderr.String(), "unsupported_schema at") ||
				!strings.Contains(stderr.String(), "'fartapp help "+strings.Join(test.command, " ")+"'") {
				t.Fatalf("text refusal = %d, %q, %q", code, stdout.String(), stderr.String())
			}
			stdout.Reset()
			stderr.Reset()
			if code := Run(append(args, "--format=json"), bytes.NewReader(input), &stdout, &stderr); code != 1 || stderr.Len() != 0 {
				t.Fatalf("JSON refusal = %d, %q", code, stderr.String())
			}
			want, err := json.Marshal(test.report)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(stdout.Bytes(), append(want, '\n')) {
				t.Fatalf("presentation altered the typed refusal: %s", stdout.String())
			}
		})
	}
}

func TestPredictionDiagnosticsGuideApplicableRecovery(t *testing.T) {
	for _, test := range []struct {
		args   []string
		reason string
		hint   string
	}{
		{[]string{"restriction", "history", "missing-history.json"}, "input_not_found", "use '-' to read standard input"},
		{[]string{"walk", "simulate", "testdata/walk/atemporal-no-dimension.json"}, "incompatible_law_context", "earth.continuum.si@v0alpha1"},
		{[]string{"walk", "reconstruct", "testdata/walk/ordinary-low-pressure.json"}, "missing_member", "Retain a 'walk witness' value"},
		{[]string{"walk", "refine", "testdata/walk/isothermal-choked.json", "--relative-tolerance=1e-8", "--max-evaluations=100000"}, "unsupported_accuracy_configuration", "step.max_time_s = 0"},
	} {
		var stdout, stderr bytes.Buffer
		code := Run(append([]string{"fartapp"}, test.args...), tripwireReader{}, &stdout, &stderr)
		if code != 1 || stdout.Len() != 0 || !strings.Contains(stderr.String(), test.reason) ||
			!strings.Contains(stderr.String(), test.hint) || strings.Count(stderr.String(), "\n") != 2 {
			t.Errorf("%q = %d, %q, %q", test.args, code, stdout.String(), stderr.String())
		}
	}
}

func TestPredictionDiagnosticFieldsAreBoundedAndEscaped(t *testing.T) {
	for _, hostile := range []string{
		strings.Repeat("x", 4096) + "hidden-tail",
		strings.Repeat("\x1b\n\r\x00\xff", 1024) + "hidden-tail",
		strings.Repeat("\u2028\u2029", 1024) + "hidden-tail",
	} {
		var output bytes.Buffer
		writePredictionDiagnostic(&output, hostile, hostile, hostile, hostile)
		if output.Len() > 2048 || strings.Count(output.String(), "\n") != 2 ||
			strings.ContainsAny(output.String(), "\r\x00\x1b\u2028\u2029") ||
			strings.Contains(output.String(), "hidden-tail") || !strings.Contains(output.String(), "...") ||
			!strings.Contains(output.String(), "'fartapp help'") {
			t.Fatalf("unbounded or unsafe diagnostic: %q", output.String())
		}
	}
	var output bytes.Buffer
	writePredictionDiagnostic(&output, "walk predict", "FART-E-MODEL-0006", "negative_area", "/restriction/area/prescribed_m2")
	if !strings.Contains(output.String(), "FART-E-MODEL-0006 negative_area at \"/restriction/area/prescribed_m2\"") ||
		!strings.Contains(output.String(), "nonnegative area in square metres") {
		t.Fatal("safe identifiers, exact path, or applicable recovery were lost")
	}
	output.Reset()
	writePredictionDiagnostic(&output, "walk refine", "", "", "")
	if !strings.Contains(output.String(), `failed: "" "" at ""`) {
		t.Fatal("empty diagnostic fields became ambiguous")
	}
}
