package assurance

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestBuiltInRetainsOriginalInvariantIdentitiesAndCandidateScopes(t *testing.T) {
	registry := mustRegistry(t)
	want := strings.Fields("CLI-001 CLI-002 CLI-003 CLI-004 ONT-001 CAP-001 CAP-002 CAP-003 CAP-004 CAP-005 CAP-006 CAP-007 CAP-008 SCN-001 SCN-002 SCN-003 SCN-004 JSON-001 PHY-001 OBS-001 ACC-001 PHY-002 PHY-003 PHY-004 PHY-005 CLI-005 ID-001 ID-002 DET-001 SURF-001 ART-001 LOC-001 CMP-001 CMP-002")
	slices.Sort(want)
	var got []string
	planned, executable := 0, 0
	for _, invariant := range registry.List() {
		got = append(got, invariant.ID)
		if invariant.Lifecycle == DesignCandidate {
			planned++
		} else if invariant.Lifecycle == ExecutableCandidate {
			executable++
		}
	}
	if !slices.Equal(got, want) || planned != 11 || executable != 23 {
		t.Fatalf("identity or lifecycle drift: %v; design=%d executable=%d", got, planned, executable)
	}
	for _, id := range strings.Fields("OBS-001 ACC-001 PHY-003 ID-001 ID-002 DET-001 SURF-001 ART-001 LOC-001 CMP-001 CMP-002") {
		invariant, _ := registry.Inspect(id)
		if invariant.Lifecycle != DesignCandidate || len(invariant.Checks) != 0 || len(invariant.Counterexamples) != 0 {
			t.Fatalf("planned scope was promoted: %+v", invariant)
		}
	}
	ontology, _ := registry.Inspect("ONT-001")
	if len(ontology.RelatedBenchmarks) != 1 || ontology.RelatedBenchmarks[0].Namespace != "verification-benchmark" || ontology.RelatedBenchmarks[0].Relationship != "partial-support" {
		t.Fatalf("invariant and benchmark namespaces collapsed: %+v", ontology)
	}
	if _, err := registry.Inspect("ont-001"); !errors.Is(err, ErrUnknownInvariant) {
		t.Fatal(err)
	}
	if _, err := registry.Inspect("CLI-001\n"); !errors.Is(err, ErrUnknownInvariant) {
		t.Fatal(err)
	}
}

func TestRegistryCopiesEveryMutableCollection(t *testing.T) {
	registry := mustRegistry(t)
	before, _ := registry.ListJSON()
	values := registry.List()
	for index := range values {
		values[index].ID = "BAD-000"
		mutateCollections(&values[index])
	}
	for _, id := range []string{"ONT-001", "CLI-001"} {
		value, _ := registry.Inspect(id)
		mutateCollections(&value)
	}
	after, _ := registry.ListJSON()
	if !bytes.Equal(before, after) {
		t.Fatal("returned values mutated the registry")
	}
	other := mustRegistry(t)
	other.invariants[0].ID = "BAD-000"
	after, _ = registry.ListJSON()
	if !bytes.Equal(before, after) {
		t.Fatal("separate built-in registries share state")
	}
}

func mutateCollections(value *Invariant) {
	if len(value.Checks) > 0 {
		value.Checks[0].Name = "TestChanged"
	}
	if len(value.Evidence) > 0 {
		value.Evidence[0].Path = "changed"
	}
	if len(value.Counterexamples) > 0 {
		value.Counterexamples[0].Description = "changed"
	}
	if len(value.RelatedBenchmarks) > 0 {
		value.RelatedBenchmarks[0].Scope = "changed"
	}
}

