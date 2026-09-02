package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/blisspixel/fartapp/internal/scenarioprobe"
)

const expectedAtemporalScenarioText = `SCENARIO PROBE DOCUMENT VALID

Document schema: fart.scenario-probe/v0alpha1
Law context: conformance.relation.atemporal@v0alpha1
Scope: s0
Realization admission: not-evaluated (admission_policy_unratified)
Realization: not-performed (validation_only)
Validation stages:
  syntax:                 valid
  schema:                 valid
  law resolution:         resolved
  capability resolution:  resolved
Validator inputs: document_bytes, built_in_law_catalog
Ambient inputs: none

CAPABILITY REQUESTS

catalog.inspect [resolved]
  law definition:      not-applicable (application_capability)
  implementation:      available
  closure:             not-required
  applicability:       applicable
  evidence:            verified (software_fixture)
  evidence references: test:law-catalog-inspection, test:law-cli-fixtures
  trust:               built-in-candidate
  backend feasibility: not-required (application_capability)
  resource feasibility: within-default-budget

CAPABILITY EVIDENCE REGISTRY
test:law-catalog-inspection [software/go-test]
  go test ./internal/lawcatalog -run ^TestBuiltInCatalog$
test:law-cli-fixtures [software/go-test]
  go test . -run ^TestLawCLITextAndJSONFixtures$
`

const expectedAtemporalScenarioJSONSHA256 = "cf509eba1b2a85b8e812c9dd3fdb18f7f3d80c36d1fb4cab5109f5ec5e3acc76"
const expectedEarthScenarioJSONSHA256 = "336a73ea0e22b8c520e393aafee75fcc7675466673b08f3f02a925dabdfaab96"

func TestScenarioCLITextAndJSONFixtures(t *testing.T) {
	input := readScenarioFixture(t, "atemporal-probe.json")
	for _, test := range []struct {
		name string
		args []string
	}{
		{name: "stdin", args: []string{"fartapp", "scenario", "validate", "-"}},
		{name: "stdin after terminator", args: []string{"fartapp", "scenario", "validate", "--", "-"}},
		{name: "file", args: []string{"fartapp", "scenario", "validate", filepath.FromSlash("testdata/scenarios/atemporal-probe.json")}},
	} {
		t.Run(test.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			if code := runWithInput(test.args, bytes.NewReader(input), &stdout, &stderr); code != 0 {
				t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
			}
			if got := stdout.String(); got != expectedAtemporalScenarioText {
				t.Fatalf("stdout = %q, want %q", got, expectedAtemporalScenarioText)
			}
			if stderr.Len() != 0 {
				t.Fatalf("stderr = %q", stderr.String())
			}
		})
	}

	invokeJSON := func(args []string, stdin []byte) []byte {
		t.Helper()
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		if code := runWithInput(args, bytes.NewReader(stdin), &stdout, &stderr); code != 0 {
			t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
		}
		return bytes.Clone(stdout.Bytes())
	}
	args := []string{"fartapp", "scenario", "validate", "-", "--format=json"}
	first := invokeJSON(args, input)
	second := invokeJSON(args, input)
	if !bytes.Equal(first, second) {
		t.Fatal("JSON output is not byte deterministic")
	}
	digest := sha256.Sum256(first)
	if got := hex.EncodeToString(digest[:]); got != expectedAtemporalScenarioJSONSHA256 {
		t.Fatalf("JSON SHA-256 = %s, want %s", got, expectedAtemporalScenarioJSONSHA256)
	}
	var report scenarioprobe.Report
	if err := json.Unmarshal(first, &report); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if !report.Valid() || report.Realization.Status != "not-performed" {
		t.Fatalf("report = %#v", report)
	}
	for _, forbidden := range []string{"emitter", "locale", "observer", "presentation", "time", "unit"} {
		if bytes.Contains(first, []byte(forbidden)) {
			t.Errorf("JSON unexpectedly contains %q", forbidden)
		}
	}
}

