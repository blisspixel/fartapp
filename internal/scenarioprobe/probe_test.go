package scenarioprobe

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
)

const atemporalProbe = `{
  "schema": "fart.scenario-probe/v0alpha1",
  "law_context_set": {
    "contexts": [{
      "id": "conformance.relation.atemporal",
      "version": "v0alpha1",
      "scope_id": "s0"
    }]
  },
  "scope": {"id": "s0"},
  "capability_requests": [{"id": "catalog.inspect"}]
}`

const earthProbe = `{
  "schema": "fart.scenario-probe/v0alpha1",
  "law_context_set": {
    "contexts": [{
      "id": "earth.continuum.si",
      "version": "v0alpha1",
      "scope_id": "s0"
    }]
  },
  "scope": {"id": "s0"},
  "capability_requests": [{"id": "flow.subsonic"}]
}`

func TestAtemporalProbeHasNoAmbientOrEarthRequirements(t *testing.T) {
	report := Validate([]byte(atemporalProbe))
	if !report.Valid() {
		t.Fatalf("Validate: %#v", report.Diagnostics)
	}
	if report.LawContext.ID != "conformance.relation.atemporal" ||
		report.LawContext.Version != "v0alpha1" {
		t.Fatalf("law context = %#v", report.LawContext)
	}
	if len(report.Environment.AmbientInputs) != 0 {
		t.Fatalf("ambient inputs = %q, want empty", report.Environment.AmbientInputs)
	}
	if got, want := report.Environment.ConsultedInputs, []string{"document_bytes", "built_in_law_catalog"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("consulted inputs = %q, want %q", got, want)
	}
	if report.Realization != (StageAssessment{Status: "not-performed", ReasonCode: "validation_only"}) {
		t.Fatalf("realization = %#v", report.Realization)
	}
	if report.Admission != (StageAssessment{Status: "not-evaluated", ReasonCode: "admission_policy_unratified"}) {
		t.Fatalf("admission = %#v", report.Admission)
	}
	wantStages := successfulValidationStages()
	if report.ValidationStages != wantStages {
		t.Fatalf("validation stages = %#v, want %#v", report.ValidationStages, wantStages)
	}
	encoded, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	for _, forbidden := range []string{
		"emitter", "exterior", "geometry", "locale", "observer", "pressure",
		"presentation", "sound", "time", "unit", "world",
	} {
		if bytes.Contains(encoded, []byte(forbidden)) {
			t.Errorf("report unexpectedly contains %q: %s", forbidden, encoded)
		}
	}
}

func TestEarthUnavailableCapabilityPreservesEveryAxis(t *testing.T) {
	report := Validate([]byte(earthProbe))
	if !report.Valid() || len(report.Capabilities) != 1 {
		t.Fatalf("Validate: %#v", report)
	}
	capability := report.Capabilities[0]
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
			t.Errorf("axis %d is empty", index)
		}
	}
	if capability.Implementation.Status != "unavailable" ||
		capability.Applicability.Status != "undetermined" {
		t.Fatalf("capability = %#v", capability)
	}
	if report.Realization.Status != "not-performed" {
		t.Fatalf("realization = %#v", report.Realization)
	}
}

func TestSemanticInputOrderDoesNotChangeReport(t *testing.T) {
	reordered := `{"capability_requests":[{"id":"catalog.inspect"}],"scope":{"id":"s0"},"law_context_set":{"contexts":[{"scope_id":"s0","version":"v0alpha1","id":"conformance.relation.atemporal"}]},"schema":"fart.scenario-probe/v0alpha1"}`
	first := Validate([]byte(atemporalProbe))
	second := Validate([]byte(reordered))
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("reports differ:\n%#v\n%#v", first, second)
	}

	requestsOne := strings.Replace(
		earthProbe,
		`[{"id": "flow.subsonic"}]`,
		`[{"id":"flow.subsonic"},{"id":"catalog.inspect"}]`,
		1,
	)
	requestsTwo := strings.Replace(
		earthProbe,
		`[{"id": "flow.subsonic"}]`,
		`[{"id":"catalog.inspect"},{"id":"flow.subsonic"}]`,
		1,
	)
	if first, second = Validate([]byte(requestsOne)), Validate([]byte(requestsTwo)); !reflect.DeepEqual(first, second) {
		t.Fatalf("capability request order changed report:\n%#v\n%#v", first, second)
	}
}

