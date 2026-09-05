package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestAssuranceInspectionIsEmbeddedAndReadOnly(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, args := range [][]string{
		{"list"}, {"inspect", "CLI-001"}, {"inspect", "OBS-001"},
		{"list", "--format", "json"}, {"inspect", "PHY-001", "--format=json"},
		{"inspect", "ID-001", "--format", "json"},
	} {
		var stdout, stderr bytes.Buffer
		command := append([]string{"fartapp", "assurance"}, args...)
		if code := Run(command, tripwireReader{}, &stdout, &stderr); code != 0 || stderr.Len() != 0 {
			t.Fatalf("%q = (%d, %q, %q)", args, code, stdout.String(), stderr.String())
		}
		first := bytes.Clone(stdout.Bytes())
		stdout.Reset()
		if code := Run(command, tripwireReader{}, &stdout, &stderr); code != 0 || !bytes.Equal(first, stdout.Bytes()) {
			t.Fatalf("inspection changed without an input change: %q", args)
		}
		if len(args) > 2 && (strings.Contains(strings.Join(args, " "), "json")) {
			if !json.Valid(first) || bytes.Count(first, []byte("\n")) != 1 {
				t.Fatalf("invalid JSON framing: %q", first)
			}
		}
		if len(first) == 0 || len(first) > 1<<20 {
			t.Fatal("missing or unbounded assurance metadata")
		}
	}
}

func TestAssuranceHelpAndMalformedArgumentsDoNotReadInput(t *testing.T) {
	for _, args := range [][]string{
		{"help", "assurance"}, {"help", "assurance", "list"}, {"help", "assurance", "inspect"},
		{"assurance", "--help"}, {"assurance", "help"}, {"assurance", "list", "-h"},
		{"assurance", "inspect", "unknown", "--format", "json", "--help"},
	} {
		var stdout, stderr bytes.Buffer
		code := Run(append([]string{"fartapp"}, args...), tripwireReader{}, &stdout, &stderr)
		if code != 0 || stdout.String() != assuranceHelp || stderr.Len() != 0 {
			t.Fatalf("help %q = (%d, %q, %q)", args, code, stdout.String(), stderr.String())
		}
	}
	for _, args := range [][]string{
		{}, {"missing"}, {"list", "extra"}, {"inspect"}, {"inspect", "CLI-001", "extra"},
		{"list", "--format", "yaml"}, {"list", "--format"}, {"list", "--format", "--help"},
		{"list", "--format=json", "--format=text"}, {"list", "--help", "--help"},
		{"inspect", "not-an-invariant"}, {"inspect", "UNKNOWN-999", "--format", "json"},
		{"inspect", strings.Repeat("x", 4096) + "\x1b\n", "--format", "json"},
	} {
		var stdout, stderr bytes.Buffer
		code := Run(append([]string{"fartapp", "assurance"}, args...), tripwireReader{}, &stdout, &stderr)
		if code != 1 || stdout.Len() != 0 || stderr.Len() == 0 || stderr.Len() > 512 ||
			bytes.ContainsRune(stderr.Bytes(), '\x1b') {
			t.Fatalf("bad arguments %q = (%d, %q, %q)", args, code, stdout.String(), stderr.String())
		}
	}
}

func TestAssuranceOutputFailuresRemainObservable(t *testing.T) {
	for _, args := range [][]string{{"list"}, {"inspect", "CLI-001", "--format", "json"}, {"--help"}} {
		var stderr bytes.Buffer
		if code := Run(append([]string{"fartapp", "assurance"}, args...), tripwireReader{}, failingWriter{}, &stderr); code != 1 {
			t.Fatalf("output failure %q = %d", args, code)
		}
	}
}
