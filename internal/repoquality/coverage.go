package repoquality

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type coverageOptions struct {
	profile   string
	aggregate float64
	pkg       float64
}

func runCoverage(args []string, stdout, stderr io.Writer) int {
	options := coverageOptions{profile: "coverage.out", aggregate: 90, pkg: 80}
	for index := 0; index < len(args); index++ {
		argument := args[index]
		switch {
		case argument == "-h" || argument == "--help":
			_, _ = io.WriteString(stdout, "usage: repoquality coverage [--profile coverage.out] [--aggregate 90] [--package 80]\n")
			return 0
		case argument == "--profile":
			index++
			if index == len(args) {
				writeDiagnostic(stderr, "coverage: --profile requires a path\n")
				return 1
			}
			options.profile = args[index]
		case strings.HasPrefix(argument, "--profile="):
			options.profile = strings.TrimPrefix(argument, "--profile=")
		case argument == "--aggregate":
			index++
			if index == len(args) {
				writeDiagnostic(stderr, "coverage: --aggregate requires a number\n")
				return 1
			}
			value, err := strconv.ParseFloat(args[index], 64)
			if err != nil {
				writeDiagnostic(stderr, "coverage: invalid aggregate minimum\n")
				return 1
			}
			options.aggregate = value
		case strings.HasPrefix(argument, "--aggregate="):
			value, err := strconv.ParseFloat(strings.TrimPrefix(argument, "--aggregate="), 64)
			if err != nil {
				writeDiagnostic(stderr, "coverage: invalid aggregate minimum\n")
				return 1
			}
			options.aggregate = value
		case argument == "--package":
			index++
			if index == len(args) {
				writeDiagnostic(stderr, "coverage: --package requires a number\n")
				return 1
			}
			value, err := strconv.ParseFloat(args[index], 64)
			if err != nil {
				writeDiagnostic(stderr, "coverage: invalid package minimum\n")
				return 1
			}
			options.pkg = value
		case strings.HasPrefix(argument, "--package="):
			value, err := strconv.ParseFloat(strings.TrimPrefix(argument, "--package="), 64)
			if err != nil {
				writeDiagnostic(stderr, "coverage: invalid package minimum\n")
				return 1
			}
			options.pkg = value
		default:
			writeDiagnostic(stderr, "coverage: unknown option %s\n", quote(argument))
			return 1
		}
	}
	root, err := FindRoot(".")
	if err != nil {
		writeDiagnostic(stderr, "coverage: %v\n", err)
		return 1
	}
	result, err := CheckCoverage(root, options)
	if err != nil {
		writeDiagnostic(stderr, "coverage: %v\n", err)
		return 1
	}
	_, _ = io.WriteString(stdout, result.Report)
	if len(result.Failures) > 0 {
		writeDiagnostic(stderr, "coverage policy failed: %s\n", joinFailures(result.Failures))
		return 1
	}
	return 0
}

type CheckResult struct {
	Report   string
	Failures []string
}

var coverageRecord = regexp.MustCompile(`^(.+):([1-9]\d*)\.([1-9]\d*),([1-9]\d*)\.([1-9]\d*)[ \t]+(\d+)[ \t]+(\d+)$`)

