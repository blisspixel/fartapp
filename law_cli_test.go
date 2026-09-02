package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/blisspixel/fartapp/internal/lawcatalog"
)

const expectedLawListText = `LAW CONTEXTS

earth.continuum.si@v0alpha1 [design-candidate]
  Earth continuum mechanics in SI
  Biology-neutral candidate context for continuum discharge models under declared Earth conditions; no solver is implemented yet.
conformance.relation.atemporal@v0alpha1 [schema-conformance]
  Atemporal relation conformance context
  A relation-only context with no required ordering, geometry, units, source, or observer.
`

const expectedLawListJSON = `{"schema":"fart.law-context-list/v0alpha1","law_contexts":[{"id":"earth.continuum.si","version":"v0alpha1","maturity":"design-candidate","presentations":[{"locale":"en","message_key":"law.earth-continuum-si","name":"Earth continuum mechanics in SI","description":"Biology-neutral candidate context for continuum discharge models under declared Earth conditions; no solver is implemented yet."}]},{"id":"conformance.relation.atemporal","version":"v0alpha1","maturity":"schema-conformance","presentations":[{"locale":"en","message_key":"law.conformance-relation-atemporal","name":"Atemporal relation conformance context","description":"A relation-only context with no required ordering, geometry, units, source, or observer."}]}]}
`

const expectedEarthInspectionText = `LAW CONTEXT

ID: earth.continuum.si
Version: v0alpha1
Maturity: design-candidate
Name: Earth continuum mechanics in SI
Description: Biology-neutral candidate context for continuum discharge models under declared Earth conditions; no solver is implemented yet.
Presentation locale: en
Structural modules: ordering, state, dimension, topology, metric, locality, fields, units, equations, symmetries, invariants, conserved-currents
Context extension roles: emitter, interface, exterior, payload

CAPABILITY REPORT
Law context: earth.continuum.si@v0alpha1

catalog.inspect
  Read the built-in candidate law description and capability report.
  law definition:      not-applicable (application_capability)
  implementation:      available
  closure:             not-required
  applicability:       applicable
  evidence:            verified (software_fixture)
  evidence references: test:law-catalog-inspection, test:law-cli-fixtures
  trust:               built-in-candidate
  backend feasibility: not-required (application_capability)
  resource feasibility: within-default-budget

thermodynamics.finite-reservoir
  Ideal-mixture finite-reservoir mass and energy balance.
  law definition:      candidate
  implementation:      unavailable (not_implemented)
  closure:             undetermined (scenario_not_evaluated)
  applicability:       undetermined (scenario_not_evaluated)
  evidence:            design-only (implementation_evidence_unavailable)
  trust:               undetermined (operation_not_evaluated)
  backend feasibility: not-applicable (implementation_unavailable)
  resource feasibility: not-applicable (implementation_unavailable)

flow.subsonic
  Prescribed-area and compliant-interface subsonic discharge.
  law definition:      candidate
  implementation:      unavailable (not_implemented)
  closure:             undetermined (scenario_not_evaluated)
  applicability:       undetermined (scenario_not_evaluated)
  evidence:            design-only (implementation_evidence_unavailable)
  trust:               undetermined (operation_not_evaluated)
  backend feasibility: not-applicable (implementation_unavailable)
  resource feasibility: not-applicable (implementation_unavailable)

flow.choking-boundary
  Analytical choking boundary with explicit assumptions.
  law definition:      candidate
  implementation:      unavailable (not_implemented)
  closure:             undetermined (scenario_not_evaluated)
  applicability:       undetermined (scenario_not_evaluated)
  evidence:            design-only (implementation_evidence_unavailable)
  trust:               undetermined (operation_not_evaluated)
  backend feasibility: not-applicable (implementation_unavailable)
  resource feasibility: not-applicable (implementation_unavailable)

EVIDENCE REGISTRY
test:law-catalog-inspection [software/go-test]
  go test ./internal/lawcatalog -run ^TestBuiltInCatalog$
test:law-cli-fixtures [software/go-test]
  go test . -run ^TestLawCLITextAndJSONFixtures$
`

const expectedAtemporalInspectionText = `LAW CONTEXT

ID: conformance.relation.atemporal
Version: v0alpha1
Maturity: schema-conformance
Name: Atemporal relation conformance context
Description: A relation-only context with no required ordering, geometry, units, source, or observer.
Presentation locale: en
Structural modules: relations

CAPABILITY REPORT
Law context: conformance.relation.atemporal@v0alpha1

catalog.inspect
  Read the built-in candidate law description and capability report.
  law definition:      not-applicable (application_capability)
  implementation:      available
  closure:             not-required
  applicability:       applicable
  evidence:            verified (software_fixture)
  evidence references: test:law-catalog-inspection, test:law-cli-fixtures
  trust:               built-in-candidate
  backend feasibility: not-required (application_capability)
  resource feasibility: within-default-budget

EVIDENCE REGISTRY
test:law-catalog-inspection [software/go-test]
  go test ./internal/lawcatalog -run ^TestBuiltInCatalog$
test:law-cli-fixtures [software/go-test]
  go test . -run ^TestLawCLITextAndJSONFixtures$
`

