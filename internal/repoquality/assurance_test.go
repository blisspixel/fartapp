package repoquality

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/blisspixel/fartapp/internal/assurance"
)

func TestAssuranceChecksDeclarationsAndDriftWithoutExecutingThem(t *testing.T) {
	root, invariant := assuranceFixture(t)
	result, err := CheckAssurance(root)
	if err != nil || len(result.Failures) != 0 || !strings.Contains(result.Report, "1 invariants (1 executable candidates, 0 design candidates), 2 distinct Go declarations; checks not executed") {
		t.Fatalf("valid metadata check = %+v, %v", result, err)
	}
	// The fixture has both panic-bearing initialization and panic-bearing test
	// bodies. Merely validating their declarations must neither compile nor run.
	retainedPath := filepath.Join(root, filepath.FromSlash(assurance.GeneratedPath))
	original, _ := os.ReadFile(retainedPath)
	writeFile(t, retainedPath, "edited generated text\n")
	result, err = CheckAssurance(root)
	if err != nil || len(result.Failures) != 1 || !strings.Contains(result.Failures[0], "drift") {
		t.Fatalf("reference drift escaped: %+v, %v", result, err)
	}
	if data, _ := os.ReadFile(retainedPath); string(data) != "edited generated text\n" {
		t.Fatal("read-only gate rewrote documentation")
	}
	result, err = checkAssurance(root, true)
	if err != nil || len(result.Failures) != 0 || !strings.Contains(result.Report, "generated docs/INVARIANTS.md") {
		t.Fatalf("regeneration = %+v, %v", result, err)
	}
	if data, _ := os.ReadFile(retainedPath); !bytes.Equal(data, original) {
		t.Fatal("regeneration did not restore exact source-derived bytes")
	}
	writeFile(t, retainedPath, strings.ReplaceAll(string(original), "\n", "\r\n"))
	if result, err := CheckAssurance(root); err != nil || len(result.Failures) != 0 {
		t.Fatalf("Git CRLF worktree rejected: %+v, %v", result, err)
	}
	// Referencing a separate benchmark ID is required even if the same token is
	// a valid invariant ID elsewhere; namespace coincidence supplies no evidence.
	invariant.RelatedBenchmarks[0].ID = "CLI-001"
	writeAssuranceFixtureRegistry(t, root, invariant)
	if _, err := checkAssurance(root, true); err == nil || !strings.Contains(err.Error(), "missing separate verification benchmark") {
		t.Fatalf("benchmark namespace collapsed: %v", err)
	}
	if data, _ := os.ReadFile(retainedPath); string(data) != strings.ReplaceAll(string(original), "\n", "\r\n") {
		t.Fatal("failed validation rewrote the generated reference")
	}
}

