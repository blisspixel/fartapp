package repoquality

import (
	"fmt"
	"io"
	"strings"
)

func runRepository(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 {
		argument := args[0]
		if argument == "-h" || argument == "--help" {
			_, _ = io.WriteString(stdout, "usage: repoquality repository\n")
			return 0
		}
		writeDiagnostic(stderr, "repository: unknown option %s\n", quote(argument))
		return 1
	}
	root, err := FindRoot(".")
	if err != nil {
		writeDiagnostic(stderr, "repository: %v\n", err)
		return 1
	}
	result, err := CheckRepository(root)
	if err != nil {
		writeDiagnostic(stderr, "repository: %v\n", err)
		return 1
	}
	_, _ = io.WriteString(stdout, result.Report)
	if len(result.Failures) > 0 {
		writeDiagnostic(stderr, "repository policy failed: %s\n", joinFailures(result.Failures))
		return 1
	}
	return 0
}

func CheckRepository(root string) (CheckResult, error) {
	var report strings.Builder
	var failures []string
	checks := []struct {
		name string
		run  func(string) (CheckResult, error)
	}{
		{"dependencies", CheckDependencies},
		{"links", CheckLinks},
		{"media", CheckMedia},
		{"plugin", CheckPlugin},
		{"assurance", CheckAssurance},
	}
	for _, check := range checks {
		result, err := check.run(root)
		if err != nil {
			return CheckResult{}, fmt.Errorf("%s: %w", check.name, err)
		}
		report.WriteString(result.Report)
		failures = append(failures, result.Failures...)
	}
	return CheckResult{Report: report.String(), Failures: failures}, nil
}
