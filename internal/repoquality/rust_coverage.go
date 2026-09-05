package repoquality

import (
	"fmt"
	"io"
	"io/fs"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

var rustCoverageCrates = [...]string{"fart-domain", "fart-core", "fart-services", "fart-cli"}

func runRustCoverage(args []string, stdout, stderr io.Writer) int {
	options := coverageOptions{aggregate: 90, pkg: 80}
	seen := map[string]bool{}
	for index := 0; index < len(args); index++ {
		argument := args[index]
		if argument == "--help" || argument == "-h" {
			_, err := io.WriteString(stdout, "usage: repoquality rust-coverage --profile llvm-coverage.json [--aggregate 90] [--package 80]\n")
			if err != nil {
				writeDiagnostic(stderr, "rust-coverage: cannot write help: %v\n", err)
				return 1
			}
			return 0
		}
		name, value, inline := strings.Cut(argument, "=")
		if name != "--profile" && name != "--aggregate" && name != "--package" {
			writeDiagnostic(stderr, "rust-coverage: unknown option %s\n", quote(argument))
			return 1
		}
		if seen[name] {
			writeDiagnostic(stderr, "rust-coverage: repeated option %s\n", name)
			return 1
		}
		seen[name] = true
		if !inline {
			index++
			if index == len(args) {
				writeDiagnostic(stderr, "rust-coverage: %s requires a value\n", name)
				return 1
			}
			value = args[index]
		}
		if name == "--profile" {
			options.profile = value
			continue
		}
		minimum, err := strconv.ParseFloat(value, 64)
		if err != nil || !validCoverageMinimum(minimum) {
			writeDiagnostic(stderr, "rust-coverage: %s must be finite and between 0 and 100\n", name)
			return 1
		}
		if name == "--aggregate" {
			options.aggregate = minimum
		} else {
			options.pkg = minimum
		}
	}
	if options.profile == "" {
		writeDiagnostic(stderr, "rust-coverage: --profile is required and must not be empty\n")
		return 1
	}
	root, err := FindRoot(".")
	if err != nil {
		writeDiagnostic(stderr, "rust-coverage: %v\n", err)
		return 1
	}
	result, err := CheckRustCoverage(root, options)
	if err != nil {
		writeDiagnostic(stderr, "rust-coverage: %v\n", err)
		return 1
	}
	if _, err := io.WriteString(stdout, result.Report); err != nil {
		writeDiagnostic(stderr, "rust-coverage: cannot write report: %v\n", err)
		return 1
	}
	if len(result.Failures) > 0 {
		writeDiagnostic(stderr, "Rust coverage policy failed: %s\n", joinFailures(result.Failures))
		return 1
	}
	return 0
}

type rustCoverageExport struct {
	Type    string `json:"type"`
	Version string `json:"version"`
	Data    []struct {
		Files []struct {
			Filename string `json:"filename"`
			Summary  struct {
				Lines *struct {
					Count   *int64 `json:"count"`
					Covered *int64 `json:"covered"`
				} `json:"lines"`
			} `json:"summary"`
		} `json:"files"`
	} `json:"data"`
}

type rustLineCounts struct{ count, covered int64 }

// CheckRustCoverage checks LLVM per-file source line evidence against the actual
// four-crate source tree. Exported totals and percentages are never trusted.
// Every source file requires evidence, including files inside src/test, vendor,
// or generated directories. A zero-line file is allowed; a zero-line crate is
// not. The 4 MiB strict policy JSON limit requires a summary export.
func CheckRustCoverage(root string, options coverageOptions) (CheckResult, error) {
	if options.profile == "" || !validCoverageMinimum(options.aggregate) || !validCoverageMinimum(options.pkg) {
		return CheckResult{}, fmt.Errorf("a profile and finite coverage minima between 0 and 100 are required")
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return CheckResult{}, err
	}
	sources, err := rustCoverageSources(root)
	if err != nil {
		return CheckResult{}, err
	}
	profile := options.profile
	if !filepath.IsAbs(profile) {
		profile = filepath.Join(root, profile)
	}
	var export rustCoverageExport
	if err := readPolicyJSON(profile, &export); err != nil {
		return CheckResult{}, err
	}
	if export.Type != "llvm.coverage.json.export" || export.Version != "3.1.0" || len(export.Data) == 0 {
		return CheckResult{}, fmt.Errorf("expected nonempty LLVM coverage JSON export version 3.1.0")
	}
	seen := map[string]bool{}
	accounted := map[string]bool{}
	crates := map[string]rustLineCounts{}
	var total rustLineCounts
	for _, data := range export.Data {
		for _, file := range data.Files {
			lines := file.Summary.Lines
			if lines == nil || lines.Count == nil || lines.Covered == nil || *lines.Count < 0 || *lines.Covered < 0 || *lines.Covered > *lines.Count {
				return CheckResult{}, fmt.Errorf("invalid or missing line counts for %s", quote(file.Filename))
			}
			local, err := rustCoveragePath(root, file.Filename)
			if err != nil {
				return CheckResult{}, err
			}
			key := rustPathKey(local)
			if seen[key] {
				return CheckResult{}, fmt.Errorf("duplicate coverage file %s", quote(file.Filename))
			}
			seen[key] = true
			crate, own := sources[key]
			if !own {
				// Dependency and integration-test evidence does not contribute.
				// Unknown files in an owned source directory cannot stand in for
				// real source files or disappear through an exclusion rule.
				relative, err := filepath.Rel(root, local)
				if err == nil && rustOwnedSource(relative) {
					return CheckResult{}, fmt.Errorf("covered Rust source is absent from the source tree: %s", quote(file.Filename))
				}
				continue
			}
			accounted[key] = true
			entry := crates[crate]
			const maxCount = int64(1<<63 - 1)
			if *lines.Count > maxCount-total.count {
				return CheckResult{}, fmt.Errorf("coverage line total overflows int64")
			}
			entry.count += *lines.Count
			entry.covered += *lines.Covered
			crates[crate] = entry
			total.count += *lines.Count
			total.covered += *lines.Covered
		}
	}
	var missing []string
	for source := range sources {
		if !accounted[source] {
			relative, err := filepath.Rel(root, source)
			if err != nil {
				return CheckResult{}, err
			}
			missing = append(missing, filepath.ToSlash(relative))
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return CheckResult{}, fmt.Errorf("coverage evidence omits %d Rust source files, first: %s", len(missing), quote(missing[0]))
	}
	var result CheckResult
	var report strings.Builder
	for _, crate := range rustCoverageCrates {
		entry := crates[crate]
		if entry.count == 0 {
			return CheckResult{}, fmt.Errorf("crate %s has no executable line evidence", crate)
		}
		fmt.Fprintf(&report, "Rust crate line coverage: %s %.1f%% (%d/%d)\n", crate, 100*float64(entry.covered)/float64(entry.count), entry.covered, entry.count)
		if rustCoverageBelow(entry, options.pkg) {
			result.Failures = append(result.Failures, fmt.Sprintf("%s has %d/%d covered lines, below %g%% minimum", crate, entry.covered, entry.count, options.pkg))
		}
	}
	fmt.Fprintf(&report, "aggregate Rust source line coverage: %.1f%% (%d/%d)\n", 100*float64(total.covered)/float64(total.count), total.covered, total.count)
	if rustCoverageBelow(total, options.aggregate) {
		result.Failures = append(result.Failures, fmt.Sprintf("aggregate has %d/%d covered lines, below %g%% minimum", total.covered, total.count, options.aggregate))
	}
	result.Report = report.String()
	return result, nil
}

func rustCoverageBelow(counts rustLineCounts, minimum float64) bool {
	actual := new(big.Rat).SetFrac(big.NewInt(counts.covered), big.NewInt(counts.count))
	actual.Mul(actual, big.NewRat(100, 1))
	return actual.Cmp(new(big.Rat).SetFloat64(minimum)) < 0
}

func rustCoverageSources(root string) (map[string]string, error) {
	sources := map[string]string{}
	for _, crate := range rustCoverageCrates {
		relative := "crates/" + crate + "/src"
		directory := filepath.Join(root, filepath.FromSlash(relative))
		resolved, err := containedPath(root, directory)
		if err != nil || rustPathKey(resolved) != rustPathKey(filepath.FromSlash(relative)) {
			return nil, fmt.Errorf("source directory escapes or aliases its declared location: %s", relative)
		}
		info, err := os.Lstat(directory)
		if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("missing regular Rust source directory: %s", relative)
		}
		before := len(sources)
		err = filepath.WalkDir(directory, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("symlink in Rust source tree: %s", quote(path))
			}
			if entry.IsDir() || !strings.EqualFold(filepath.Ext(path), ".rs") {
				return nil
			}
			info, err := entry.Info()
			if err != nil || !info.Mode().IsRegular() {
				return fmt.Errorf("source must be a regular file: %s", quote(path))
			}
			sources[rustPathKey(path)] = crate
			return nil
		})
		if err != nil {
			return nil, err
		}
		if len(sources) == before {
			return nil, fmt.Errorf("crate %s has no Rust source files", crate)
		}
	}
	return sources, nil
}