func TestAssuranceRejectsMissingMalformedAndAmbiguousSourceEvidence(t *testing.T) {
	for _, mutation := range []struct {
		name   string
		change func(*testing.T, string, *assurance.Invariant)
	}{
		{"missing registry", func(t *testing.T, root string, _ *assurance.Invariant) {
			if err := os.Remove(filepath.Join(root, filepath.FromSlash(assurance.RegistryPath))); err != nil {
				t.Fatal(err)
			}
		}},
		{"malformed registry", func(t *testing.T, root string, _ *assurance.Invariant) {
			writeFile(t, filepath.Join(root, filepath.FromSlash(assurance.RegistryPath)), `{"schema":null}`)
		}},
		{"missing benchmark document", func(t *testing.T, root string, _ *assurance.Invariant) {
			if err := os.Remove(filepath.Join(root, "docs", "VERIFICATION.md")); err != nil {
				t.Fatal(err)
			}
		}},
		{"duplicate benchmark IDs", func(t *testing.T, root string, _ *assurance.Invariant) {
			writeFile(t, filepath.Join(root, "docs", "VERIFICATION.md"), "| ONT-001 | first |\n| ONT-001 | second |\n")
		}},
		{"missing roadmap", func(t *testing.T, root string, _ *assurance.Invariant) {
			if err := os.Remove(filepath.Join(root, "ROADMAP.md")); err != nil {
				t.Fatal(err)
			}
		}},
		{"duplicate milestone", func(t *testing.T, root string, _ *assurance.Invariant) {
			writeFile(t, filepath.Join(root, "ROADMAP.md"), "## v0.8: First\n## v0.8: Second\n")
		}},
		{"missing milestone", func(t *testing.T, root string, invariant *assurance.Invariant) {
			invariant.Milestone = "v99.99"
			writeAssuranceFixtureRegistry(t, root, *invariant)
		}},
		{"missing evidence", func(t *testing.T, root string, invariant *assurance.Invariant) {
			invariant.Evidence[0].Path = "missing.md"
			writeAssuranceFixtureRegistry(t, root, *invariant)
		}},
		{"directory evidence", func(t *testing.T, root string, invariant *assurance.Invariant) {
			invariant.Evidence[0].Path = "internal/example"
			writeAssuranceFixtureRegistry(t, root, *invariant)
		}},
		{"escaping evidence", func(t *testing.T, root string, invariant *assurance.Invariant) {
			invariant.Evidence[0].Path = "../outside"
			writeAssuranceFixtureRegistry(t, root, *invariant)
		}},
		{"missing declaration file", func(t *testing.T, root string, invariant *assurance.Invariant) {
			invariant.Checks[0].File = "internal/example/absent_test.go"
			writeAssuranceFixtureRegistry(t, root, *invariant)
		}},
		{"absent exact declaration", func(t *testing.T, root string, invariant *assurance.Invariant) {
			invariant.Checks[0].Name = "TestAbsent"
			writeAssuranceFixtureRegistry(t, root, *invariant)
		}},
		{"invalid Go syntax", func(t *testing.T, root string, _ *assurance.Invariant) {
			writeFile(t, filepath.Join(root, "internal", "example", "example_test.go"), "package example\nfunc TestDeclared(")
		}},
		{"missing generated reference", func(t *testing.T, root string, _ *assurance.Invariant) {
			if err := os.Remove(filepath.Join(root, filepath.FromSlash(assurance.GeneratedPath))); err != nil {
				t.Fatal(err)
			}
		}},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			root, invariant := assuranceFixture(t)
			mutation.change(t, root, &invariant)
			if _, err := CheckAssurance(root); err == nil {
				t.Fatal("invalid metadata or source evidence passed")
			}
		})
	}
}

func TestAssuranceKeepsPlannedMetadataSeparateFromExecutableDeclarations(t *testing.T) {
	root, _ := assuranceFixture(t)
	registry, err := assurance.BuiltIn()
	if err != nil {
		t.Fatal(err)
	}
	planned, err := registry.Inspect("ACC-001")
	if err != nil {
		t.Fatal(err)
	}
	planned.Evidence = []assurance.EvidenceReference{{Path: "docs/evidence.md", Description: "Owning planned design only."}}
	writeAssuranceFixtureRegistry(t, root, planned)
	result, err := checkAssurance(root, true)
	if err != nil || len(result.Failures) != 0 || !strings.Contains(result.Report, "0 executable candidates, 1 design candidates") || !strings.Contains(result.Report, "0 distinct Go declarations; checks not executed") {
		t.Fatalf("planned metadata acquired execution claims: %+v, %v", result, err)
	}
}

