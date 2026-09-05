package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/blisspixel/fartapp/internal/reservoirprediction"
)

const expectedReservoirText = `RESERVOIR ENDPOINT PREDICTED

Model: continuum.rigid-calorically-perfect-ideal-mixture@v0alpha1
Implementation: go-oracle.reservoir/v0alpha2
Quantity system: si (explicit)
Closure: rigid-adiabatic
Prescribed withdrawal fraction: 0.75

INITIAL STATE
  component.a          1 kg; R 200 J/(kg K); cv 400 J/(kg K)
  component.b          3 kg; R 400 J/(kg K); cv 800 J/(kg K)
  Total mass:      4 kg
  Volume:          1 m^3
  Temperature:     400 K
  Pressure:        560000 Pa
  Internal energy: 1.12e+06 J
  R_mix:           350 J/(kg K)
  cv_mix:          700 J/(kg K)
  cp_mix:          1050 J/(kg K)
  gamma:           1.5

FINAL STATE
  component.a          0.25 kg; R 200 J/(kg K); cv 400 J/(kg K)
  component.b          0.75 kg; R 400 J/(kg K); cv 800 J/(kg K)
  Total mass:      1 kg
  Volume:          1 m^3
  Temperature:     200 K
  Pressure:        70000 Pa
  Internal energy: 140000 J
  R_mix:           350 J/(kg K)
  cv_mix:          700 J/(kg K)
  cp_mix:          1050 J/(kg K)
  gamma:           1.5

TRANSFERS
  component.a          mass out 0.75 kg
  component.b          mass out 2.25 kg
  Total mass out:          3 kg
  Integrated enthalpy out: 980000 J
  Heat into reservoir:     0 J
  Boundary work:           0 J

BALANCE CLAIMS
  reservoir.total-mass-balance: satisfied-within-roundoff; residual 0 kg; tolerance 5.684341886080802e-14 kg
  reservoir.energy-balance: satisfied-within-roundoff; residual 0 J; tolerance 1.5916157281026244e-08 J
  reservoir.initial-equation-of-state: satisfied-within-roundoff; residual 0 J; tolerance 7.958078640513122e-09 J
  reservoir.final-equation-of-state: satisfied-within-roundoff; residual 0 J; tolerance 9.947598300641403e-10 J

Assumptions: calorically-perfect-components, homogeneous-equilibrium-state, nonreacting-mixture, single-gas-phase, perfectly-mixed-nonselective-outflow, rigid-volume, sensible-energy-datum-cv-times-temperature, adiabatic-no-heat-transfer
Model nonclaims: aperture-and-restriction-flow, elapsed-time-history, exterior-state, momentum-and-recoil, phase-change-and-reaction, plume-and-acoustics
Operation nonclaims: case-commitment, certificate-issuance
Evidence nonclaims: empirical-validation
Ambient inputs: none
`

const expectedReservoirJSONSHA256 = "1f817094c5de9cfe97e3ed5cf0157ce07984a6e6f489ee22034686daa101c50f"

func TestReservoirCLITextAndJSONFixtures(t *testing.T) {
	input := readReservoirFixture(t, "synthetic-mixture-adiabatic.json")
	for _, test := range []struct {
		name  string
		args  []string
		stdin []byte
	}{
		{name: "stdin", args: []string{"fartapp", "reservoir", "predict", "-"}, stdin: input},
		{name: "stdin after terminator", args: []string{"fartapp", "reservoir", "predict", "--", "-"}, stdin: input},
		{name: "file", args: []string{"fartapp", "reservoir", "predict", filepath.FromSlash("testdata/reservoir/synthetic-mixture-adiabatic.json")}},
	} {
		t.Run(test.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			code := Run(test.args, bytes.NewReader(test.stdin), &stdout, &stderr)
			if code != 0 || stdout.String() != expectedReservoirText || stderr.Len() != 0 {
				t.Fatalf("result = (%d, %q, %q)", code, stdout.String(), stderr.String())
			}
		})
	}

	invokeJSON := func() []byte {
		t.Helper()
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		code := Run(
			[]string{"fartapp", "reservoir", "predict", "-", "--format=json"},
			bytes.NewReader(input),
			&stdout,
			&stderr,
		)
		if code != 0 || stderr.Len() != 0 {
			t.Fatalf("JSON result = (%d, %q)", code, stderr.String())
		}
		return bytes.Clone(stdout.Bytes())
	}
	first := invokeJSON()
	second := invokeJSON()
	if !bytes.Equal(first, second) {
		t.Fatal("JSON output is not deterministic")
	}
	digest := sha256.Sum256(first)
	if got := hex.EncodeToString(digest[:]); got != expectedReservoirJSONSHA256 {
		t.Fatalf("JSON SHA-256 = %s, want %s", got, expectedReservoirJSONSHA256)
	}
	var report reservoirprediction.Report
	if err := json.Unmarshal(first, &report); err != nil || !report.Predicted() {
		t.Fatalf("decoded report = (%#v, %v)", report, err)
	}
	if !bytes.Contains(first, []byte(`"withdrawal_fraction":0.75`)) ||
		bytes.Contains(first, []byte("earth")) || bytes.Contains(first, []byte("body")) {
		t.Fatalf("JSON identity or universality boundary changed: %s", first)
	}
}