func TestScenarioCLIUnavailableCapabilityIsAValidProbe(t *testing.T) {
	input := readScenarioFixture(t, "earth-subsonic-probe.json")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithInput(
		[]string{"fartapp", "scenario", "validate", "-", "--format", "json"},
		bytes.NewReader(input),
		&stdout,
		&stderr,
	)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("result = (%d, %q)", code, stderr.String())
	}
	digest := sha256.Sum256(stdout.Bytes())
	if got := hex.EncodeToString(digest[:]); got != expectedEarthScenarioJSONSHA256 {
		t.Fatalf("JSON SHA-256 = %s, want %s", got, expectedEarthScenarioJSONSHA256)
	}
	var report scenarioprobe.Report
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	capability := report.Capabilities[0]
	if capability.Implementation.Status != "unavailable" ||
		capability.Applicability.Status != "undetermined" ||
		report.Realization.Status != "not-performed" {
		t.Fatalf("report = %#v", report)
	}
}

func TestScenarioCLIInvalidInputContracts(t *testing.T) {
	duplicate := `{"schema":"fart.scenario-probe/v0alpha1","schema":"duplicate"}`
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithInput(
		[]string{"fartapp", "scenario", "validate", "-", "--format", "json"},
		strings.NewReader(duplicate),
		&stdout,
		&stderr,
	)
	if code != 1 || stderr.Len() != 0 {
		t.Fatalf("JSON failure = (%d, %q, %q)", code, stdout.String(), stderr.String())
	}
	var report scenarioprobe.Report
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if report.Diagnostics[0].ReasonCode != "duplicate_member" {
		t.Fatalf("report = %#v", report)
	}

	stdout.Reset()
	code = runWithInput(
		[]string{"fartapp", "scenario", "validate", "-"},
		strings.NewReader(`{"schema":"wrong"}`),
		&stdout,
		&stderr,
	)
	if code != 1 || stdout.Len() != 0 {
		t.Fatalf("text failure = (%d, %q, %q)", code, stdout.String(), stderr.String())
	}
	if want := "scenario validation failed: FART-E-SCHEMA-0001 unsupported_schema at \"/schema\"\n"; stderr.String() != want {
		t.Fatalf("stderr = %q, want %q", stderr.String(), want)
	}

	for _, hostileKey := range []string{`\u001b[31mBAD`, `line\nbreak`, `\u202eBAD`} {
		input := strings.Replace(
			string(readScenarioFixture(t, "atemporal-probe.json")),
			`"schema":`,
			`"`+hostileKey+`":null,"schema":`,
			1,
		)
		stderr.Reset()
		code = runWithInput(
			[]string{"fartapp", "scenario", "validate", "-"},
			strings.NewReader(input),
			&stdout,
			&stderr,
		)
		if code != 1 || bytes.ContainsRune(stderr.Bytes(), '\x1b') ||
			strings.ContainsRune(stderr.String(), '\u202e') || strings.Count(stderr.String(), "\n") != 1 {
			t.Fatalf("hostile diagnostic for %q = (%d, %q)", hostileKey, code, stderr.String())
		}
	}
}

type scenarioErrorReader struct{}

func (scenarioErrorReader) Read([]byte) (int, error) {
	return 0, errors.New("test read failure")
}

func TestScenarioCLIInputFailures(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		stdin  *bytes.Reader
		reason string
		input  string
	}{
		{name: "missing file", args: []string{"fartapp", "scenario", "validate", filepath.Join(t.TempDir(), "missing.json")}, stdin: bytes.NewReader(nil), reason: "input_not_found", input: "named_input"},
		{name: "too large", args: []string{"fartapp", "scenario", "validate", "-"}, stdin: bytes.NewReader(bytes.Repeat([]byte{'x'}, scenarioprobe.MaxInputBytes+1)), reason: "input_too_large", input: "standard_input"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			code := runWithInput(append(test.args, "--format", "json"), test.stdin, &stdout, &stderr)
			if code != 1 || stderr.Len() != 0 {
				t.Fatalf("result = (%d, %q, %q)", code, stdout.String(), stderr.String())
			}
			var report scenarioprobe.Report
			if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}
			if report.Diagnostics[0].ReasonCode != test.reason {
				t.Fatalf("report = %#v", report)
			}
			if got := report.Environment.ConsultedInputs; len(got) != 1 || got[0] != test.input {
				t.Fatalf("consulted inputs = %q, want %q", got, test.input)
			}
		})
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithInput(
		[]string{"fartapp", "scenario", "validate", "-", "--format", "json"},
		scenarioErrorReader{},
		&stdout,
		&stderr,
	)
	if code != 1 || stderr.Len() != 0 {
		t.Fatalf("read failure = (%d, %q, %q)", code, stdout.String(), stderr.String())
	}
	var report scenarioprobe.Report
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil ||
		len(report.Diagnostics) != 1 || report.Diagnostics[0].ReasonCode != "input_unavailable" ||
		len(report.Environment.ConsultedInputs) != 1 ||
		report.Environment.ConsultedInputs[0] != "standard_input" {
		t.Fatalf("read failure report = (%v, %#v)", err, report)
	}
}