func TestAssuranceRequiresPortableRealGoTestAndFuzzDeclarations(t *testing.T) {
	for _, test := range []struct {
		name, file, source, check string
		accepted                  bool
	}{
		{"test", "example_test.go", "import \"testing\"\nfunc TestDeclared(t *testing.T) { t.Log(\"reference only\") }", "TestDeclared", true},
		{"aliased testing", "example_test.go", "import qa \"testing\"\nfunc TestDeclared(t *qa.T) { t.Log(\"reference only\") }", "TestDeclared", true},
		{"unnamed parameter", "example_test.go", "import \"testing\"\nfunc TestDeclared(*testing.T) { panic(\"never run\") }", "TestDeclared", true},
		{"fuzz", "example_test.go", "import \"testing\"\nfunc FuzzDeclared(f *testing.F) { f.Add(1) }", "FuzzDeclared", true},
		{"comment spoof", "example_test.go", "// func TestDeclared(t *testing.T) {}", "TestDeclared", false},
		{"function value spoof", "example_test.go", "import \"testing\"\nvar TestDeclared = func(t *testing.T) { t.Fail() }", "TestDeclared", false},
		{"method", "example_test.go", "import \"testing\"\ntype X struct{}\nfunc (x X) TestDeclared(t *testing.T) { t.Fail() }", "TestDeclared", false},
		{"wrong testing type", "example_test.go", "import \"testing\"\nfunc TestDeclared(t *testing.B) { t.Fail() }", "TestDeclared", false},
		{"fuzz with T", "example_test.go", "import \"testing\"\nfunc FuzzDeclared(t *testing.T) { t.Fail() }", "FuzzDeclared", false},
		{"value parameter", "example_test.go", "import \"testing\"\nfunc TestDeclared(t testing.T) { t.Fail() }", "TestDeclared", false},
		{"variadic", "example_test.go", "import \"testing\"\nfunc TestDeclared(t ...*testing.T) { panic(0) }", "TestDeclared", false},
		{"multiple named parameters", "example_test.go", "import \"testing\"\nfunc TestDeclared(t, u *testing.T) { t.Fail() }", "TestDeclared", false},
		{"multiple parameter fields", "example_test.go", "import \"testing\"\nfunc TestDeclared(t *testing.T, n int) { t.Fail() }", "TestDeclared", false},
		{"no parameter", "example_test.go", "func TestDeclared() { panic(0) }", "TestDeclared", false},
		{"return value", "example_test.go", "import \"testing\"\nfunc TestDeclared(t *testing.T) bool { return true }", "TestDeclared", false},
		{"generic", "example_test.go", "import \"testing\"\nfunc TestDeclared[T any](t *testing.T) { t.Fail() }", "TestDeclared", false},
		{"missing body", "example_test.go", "import \"testing\"\nfunc TestDeclared(t *testing.T)", "TestDeclared", false},
		{"empty body", "example_test.go", "import \"testing\"\nfunc TestDeclared(t *testing.T) {}", "TestDeclared", false},
		{"duplicate function", "example_test.go", "import \"testing\"\nfunc TestDeclared(t *testing.T) { t.Fail() }\nfunc TestDeclared(t *testing.T) { t.Fail() }", "TestDeclared", false},
		{"nonstandard testing import", "example_test.go", "import testing \"example/testing\"\nfunc TestDeclared(t *testing.T) { t.Fail() }", "TestDeclared", false},
		{"mismatched qualifier", "example_test.go", "import qa \"testing\"\nfunc TestDeclared(t *testing.T) { t.Fail() }", "TestDeclared", false},
		{"dot import", "example_test.go", "import . \"testing\"\nfunc TestDeclared(t *T) { t.Fail() }", "TestDeclared", false},
		{"blank import", "example_test.go", "import _ \"testing\"\nfunc TestDeclared(t *testing.T) { t.Fail() }", "TestDeclared", false},
		{"unqualified fake T", "example_test.go", "import \"testing\"\ntype T struct{}\nfunc TestDeclared(t *T) { panic(0) }", "TestDeclared", false},
		{"non-test suffix", "example.go", "import \"testing\"\nfunc TestDeclared(t *testing.T) { t.Fail() }", "TestDeclared", false},
		{"ignored underscore file", "_example_test.go", "import \"testing\"\nfunc TestDeclared(t *testing.T) { t.Fail() }", "TestDeclared", false},
		{"ignored dot file", ".example_test.go", "import \"testing\"\nfunc TestDeclared(t *testing.T) { t.Fail() }", "TestDeclared", false},
		{"platform-only suffix", "example_linux_test.go", "import \"testing\"\nfunc TestDeclared(t *testing.T) { t.Fail() }", "TestDeclared", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			file, err := parseAssuranceTest("internal/example/"+test.file, []byte("package example\n"+test.source+"\n"))
			accepted := err == nil && hasAssuranceTest(file, test.check)
			if accepted != test.accepted {
				t.Fatalf("declaration acceptance=%v want %v; error=%v", accepted, test.accepted, err)
			}
		})
	}
	for _, constraint := range []string{"//go:build ignore", "//go:build custom", "//go:build cgo", "//go:build windows", "//go:build !windows", "//go:build [invalid]", "//go:build linux\n//go:build linux"} {
		data := []byte(constraint + "\n\npackage example\nimport \"testing\"\nfunc TestDeclared(t *testing.T) { t.Fail() }\n")
		if _, err := parseAssuranceTest("internal/example/example_test.go", data); err == nil {
			t.Fatalf("excluded or malformed constraint admitted: %s", constraint)
		}
	}
}