func TestReservoirCLIFailures(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "family usage", args: []string{"fartapp", "reservoir"}, want: "usage: fartapp reservoir predict <request.json|-> [--format text|json]\n"},
		{name: "unknown command", args: []string{"fartapp", "reservoir", "release"}, want: "unknown reservoir command \"release\"\n"},
		{name: "missing source", args: []string{"fartapp", "reservoir", "predict"}, want: "usage: fartapp reservoir predict <request.json|-> [--format text|json]\n"},
		{name: "extra source", args: []string{"fartapp", "reservoir", "predict", "a", "b"}, want: "usage: fartapp reservoir predict <request.json|-> [--format text|json]\n"},
		{name: "bad option", args: []string{"fartapp", "reservoir", "predict", "-", "--loud"}, want: "invalid reservoir predict: unknown option \"--loud\"\n"},
		{name: "missing file", args: []string{"fartapp", "reservoir", "predict", "missing.json"}, want: "reservoir prediction failed: FART-E-IO-0002 input_not_found at \"/\"\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			if code := Run(test.args, strings.NewReader(""), &stdout, &stderr); code != 1 ||
				stdout.Len() != 0 || stderr.String() != test.want {
				t.Fatalf("result = (%d, %q, %q)", code, stdout.String(), stderr.String())
			}
		})
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run(
		[]string{"fartapp", "reservoir", "predict", "-", "--format", "json"},
		strings.NewReader(`{"schema":"wrong"}`),
		&stdout,
		&stderr,
	)
	if code != 1 || stderr.Len() != 0 {
		t.Fatalf("JSON failure streams = (%d, %q, %q)", code, stdout.String(), stderr.String())
	}
	var report reservoirprediction.Report
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil || report.Predicted() || len(report.Diagnostics) != 1 {
		t.Fatalf("JSON failure report = (%#v, %v)", report, err)
	}
	if report.WithdrawalFraction != nil {
		t.Fatalf("JSON failure fabricated a withdrawal fraction: %#v", report)
	}
}

func TestReservoirCLIInputAndOutputFailures(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run(
		[]string{"fartapp", "reservoir", "predict", "-"},
		reservoirErrorReader{},
		&stdout,
		&stderr,
	)
	if code != 1 || stdout.Len() != 0 ||
		stderr.String() != "reservoir prediction failed: FART-E-IO-0002 input_unavailable at \"/\"\n" {
		t.Fatalf("read failure = (%d, %q, %q)", code, stdout.String(), stderr.String())
	}

	stderr.Reset()
	code = Run(
		[]string{"fartapp", "reservoir", "predict", "-"},
		bytes.NewReader(readReservoirFixture(t, "synthetic-mixture-adiabatic.json")),
		failingWriter{},
		&stderr,
	)
	if code != 1 || !strings.HasPrefix(stderr.String(), "write output: ") {
		t.Fatalf("write failure = (%d, %q)", code, stderr.String())
	}

	if got := classifyReservoirInputError(os.ErrPermission); got != "input_permission_denied" {
		t.Fatalf("permission classification = %q", got)
	}
	if got := classifyReservoirInputError(errors.New("other")); got != "input_unavailable" {
		t.Fatalf("other classification = %q", got)
	}
	if got := formatScientificValue(1.25); got != "1.25" {
		t.Fatalf("formatScientificValue = %q", got)
	}
}

type reservoirErrorReader struct{}

func (reservoirErrorReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }

func readReservoirFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "reservoir", name))
	if err != nil {
		t.Fatalf("read reservoir fixture: %v", err)
	}
	return data
}