func TestClassifyScenarioInputError(t *testing.T) {
	tests := []struct {
		err    error
		reason string
	}{
		{err: os.ErrNotExist, reason: "input_not_found"},
		{err: errors.Join(errors.New("open"), os.ErrPermission), reason: "input_permission_denied"},
		{err: errors.New("other"), reason: "input_unavailable"},
	}
	for _, test := range tests {
		if got := classifyScenarioInputError(test.err); got != test.reason {
			t.Errorf("classifyScenarioInputError(%v) = %q, want %q", test.err, got, test.reason)
		}
	}
}

func TestScenarioCLIHelpAndUsage(t *testing.T) {
	for _, required := range []string{
		scenarioprobe.DocumentSchema,
		"65,536 bytes",
		"Exit status:",
		"scenario validate testdata/scenarios/atemporal-probe.json",
		"scenario validate testdata/scenarios/atemporal-probe.json --format json",
		"Recovery:",
	} {
		if !strings.Contains(scenarioValidateHelp, required) {
			t.Errorf("scenario validate help omits %q", required)
		}
	}

	tests := []struct {
		args       []string
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{args: []string{"fartapp", "scenario", "help"}, wantStdout: scenarioHelp},
		{args: []string{"fartapp", "scenario", "--help"}, wantStdout: scenarioHelp},
		{args: []string{"fartapp", "scenario", "validate", "--help"}, wantStdout: scenarioValidateHelp},
		{args: []string{"fartapp", "scenario", "validate", "missing", "-h"}, wantStdout: scenarioValidateHelp},
		{args: []string{"fartapp", "scenario"}, wantCode: 1, wantStderr: "usage: fartapp scenario validate <scenario.json|-> [--format text|json]\n"},
		{args: []string{"fartapp", "scenario", "unknown"}, wantCode: 1, wantStderr: "unknown scenario command \"unknown\"\n"},
		{args: []string{"fartapp", "scenario", "validate"}, wantCode: 1, wantStderr: "usage: fartapp scenario validate <scenario.json|-> [--format text|json]\n"},
		{args: []string{"fartapp", "scenario", "validate", "-", "extra"}, wantCode: 1, wantStderr: "usage: fartapp scenario validate <scenario.json|-> [--format text|json]\n"},
		{args: []string{"fartapp", "scenario", "validate", "-", "--unknown"}, wantCode: 1, wantStderr: "invalid scenario validate: unknown option \"--unknown\"\n"},
		{args: []string{"fartapp", "scenario", "validate", "-", "--unknown", "--help"}, wantCode: 1, wantStderr: "invalid scenario validate: unknown option \"--unknown\"\n"},
		{args: []string{"fartapp", "scenario", "unknown", "--help"}, wantCode: 1, wantStderr: "unknown scenario command \"unknown\"\n"},
		{args: []string{"fartapp", "scenario", "validate", "--format", "--help"}, wantCode: 1, wantStderr: "invalid scenario validate: unsupported format \"--help\"; expected text or json\n"},
		{args: []string{"fartapp", "scenario", "validate", "--format", "xml", "--help"}, wantCode: 1, wantStderr: "invalid scenario validate: unsupported format \"xml\"; expected text or json\n"},
		{args: []string{"fartapp", "scenario", "validate", "-", "extra", "--help"}, wantCode: 1, wantStderr: "usage: fartapp scenario validate <scenario.json|-> [--format text|json]\n"},
	}
	for _, test := range tests {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		code := runWithInput(test.args, scenarioErrorReader{}, &stdout, &stderr)
		if code != test.wantCode || stdout.String() != test.wantStdout || stderr.String() != test.wantStderr {
			t.Errorf("run(%q) = (%d, %q, %q), want (%d, %q, %q)", test.args, code, stdout.String(), stderr.String(), test.wantCode, test.wantStdout, test.wantStderr)
		}
	}
}