func TestUnknownCapabilityPathUsesSourceOrder(t *testing.T) {
	for _, test := range []struct {
		requests string
		path     string
	}{
		{requests: `[{"id":"unknown.capability"},{"id":"catalog.inspect"}]`, path: "/capability_requests/0/id"},
		{requests: `[{"id":"catalog.inspect"},{"id":"unknown.capability"}]`, path: "/capability_requests/1/id"},
	} {
		input := strings.Replace(earthProbe, `[{"id": "flow.subsonic"}]`, test.requests, 1)
		report := Validate([]byte(input))
		if report.Valid() || len(report.Diagnostics) != 1 || report.Diagnostics[0].Path != test.path {
			t.Errorf("requests %s report = %#v", test.requests, report)
		}
	}
}

func TestValidationFailures(t *testing.T) {
	valid := atemporalProbe
	tests := []struct {
		name   string
		input  string
		code   string
		path   string
		reason string
	}{
		{name: "empty", input: " \n\t", code: "FART-E-SYNTAX-0001", path: "/", reason: "empty_input"},
		{name: "malformed", input: "{", code: "FART-E-SYNTAX-0001", path: "/", reason: "malformed_json"},
		{name: "unpaired high surrogate", input: `{"x":"\ud800"}`, code: "FART-E-SYNTAX-0001", path: "/", reason: "malformed_json"},
		{name: "unpaired low surrogate", input: `{"x":"\udc00"}`, code: "FART-E-SYNTAX-0001", path: "/", reason: "malformed_json"},
		{name: "malformed array", input: "[", code: "FART-E-SYNTAX-0001", path: "/", reason: "malformed_json"},
		{name: "unexpected close", input: "}", code: "FART-E-SYNTAX-0001", path: "/", reason: "malformed_json"},
		{name: "trailing", input: valid + " {}", code: "FART-E-SYNTAX-0002", path: "/", reason: "trailing_json_value"},
		{name: "null root", input: "null", code: "FART-E-SCHEMA-0005", path: "/", reason: "wrong_type"},
		{name: "duplicate root", input: strings.Replace(valid, `"schema":`, `"schema":"first","schema":`, 1), code: "FART-E-SCHEMA-0002", path: "/schema", reason: "duplicate_member"},
		{name: "escaped equivalent duplicate root", input: strings.Replace(valid, `"schema":`, `"\u0073chema":"first","schema":`, 1), code: "FART-E-SCHEMA-0002", path: "/schema", reason: "duplicate_member"},
		{name: "duplicate nested", input: strings.Replace(valid, `"scope_id": "s0"`, `"scope_id":"s0","scope_id":"s0"`, 1), code: "FART-E-SCHEMA-0002", path: "/law_context_set/contexts/0/scope_id", reason: "duplicate_member"},
		{name: "unsupported schema", input: strings.Replace(valid, DocumentSchema, "other.schema/v1", 1), code: "FART-E-SCHEMA-0001", path: "/schema", reason: "unsupported_schema"},
		{name: "unknown root", input: strings.Replace(valid, `"schema":`, `"extra":true,"schema":`, 1), code: "FART-E-SCHEMA-0003", path: "/extra", reason: "unknown_member"},
		{name: "missing root", input: strings.Replace(valid, `  "scope": {"id": "s0"},`, "", 1), code: "FART-E-SCHEMA-0004", path: "/scope", reason: "missing_member"},
		{name: "wrong schema type", input: strings.Replace(valid, `"fart.scenario-probe/v0alpha1"`, "null", 1), code: "FART-E-SCHEMA-0005", path: "/schema", reason: "wrong_type"},
		{name: "law set wrong type", input: `{"schema":"fart.scenario-probe/v0alpha1","law_context_set":null,"scope":{"id":"s0"},"capability_requests":[{"id":"catalog.inspect"}]}`, code: "FART-E-SCHEMA-0005", path: "/law_context_set", reason: "wrong_type"},
		{name: "law set missing contexts", input: `{"schema":"fart.scenario-probe/v0alpha1","law_context_set":{},"scope":{"id":"s0"},"capability_requests":[{"id":"catalog.inspect"}]}`, code: "FART-E-SCHEMA-0004", path: "/law_context_set/contexts", reason: "missing_member"},
		{name: "contexts wrong type", input: `{"schema":"fart.scenario-probe/v0alpha1","law_context_set":{"contexts":null},"scope":{"id":"s0"},"capability_requests":[{"id":"catalog.inspect"}]}`, code: "FART-E-SCHEMA-0005", path: "/law_context_set/contexts", reason: "wrong_type"},
		{name: "context wrong type", input: `{"schema":"fart.scenario-probe/v0alpha1","law_context_set":{"contexts":[null]},"scope":{"id":"s0"},"capability_requests":[{"id":"catalog.inspect"}]}`, code: "FART-E-SCHEMA-0005", path: "/law_context_set/contexts/0", reason: "wrong_type"},
		{name: "context missing id", input: strings.Replace(valid, `"id": "conformance.relation.atemporal",`, `"other": "conformance.relation.atemporal",`, 1), code: "FART-E-SCHEMA-0004", path: "/law_context_set/contexts/0/id", reason: "missing_member"},
		{name: "context id wrong type", input: strings.Replace(valid, `"id": "conformance.relation.atemporal",`, `"id": null,`, 1), code: "FART-E-SCHEMA-0005", path: "/law_context_set/contexts/0/id", reason: "wrong_type"},
		{name: "context version wrong type", input: strings.Replace(valid, `"version": "v0alpha1",`, `"version": null,`, 1), code: "FART-E-SCHEMA-0005", path: "/law_context_set/contexts/0/version", reason: "wrong_type"},
		{name: "context version missing", input: strings.Replace(valid, "      \"version\": \"v0alpha1\",\n", "", 1), code: "FART-E-SCHEMA-0004", path: "/law_context_set/contexts/0/version", reason: "missing_member"},
		{name: "combined id and version forbidden", input: strings.Replace(valid, "conformance.relation.atemporal", "conformance.relation.atemporal@v0alpha1", 1), code: "FART-E-SCHEMA-0006", path: "/law_context_set/contexts/0/id", reason: "invalid_token"},
		{name: "context scope wrong type", input: strings.Replace(valid, `"scope_id": "s0"`, `"scope_id": null`, 1), code: "FART-E-SCHEMA-0005", path: "/law_context_set/contexts/0/scope_id", reason: "wrong_type"},
		{name: "invalid token", input: strings.Replace(valid, `"scope_id": "s0"`, `"scope_id": "Bad Scope"`, 1), code: "FART-E-SCHEMA-0006", path: "/law_context_set/contexts/0/scope_id", reason: "invalid_token"},
		{name: "scope wrong type", input: strings.Replace(valid, `"scope": {"id": "s0"}`, `"scope": null`, 1), code: "FART-E-SCHEMA-0005", path: "/scope", reason: "wrong_type"},
		{name: "scope missing id", input: strings.Replace(valid, `"scope": {"id": "s0"}`, `"scope": {}`, 1), code: "FART-E-SCHEMA-0004", path: "/scope/id", reason: "missing_member"},
		{name: "scope id wrong type", input: strings.Replace(valid, `"scope": {"id": "s0"}`, `"scope": {"id":null}`, 1), code: "FART-E-SCHEMA-0005", path: "/scope/id", reason: "wrong_type"},
		{name: "scope mismatch", input: strings.Replace(valid, `"scope": {"id": "s0"}`, `"scope": {"id": "other"}`, 1), code: "FART-E-SCHEMA-0008", path: "/law_context_set/contexts/0/scope_id", reason: "scope_reference_unresolved"},
		{name: "no context", input: strings.Replace(valid, strings.Split(strings.Split(valid, `"contexts": [`)[1], `]`)[0], "", 1), code: "FART-E-SCHEMA-0004", path: "/law_context_set/contexts", reason: "missing_law_context"},
		{name: "two contexts", input: strings.Replace(valid, "    }]", "    },{\"id\":\"conformance.relation.atemporal\",\"version\":\"v0alpha1\",\"scope_id\":\"s0\"}]", 1), code: "FART-E-ONTOLOGY-0001", path: "/law_context_set/contexts", reason: "multi_law_not_supported"},
		{name: "no request", input: strings.Replace(valid, `[{"id": "catalog.inspect"}]`, `[]`, 1), code: "FART-E-SCHEMA-0004", path: "/capability_requests", reason: "missing_capability_request"},
		{name: "requests wrong type", input: strings.Replace(valid, `[{"id": "catalog.inspect"}]`, `null`, 1), code: "FART-E-SCHEMA-0005", path: "/capability_requests", reason: "wrong_type"},
		{name: "request wrong type", input: strings.Replace(valid, `[{"id": "catalog.inspect"}]`, `[null]`, 1), code: "FART-E-SCHEMA-0005", path: "/capability_requests/0", reason: "wrong_type"},
		{name: "request missing id", input: strings.Replace(valid, `[{"id": "catalog.inspect"}]`, `[{}]`, 1), code: "FART-E-SCHEMA-0004", path: "/capability_requests/0/id", reason: "missing_member"},
		{name: "request id wrong type", input: strings.Replace(valid, `[{"id": "catalog.inspect"}]`, `[{"id":null}]`, 1), code: "FART-E-SCHEMA-0005", path: "/capability_requests/0/id", reason: "wrong_type"},
		{name: "duplicate request", input: strings.Replace(valid, `[{"id": "catalog.inspect"}]`, `[{"id":"catalog.inspect"},{"id":"catalog.inspect"}]`, 1), code: "FART-E-SCHEMA-0009", path: "/capability_requests/1/id", reason: "duplicate_capability_request"},
		{name: "unknown law", input: strings.Replace(valid, "conformance.relation.atemporal", "unknown.context", 1), code: "FART-E-LAW-0001", path: "/law_context_set/contexts/0", reason: "law_context_not_found"},
		{name: "unknown version", input: strings.Replace(valid, `"version": "v0alpha1"`, `"version": "v9"`, 1), code: "FART-E-LAW-0001", path: "/law_context_set/contexts/0", reason: "law_context_not_found"},
		{name: "unknown capability", input: strings.Replace(valid, "catalog.inspect", "relations.unknown", 1), code: "FART-E-CAP-0001", path: "/capability_requests/0/id", reason: "capability_not_defined"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := Validate([]byte(tt.input))
			wantStatus := "invalid"
			if tt.reason == "unsupported_schema" {
				wantStatus = "unsupported-schema"
			}
			if report.Valid() || report.DocumentStatus != wantStatus || len(report.Diagnostics) != 1 {
				t.Fatalf("report = %#v", report)
			}
			diagnostic := report.Diagnostics[0]
			if diagnostic.Code != tt.code || diagnostic.Path != tt.path || diagnostic.ReasonCode != tt.reason {
				t.Fatalf("diagnostic = %#v, want code %s path %s reason %s", diagnostic, tt.code, tt.path, tt.reason)
			}
			if report.Realization.Status != "not-performed" {
				t.Fatalf("realization = %#v", report.Realization)
			}
			wantInputs := []string{"document_bytes"}
			if diagnostic.Stage == "law-resolution" || diagnostic.Stage == "capability-resolution" {
				wantInputs = []string{"document_bytes", "built_in_law_catalog"}
			}
			if !reflect.DeepEqual(report.Environment.ConsultedInputs, wantInputs) {
				t.Fatalf("consulted inputs = %q, want %q", report.Environment.ConsultedInputs, wantInputs)
			}
			stages := report.ValidationStages
			switch diagnostic.Stage {
			case "syntax":
				if stages.Syntax.Status != "invalid" || stages.Schema.Status != "not-evaluated" {
					t.Fatalf("validation stages = %#v", stages)
				}
			case "schema":
				if stages.Syntax.Status != "valid" ||
					(stages.Schema.Status != "invalid" && stages.Schema.Status != "unsupported") ||
					stages.LawResolution.Status != "not-evaluated" {
					t.Fatalf("validation stages = %#v", stages)
				}
			case "law-resolution":
				if stages.Schema.Status != "valid" || stages.LawResolution.Status != "unresolved" {
					t.Fatalf("validation stages = %#v", stages)
				}
			case "capability-resolution":
				if stages.LawResolution.Status != "resolved" || stages.CapabilityResolution.Status != "unresolved" {
					t.Fatalf("validation stages = %#v", stages)
				}
			}
		})
	}
}