func TestParseRejectsClosedShapeSyntaxAndBounds(t *testing.T) {
	valid := minimalDocument(t)
	encoded := encodeDocument(t, valid)
	for _, data := range [][]byte{
		nil, []byte("null"), []byte("[]"), []byte("{}"), []byte("{broken"), append(slices.Clone(encoded), []byte("{}")...),
		[]byte(`{"schema":"a","schema":"b","invariants":[]}`), []byte(`{"\u0073chema":"a","schema":"b"}`),
		[]byte(`{"schema":"\ud800"}`), []byte("{\"schema\":\"\xff\"}"),
		[]byte(`{"` + strings.Repeat("x", 65) + `":0}`), []byte(strings.Repeat("[", 14) + "0" + strings.Repeat("]", 14)),
		bytes.Repeat([]byte(" "), MaxRegistryBytes+1),
	} {
		if _, err := Parse(data); !errors.Is(err, ErrInvalidRegistry) {
			t.Fatalf("invalid encoded input admitted (%d bytes): %v", len(data), err)
		}
	}
	for _, mutation := range []struct {
		name   string
		change func(map[string]any)
	}{
		{"unknown root", func(d map[string]any) { d["unknown"] = true }},
		{"case folded member", func(d map[string]any) { d["Schema"] = d["schema"]; delete(d, "schema") }},
		{"null array", func(d map[string]any) { d["invariants"] = nil }},
		{"unknown nested", func(d map[string]any) { firstJSONInvariant(d)["command"] = "arbitrary command" }},
		{"null nested", func(d map[string]any) { firstJSONInvariant(d)["tolerance"] = nil }},
		{"numeric owner", func(d map[string]any) { firstJSONInvariant(d)["owner"] = 1 }},
		{"missing checks", func(d map[string]any) { delete(firstJSONInvariant(d), "checks") }},
		{"null check", func(d map[string]any) { firstJSONInvariant(d)["checks"] = []any{nil} }},
		{"unknown check", func(d map[string]any) { firstJSONInvariant(d)["checks"].([]any)[0].(map[string]any)["run"] = true }},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			var generic map[string]any
			if err := json.Unmarshal(encoded, &generic); err != nil {
				t.Fatal(err)
			}
			mutation.change(generic)
			data, _ := json.Marshal(generic)
			if _, err := Parse(data); !errors.Is(err, ErrInvalidRegistry) {
				t.Fatalf("invalid shape accepted: %v", err)
			}
		})
	}
}