func TestScenarioCLIOutputFailures(t *testing.T) {
	input := readScenarioFixture(t, "atemporal-probe.json")
	for _, writer := range []interface{ Write([]byte) (int, error) }{failingWriter{}, shortWriter{}} {
		var stderr bytes.Buffer
		if code := runWithInput(
			[]string{"fartapp", "scenario", "validate", "-"},
			bytes.NewReader(input),
			writer,
			&stderr,
		); code != 1 {
			t.Fatalf("exit code = %d", code)
		}
		if !strings.HasPrefix(stderr.String(), "write output: ") {
			t.Fatalf("stderr = %q", stderr.String())
		}
	}
}

func TestScenarioExecutable(t *testing.T) {
	binaryName := "fartapp"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	binaryPath := filepath.Join(t.TempDir(), binaryName)
	build := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build executable: %v\n%s", err, output)
	}
	command := exec.Command(binaryPath, "scenario", "validate", "-")
	command.Stdin = bytes.NewReader(readScenarioFixture(t, "atemporal-probe.json"))
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("run executable: %v, stderr = %q", err, stderr.String())
	}
	if stdout.String() != expectedAtemporalScenarioText || stderr.Len() != 0 {
		t.Fatalf("output = (%q, %q)", stdout.String(), stderr.String())
	}

	command = exec.Command(binaryPath, "scenario", "validate", "-")
	command.Stdin = strings.NewReader(`{"schema":"wrong"}`)
	stdout.Reset()
	stderr.Reset()
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) || exitError.ExitCode() != 1 || stdout.Len() != 0 ||
		!strings.Contains(stderr.String(), "FART-E-SCHEMA-0001") {
		t.Fatalf("invalid executable result = (%v, %q, %q)", err, stdout.String(), stderr.String())
	}

	command = exec.Command(binaryPath, "scenario", "validate", filepath.Join(t.TempDir(), "missing.json"), "--format", "json")
	stdout.Reset()
	stderr.Reset()
	command.Stdout = &stdout
	command.Stderr = &stderr
	err = command.Run()
	if !errors.As(err, &exitError) || exitError.ExitCode() != 1 || stderr.Len() != 0 {
		t.Fatalf("missing executable result = (%v, %q, %q)", err, stdout.String(), stderr.String())
	}
	var report scenarioprobe.Report
	if unmarshalErr := json.Unmarshal(stdout.Bytes(), &report); unmarshalErr != nil ||
		len(report.Diagnostics) != 1 || report.Diagnostics[0].ReasonCode != "input_not_found" {
		t.Fatalf("missing executable report = (%v, %#v)", unmarshalErr, report)
	}

	dashDirectory := t.TempDir()
	if writeErr := os.WriteFile(
		filepath.Join(dashDirectory, "-probe.json"),
		readScenarioFixture(t, "atemporal-probe.json"),
		0o600,
	); writeErr != nil {
		t.Fatalf("write dash-prefixed fixture: %v", writeErr)
	}
	command = exec.Command(binaryPath, "scenario", "validate", "--", "-probe.json")
	command.Dir = dashDirectory
	stdout.Reset()
	stderr.Reset()
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err = command.Run(); err != nil || stdout.String() != expectedAtemporalScenarioText || stderr.Len() != 0 {
		t.Fatalf("dash-prefixed executable result = (%v, %q, %q)", err, stdout.String(), stderr.String())
	}
}

func readScenarioFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "scenarios", name))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return data
}
