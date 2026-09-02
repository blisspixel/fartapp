package lawcatalog

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBuiltInCatalog(t *testing.T) {
	listing := List()
	if listing.Schema != ListSchema {
		t.Fatalf("list schema = %q, want %q", listing.Schema, ListSchema)
	}
	if len(listing.LawContexts) != 2 {
		t.Fatalf("law context count = %d, want 2", len(listing.LawContexts))
	}

	for _, reference := range []string{
		"earth.continuum.si",
		"earth.continuum.si@v0alpha1",
		"conformance.relation.atemporal",
		"conformance.relation.atemporal@v0alpha1",
	} {
		inspection, found := Inspect(reference)
		if !found {
			t.Fatalf("Inspect(%q) did not find a context", reference)
		}
		if err := ValidateInspection(inspection); err != nil {
			t.Fatalf("ValidateInspection(%q): %v", reference, err)
		}
	}
	if _, found := Inspect("missing.context"); found {
		t.Fatal("missing context was found")
	}
}

func TestCatalogReturnsDefensiveCopies(t *testing.T) {
	listing := List()
	listing.LawContexts[0].ID = "mutated"
	listing.LawContexts[0].Presentations[0].Name = "mutated"
	freshList := List()
	if freshList.LawContexts[0].ID != "earth.continuum.si" ||
		freshList.LawContexts[0].Presentations[0].Name != "Earth continuum mechanics in SI" {
		t.Fatalf("catalog list was mutated: %#v", freshList.LawContexts[0])
	}

	inspection, _ := Inspect("earth.continuum.si")
	inspection.LawContext.Presentations[0].Name = "mutated"
	inspection.LawContext.StructuralModules[0].ID = "mutated"
	inspection.LawContext.ExtensionRoles[0] = "mutated"
	inspection.CapabilityReport.Capabilities[0].ID = "mutated"
	inspection.CapabilityReport.Capabilities[0].Presentations[0].Description = "mutated"
	inspection.CapabilityReport.Capabilities[0].EvidenceReferences[0] = "mutated"
	inspection.CapabilityReport.EvidenceRegistry[0].GoTest = "mutated"

	fresh, _ := Inspect("earth.continuum.si")
	if fresh.LawContext.Presentations[0].Name != "Earth continuum mechanics in SI" ||
		fresh.LawContext.StructuralModules[0].ID != "ordering" ||
		fresh.LawContext.ExtensionRoles[0] != "emitter" ||
		fresh.CapabilityReport.Capabilities[0].ID != "catalog.inspect" ||
		fresh.CapabilityReport.Capabilities[0].Presentations[0].Description == "mutated" ||
		fresh.CapabilityReport.Capabilities[0].EvidenceReferences[0] != "test:law-catalog-inspection" ||
		fresh.CapabilityReport.EvidenceRegistry[0].GoTest != "TestBuiltInCatalog" {
		t.Fatalf("catalog inspection was mutated: %#v", fresh)
	}
}

func TestInspectRejectsAmbiguousUnversionedReference(t *testing.T) {
	original := builtInInspections
	defer func() { builtInInspections = original }()
	duplicate := cloneInspection(builtInInspections[0])
	duplicate.LawContext.Version = "v0alpha2"
	duplicate.CapabilityReport.LawContext.Version = "v0alpha2"
	builtInInspections = append(builtInInspections, duplicate)

	if _, found := Inspect("earth.continuum.si"); found {
		t.Fatal("ambiguous unversioned reference was resolved")
	}
	inspection, found := Inspect("earth.continuum.si@v0alpha2")
	if !found || inspection.LawContext.Version != "v0alpha2" {
		t.Fatalf("exact version resolution = (%#v, %v)", inspection.LawContext, found)
	}
}

func TestAtemporalRelationContextNeedsNoEarthOrLanguageFields(t *testing.T) {
	inspection := neutralInspection()
	if err := ValidateInspection(inspection); err != nil {
		t.Fatalf("ValidateInspection: %v", err)
	}
	encoded, err := json.Marshal(inspection)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	for _, forbidden := range []string{
		"description", "emitter", "exterior", "geometry", "locale", "name",
		"observer", "pressure", "sound", "time", "unit", "world",
	} {
		if strings.Contains(string(encoded), forbidden) {
			t.Errorf("neutral inspection unexpectedly contains %q: %s", forbidden, encoded)
		}
	}
}