func TestUnicodePreflight(t *testing.T) {
	valid := []string{
		`{"x":"plain"}`,
		`{"x":"\u0061"}`,
		`{"x":"\ud83d\udca8"}`,
		`{"x":"\\ud800"}`,
	}
	for _, input := range valid {
		if diagnostic := preflightJSON([]byte(input)); diagnostic != nil {
			t.Errorf("preflightJSON(%q) = %#v", input, diagnostic)
		}
	}
	invalid := [][]byte{
		[]byte(`{"x":"\ud800\u0061"}`),
		[]byte(`{"x":"\udc00"}`),
		[]byte(`{"x":"\ud800"}`),
		[]byte(`{"x":"\ud800\u"}`),
		{'{', '"', 'x', '"', ':', '"', 0xff, '"', '}'},
	}
	for _, input := range invalid {
		if diagnostic := preflightJSON(input); diagnostic == nil || diagnostic.ReasonCode != "malformed_json" {
			t.Errorf("preflightJSON(%q) = %#v", input, diagnostic)
		}
	}
}

func TestJSONWhitespaceIsExact(t *testing.T) {
	for _, suffix := range []string{" ", "\t", "\n", "\r\n"} {
		if diagnostic := preflightJSON([]byte(atemporalProbe + suffix)); diagnostic != nil {
			t.Errorf("JSON whitespace %q rejected: %#v", suffix, diagnostic)
		}
	}
	for _, suffix := range []string{"\v", "\f", "\u00a0"} {
		diagnostic := preflightJSON([]byte(atemporalProbe + suffix))
		if diagnostic == nil || diagnostic.ReasonCode != "trailing_json_value" {
			t.Errorf("non-JSON whitespace %q = %#v", suffix, diagnostic)
		}
	}
}