func rustOwnedSource(relative string) bool {
	path := filepath.ToSlash(rustPathKey(relative))
	for _, crate := range rustCoverageCrates {
		if strings.HasPrefix(path, "crates/"+crate+"/src/") && strings.EqualFold(filepath.Ext(path), ".rs") {
			return true
		}
	}
	return false
}

func rustCoveragePath(root, name string) (string, error) {
	portableName := strings.ReplaceAll(name, "\\", "/")
	if name == "" || strings.IndexFunc(name, unicode.IsControl) >= 0 || strings.HasPrefix(portableName, "//?/") || strings.HasPrefix(portableName, "//./") {
		return "", fmt.Errorf("invalid coverage filename %s", quote(name))
	}
	if !filepath.IsAbs(name) {
		local, err := repositoryPath(root, name)
		if err != nil {
			return "", err
		}
		if _, err := containedPath(root, local); err != nil {
			return "", fmt.Errorf("coverage source path escapes repository: %s", quote(name))
		}
		return local, nil
	}
	// Absolute native paths are emitted by LLVM. Strip only the native volume
	// and root separator before checking portable, unambiguous components.
	portable := filepath.ToSlash(name[len(filepath.VolumeName(name)):])
	relative := strings.TrimPrefix(portable, "/")
	if !fs.ValidPath(relative) || relative == "." || strings.ContainsAny(relative, "\\:") {
		return "", fmt.Errorf("invalid absolute coverage filename %s", quote(name))
	}
	local := filepath.Clean(name)
	if inside, err := filepath.Rel(root, local); err == nil && inside != ".." && !strings.HasPrefix(inside, ".."+string(os.PathSeparator)) {
		if _, err := containedPath(root, local); err != nil {
			return "", fmt.Errorf("coverage source path escapes repository: %s", quote(name))
		}
	}
	return local, nil
}

func rustPathKey(path string) string {
	key := filepath.Clean(path)
	if runtime.GOOS == "windows" {
		key = strings.ToLower(key)
	}
	return key
}
