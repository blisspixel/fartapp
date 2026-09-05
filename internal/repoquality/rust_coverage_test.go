package repoquality

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRustCoverageRecomputesSourceOnlyThresholds(t *testing.T) {
	root, files := rustCoverageFixture(t)
	for _, file := range files {
		rustFileLines(file)["covered"] = 90
		rustFileLines(file)["percent"] = 0
	}
	files = append(files, rustCoverageFile(filepath.Join(t.TempDir(), "dependency.rs"), 1000000, 0),
		rustCoverageFile("crates/fart-core/tests/integration.rs", 1000000, 0))
	writeRustCoverage(t, root, files)
	result, err := CheckRustCoverage(root, coverageOptions{profile: "coverage.json", aggregate: 90, pkg: 80})
	if err != nil || len(result.Failures) != 0 || !strings.Contains(result.Report, "90.0% (360/400)") {
		t.Fatalf("source-only recomputation = %+v, %v", result, err)
	}
	for _, crate := range rustCoverageCrates {
		if !strings.Contains(result.Report, crate+" 90.0% (90/100)") {
			t.Fatalf("missing crate evidence: %s", result.Report)
		}
	}
	rustFileLines(files[0])["covered"] = 50
	writeRustCoverage(t, root, files)
	result, err = CheckRustCoverage(root, coverageOptions{profile: filepath.Join(root, "coverage.json"), aggregate: 90, pkg: 80})
	if err != nil || len(result.Failures) != 2 || !strings.Contains(result.Report, "80.0% (320/400)") {
		t.Fatalf("real below-threshold accounting = %+v, %v", result, err)
	}
}

func TestRustCoverageRequiresEverySourceWithoutExclusionShortcuts(t *testing.T) {
	for _, nested := range []string{"test/uncovered.rs", "tests/uncovered.rs", "vendor/uncovered.rs", "generated.rs"} {
		t.Run(nested, func(t *testing.T) {
			root, files := rustCoverageFixture(t)
			path := "crates/fart-core/src/" + nested
			writeFile(t, filepath.Join(root, filepath.FromSlash(path)), "pub fn uncovered() {}\n")
			writeRustCoverage(t, root, files)
			if _, err := CheckRustCoverage(root, coverageOptions{profile: "coverage.json"}); err == nil || !strings.Contains(err.Error(), "omits 1 Rust source") {
				t.Fatalf("missing own source passed: %v", err)
			}
			files = append(files, rustCoverageFile(path, 100, 0))
			writeRustCoverage(t, root, files)
			result, err := CheckRustCoverage(root, coverageOptions{profile: "coverage.json", aggregate: 90, pkg: 80})
			if err != nil || len(result.Failures) != 2 || !strings.Contains(result.Report, "fart-core 50.0% (100/200)") {
				t.Fatalf("uncovered nested source was excluded: %+v, %v", result, err)
			}
		})
	}
	root, files := rustCoverageFixture(t)
	writeRustCoverage(t, root, files[:3])
	if _, err := CheckRustCoverage(root, coverageOptions{profile: "coverage.json"}); err == nil {
		t.Fatal("omitted crate passed")
	}
	writeFile(t, filepath.Join(root, "crates", "fart-core", "src", "module.rs"), "// no executable lines\n")
	files = append(files, rustCoverageFile("crates/fart-core/src/module.rs", 0, 0))
	writeRustCoverage(t, root, files)
	if result, err := CheckRustCoverage(root, coverageOptions{profile: "coverage.json", aggregate: 100, pkg: 100}); err != nil || len(result.Failures) != 0 {
		t.Fatalf("explicit zero-line file should be admissible: %+v, %v", result, err)
	}
}

