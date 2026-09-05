package cli

import (
	"bytes"
	"errors"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestIntensityDomain(t *testing.T) {
	want := []emission{
		{sound: "pfft", rating: "gentle"},
		{sound: "toot", rating: "respectable"},
		{sound: "braaap", rating: "respectable"},
		{sound: "blorp", rating: "respectable"},
		{sound: "KABLAM", rating: "mighty"},
	}

	for value, wantEvent := range want {
		level, err := newIntensity(value + minimumIntensity)
		if err != nil {
			t.Fatalf("newIntensity(%d): %v", value+minimumIntensity, err)
		}
		if got := level.emission(); got != wantEvent {
			t.Errorf("intensity %d emission = %#v, want %#v", value+1, got, wantEvent)
		}
	}

	for _, value := range []int{-1, 0, 6, 7} {
		if _, err := newIntensity(value); err == nil {
			t.Errorf("newIntensity(%d) succeeded, want error", value)
		}
	}
}

func TestCLIFixtures(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{name: "one", args: []string{"fartapp", "1"}, wantStdout: "pfft (gentle)\n"},
		{name: "two", args: []string{"fartapp", "2"}, wantStdout: "toot (respectable)\n"},
		{name: "three", args: []string{"fartapp", "3"}, wantStdout: "braaap (respectable)\n"},
		{name: "four", args: []string{"fartapp", "4"}, wantStdout: "blorp (respectable)\n"},
		{name: "five", args: []string{"fartapp", "5"}, wantStdout: "KABLAM (mighty)\n"},
		{name: "missing intensity", args: []string{"fartapp"}, wantCode: 1, wantStderr: "usage: fartapp <intensity>\n"},
		{name: "extra argument", args: []string{"fartapp", "3", "extra"}, wantCode: 1, wantStderr: "usage: fartapp <intensity>\n"},
		{
			name:       "non-integer",
			args:       []string{"fartapp", "nope"},
			wantCode:   1,
			wantStderr: "invalid intensity \"nope\": must be a canonical integer from 1 to 5\n",
		},
		{name: "below range", args: []string{"fartapp", "0"}, wantCode: 1, wantStderr: "invalid intensity 0 must be from 1 to 5\n"},
		{name: "above range", args: []string{"fartapp", "6"}, wantCode: 1, wantStderr: "invalid intensity 6 must be from 1 to 5\n"},
		{
			name:       "leading plus",
			args:       []string{"fartapp", "+1"},
			wantCode:   1,
			wantStderr: "invalid intensity \"+1\": must be a canonical integer from 1 to 5\n",
		},
		{
			name:       "leading zero",
			args:       []string{"fartapp", "01"},
			wantCode:   1,
			wantStderr: "invalid intensity \"01\": must be a canonical integer from 1 to 5\n",
		},
		{
			name:       "surrounding whitespace",
			args:       []string{"fartapp", " 1 "},
			wantCode:   1,
			wantStderr: "invalid intensity \" 1 \": must be a canonical integer from 1 to 5\n",
		},
		{
			name:       "overflow",
			args:       []string{"fartapp", "999999999999999999999999"},
			wantCode:   1,
			wantStderr: "invalid intensity \"999999999999999999999999\": must be a canonical integer from 1 to 5\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			gotCode := run(tt.args, &stdout, &stderr)
			if gotCode != tt.wantCode {
				t.Errorf("exit code = %d, want %d", gotCode, tt.wantCode)
			}
			if got := stdout.String(); got != tt.wantStdout {
				t.Errorf("stdout = %q, want %q", got, tt.wantStdout)
			}
			if got := stderr.String(); got != tt.wantStderr {
				t.Errorf("stderr = %q, want %q", got, tt.wantStderr)
			}
		})
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("test write failure")
}