func TestParseRejectsFalseLifecycleAndBrokenSemanticReferences(t *testing.T) {
	for _, mutation := range []struct {
		name   string
		change func(*document)
	}{
		{"wrong schema", func(d *document) { d.Schema = "fart.assurance-registry/v1" }},
		{"no invariants", func(d *document) { d.Invariants = []Invariant{} }},
		{"too many invariants", func(d *document) {
			for index := 1; index <= MaxInvariants; index++ {
				extra := cloneInvariant(d.Invariants[0])
				extra.ID = fmt.Sprintf("NEW-%03d", index)
				d.Invariants = append(d.Invariants, extra)
			}
		}},
		{"duplicate invariant", func(d *document) { d.Invariants = append(d.Invariants, d.Invariants[0]) }},
		{"invalid ID", func(d *document) { d.Invariants[0].ID = "CLI-1" }},
		{"overlong unsafe ID", func(d *document) { d.Invariants[0].ID = "\x1b" + strings.Repeat("x", 4096) }},
		{"owner absent", func(d *document) { d.Invariants[0].Owner = "" }},
		{"owner whitespace", func(d *document) { d.Invariants[0].Owner = " owner" }},
		{"invalid milestone", func(d *document) { d.Invariants[0].Milestone = "whenever" }},
		{"statement absent", func(d *document) { d.Invariants[0].Statement = "" }},
		{"scope absent", func(d *document) { d.Invariants[0].Applicability = "" }},
		{"direction absent", func(d *document) { d.Invariants[0].Direction = "" }},
		{"overlong prose", func(d *document) { d.Invariants[0].Statement = strings.Repeat("x", 4097) }},
		{"terminal escape", func(d *document) { d.Invariants[0].Statement = "x\x1b[31m" }},
		{"bidi formatting", func(d *document) { d.Invariants[0].Statement = "x\u202e" }},
		{"tolerance absent", func(d *document) { d.Invariants[0].Tolerance.ID = "" }},
		{"conflicting tolerance", func(d *document) {
			extra := cloneInvariant(d.Invariants[0])
			extra.ID = "CLI-099"
			extra.Tolerance.Description = "conflicting"
			d.Invariants = append(d.Invariants, extra)
		}},
		{"ratified claim", func(d *document) { d.Invariants[0].Lifecycle = "ratified-internal" }},
		{"passing claim", func(d *document) { d.Invariants[0].Lifecycle = "pass" }},
		{"executable without checks", func(d *document) { d.Invariants[0].Checks = []Check{} }},
		{"executable without counterexamples", func(d *document) { d.Invariants[0].Counterexamples = []EvidenceReference{} }},
		{"executable with planned tolerance", func(d *document) { d.Invariants[0].Tolerance.ID = "planned-v0alpha1" }},
		{"design with checks", func(d *document) { d.Invariants[0].Lifecycle = DesignCandidate }},
		{"omitted evidence", func(d *document) { d.Invariants[0].Evidence = nil }},
		{"empty evidence", func(d *document) { d.Invariants[0].Evidence = []EvidenceReference{} }},
		{"omitted benchmark array", func(d *document) { d.Invariants[0].RelatedBenchmarks = nil }},
		{"excessive evidence", func(d *document) { d.Invariants[0].Evidence = make([]EvidenceReference, MaxReferences+1) }},
		{"duplicate evidence", func(d *document) {
			d.Invariants[0].Evidence = append(d.Invariants[0].Evidence, d.Invariants[0].Evidence[0])
		}},
		{"missing evidence description", func(d *document) { d.Invariants[0].Evidence[0].Description = "" }},
		{"escaping evidence", func(d *document) { d.Invariants[0].Evidence[0].Path = "../outside" }},
		{"escaping counterexample", func(d *document) { d.Invariants[0].Counterexamples[0].Path = "/outside" }},
		{"package mismatch", func(d *document) { d.Invariants[0].Checks[0].Package = "./internal/other" }},
		{"non-test file", func(d *document) { d.Invariants[0].Checks[0].File = "internal/cli/main.go" }},
		{"test pattern", func(d *document) { d.Invariants[0].Checks[0].Name = "Test.*" }},
		{"shell check ID", func(d *document) { d.Invariants[0].Checks[0].ID = "run; arbitrary" }},
		{"duplicate check", func(d *document) { d.Invariants[0].Checks = append(d.Invariants[0].Checks, d.Invariants[0].Checks[0]) }},
		{"duplicate declaration", func(d *document) {
			extra := d.Invariants[0].Checks[0]
			extra.ID = "other-check"
			d.Invariants[0].Checks = append(d.Invariants[0].Checks, extra)
		}},
		{"conflicting check ID", func(d *document) {
			extra := cloneInvariant(d.Invariants[0])
			extra.ID = "CLI-099"
			extra.Checks[0].Name = "TestOther"
			d.Invariants = append(d.Invariants, extra)
		}},
		{"too many checks", func(d *document) { d.Invariants[0].Checks = make([]Check, MaxReferences+1) }},
		{"too many benchmarks", func(d *document) { d.Invariants[0].RelatedBenchmarks = make([]BenchmarkReference, MaxReferences+1) }},
		{"ambiguous benchmark namespace", func(d *document) {
			d.Invariants[0].RelatedBenchmarks = []BenchmarkReference{{"invariant", "CLI-001", "partial-support", "scope"}}
		}},
		{"claimed full conformance", func(d *document) {
			d.Invariants[0].RelatedBenchmarks = []BenchmarkReference{{"verification-benchmark", "ONT-001", "passing", "scope"}}
		}},
		{"duplicate benchmark", func(d *document) {
			reference := BenchmarkReference{"verification-benchmark", "ONT-001", "partial-support", "scope"}
			d.Invariants[0].RelatedBenchmarks = []BenchmarkReference{reference, reference}
		}},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			doc := minimalDocument(t)
			mutation.change(&doc)
			if _, err := Parse(encodeDocument(t, doc)); !errors.Is(err, ErrInvalidRegistry) {
				t.Fatalf("invalid semantics accepted: %v", err)
			} else if len(err.Error()) > 512 || strings.ContainsRune(err.Error(), '\x1b') {
				t.Fatalf("unbounded or unsafe diagnostic: %q", err)
			}
		})
	}
	registry := mustRegistry(t)
	planned, _ := registry.Inspect("OBS-001")
	planned.RelatedBenchmarks[0].Relationship = "partial-support"
	if _, err := Parse(encodeDocument(t, document{Schema, []Invariant{planned}})); !errors.Is(err, ErrInvalidRegistry) {
		t.Fatal("design candidate claimed benchmark support")
	}
}

func TestPortableSourcePathGrammar(t *testing.T) {
	for _, value := range []string{"a", "docs/QUALITY.md", "internal/example_test.go", "testdata/v0alpha1/a-b_c.json"} {
		if !ValidRepositoryPath(value) {
			t.Fatalf("valid path refused: %s", value)
		}
	}
	for _, value := range []string{"", ".", "..", "../outside", "/absolute", "C:/absolute", "a\\b", "a//b", "a/../b", "a/./b", "a#anchor", "a?query", "a b", "a\x00b", "a\n", "a\u202e", strings.Repeat("a", 241)} {
		if ValidRepositoryPath(value) {
			t.Fatalf("invalid path admitted: %q", value)
		}
	}
}