func TestRustCoverageRefusesMalformedAndAmbiguousEvidence(t *testing.T) {
	for _, mutation := range []struct {
		name string
		edit func([]map[string]any) []map[string]any
	}{
		{"negative count", func(f []map[string]any) []map[string]any { rustFileLines(f[0])["count"] = -1; return f }},
		{"negative covered", func(f []map[string]any) []map[string]any { rustFileLines(f[0])["covered"] = -1; return f }},
		{"covered exceeds count", func(f []map[string]any) []map[string]any { rustFileLines(f[0])["covered"] = 101; return f }},
		{"fractional count", func(f []map[string]any) []map[string]any { rustFileLines(f[0])["count"] = 100.5; return f }},
		{"string count", func(f []map[string]any) []map[string]any { rustFileLines(f[0])["count"] = "100"; return f }},
		{"overflow count", func(f []map[string]any) []map[string]any {
			rustFileLines(f[0])["count"] = json.Number("9223372036854775808")
			return f
		}},
		{"overflow covered", func(f []map[string]any) []map[string]any {
			rustFileLines(f[0])["covered"] = json.Number("1e100")
			return f
		}},
		{"missing count", func(f []map[string]any) []map[string]any { delete(rustFileLines(f[0]), "count"); return f }},
		{"null count", func(f []map[string]any) []map[string]any { rustFileLines(f[0])["count"] = nil; return f }},
		{"missing lines", func(f []map[string]any) []map[string]any { f[0]["summary"] = map[string]any{}; return f }},
		{"empty crate", func(f []map[string]any) []map[string]any {
			rustFileLines(f[0])["count"] = 0
			rustFileLines(f[0])["covered"] = 0
			return f
		}},
		{"duplicate source", func(f []map[string]any) []map[string]any { return append(f, f[0]) }},
		{"duplicate relative alias", func(f []map[string]any) []map[string]any {
			return append(f, rustCoverageFile("crates/fart-domain/src/lib.rs", 100, 100))
		}},
		{"absent source", func(f []map[string]any) []map[string]any {
			return append(f, rustCoverageFile("crates/fart-core/src/missing.rs", 100, 100))
		}},
		{"total overflow", func(f []map[string]any) []map[string]any {
			rustFileLines(f[0])["count"] = int64(math.MaxInt64)
			return f
		}},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			root, files := rustCoverageFixture(t)
			writeRustCoverage(t, root, mutation.edit(files))
			if _, err := CheckRustCoverage(root, coverageOptions{profile: "coverage.json"}); err == nil {
				t.Fatal("invalid evidence passed")
			}
		})
	}
	root, files := rustCoverageFixture(t)
	for _, profile := range []string{
		`{`, `null`, `{"type":"llvm.coverage.json.export","version":"3.1.0","data":[]}`,
		strings.Replace(rustCoverageJSON(t, files), "3.1.0", "4.0.0", 1),
		strings.Replace(rustCoverageJSON(t, files), "llvm.coverage.json.export", "unknown", 1),
		strings.Replace(rustCoverageJSON(t, files), `"version":"3.1.0"`, `"version":"3.1.0","version":"3.1.0"`, 1),
		strings.Replace(rustCoverageJSON(t, files), `"covered":100`, `"covered":100,"covered":100`, 1),
	} {
		writeFile(t, filepath.Join(root, "coverage.json"), profile)
		if _, err := CheckRustCoverage(root, coverageOptions{profile: "coverage.json"}); err == nil {
			t.Fatal("invalid or ambiguous JSON passed")
		}
	}
}

func TestRustCoverageRejectsPathAndSymlinkEscapes(t *testing.T) {
	root, files := rustCoverageFixture(t)
	for _, name := range []string{"", "../outside.rs", "crates/fart-core/src/../src/lib.rs", "crates//fart-core/src/lib.rs", "bad\nfile.rs", "bad\x00file.rs", `relative\file.rs`, `\\?\C:\alias.rs`, `\\.\C:\alias.rs`,
		filepath.ToSlash(root) + "/crates/fart-core/src/../src/lib.rs", filepath.ToSlash(root) + "//crates/fart-core/src/lib.rs"} {
		copyFiles := append([]map[string]any(nil), files...)
		copyFiles = append(copyFiles, rustCoverageFile(name, 1, 1))
		writeRustCoverage(t, root, copyFiles)
		if _, err := CheckRustCoverage(root, coverageOptions{profile: "coverage.json"}); err == nil {
			t.Errorf("invalid path %q passed", name)
		}
	}
	outside := filepath.Join(t.TempDir(), "outside.rs")
	writeFile(t, outside, "pub fn outside() {}\n")
	link := filepath.Join(root, "crates", "fart-core", "src", "escape.rs")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	writeRustCoverage(t, root, append(files, rustCoverageFile(link, 1, 1)))
	if _, err := CheckRustCoverage(root, coverageOptions{profile: "coverage.json"}); err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("source symlink escape passed: %v", err)
	}
	if _, err := rustCoveragePath(root, link); err == nil {
		t.Fatal("absolute symlink escape passed path checking")
	}
	if _, err := rustCoveragePath(root, "crates/fart-core/src/escape.rs"); err == nil {
		t.Fatal("relative symlink escape passed path checking")
	}
}

