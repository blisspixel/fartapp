package main

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

	"github.com/blisspixel/fartapp/internal/restrictionprediction"
)

const expectedRestrictionText = `RESTRICTION STATE PREDICTED

Model: continuum.quasi-steady-isentropic-converging-restriction@v0alpha1
Implementation: go-oracle/v0alpha1
Quantity system: si (explicit)
Regime: choked

STAGNATION
  Pressure:     125000 Pa
  Temperature:  400 K
  R:            200 J/(kg K)
  gamma:        1.5

RESTRICTION
  Area law:     prescribed
  Prescribed:   0.01 m^2
  Effective:    0.01 m^2
  Cd:           1
  Back pressure: 50000 Pa

CONTROL SECTION
  Critical ratio: 0.512
  Back ratio:     0.4
  Throat Mach:    1
  Exit pressure:  64000 Pa
  Exit temperature: 320 K
  Exit speed:     309.83866769659335 m/s
  Mass flow:      3.0983866769659336 kg/s
  Sonic mass flow: 3.0983866769659336 kg/s
  Thrust:         1100 N
  Recoil:         -1100 N

BALANCE CLAIMS
  restriction.mass-flow-definition: satisfied-within-roundoff; residual 0 kg/s; tolerance 4.4030722917880405e-14 kg/s
  restriction.thrust-control-surface: satisfied-within-roundoff; residual 0 N; tolerance 1.5631940186722204e-11 N
  restriction.recoil-action-reaction: satisfied-within-roundoff; residual 0 N; tolerance 1.5631940186722204e-11 N

Assumptions: calorically-perfect-gas, quasi-steady-flow, isentropic-control-section, converging-restriction-only, discharge-coefficient-scales-mass-flow-only, no-reverse-flow, no-shock-inside-restriction, prescribed-or-linear-quasi-static-area
Model nonclaims: elapsed-time-history, reservoir-mass-energy-coupling, shock-containing-or-underexpanded-plume, viscous-resolved-vena-contracta, phase-change-and-reaction, acoustics
Operation nonclaims: case-commitment, certificate-issuance
Evidence nonclaims: empirical-validation
Ambient inputs: none
`

const expectedRestrictionJSONSHA256 = "51d393906ab6eee7e07cbc730a2058cdbad63431dbc7f400d0624119fe3db0e4"

func TestRestrictionCLITextAndJSONFixtures(t *testing.T) {
	input := readRestrictionFixture(t, "gamma15-choked.json")
	for _, test := range []struct {
		name  string
		args  []string
		stdin []byte
	}{
		{name: "stdin", args: []string{"fartapp", "restriction", "predict", "-"}, stdin: input},
		{name: "stdin after terminator", args: []string{"fartapp", "restriction", "predict", "--", "-"}, stdin: input},
		{name: "file", args: []string{"fartapp", "restriction", "predict", filepath.FromSlash("testdata/restriction/gamma15-choked.json")}},
	} {
		t.Run(test.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			code := runWithInput(test.args, bytes.NewReader(test.stdin), &stdout, &stderr)
			if code != 0 || stdout.String() != expectedRestrictionText || stderr.Len() != 0 {
				t.Fatalf("result = (%d, %q, %q)", code, stdout.String(), stderr.String())
			}
		})
	}

	invokeJSON := func() []byte {
		t.Helper()
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		code := runWithInput(
			[]string{"fartapp", "restriction", "predict", "-", "--format=json"},
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
	if got := hex.EncodeToString(digest[:]); got != expectedRestrictionJSONSHA256 {
		t.Fatalf("JSON SHA-256 = %s, want %s", got, expectedRestrictionJSONSHA256)
	}
	var report restrictionprediction.Report
	if err := json.Unmarshal(first, &report); err != nil || !report.Predicted() {
		t.Fatalf("decoded report = (%#v, %v)", report, err)
	}
	if !bytes.Contains(first, []byte(`"regime":"choked"`)) ||
		bytes.Contains(first, []byte("earth")) || bytes.Contains(first, []byte("body")) {
		t.Fatalf("JSON identity or universality boundary changed: %s", first)
	}
}

func TestRestrictionCLIFailures(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "family usage", args: []string{"fartapp", "restriction"}, want: "usage: fartapp restriction <predict|history> <request.json|-> [--format text|json]\n"},
		{name: "unknown command", args: []string{"fartapp", "restriction", "release"}, want: "unknown restriction command \"release\"\n"},
		{name: "missing source", args: []string{"fartapp", "restriction", "predict"}, want: "usage: fartapp restriction predict <request.json|-> [--format text|json]\n"},
		{name: "extra source", args: []string{"fartapp", "restriction", "predict", "a", "b"}, want: "usage: fartapp restriction predict <request.json|-> [--format text|json]\n"},
		{name: "bad option", args: []string{"fartapp", "restriction", "predict", "-", "--loud"}, want: "invalid restriction predict: unknown option \"--loud\"\n"},
		{name: "missing file", args: []string{"fartapp", "restriction", "predict", "missing.json"}, want: "restriction prediction failed: FART-E-IO-0003 input_not_found at \"/\"\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			if code := runWithInput(test.args, strings.NewReader(""), &stdout, &stderr); code != 1 ||
				stdout.Len() != 0 || stderr.String() != test.want {
				t.Fatalf("result = (%d, %q, %q)", code, stdout.String(), stderr.String())
			}
		})
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithInput(
		[]string{"fartapp", "restriction", "predict", "-", "--format", "json"},
		strings.NewReader(`{"schema":"wrong"}`),
		&stdout,
		&stderr,
	)
	if code != 1 || stderr.Len() != 0 {
		t.Fatalf("JSON failure streams = (%d, %q, %q)", code, stdout.String(), stderr.String())
	}
	var report restrictionprediction.Report
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil || report.Predicted() || len(report.Diagnostics) != 1 {
		t.Fatalf("JSON failure report = (%#v, %v)", report, err)
	}
	if report.Flow != nil {
		t.Fatalf("JSON failure fabricated flow: %#v", report)
	}
}

