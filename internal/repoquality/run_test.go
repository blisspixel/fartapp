package repoquality

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunDispatchesAndRejectsUnknownCommands(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run(nil, &stdout, &stderr); code != 1 || !strings.Contains(stderr.String(), "usage:") {
		t.Fatalf("empty = (%d, %q, %q)", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"help"}, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "repository") {
		t.Fatalf("help = (%d, %q, %q)", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"wat"}, &stdout, &stderr); code != 1 || !strings.Contains(stderr.String(), "unknown") {
		t.Fatalf("unknown = (%d, %q, %q)", code, stdout.String(), stderr.String())
	}
}

func TestRunCoverageAndRepositoryAgainstTempAndCurrentTrees(t *testing.T) {
	root, err := FindRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"repository"}, &stdout, &stderr); code != 0 {
		t.Fatalf("repository = (%d, %q, %q)", code, stdout.String(), stderr.String())
	}
	if code := Run([]string{"coverage", "--help"}, &stdout, &stderr); code != 0 {
		t.Fatalf("coverage help = (%d, %q)", code, stderr.String())
	}
	if code := Run([]string{"fuzz", "--help"}, &stdout, &stderr); code != 0 {
		t.Fatalf("fuzz help = (%d, %q)", code, stderr.String())
	}
	if code := Run([]string{"repository", "--wat"}, &stdout, &stderr); code != 1 {
		t.Fatalf("repository unknown = %d", code)
	}
	if code := Run([]string{"coverage", "--profile"}, &stdout, &stderr); code != 1 {
		t.Fatalf("coverage missing profile = %d", code)
	}
	if code := Run([]string{"fuzz", "--time", "nope"}, &stdout, &stderr); code != 1 {
		t.Fatalf("fuzz bad time = %d", code)
	}

	temp := t.TempDir()
	writeFile(t, filepath.Join(temp, "go.mod"), "module example.com/lab\n")
	writeFile(t, filepath.Join(temp, "alpha.go"), "package alpha\n")
	writeFile(t, filepath.Join(temp, "coverage.out"), "mode: set\nexample.com/lab/alpha.go:1.1,2.2 10 1\n")
	if err := os.Chdir(temp); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	code := Run([]string{"coverage", "--profile", "coverage.out", "--aggregate=90", "--package=80"}, &stdout, &stderr)
	if err := os.Chdir(cwd); err != nil {
		t.Fatal(err)
	}
	if code != 0 {
		t.Fatalf("temp coverage = (%d, %q, %q)", code, stdout.String(), stderr.String())
	}
}

func TestQuoteAndDurationHelpers(t *testing.T) {
	if quote(strings.Repeat("x", 40)) == strings.Repeat("x", 40) {
		t.Fatal("quote did not bound the value")
	}
	if _, err := parseDurationArgument("0s"); err == nil {
		t.Fatal("zero duration was accepted")
	}
	if _, err := parseDurationArgument("2s"); err != nil {
		t.Fatal(err)
	}
}