func TestValidateInspectionRejectsMalformedValues(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Inspection)
		want   string
	}{
		{name: "inspection schema", mutate: func(value *Inspection) { value.Schema = "wrong" }, want: "inspection schema"},
		{name: "context id", mutate: func(value *Inspection) { value.LawContext.ID = "Bad" }, want: "law context id"},
		{name: "context version", mutate: func(value *Inspection) { value.LawContext.Version = "" }, want: "law context version"},
		{name: "maturity", mutate: func(value *Inspection) { value.LawContext.Maturity = "Design Candidate" }, want: "law context maturity"},
		{name: "context presentation", mutate: func(value *Inspection) {
			value.LawContext.Presentations = []LocalizedPresentation{{Locale: "?", MessageKey: "key", Name: "name"}}
		}, want: "presentation"},
		{name: "report schema", mutate: func(value *Inspection) { value.CapabilityReport.Schema = "wrong" }, want: "capability report schema"},
		{name: "reference id", mutate: func(value *Inspection) { value.CapabilityReport.LawContext.ID = "Bad" }, want: "reference id"},
		{name: "reference version", mutate: func(value *Inspection) { value.CapabilityReport.LawContext.Version = "" }, want: "reference version"},
		{name: "reference mismatch", mutate: func(value *Inspection) { value.CapabilityReport.LawContext.ID = "other.context" }, want: "does not match"},
		{name: "invalid evidence record id", mutate: func(value *Inspection) { value.CapabilityReport.EvidenceRegistry[0].ID = "Bad" }, want: "evidence record id"},
		{name: "duplicate evidence record", mutate: func(value *Inspection) {
			value.CapabilityReport.EvidenceRegistry = append(value.CapabilityReport.EvidenceRegistry, value.CapabilityReport.EvidenceRegistry[0])
		}, want: "duplicated"},
		{name: "invalid evidence scope", mutate: func(value *Inspection) { value.CapabilityReport.EvidenceRegistry[0].Scope = "physical" }, want: "unsupported scope"},
		{name: "invalid evidence kind", mutate: func(value *Inspection) { value.CapabilityReport.EvidenceRegistry[0].Kind = "paper" }, want: "unsupported kind"},
		{name: "invalid evidence package", mutate: func(value *Inspection) { value.CapabilityReport.EvidenceRegistry[0].GoPackage = "../outside" }, want: "go package"},
		{name: "invalid evidence test", mutate: func(value *Inspection) { value.CapabilityReport.EvidenceRegistry[0].GoTest = "ExampleBad" }, want: "go test"},
		{name: "invalid evidence test character", mutate: func(value *Inspection) { value.CapabilityReport.EvidenceRegistry[0].GoTest = "Test-Bad" }, want: "ASCII identifier"},
		{name: "invalid module", mutate: func(value *Inspection) { value.LawContext.StructuralModules[0].ID = "Bad" }, want: "structural module id"},
		{name: "duplicate module", mutate: func(value *Inspection) {
			value.LawContext.StructuralModules = append(value.LawContext.StructuralModules, StructuralModule{ID: "relations"})
		}, want: "duplicated"},
		{name: "invalid role", mutate: func(value *Inspection) { value.LawContext.ExtensionRoles = []string{"Bad"} }, want: "extension role"},
		{name: "duplicate role", mutate: func(value *Inspection) { value.LawContext.ExtensionRoles = []string{"role", "role"} }, want: "duplicated"},
		{name: "duplicate capability", mutate: func(value *Inspection) {
			value.CapabilityReport.Capabilities = append(value.CapabilityReport.Capabilities, value.CapabilityReport.Capabilities[0])
		}, want: "capability id"},
		{name: "invalid capability id", mutate: func(value *Inspection) { value.CapabilityReport.Capabilities[0].ID = "Bad" }, want: "capability id"},
		{name: "capability presentation", mutate: func(value *Inspection) {
			value.CapabilityReport.Capabilities[0].Presentations = []LocalizedPresentation{{Locale: "en", MessageKey: "", Name: "name"}}
		}, want: "message key"},
		{name: "invalid evidence reference", mutate: func(value *Inspection) { value.CapabilityReport.Capabilities[0].EvidenceReferences = []string{"Bad"} }, want: "evidence reference"},
		{name: "duplicate evidence reference", mutate: func(value *Inspection) {
			value.CapabilityReport.Capabilities[0].EvidenceReferences = []string{"test:one", "test:one"}
		}, want: "duplicated"},
		{name: "unresolved evidence reference", mutate: func(value *Inspection) {
			value.CapabilityReport.Capabilities[0].EvidenceReferences = []string{"test:missing"}
		}, want: "does not resolve"},
		{name: "verified without evidence", mutate: func(value *Inspection) {
			value.CapabilityReport.Capabilities[0].EvidenceReferences = nil
		}, want: "requires a reference"},
		{name: "missing axis", mutate: func(value *Inspection) { value.CapabilityReport.Capabilities[0].Closure.Status = "" }, want: "closure status"},
		{name: "invalid reason token", mutate: func(value *Inspection) { value.CapabilityReport.Capabilities[0].Closure.ReasonCode = "Not Valid" }, want: "closure reason code"},
		{name: "cross-axis status", mutate: func(value *Inspection) {
			value.CapabilityReport.Capabilities[0].Closure = Assessment{Status: "available", ReasonCode: "not_implemented"}
		}, want: "does not permit"},
		{name: "long token", mutate: func(value *Inspection) { value.LawContext.ID = strings.Repeat("a", 129) }, want: "1 to 128"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := neutralInspection()
			tt.mutate(&value)
			err := ValidateInspection(value)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestPresentationValidation(t *testing.T) {
	tests := []struct {
		name  string
		value LocalizedPresentation
		want  string
	}{
		{name: "empty locale", value: LocalizedPresentation{MessageKey: "key", Name: "name"}, want: "locale"},
		{name: "leading hyphen", value: LocalizedPresentation{Locale: "-en", MessageKey: "key", Name: "name"}, want: "locale"},
		{name: "trailing hyphen", value: LocalizedPresentation{Locale: "en-", MessageKey: "key", Name: "name"}, want: "locale"},
		{name: "double hyphen", value: LocalizedPresentation{Locale: "en--US", MessageKey: "key", Name: "name"}, want: "locale"},
		{name: "long locale", value: LocalizedPresentation{Locale: strings.Repeat("a", 64), MessageKey: "key", Name: "name"}, want: "locale"},
		{name: "empty text", value: LocalizedPresentation{Locale: "en", MessageKey: "key"}, want: "name or description"},
		{name: "invalid utf8", value: LocalizedPresentation{Locale: "en", MessageKey: "key", Name: string([]byte{0xff})}, want: "valid UTF-8"},
		{name: "long text", value: LocalizedPresentation{Locale: "en", MessageKey: "key", Description: strings.Repeat("a", 4097)}, want: "4096"},
		{name: "control", value: LocalizedPresentation{Locale: "en", MessageKey: "key", Name: "bad\nname"}, want: "control"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePresentations([]LocalizedPresentation{tt.value})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want containing %q", err, tt.want)
			}
		})
	}

	values := []LocalizedPresentation{
		{Locale: "en", MessageKey: "key.en", Name: "Name"},
		{Locale: "zh-Hant", MessageKey: "key.zh", Description: "description"},
	}
	if err := validatePresentations(values); err != nil {
		t.Fatalf("valid presentations: %v", err)
	}
	values[1].Locale = "EN"
	if err := validatePresentations(values); err == nil || !strings.Contains(err.Error(), "duplicated") {
		t.Fatalf("duplicate locale error = %v", err)
	}
}