func TestRestrictionCLIInputAndOutputFailures(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithInput(
		[]string{"fartapp", "restriction", "predict", "-"},
		restrictionErrorReader{},
		&stdout,
		&stderr,
	)
	if code != 1 || stdout.Len() != 0 ||
		stderr.String() != "restriction prediction failed: FART-E-IO-0003 input_unavailable at \"/\"\n" {
		t.Fatalf("read failure = (%d, %q, %q)", code, stdout.String(), stderr.String())
	}

	stderr.Reset()
	code = runWithInput(
		[]string{"fartapp", "restriction", "predict", "-"},
		bytes.NewReader(readRestrictionFixture(t, "gamma15-choked.json")),
		failingWriter{},
		&stderr,
	)
	if code != 1 || !strings.HasPrefix(stderr.String(), "write output: ") {
		t.Fatalf("write failure = (%d, %q)", code, stderr.String())
	}

	if got := classifyRestrictionInputError(os.ErrPermission); got != "input_permission_denied" {
		t.Fatalf("permission classification = %q", got)
	}
	if got := classifyRestrictionInputError(errors.New("other")); got != "input_unavailable" {
		t.Fatalf("other classification = %q", got)
	}

	stdout.Reset()
	stderr.Reset()
	code = runWithInput(
		[]string{"fartapp", "restriction", "predict", filepath.FromSlash("testdata/restriction/linear-compliance-choked.json")},
		strings.NewReader(""),
		&stdout,
		&stderr,
	)
	if code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), "linear-compliance") ||
		!strings.Contains(stdout.String(), "Compliance:") {
		t.Fatalf("compliance text = (%d, %q, %q)", code, stdout.String(), stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = runWithInput(
		[]string{"fartapp", "restriction", "history", filepath.FromSlash("testdata/restriction/gamma15-choked-history.json")},
		strings.NewReader(""),
		&stdout,
		&stderr,
	)
	if code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), "RESTRICTION HISTORY PREDICTED") ||
		!strings.Contains(stdout.String(), "Mass out:") {
		t.Fatalf("history text = (%d, %q, %q)", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := runWithInput(
		[]string{"fartapp", "restriction", "history", "missing.json"},
		strings.NewReader(""),
		&stdout,
		&stderr,
	); code != 1 || !strings.Contains(stderr.String(), "restriction history failed") {
		t.Fatalf("history missing = (%d, %q, %q)", code, stdout.String(), stderr.String())
	}
}

type restrictionErrorReader struct{}

func (restrictionErrorReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }

func readRestrictionFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "restriction", name))
	if err != nil {
		t.Fatalf("read restriction fixture: %v", err)
	}
	return data
}