func CheckCoverage(root string, options coverageOptions) (CheckResult, error) {
	if !validCoverageMinimum(options.aggregate) || !validCoverageMinimum(options.pkg) {
		return CheckResult{}, fmt.Errorf("coverage minima must be finite numbers between 0 and 100")
	}
	if options.profile == "" {
		return CheckResult{}, fmt.Errorf("coverage profile path must not be empty")
	}
	modulePath, err := modulePathFromGoMod(root)
	if err != nil {
		return CheckResult{}, err
	}
	profilePath := options.profile
	if !filepath.IsAbs(profilePath) {
		profilePath = filepath.Join(root, profilePath)
	}
	data, err := readFileLimited(profilePath, maxCoverageBytes)
	if err != nil {
		return CheckResult{}, fmt.Errorf("cannot read coverage profile %s: %w", options.profile, err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	if !scanner.Scan() {
		return CheckResult{}, fmt.Errorf("invalid Go coverage profile: %s", options.profile)
	}
	mode := scanner.Text()
	if mode != "mode: set" && mode != "mode: count" && mode != "mode: atomic" {
		return CheckResult{}, fmt.Errorf("invalid Go coverage profile: %s", options.profile)
	}

	type packageCounts struct {
		statements int64
		covered    int64
	}
	packages := map[string]*packageCounts{}
	generated := map[string]bool{}
	blocks := map[string]struct{}{}
	var totalStatements int64
	var totalCovered int64
	const maximumCount = int64(1<<63 - 1)
	for scanner.Scan() {
		line := scanner.Text()
		matches := coverageRecord.FindStringSubmatch(line)
		if matches == nil {
			return CheckResult{}, fmt.Errorf("invalid Go coverage record: %s", line)
		}
		profileFile := matches[1]
		var values [6]int64
		for index := range values {
			value, err := strconv.ParseInt(matches[index+2], 10, 64)
			if err != nil {
				return CheckResult{}, fmt.Errorf("coverage record contains an overflowing number: %s", line)
			}
			values[index] = value
		}
		if values[2] < values[0] || values[2] == values[0] && values[3] < values[1] {
			return CheckResult{}, fmt.Errorf("coverage record has a reversed source range: %s", line)
		}
		statements, count := values[4], values[5]
		if mode == "mode: set" && count > 1 {
			return CheckResult{}, fmt.Errorf("set-mode coverage count exceeds one: %s", line)
		}
		prefix := modulePath + "/"
		if !strings.HasPrefix(profileFile, prefix) {
			return CheckResult{}, fmt.Errorf("coverage record is outside the module: %s", profileFile)
		}
		relativeFile := strings.TrimPrefix(profileFile, prefix)
		localFile, err := repositoryPath(root, relativeFile)
		if err != nil {
			return CheckResult{}, err
		}
		if _, err := containedPath(root, localFile); err != nil {
			return CheckResult{}, fmt.Errorf("covered source escapes the module: %s", relativeFile)
		}
		if info, err := os.Stat(localFile); err != nil || !info.Mode().IsRegular() || filepath.Ext(localFile) != ".go" {
			return CheckResult{}, fmt.Errorf("covered source file not found: %s", relativeFile)
		}
		block := strings.Join(matches[1:6], ":")
		if _, exists := blocks[block]; exists {
			return CheckResult{}, fmt.Errorf("duplicate Go coverage block: %s", line)
		}
		blocks[block] = struct{}{}
		isGenerated, known := generated[localFile]
		if !known {
			isGenerated, err = generatedSource(localFile)
			if err != nil {
				return CheckResult{}, err
			}
			generated[localFile] = isGenerated
		}
		if isGenerated {
			continue
		}
		separator := strings.LastIndex(profileFile, "/")
		packagePath := profileFile[:separator]
		entry := packages[packagePath]
		if entry == nil {
			entry = &packageCounts{}
			packages[packagePath] = entry
		}
		if statements > maximumCount-totalStatements {
			return CheckResult{}, fmt.Errorf("coverage statement total overflows int64")
		}
		entry.statements += statements
		totalStatements += statements
		if count > 0 {
			entry.covered += statements
			totalCovered += statements
		}
	}
	if err := scanner.Err(); err != nil {
		return CheckResult{}, err
	}
	if totalStatements == 0 {
		return CheckResult{}, fmt.Errorf("coverage profile contains no non-generated package statements")
	}

	names := make([]string, 0, len(packages))
	for name := range packages {
		names = append(names, name)
	}
	sort.Strings(names)

	var report strings.Builder
	var failures []string
	for _, name := range names {
		entry := packages[name]
		if entry.statements == 0 {
			continue
		}
		coverage := 100.0 * float64(entry.covered) / float64(entry.statements)
		fmt.Fprintf(&report, "package coverage: %s %.1f%% (%d/%d)\n", name, coverage, entry.covered, entry.statements)
		if coverage < options.pkg {
			failures = append(failures, fmt.Sprintf("%s is %.1f%%, below %.1f%%", name, coverage, options.pkg))
		}
	}
	aggregate := 100.0 * float64(totalCovered) / float64(totalStatements)
	fmt.Fprintf(&report, "aggregate non-generated coverage: %.1f%% (%d/%d)\n", aggregate, totalCovered, totalStatements)
	if aggregate < options.aggregate {
		failures = append(failures, fmt.Sprintf("aggregate is %.1f%%, below %.1f%%", aggregate, options.aggregate))
	}
	return CheckResult{Report: report.String(), Failures: failures}, nil
}

func validCoverageMinimum(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0 && value <= 100
}

func modulePathFromGoMod(root string) (string, error) {
	data, err := readFileLimited(filepath.Join(root, "go.mod"), maxPolicyFileBytes)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "module" {
			modulePath := fields[1]
			if strings.HasPrefix(modulePath, `"`) || strings.HasPrefix(modulePath, "`") {
				modulePath, err = strconv.Unquote(modulePath)
				if err != nil {
					return "", fmt.Errorf("invalid quoted module path: %w", err)
				}
			}
			if modulePath != "" {
				return modulePath, nil
			}
		}
	}
	return "", fmt.Errorf("go.mod has no module path")
}

func generatedSource(path string) (bool, error) {
	data, err := readFileLimited(path, maxPolicyFileBytes)
	if err != nil {
		return false, err
	}
	file, err := parser.ParseFile(token.NewFileSet(), path, data, parser.PackageClauseOnly|parser.ParseComments)
	if err != nil {
		return false, fmt.Errorf("cannot inspect covered source %s: %w", path, err)
	}
	return ast.IsGenerated(file), nil
}