func TestRenderingsExposeMetadataWithoutExecutionOrNamespacePromotion(t *testing.T) {
	registry := mustRegistry(t)
	data, err := registry.ListJSON()
	var report listReport
	if err != nil || json.Unmarshal(data, &report) != nil || report.Schema != ListSchema || report.RegistrySchema != Schema || report.EvidenceStatus != "not-executed" || report.ApplicabilityStatus != "not-evaluated" || len(report.Invariants) != 34 {
		t.Fatalf("invalid list report: %v", err)
	}
	if bytes.HasSuffix(data, []byte("\n")) {
		t.Fatal("framing belongs to the CLI writer")
	}
	for _, id := range []string{"ONT-001", "OBS-001", "CLI-001"} {
		data, err := registry.InspectJSON(id)
		var report inspectionReport
		if err != nil || json.Unmarshal(data, &report) != nil || report.Invariant.ID != id || report.Schema != InspectionSchema || report.EvidenceStatus != "not-executed" || report.ApplicabilityStatus != "not-evaluated" {
			t.Fatalf("invalid inspection: %v", err)
		}
		text, err := registry.InspectText(id)
		if err != nil || !strings.HasPrefix(text, "ASSURANCE "+id+"\n"+metadataNotice) || !strings.Contains(text, "Declared Go checks:") || !strings.Contains(text, "Open direction:") {
			t.Fatalf("unreadable inspection: %q %v", text, err)
		}
	}
	if !strings.HasPrefix(registry.ListText(), "ASSURANCE REGISTRY\n"+metadataNotice) || !strings.HasSuffix(registry.ListText(), "Inspect: fartapp assurance inspect <ID>\n") {
		t.Fatal("missing list hierarchy or next step")
	}
	if _, err := registry.InspectJSON("BAD-001"); !errors.Is(err, ErrUnknownInvariant) {
		t.Fatal(err)
	}
	if _, err := registry.InspectText("BAD-001"); !errors.Is(err, ErrUnknownInvariant) {
		t.Fatal(err)
	}
	var zero Registry
	if _, err := zero.ListJSON(); !errors.Is(err, ErrInvalidRegistry) {
		t.Fatal(err)
	}
	if _, err := zero.Inspect("CLI-001"); !errors.Is(err, ErrInvalidRegistry) {
		t.Fatal(err)
	}
	if got := zero.List(); got == nil || len(got) != 0 {
		t.Fatalf("zero registry list = %#v", got)
	}
	markdown := registry.Markdown()
	if markdown != registry.Markdown() || !strings.Contains(markdown, "## ONT-001") || !strings.Contains(markdown, "verification-benchmark:ONT-001") || !strings.Contains(markdown, "implementation remains planned") {
		t.Fatal("incomplete or nondeterministic reference rendering")
	}
	doc := minimalDocument(t)
	doc.Invariants[0].Statement = "Markup | [link](outside) <tag> & `code` *text* _text_ \\"
	marked, err := Parse(encodeDocument(t, doc))
	if err != nil {
		t.Fatal(err)
	}
	if got := marked.Markdown(); strings.Contains(got, "<tag>") || strings.Contains(got, "[link](outside)") || !strings.Contains(got, "\\|") || !strings.Contains(got, "&lt;tag&gt;") {
		t.Fatalf("unescaped authored Markdown: %s", got)
	}
}

func FuzzParse(f *testing.F) {
	f.Add([]byte(`{"schema":"fart.assurance-registry/v0alpha1","invariants":[]}`))
	f.Add(slices.Clone(builtInJSON))
	f.Fuzz(func(t *testing.T, data []byte) {
		registry, err := Parse(data)
		if err != nil {
			return
		}
		if len(data) > MaxRegistryBytes || len(registry.List()) == 0 || len(registry.List()) > MaxInvariants {
			t.Fatal("accepted registry exceeds bounds")
		}
		reencoded := encodeDocument(t, document{Schema, registry.List()})
		roundTrip, err := Parse(reencoded)
		if err != nil || !reflect.DeepEqual(registry.List(), roundTrip.List()) {
			t.Fatalf("accepted metadata does not round-trip: %v", err)
		}
		if _, err := registry.ListJSON(); err != nil {
			t.Fatal(err)
		}
	})
}

func mustRegistry(t *testing.T) Registry {
	t.Helper()
	registry, err := BuiltIn()
	if err != nil {
		t.Fatal(err)
	}
	return registry
}

func minimalDocument(t *testing.T) document {
	t.Helper()
	value, err := mustRegistry(t).Inspect("CLI-001")
	if err != nil {
		t.Fatal(err)
	}
	return document{Schema, []Invariant{value}}
}

func encodeDocument(t *testing.T, value document) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func firstJSONInvariant(value map[string]any) map[string]any {
	return value["invariants"].([]any)[0].(map[string]any)
}