func TestNestingAndCollectionLimits(t *testing.T) {
	oversized := Validate(bytes.Repeat([]byte{'x'}, MaxInputBytes+1))
	if got := oversized.Diagnostics[0].ReasonCode; got != "input_too_large" {
		t.Fatalf("oversized reason = %q", got)
	}

	deep := strings.Repeat("[", maximumJSONDepth+2) + "0" + strings.Repeat("]", maximumJSONDepth+2)
	report := Validate([]byte(deep))
	if got := report.Diagnostics[0].ReasonCode; got != "maximum_nesting_exceeded" {
		t.Fatalf("deep reason = %q", got)
	}

	requests := make([]string, maximumCapabilityRequests+1)
	for index := range requests {
		requests[index] = `{"id":"capability.` + strings.Repeat("a", index+1) + `"}`
	}
	large := strings.Replace(atemporalProbe, `[{"id": "catalog.inspect"}]`, "["+strings.Join(requests, ",")+"]", 1)
	report = Validate([]byte(large))
	if got := report.Diagnostics[0].ReasonCode; got != "collection_limit_exceeded" {
		t.Fatalf("collection reason = %q", got)
	}
}

func TestDiagnosticPathsAreBoundedAndEscaped(t *testing.T) {
	longKey := strings.Repeat("界", 50)
	input := strings.Replace(atemporalProbe, `"schema":`, `"`+longKey+`":true,"schema":`, 1)
	report := Validate([]byte(input))
	if diagnostic := report.Diagnostics[0]; diagnostic.Path != "/" ||
		diagnostic.ReasonCode != "member_name_too_long" {
		t.Fatalf("long member diagnostic = %#v", diagnostic)
	}

	input = strings.Replace(atemporalProbe, `"schema":`, `"a~/b":true,"schema":`, 1)
	report = Validate([]byte(input))
	if got, want := report.Diagnostics[0].Path, "/a~0~1b"; got != want {
		t.Fatalf("escaped path = %q, want %q", got, want)
	}
}

