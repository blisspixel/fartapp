package repoquality

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/blisspixel/fartapp/internal/assurance"
)

const (
	maxAssuranceSourceFiles = 512
	maxAssuranceSourceBytes = 32 << 20
)

var benchmarkRow = regexp.MustCompile(`(?m)^\| ([A-Z]{2,8}-[0-9]{3}) \|`)
var assuranceMilestoneHeading = regexp.MustCompile(`(?m)^## (v[0-9]{1,2}\.[0-9]{1,2}):`)

func runAssurance(args []string, stdout, stderr io.Writer) int {
	write := false
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		if _, err := io.WriteString(stdout, "usage: repoquality assurance [--write]\nChecks metadata references without executing declared checks.\n--write regenerates only docs/INVARIANTS.md after reference validation.\n"); err != nil {
			return 1
		}
		return 0
	}
	if len(args) == 1 && args[0] == "--write" {
		write = true
	} else if len(args) != 0 {
		writeDiagnostic(stderr, "assurance: expected no arguments or --write\n")
		return 1
	}
	root, err := FindRoot(".")
	if err != nil {
		writeDiagnostic(stderr, "assurance: %v\n", err)
		return 1
	}
	result, err := checkAssurance(root, write)
	if err != nil {
		writeDiagnostic(stderr, "assurance: %v\n", err)
		return 1
	}
	if _, err := io.WriteString(stdout, result.Report); err != nil {
		writeDiagnostic(stderr, "assurance: output failed: %v\n", err)
		return 1
	}
	if len(result.Failures) != 0 {
		writeDiagnostic(stderr, "assurance policy failed: %s\n", joinFailures(result.Failures))
		return 1
	}
	return 0
}

// CheckAssurance validates declared metadata against bounded, contained source
// files. It does not execute tests, infer case applicability, or ratify evidence.
func CheckAssurance(root string) (CheckResult, error) { return checkAssurance(root, false) }

func checkAssurance(root string, write bool) (CheckResult, error) {
	reader := assuranceReader{root: root, files: make(map[string][]byte)}
	data, err := reader.read(assurance.RegistryPath, assurance.MaxRegistryBytes)
	if err != nil {
		return CheckResult{}, err
	}
	registry, err := assurance.Parse(data)
	if err != nil {
		return CheckResult{}, err
	}
	benchmarks, err := reader.read("docs/VERIFICATION.md", maxPolicyFileBytes)
	if err != nil {
		return CheckResult{}, err
	}
	benchmarkIDs := make(map[string]bool)
	for _, row := range benchmarkRow.FindAllSubmatch(benchmarks, -1) {
		id := string(row[1])
		if benchmarkIDs[id] {
			return CheckResult{}, fmt.Errorf("duplicate verification benchmark ID: %s", id)
		}
		benchmarkIDs[id] = true
	}
	roadmap, err := reader.read("ROADMAP.md", maxPolicyFileBytes)
	if err != nil {
		return CheckResult{}, err
	}
	milestones := make(map[string]bool)
	for _, heading := range assuranceMilestoneHeading.FindAllSubmatch(roadmap, -1) {
		id := string(heading[1])
		if milestones[id] {
			return CheckResult{}, fmt.Errorf("duplicate roadmap milestone: %s", id)
		}
		milestones[id] = true
	}
	declarations := make(map[string]*ast.File)
	uniqueChecks := make(map[string]bool)
	executable, planned := 0, 0
	for _, invariant := range registry.List() {
		if !milestones[invariant.Milestone] {
			return CheckResult{}, fmt.Errorf("%s: missing roadmap milestone %s", invariant.ID, invariant.Milestone)
		}
		if invariant.Lifecycle == assurance.ExecutableCandidate {
			executable++
		} else {
			planned++
		}
		for _, reference := range append(invariant.Evidence, invariant.Counterexamples...) {
			if _, err := reader.read(reference.Path, maxPolicyFileBytes); err != nil {
				return CheckResult{}, fmt.Errorf("%s evidence: %w", invariant.ID, err)
			}
		}
		for _, reference := range invariant.RelatedBenchmarks {
			if !benchmarkIDs[reference.ID] {
				return CheckResult{}, fmt.Errorf("%s: missing separate verification benchmark %s", invariant.ID, reference.ID)
			}
		}
		for _, check := range invariant.Checks {
			file, exists := declarations[check.File]
			if !exists {
				data, err := reader.read(check.File, maxPolicyFileBytes)
				if err != nil {
					return CheckResult{}, fmt.Errorf("%s check: %w", invariant.ID, err)
				}
				file, err = parseAssuranceTest(check.File, data)
				if err != nil {
					return CheckResult{}, fmt.Errorf("%s: %w", invariant.ID, err)
				}
				declarations[check.File] = file
			}
			if !hasAssuranceTest(file, check.Name) {
				return CheckResult{}, fmt.Errorf("%s: %s lacks one real Go test declaration %s", invariant.ID, check.File, check.Name)
			}
			uniqueChecks[check.File+":"+check.Name] = true
		}
	}
	generated := []byte(registry.Markdown())
	var failures []string
	if write {
		if err := writeAssuranceReference(root, generated); err != nil {
			return CheckResult{}, err
		}
	} else {
		retained, err := reader.read(assurance.GeneratedPath, maxPolicyFileBytes)
		if err != nil {
			return CheckResult{}, err
		}
		// Git may use CRLF in a worktree. This is the only normalization; all
		// content, whitespace and the final newline otherwise remain exact.
		if !bytes.Equal(bytes.ReplaceAll(retained, []byte("\r\n"), []byte("\n")), generated) {
			failures = append(failures, "generated invariant reference drift; run go run ./tools/repoquality assurance --write")
		}
	}
	report := fmt.Sprintf("assurance metadata references verified: %d invariants (%d executable candidates, %d design candidates), %d distinct Go declarations; checks not executed\n", executable+planned, executable, planned, len(uniqueChecks))
	if write {
		report += "generated docs/INVARIANTS.md from the validated registry\n"
	}
	return CheckResult{Report: report, Failures: failures}, nil
}