func TestAssuranceSymlinkEscapesAndAliasesCannotSupplyEvidenceOrOverwriteFiles(t *testing.T) {
	for _, kind := range []string{"external evidence", "internal alias", "generated escape", "source escape"} {
		t.Run(kind, func(t *testing.T) {
			root, invariant := assuranceFixture(t)
			external := filepath.Join(t.TempDir(), "outside.txt")
			writeFile(t, external, "outside must remain intact")
			link, target := filepath.Join(root, "evidence-link.md"), external
			switch kind {
			case "internal alias":
				target = filepath.Join(root, "docs", "evidence.md")
			case "generated escape":
				link = filepath.Join(root, filepath.FromSlash(assurance.GeneratedPath))
			case "source escape":
				link = filepath.Join(root, "internal", "example", "example_test.go")
			}
			if kind == "generated escape" || kind == "source escape" {
				if err := os.Remove(link); err != nil {
					t.Fatal(err)
				}
			} else {
				invariant.Evidence[0].Path = "evidence-link.md"
				writeAssuranceFixtureRegistry(t, root, invariant)
			}
			if err := os.Symlink(target, link); err != nil {
				t.Skipf("symlinks unavailable: %v", err)
			}
			if _, err := CheckAssurance(root); err == nil {
				t.Fatal("symlink evidence accepted")
			}
			if kind == "generated escape" {
				if _, err := checkAssurance(root, true); err == nil {
					t.Fatal("regeneration followed external link")
				}
			}
			if data, _ := os.ReadFile(external); string(data) != "outside must remain intact" {
				t.Fatal("external target changed")
			}
		})
	}
}

func TestAssuranceBoundsTotalReferencedSourceWork(t *testing.T) {
	root := t.TempDir()
	reader := assuranceReader{root: root, files: make(map[string][]byte)}
	for index := 0; index < maxAssuranceSourceFiles; index++ {
		reader.files[fmt.Sprintf("file-%d", index)] = nil
	}
	if _, err := reader.read("extra", maxPolicyFileBytes); err == nil || !strings.Contains(err.Error(), "source count") {
		t.Fatalf("source-count budget ignored: %v", err)
	}
	reader.files = make(map[string][]byte)
	reader.bytes = maxAssuranceSourceBytes - 1
	writeFile(t, filepath.Join(root, "small"), "12")
	if _, err := reader.read("small", maxPolicyFileBytes); err == nil || !strings.Contains(err.Error(), "source bytes") {
		t.Fatalf("source-byte budget ignored: %v", err)
	}
	reader.bytes = 0
	if _, err := reader.read("small", 1); err == nil {
		t.Fatal("per-file limit ignored")
	}
	if _, err := reader.read("../outside", 1); err == nil {
		t.Fatal("lexical source escape ignored")
	}
}

func TestAssuranceMaintainerCommandIsBoundedAndPublicationIsExplicit(t *testing.T) {
	for _, args := range [][]string{{"--unknown"}, {"--write", "--write"}, {"--help", "ignored"}} {
		var stdout, stderr bytes.Buffer
		if code := runAssurance(args, &stdout, &stderr); code == 0 || stdout.Len() != 0 || stderr.Len() == 0 {
			t.Fatalf("bad args accepted: %v", args)
		}
	}
	var stdout, stderr bytes.Buffer
	if code := runAssurance([]string{"--help"}, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "without executing") || stderr.Len() != 0 {
		t.Fatal("bad command help")
	}
	if code := runAssurance([]string{"--help"}, rustCoverageFailedWriter{}, &stderr); code == 0 {
		t.Fatal("help output failure ignored")
	}
	t.Chdir(t.TempDir())
	if code := runAssurance(nil, &stdout, &stderr); code == 0 {
		t.Fatal("accepted a nonrepository working directory")
	}
	root, _ := assuranceFixture(t)
	t.Chdir(root)
	stdout.Reset()
	stderr.Reset()
	if code := runAssurance(nil, &stdout, &stderr); code != 0 || stderr.Len() != 0 {
		t.Fatalf("valid command failed: %s", stderr.String())
	}
	if code := runAssurance(nil, rustCoverageFailedWriter{}, &stderr); code == 0 {
		t.Fatal("command output failure ignored")
	}
	writeFile(t, filepath.Join(root, filepath.FromSlash(assurance.GeneratedPath)), "drift\n")
	if code := runAssurance(nil, &stdout, &stderr); code == 0 {
		t.Fatal("drift passed through command")
	}
	if code := runAssurance([]string{"--write"}, &stdout, &stderr); code != 0 {
		t.Fatalf("explicit regeneration failed: %s", stderr.String())
	}
	writeFile(t, filepath.Join(root, filepath.FromSlash(assurance.RegistryPath)), "invalid")
	if code := runAssurance(nil, &stdout, &stderr); code == 0 {
		t.Fatal("invalid source registry passed through command")
	}
}