func TestAssessmentVocabulary(t *testing.T) {
	valid := []struct {
		axis       string
		assessment Assessment
	}{
		{"law definition", Assessment{Status: "not-declared", ReasonCode: "law_does_not_define"}},
		{"implementation", Assessment{Status: "unknown", ReasonCode: "not_evaluated"}},
		{"closure", Assessment{Status: "unavailable", ReasonCode: "closure_unavailable"}},
		{"applicability", Assessment{Status: "not-applicable", ReasonCode: "outside_validity"}},
		{"evidence", Assessment{Status: "validated", ReasonCode: "empirical_reference"}},
		{"trust", Assessment{Status: "refused", ReasonCode: "forbidden_by_policy"}},
		{"backend feasibility", Assessment{Status: "unavailable", ReasonCode: "backend_unavailable"}},
		{"resource feasibility", Assessment{Status: "insufficient", ReasonCode: "resource_budget_exceeded"}},
	}
	for _, value := range valid {
		if err := validateAssessment("test.capability", value.axis, value.assessment); err != nil {
			t.Errorf("valid %s assessment: %v", value.axis, err)
		}
	}
	if validAssessmentPair("unknown axis", "available", "") {
		t.Fatal("unknown axis was accepted")
	}
}

func neutralInspection() Inspection {
	capability := Capability{
		ID:                  "relations.compare",
		LawDefinition:       Assessment{Status: "declared"},
		Implementation:      Assessment{Status: "available"},
		Closure:             Assessment{Status: "not-required"},
		Applicability:       Assessment{Status: "applicable"},
		Evidence:            Assessment{Status: "verified", ReasonCode: "software_fixture"},
		EvidenceReferences:  []string{"test:relation-fixture"},
		Trust:               Assessment{Status: "built-in-candidate"},
		BackendFeasibility:  Assessment{Status: "available"},
		ResourceFeasibility: Assessment{Status: "within-default-budget"},
	}
	return Inspection{
		Schema: InspectionSchema,
		LawContext: Context{
			ID:                "relation.atemporal",
			Version:           "v1",
			Maturity:          "test-fixture",
			StructuralModules: []StructuralModule{{ID: "relations"}},
		},
		CapabilityReport: CapabilityReport{
			Schema:     ReportSchema,
			LawContext: LawContextRef{ID: "relation.atemporal", Version: "v1"},
			EvidenceRegistry: []EvidenceRecord{{
				ID:        "test:relation-fixture",
				Scope:     "software",
				Kind:      "go-test",
				GoPackage: ".",
				GoTest:    "TestAtemporalRelationContextNeedsNoEarthOrLanguageFields",
			}},
			Capabilities: []Capability{capability},
		},
	}
}