type assuranceReader struct {
	root  string
	files map[string][]byte
	bytes int
}

func (reader *assuranceReader) read(relative string, maximum int64) ([]byte, error) {
	if data, exists := reader.files[relative]; exists {
		return data, nil
	}
	if len(reader.files) >= maxAssuranceSourceFiles {
		return nil, fmt.Errorf("assurance source count exceeds %d", maxAssuranceSourceFiles)
	}
	candidate, err := assuranceSourcePath(reader.root, relative)
	if err != nil {
		return nil, err
	}
	data, err := readFileLimited(candidate, maximum)
	if err != nil {
		return nil, err
	}
	if len(data) > maxAssuranceSourceBytes-reader.bytes {
		return nil, fmt.Errorf("assurance source bytes exceed %d", maxAssuranceSourceBytes)
	}
	reader.bytes += len(data)
	reader.files[relative] = data
	return data, nil
}

func assuranceSourcePath(root, relative string) (string, error) {
	if !assurance.ValidRepositoryPath(relative) {
		return "", fmt.Errorf("invalid assurance source path: %s", quote(relative))
	}
	candidate, err := repositoryPath(root, relative)
	if err != nil {
		return "", err
	}
	resolved, err := containedPath(root, candidate)
	if err != nil {
		return "", err
	}
	if resolved != relative {
		return "", fmt.Errorf("assurance source alias is not canonical: %s", relative)
	}
	return candidate, nil
}

func parseAssuranceTest(relative string, data []byte) (*ast.File, error) {
	name := filepath.Base(relative)
	if !strings.HasSuffix(name, "_test.go") || strings.HasPrefix(name, "_") || strings.HasPrefix(name, ".") {
		return nil, fmt.Errorf("registered test file is ignored by Go: %s", relative)
	}
	// Registry checks in this revision must participate in all three portable
	// Go CI targets without extra tags or cgo. MatchFile receives only the same
	// bounded bytes later parsed by the AST checker, never another source read.
	for _, operatingSystem := range []string{"windows", "darwin", "linux"} {
		context := build.Default
		context.GOOS, context.GOARCH, context.CgoEnabled = operatingSystem, "amd64", false
		context.BuildTags, context.ToolTags = nil, nil
		context.OpenFile = func(string) (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(data)), nil }
		matched, err := context.MatchFile(".", name)
		if err != nil || !matched {
			return nil, fmt.Errorf("registered test %s is excluded or has invalid build constraints for %s: %v", relative, operatingSystem, err)
		}
	}
	file, err := parser.ParseFile(token.NewFileSet(), relative, data, parser.AllErrors)
	if err != nil {
		return nil, fmt.Errorf("invalid registered Go test file %s: %w", relative, err)
	}
	return file, nil
}

func hasAssuranceTest(file *ast.File, name string) bool {
	testingAlias := ""
	for _, imported := range file.Imports {
		importPath, err := strconv.Unquote(imported.Path.Value)
		if err != nil || importPath != "testing" {
			continue
		}
		testingAlias = "testing"
		if imported.Name != nil {
			testingAlias = imported.Name.Name
		}
	}
	if testingAlias == "" || testingAlias == "." || testingAlias == "_" {
		return false
	}
	count, valid := 0, false
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name.Name != name {
			continue
		}
		count++
		if function.Recv != nil || function.Type.TypeParams != nil || function.Type.Results != nil || function.Body == nil || len(function.Body.List) == 0 ||
			function.Type.Params == nil || len(function.Type.Params.List) != 1 {
			continue
		}
		parameter := function.Type.Params.List[0]
		if len(parameter.Names) > 1 {
			continue
		}
		pointer, ok := parameter.Type.(*ast.StarExpr)
		if !ok {
			continue
		}
		selection, ok := pointer.X.(*ast.SelectorExpr)
		if !ok {
			continue
		}
		qualifier, ok := selection.X.(*ast.Ident)
		if !ok || qualifier.Name != testingAlias {
			continue
		}
		wantType := "T"
		if strings.HasPrefix(name, "Fuzz") {
			wantType = "F"
		}
		valid = selection.Sel.Name == wantType
	}
	return count == 1 && valid
}

func writeAssuranceReference(root string, data []byte) error {
	candidate, err := assuranceSourcePath(root, assurance.GeneratedPath)
	if err != nil {
		return err
	}
	if info, err := os.Stat(candidate); err == nil && !info.Mode().IsRegular() {
		return fmt.Errorf("generated reference must be a regular file")
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}
	// Only the fixed generated leaf is writable. A same-directory temporary
	// keeps interruption from leaving half of an otherwise valid reference.
	temporary, err := os.CreateTemp(filepath.Dir(candidate), ".assurance-reference-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, candidate)
}