func TestRustCoverageRequiresRealCratesAndBoundedRegularProfiles(t *testing.T) {
	root := t.TempDir()
	if _, err := CheckRustCoverage(root, coverageOptions{profile: "coverage.json"}); err == nil {
		t.Fatal("missing source tree passed")
	}
	for _, crate := range rustCoverageCrates {
		if err := os.MkdirAll(filepath.Join(root, "crates", crate, "src"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := CheckRustCoverage(root, coverageOptions{profile: "coverage.json"}); err == nil || !strings.Contains(err.Error(), "no Rust source") {
		t.Fatalf("empty source tree passed: %v", err)
	}
	root, _ = rustCoverageFixture(t)
	for _, profile := range []string{"missing.json", "."} {
		if _, err := CheckRustCoverage(root, coverageOptions{profile: profile}); err == nil {
			t.Fatal("missing or nonregular profile passed")
		}
	}
	writeFile(t, filepath.Join(root, "coverage.json"), strings.Repeat(" ", maxPolicyFileBytes+1))
	if _, err := CheckRustCoverage(root, coverageOptions{profile: "coverage.json"}); err == nil {
		t.Fatal("oversized profile passed")
	}
	for _, options := range []coverageOptions{{}, {profile: "x", aggregate: math.NaN()}, {profile: "x", pkg: math.Inf(1)}, {profile: "x", aggregate: -1}} {
		if _, err := CheckRustCoverage(root, options); err == nil {
			t.Fatal("invalid direct options passed")
		}
	}
}

func TestRustCoverageThresholdComparisonDoesNotRoundUp(t *testing.T) {
	if !rustCoverageBelow(rustLineCounts{count: math.MaxInt64, covered: math.MaxInt64 - 1}, 100) {
		t.Fatal("incomplete very large line count rounded up to 100 percent")
	}
	if rustCoverageBelow(rustLineCounts{count: 10, covered: 9}, 90) || !rustCoverageBelow(rustLineCounts{count: 10, covered: 8}, 90) {
		t.Fatal("exact threshold equality failed")
	}
}

func TestRustCoverageCLIParsesPolicyAndReportsFailures(t *testing.T) {
	root, files := rustCoverageFixture(t)
	writeRustCoverage(t, root, files)
	t.Chdir(root)
	for _, args := range [][]string{
		{"--profile", "coverage.json"}, {"--profile=coverage.json", "--aggregate=100", "--package", "100"}, {"--help"}, {"-h"},
	} {
		var stdout, stderr bytes.Buffer
		if code := runRustCoverage(args, &stdout, &stderr); code != 0 || stdout.Len() == 0 || stderr.Len() != 0 {
			t.Fatalf("valid CLI %v = %d, %q, %q", args, code, stdout.String(), stderr.String())
		}
	}
	for _, args := range [][]string{
		nil, {"--profile"}, {"--profile="}, {"--unknown"}, {"--aggregate"}, {"--package"},
		{"--profile=x", "--profile=y"}, {"--aggregate=NaN"}, {"--package=Inf"}, {"--aggregate=-1"}, {"--package=101"}, {"--package=abc"}, {"--profile=missing"},
	} {
		var stdout, stderr bytes.Buffer
		if code := runRustCoverage(args, &stdout, &stderr); code != 1 || stderr.Len() == 0 {
			t.Fatalf("invalid CLI %v = %d, %q", args, code, stderr.String())
		}
	}
	var stderr bytes.Buffer
	if code := runRustCoverage([]string{"--help"}, rustCoverageFailedWriter{}, &stderr); code != 1 {
		t.Fatal("failed help output succeeded")
	}
	if code := runRustCoverage([]string{"--profile=coverage.json"}, rustCoverageFailedWriter{}, &stderr); code != 1 {
		t.Fatal("failed report output succeeded")
	}
	rustFileLines(files[0])["covered"] = 0
	writeRustCoverage(t, root, files)
	var stdout bytes.Buffer
	if code := runRustCoverage([]string{"--profile=coverage.json"}, &stdout, &stderr); code != 1 || !strings.Contains(stderr.String(), "policy failed") {
		t.Fatal("below-threshold CLI succeeded")
	}
	t.Chdir(t.TempDir())
	if code := runRustCoverage([]string{"--profile=coverage.json"}, &stdout, &stderr); code != 1 {
		t.Fatal("CLI outside repository succeeded")
	}
}

type rustCoverageFailedWriter struct{}

func (rustCoverageFailedWriter) Write([]byte) (int, error) { return 0, errors.New("write failure") }

func rustCoverageFixture(t *testing.T) (string, []map[string]any) {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/lab\n")
	var files []map[string]any
	for _, crate := range rustCoverageCrates {
		path := filepath.Join(root, "crates", crate, "src", "lib.rs")
		writeFile(t, path, "pub fn example() {}\n")
		files = append(files, rustCoverageFile(path, 100, 100))
	}
	return root, files
}

func rustCoverageFile(name string, count, covered any) map[string]any {
	return map[string]any{"filename": name, "summary": map[string]any{"lines": map[string]any{"count": count, "covered": covered, "percent": 100}}}
}

func rustFileLines(file map[string]any) map[string]any {
	return file["summary"].(map[string]any)["lines"].(map[string]any)
}

func rustCoverageJSON(t *testing.T, files []map[string]any) string {
	t.Helper()
	data, err := json.Marshal(map[string]any{"type": "llvm.coverage.json.export", "version": "3.1.0", "data": []any{map[string]any{
		"files": files, "totals": map[string]any{"lines": map[string]any{"count": 0, "covered": 0, "percent": 0}},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func writeRustCoverage(t *testing.T, root string, files []map[string]any) {
	t.Helper()
	writeFile(t, filepath.Join(root, "coverage.json"), rustCoverageJSON(t, files))
}