func TestCLIOutputFailure(t *testing.T) {
	var stderr bytes.Buffer
	if code := run([]string{"fartapp", "1"}, failingWriter{}, &stderr); code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if want := "write output: test write failure\n"; stderr.String() != want {
		t.Fatalf("stderr = %q, want %q", stderr.String(), want)
	}
}

func TestCLIDiagnosticFailureDoesNotPanic(t *testing.T) {
	if code := run([]string{"fartapp"}, &bytes.Buffer{}, failingWriter{}); code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
}

func TestExecutable(t *testing.T) {
	binaryName := "fartapp"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	binaryPath := filepath.Join(t.TempDir(), binaryName)

	build := exec.Command("go", "build", "-o", binaryPath, "./cmd/fartapp")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build executable: %v\n%s", err, output)
	}

	command := exec.Command(binaryPath, "3")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("run executable: %v\n%s", err, output)
	}
	if got, want := string(output), "braaap (respectable)\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}

	command = exec.Command(binaryPath, "--help")
	output, err = command.CombinedOutput()
	if err != nil || string(output) != rootHelp {
		t.Fatalf("run executable root help: %v\n%s", err, output)
	}

	command = exec.Command(binaryPath, "help", "scenario", "validate")
	output, err = command.CombinedOutput()
	if err != nil || string(output) != scenarioValidateHelp {
		t.Fatalf("run executable help: %v\n%s", err, output)
	}
}

func FuzzRun(f *testing.F) {
	for _, seed := range []string{
		"", "0", "1", "5", "6", "+1", "01", " 1 ", "nope", "text", "json", "help", "-h", "--help",
		"law", "list", "inspect", "scenario", "validate", "restriction", "predict", "walk", "simulate",
		"earth.continuum.si", "earth.continuum.si@v0alpha1",
		"conformance.relation.atemporal",
		`{"schema":"fart.scenario-probe/v0alpha1","law_context_set":{"contexts":[{"id":"conformance.relation.atemporal","version":"v0alpha1","scope_id":"s0"}]},"scope":{"id":"s0"},"capability_requests":[{"id":"catalog.inspect"}]}`,
		nestedDiagnosticSeed(), strings.Repeat("9", 200), "\x00",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		invoke := func(args []string) (int, string, string) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			code := Run(args, strings.NewReader(input), &stdout, &stderr)
			return code, stdout.String(), stderr.String()
		}

		argumentSets := [][]string{
			{"fartapp", input},
			{"fartapp", "law", "inspect", input},
			{"fartapp", "law", "list", "--format", input},
			{"fartapp", "help", input},
			{"fartapp", "scenario", "validate", "-", "--format", "json"},
		}
		for _, args := range argumentSets {
			code1, stdout1, stderr1 := invoke(args)
			code2, stdout2, stderr2 := invoke(args)
			if code1 != code2 || stdout1 != stdout2 || stderr1 != stderr2 {
				t.Fatalf("nondeterministic result for %q", args)
			}
			if code1 != 0 && code1 != 1 {
				t.Fatalf("exit code = %d, want 0 or 1", code1)
			}
			if stdout1 != "" && stderr1 != "" {
				t.Fatalf("mixed stdout %q and stderr %q", stdout1, stderr1)
			}
			limit := 256
			if code1 == 0 && len(args) >= 2 &&
				(args[1] == "help" || args[1] == "-h" || args[1] == "--help") {
				limit = 8 * 1024
			} else if len(args) >= 2 && args[1] == "scenario" {
				limit = 32 * 1024
			} else if code1 == 0 && len(args) >= 2 && args[1] == "law" {
				limit = 16 * 1024
			}
			if len(stdout1)+len(stderr1) > limit {
				t.Fatalf("output length = %d, want at most %d", len(stdout1)+len(stderr1), limit)
			}
		}
	})
}

func nestedDiagnosticSeed() string {
	key := strings.Repeat(`\u0001`, 128)
	return strings.Repeat(`{"`+key+`":`, 34) + "0" + strings.Repeat("}", 34)
}