func TestAssuranceRegenerationHandlesMissingAndInvalidDestinations(t *testing.T) {
	root, _ := assuranceFixture(t)
	target := filepath.Join(root, filepath.FromSlash(assurance.GeneratedPath))
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	if _, err := checkAssurance(root, true); err != nil {
		t.Fatalf("missing generated leaf not creatable: %v", err)
	}
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := writeAssuranceReference(root, []byte("data")); err == nil {
		t.Fatal("directory destination overwritten")
	}
	missing := filepath.Join(t.TempDir(), "missing")
	if err := writeAssuranceReference(missing, []byte("data")); err == nil {
		t.Fatal("missing root silently created")
	}
	if err := writeAssuranceReference(t.TempDir(), []byte("data")); err == nil {
		t.Fatal("missing documentation parent silently created")
	}
}

func TestRepositoryAssuranceRegistryHasConcreteCurrentReferences(t *testing.T) {
	root, err := FindRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	result, err := CheckAssurance(root)
	if err != nil || len(result.Failures) != 0 {
		t.Fatalf("repository registry references are not current: %+v, %v", result, err)
	}
}

func assuranceFixture(t *testing.T) (string, assurance.Invariant) {
	t.Helper()
	root := t.TempDir()
	registry, err := assurance.BuiltIn()
	if err != nil {
		t.Fatal(err)
	}
	invariant, err := registry.Inspect("CLI-001")
	if err != nil {
		t.Fatal(err)
	}
	invariant.Checks = []assurance.Check{
		{ID: "declared-test", Package: "./internal/example", File: "internal/example/example_test.go", Name: "TestDeclared"},
		{ID: "declared-fuzz", Package: "./internal/example", File: "internal/example/example_test.go", Name: "FuzzDeclared"},
	}
	invariant.Evidence = []assurance.EvidenceReference{{Path: "docs/evidence.md", Description: "Authored metadata fixture."}}
	invariant.Counterexamples = []assurance.EvidenceReference{{Path: "internal/example/example_test.go", Description: "Declarations are validated but must never execute during inspection."}}
	invariant.RelatedBenchmarks = []assurance.BenchmarkReference{{Namespace: "verification-benchmark", ID: "ONT-001", Relationship: "partial-support", Scope: "This fixture only verifies separate namespaces."}}
	writeFile(t, filepath.Join(root, "go.mod"), "module example\n\ngo 1.27.1\n")
	writeFile(t, filepath.Join(root, "ROADMAP.md"), "## v0.8: Fixture milestone\n")
	writeFile(t, filepath.Join(root, "docs", "evidence.md"), "Declared reference only.\n")
	writeFile(t, filepath.Join(root, "docs", "VERIFICATION.md"), "| ONT-001 | Separate verification benchmark |\n")
	writeFile(t, filepath.Join(root, "internal", "example", "example_test.go"), "package example\nimport \"testing\"\nfunc init() { panic(\"must never execute\") }\nfunc TestDeclared(t *testing.T) { panic(\"must never execute\") }\nfunc FuzzDeclared(f *testing.F) { panic(\"must never execute\") }\n")
	data := writeAssuranceFixtureRegistry(t, root, invariant)
	parsed, err := assurance.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, filepath.FromSlash(assurance.GeneratedPath)), parsed.Markdown())
	return root, invariant
}

func writeAssuranceFixtureRegistry(t *testing.T, root string, invariant assurance.Invariant) []byte {
	t.Helper()
	data, err := json.Marshal(struct {
		Schema     string                `json:"schema"`
		Invariants []assurance.Invariant `json:"invariants"`
	}{assurance.Schema, []assurance.Invariant{invariant}})
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, filepath.FromSlash(assurance.RegistryPath)), string(data))
	return data
}