func TestDomainDefaultsAreRejectedStructurally(t *testing.T) {
	for _, member := range []string{
		"body", "bridge", "clock", "conservation", "coupling", "dimension", "gas",
		"geometry", "identity", "observer", "ordering", "participant", "seed",
		"source", "state", "time", "unit",
	} {
		input := strings.Replace(atemporalProbe, `"schema":`, `"`+member+`":null,"schema":`, 1)
		report := Validate([]byte(input))
		if report.Valid() || len(report.Diagnostics) != 1 ||
			report.Diagnostics[0].ReasonCode != "unknown_member" ||
			report.Diagnostics[0].Path != "/"+member {
			t.Errorf("member %q report = %#v", member, report)
		}
	}
}

type errorReader struct {
	data []byte
	done bool
}

func (reader *errorReader) Read(destination []byte) (int, error) {
	if reader.done {
		return 0, errors.New("test read failure")
	}
	reader.done = true
	return copy(destination, reader.data), nil
}

func TestReadBounded(t *testing.T) {
	exact := bytes.Repeat([]byte{'x'}, MaxInputBytes)
	got, err := ReadBounded(bytes.NewReader(exact))
	if err != nil || !bytes.Equal(got, exact) {
		t.Fatalf("exact read = (%d bytes, %v)", len(got), err)
	}
	if _, err = ReadBounded(bytes.NewReader(append(exact, 'x'))); !errors.Is(err, ErrInputTooLarge) {
		t.Fatalf("oversize error = %v", err)
	}
	if _, err = ReadBounded(&errorReader{data: []byte("partial")}); err == nil || errors.Is(err, ErrInputTooLarge) {
		t.Fatalf("reader error = %v", err)
	}
	if _, err = ReadBounded(nil); err == nil {
		t.Fatal("nil reader succeeded")
	}
}