// The digest is an exact fixture for compact deterministic JSON inspection.
const expectedEarthInspectionJSONSHA256 = "4445c07e794b98cc1a916a06235ac2b9b1012e973f6cb4a40d74b6d3b860beef"

func TestLawCLITextAndJSONFixtures(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "list text", args: []string{"fartapp", "law", "list"}, want: expectedLawListText},
		{name: "list explicit text", args: []string{"fartapp", "law", "list", "--format", "text"}, want: expectedLawListText},
		{name: "list json", args: []string{"fartapp", "law", "list", "--format", "json"}, want: expectedLawListJSON},
		{name: "inspect earth text", args: []string{"fartapp", "law", "inspect", "earth.continuum.si"}, want: expectedEarthInspectionText},
		{name: "inspect atemporal text", args: []string{"fartapp", "law", "inspect", "conformance.relation.atemporal@v0alpha1"}, want: expectedAtemporalInspectionText},
		{name: "help", args: []string{"fartapp", "law", "help"}, want: lawHelp},
		{name: "help flag", args: []string{"fartapp", "law", "--help"}, want: lawHelp},
		{name: "short help flag", args: []string{"fartapp", "law", "-h"}, want: lawHelp},
		{name: "list help", args: []string{"fartapp", "law", "list", "--help"}, want: lawListHelp},
		{name: "list help after format", args: []string{"fartapp", "law", "list", "--format", "json", "--help"}, want: lawListHelp},
		{name: "inspect help", args: []string{"fartapp", "law", "inspect", "--help"}, want: lawInspectHelp},
		{name: "inspect short help after id", args: []string{"fartapp", "law", "inspect", "earth.continuum.si", "-h"}, want: lawInspectHelp},
		{name: "unknown help", args: []string{"fartapp", "law", "nope", "--help"}, want: lawHelp},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			if code := run(tt.args, &stdout, &stderr); code != 0 {
				t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
			}
			if got := stdout.String(); got != tt.want {
				t.Fatalf("stdout = %q, want %q", got, tt.want)
			}
			if stderr.Len() != 0 {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
		})
	}
}

func TestLawInspectionJSONFixture(t *testing.T) {
	invoke := func() []byte {
		t.Helper()
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		args := []string{"fartapp", "law", "inspect", "--format", "json", "earth.continuum.si"}
		if code := run(args, &stdout, &stderr); code != 0 {
			t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
		}
		return bytes.Clone(stdout.Bytes())
	}

	first := invoke()
	second := invoke()
	if !bytes.Equal(first, second) {
		t.Fatal("JSON inspection output is not byte deterministic")
	}
	digest := sha256.Sum256(first)
	if got := hex.EncodeToString(digest[:]); got != expectedEarthInspectionJSONSHA256 {
		t.Fatalf("JSON inspection SHA-256 = %s, want %s", got, expectedEarthInspectionJSONSHA256)
	}

	var inspection lawcatalog.Inspection
	if err := json.Unmarshal(first, &inspection); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if err := lawcatalog.ValidateInspection(inspection); err != nil {
		t.Fatalf("ValidateInspection: %v", err)
	}
	for _, capability := range inspection.CapabilityReport.Capabilities {
		statuses := []string{
			capability.LawDefinition.Status,
			capability.Implementation.Status,
			capability.Closure.Status,
			capability.Applicability.Status,
			capability.Evidence.Status,
			capability.Trust.Status,
			capability.BackendFeasibility.Status,
			capability.ResourceFeasibility.Status,
		}
		for index, status := range statuses {
			if status == "" {
				t.Errorf("capability %q axis %d is empty", capability.ID, index)
			}
		}
	}
}

func TestLanguageFreeTextFormatter(t *testing.T) {
	inspection := lawcatalog.Inspection{
		Schema: lawcatalog.InspectionSchema,
		LawContext: lawcatalog.Context{
			ID:       "relation.atemporal",
			Version:  "v1",
			Maturity: "test-fixture",
		},
		CapabilityReport: lawcatalog.CapabilityReport{
			Schema:     lawcatalog.ReportSchema,
			LawContext: lawcatalog.LawContextRef{ID: "relation.atemporal", Version: "v1"},
		},
	}
	output := formatLawInspection(inspection)
	if !strings.Contains(output, "ID: relation.atemporal") ||
		strings.Contains(output, "Name:") || strings.Contains(output, "Description:") {
		t.Fatalf("language-free inspection output = %q", output)
	}
}