func TestReportIsDefensivelyIndependent(t *testing.T) {
	first := Validate([]byte(atemporalProbe))
	brokenStages := first
	brokenStages.ValidationStages.Schema.Status = "invalid"
	if brokenStages.Valid() {
		t.Fatal("report with invalid schema stage is valid")
	}
	first.Capabilities[0].EvidenceReferences[0] = "mutated"
	first.EvidenceRegistry[0].ID = "mutated"
	first.Environment.ConsultedInputs[0] = "mutated"
	second := Validate([]byte(atemporalProbe))
	if second.Capabilities[0].EvidenceReferences[0] == "mutated" ||
		second.EvidenceRegistry[0].ID == "mutated" ||
		second.Environment.ConsultedInputs[0] == "mutated" {
		t.Fatal("report mutation escaped into later validation")
	}
}

func TestInputFailureStages(t *testing.T) {
	report := InputFailure(Diagnostic{
		Code:       "FART-E-IO-0001",
		Stage:      "input",
		Path:       "/",
		ReasonCode: "input_unavailable",
	}, "standard_input")
	if got, want := report.Environment.ConsultedInputs, []string{"standard_input"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("consulted inputs = %q, want %q", got, want)
	}
	stages := report.ValidationStages
	if stages.Syntax.ReasonCode != "input_unavailable" ||
		stages.CapabilityResolution.Status != "not-evaluated" {
		t.Fatalf("validation stages = %#v", stages)
	}
}

func FuzzValidate(f *testing.F) {
	for _, seed := range []string{atemporalProbe, earthProbe, "", "null", "{}", "{", "{}{}"} {
		f.Add([]byte(seed))
	}
	f.Add([]byte{'{', '"', 'x', '"', ':', '"', 0xff, '"', '}'})
	f.Fuzz(func(t *testing.T, input []byte) {
		first := Validate(input)
		second := Validate(input)
		if !reflect.DeepEqual(first, second) {
			t.Fatal("validation is nondeterministic")
		}
		encoded, err := json.Marshal(first)
		if err != nil {
			t.Fatalf("json.Marshal: %v", err)
		}
		if len(encoded) > 32*1024 {
			t.Fatalf("report is %d bytes", len(encoded))
		}
	})
}

var _ io.Reader = (*errorReader)(nil)