func TestLawCLIErrors(t *testing.T) {
	longID := strings.Repeat("x", 200)
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "missing command", args: []string{"fartapp", "law"}, want: "usage: fartapp law <list|inspect> [--format text|json]\n"},
		{name: "unknown command", args: []string{"fartapp", "law", "nope"}, want: "unknown law command \"nope\"\n"},
		{name: "list positional", args: []string{"fartapp", "law", "list", "extra"}, want: "usage: fartapp law list [--format text|json]\n"},
		{name: "missing format", args: []string{"fartapp", "law", "list", "--format"}, want: "invalid law list: --format requires text or json\n"},
		{name: "invalid format", args: []string{"fartapp", "law", "list", "--format", "xml"}, want: "invalid law list: unsupported format \"xml\"; expected text or json\n"},
		{name: "duplicate format", args: []string{"fartapp", "law", "list", "--format", "json", "--format", "text"}, want: "invalid law list: --format may be specified only once\n"},
		{name: "missing id", args: []string{"fartapp", "law", "inspect"}, want: "usage: fartapp law inspect <law-context-id[@version]> [--format text|json]\n"},
		{name: "extra id", args: []string{"fartapp", "law", "inspect", "a", "b"}, want: "usage: fartapp law inspect <law-context-id[@version]> [--format text|json]\n"},
		{name: "unknown id", args: []string{"fartapp", "law", "inspect", "missing.context"}, want: "unknown law context \"missing.context\"\n"},
		{name: "hostile id bounded", args: []string{"fartapp", "law", "inspect", longID}, want: "unknown law context \"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx...\"\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			if code := run(tt.args, &stdout, &stderr); code != 1 {
				t.Fatalf("exit code = %d, want 1", code)
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
			if got := stderr.String(); got != tt.want {
				t.Fatalf("stderr = %q, want %q", got, tt.want)
			}
		})
	}
}

type shortWriter struct{}

func (shortWriter) Write(value []byte) (int, error) {
	return len(value) / 2, nil
}

func TestLawCLIOutputFailures(t *testing.T) {
	for _, writer := range []io.Writer{failingWriter{}, shortWriter{}} {
		for _, args := range [][]string{
			{"fartapp", "law", "list"},
			{"fartapp", "law", "list", "--format", "json"},
		} {
			var stderr bytes.Buffer
			if code := run(args, writer, &stderr); code != 1 {
				t.Fatalf("run(%q) exit code = %d, want 1", args, code)
			}
			if !strings.HasPrefix(stderr.String(), "write output: ") {
				t.Fatalf("stderr = %q, want write failure", stderr.String())
			}
		}
	}

	var stderr bytes.Buffer
	code := writeLawValue(
		&bytes.Buffer{},
		&stderr,
		lawOutputJSON,
		make(chan int),
		func(chan int) string { return "unused" },
	)
	if code != 1 || !strings.HasPrefix(stderr.String(), "encode output: json: unsupported type:") {
		t.Fatalf("encode failure = (%d, %q)", code, stderr.String())
	}

	if code := run([]string{"fartapp", "law", "nope"}, &bytes.Buffer{}, failingWriter{}); code != 1 {
		t.Fatalf("diagnostic failure exit code = %d, want 1", code)
	}
}

func TestLawHelpers(t *testing.T) {
	positional, format, err := parseLawFormat([]string{"earth.continuum.si", "--format", "json"})
	if err != nil || len(positional) != 1 || positional[0] != "earth.continuum.si" || format != lawOutputJSON {
		t.Fatalf("parseLawFormat = (%q, %v, %v)", positional, format, err)
	}
	if got := joinModuleIDs(nil); got != "" {
		t.Fatalf("joinModuleIDs(nil) = %q, want empty", got)
	}
	presentation, found := selectPresentation([]lawcatalog.LocalizedPresentation{{
		Locale: "fr",
		Name:   "Nom",
	}})
	if !found || presentation.Locale != "fr" {
		t.Fatalf("selectPresentation fallback = (%#v, %v)", presentation, found)
	}
	if _, found := selectPresentation(nil); found {
		t.Fatal("selectPresentation(nil) found a value")
	}
}

func TestLawExecutable(t *testing.T) {
	binaryName := "fartapp"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	binaryPath := filepath.Join(t.TempDir(), binaryName)
	build := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build executable: %v\n%s", err, output)
	}
	command := exec.Command(binaryPath, "law", "list", "--format", "json")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("run executable: %v\n%s", err, output)
	}
	if got := string(output); got != expectedLawListJSON {
		t.Fatalf("output = %q, want %q", got, expectedLawListJSON)
	}
}
