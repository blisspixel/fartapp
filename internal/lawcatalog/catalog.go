// Package lawcatalog exposes the Go oracle's read-only candidate law catalog.
// It deliberately describes capabilities without requiring Earth-specific or
// human-language concepts in the machine report.
package lawcatalog

import (
	"fmt"
	"path"
	"slices"
	"strings"
	"unicode/utf8"
)

const (
	ListSchema       = "fart.law-context-list/v0alpha1"
	InspectionSchema = "fart.law-context-inspection/v0alpha1"
	ReportSchema     = "fart.law-capability-report/v0alpha1"
)

// ValidateMachineToken checks the bounded locale-invariant token grammar used
// by candidate catalog and scenario schemas.
func ValidateMachineToken(value string) error {
	return validateToken("machine token", value)
}

// Assessment is one independently reported capability axis. Status vocabularies
// are axis-specific; ReasonCode is a stable machine-oriented explanation.
type Assessment struct {
	Status     string `json:"status"`
	ReasonCode string `json:"reason_code,omitempty"`
}

// LocalizedPresentation is an optional user-facing projection. Locale is a
// provisional BCP 47-oriented token until the localization schema is ratified.
// Machine clients never need this record to identify a context or capability.
type LocalizedPresentation struct {
	Locale      string `json:"locale"`
	MessageKey  string `json:"message_key"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// Capability keeps the eight report axes separate. A missing implementation can
// therefore never masquerade as an undefined law concept or a resource refusal.
type Capability struct {
	ID                  string                  `json:"id"`
	Presentations       []LocalizedPresentation `json:"presentations,omitempty"`
	LawDefinition       Assessment              `json:"law_definition"`
	Implementation      Assessment              `json:"implementation"`
	Closure             Assessment              `json:"closure"`
	Applicability       Assessment              `json:"applicability"`
	Evidence            Assessment              `json:"evidence"`
	EvidenceReferences  []string                `json:"evidence_references,omitempty"`
	Trust               Assessment              `json:"trust"`
	BackendFeasibility  Assessment              `json:"backend_feasibility"`
	ResourceFeasibility Assessment              `json:"resource_feasibility"`
}

// LawContextRef identifies one exact candidate context revision.
type LawContextRef struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

// EvidenceRecord resolves a stable evidence ID to an executable repository
// fixture. It is software evidence only and never implies physical validation.
type EvidenceRecord struct {
	ID        string `json:"id"`
	Scope     string `json:"scope"`
	Kind      string `json:"kind"`
	GoPackage string `json:"go_package"`
	GoTest    string `json:"go_test"`
}

// CapabilityReport is intentionally a singular per-law catalog report. The
// future occurrence report will use the ratified LawContextSet, scope
// assignments, and inter-law coupling contract instead of extending this type
// with an incomplete list of IDs.
type CapabilityReport struct {
	Schema           string           `json:"schema"`
	LawContext       LawContextRef    `json:"law_context"`
	Capabilities     []Capability     `json:"capabilities"`
	EvidenceRegistry []EvidenceRecord `json:"evidence_registry,omitempty"`
}

// StructuralModule declares that a law context defines a concept. Absence from
// this list does not become a fabricated zero or an Earth default.
type StructuralModule struct {
	ID string `json:"id"`
}

// Context describes one candidate law context. ExtensionRoles are local to the
// context and never required by the catalog or capability-report schemas.
type Context struct {
	ID                string                  `json:"id"`
	Version           string                  `json:"version"`
	Maturity          string                  `json:"maturity"`
	Presentations     []LocalizedPresentation `json:"presentations,omitempty"`
	StructuralModules []StructuralModule      `json:"structural_modules,omitempty"`
	ExtensionRoles    []string                `json:"extension_roles,omitempty"`
}

// Summary is the compact catalog representation of a law context.
type Summary struct {
	ID            string                  `json:"id"`
	Version       string                  `json:"version"`
	Maturity      string                  `json:"maturity"`
	Presentations []LocalizedPresentation `json:"presentations,omitempty"`
}

// ListDocument is the machine-readable catalog listing.
type ListDocument struct {
	Schema      string    `json:"schema"`
	LawContexts []Summary `json:"law_contexts"`
}

// Inspection combines a context description with its report so localized text
// and locale-invariant machine renderers consume the same typed value.
type Inspection struct {
	Schema           string           `json:"schema"`
	LawContext       Context          `json:"law_context"`
	CapabilityReport CapabilityReport `json:"capability_report"`
}

var builtInInspections = []Inspection{
	{
		Schema: InspectionSchema,
		LawContext: Context{
			ID:       "earth.continuum.si",
			Version:  "v0alpha1",
			Maturity: "design-candidate",
			Presentations: presentations(
				"law.earth-continuum-si",
				"Earth continuum mechanics in SI",
				"Biology-neutral candidate context for continuum discharge models under "+
					"declared Earth conditions; no solver is implemented yet.",
			),
			StructuralModules: modules(
				"ordering", "state", "dimension", "topology", "metric",
				"locality", "fields", "units", "equations", "symmetries",
				"invariants", "conserved-currents",
			),
			ExtensionRoles: []string{"emitter", "interface", "exterior", "payload"},
		},
		CapabilityReport: CapabilityReport{
			Schema:           ReportSchema,
			LawContext:       LawContextRef{ID: "earth.continuum.si", Version: "v0alpha1"},
			EvidenceRegistry: softwareEvidenceRegistry(),
			Capabilities: []Capability{
				availableMetadataCapability(),
				plannedPhysicsCapability(
					"thermodynamics.finite-reservoir",
					"Ideal-mixture finite-reservoir mass and energy balance.",
				),
				plannedPhysicsCapability(
					"flow.subsonic",
					"Prescribed-area and compliant-interface subsonic discharge.",
				),
				plannedPhysicsCapability(
					"flow.choking-boundary",
					"Analytical choking boundary with explicit assumptions.",
				),
			},
		},
	},
	{
		Schema: InspectionSchema,
		LawContext: Context{
			ID:       "conformance.relation.atemporal",
			Version:  "v0alpha1",
			Maturity: "schema-conformance",
			Presentations: presentations(
				"law.conformance-relation-atemporal",
				"Atemporal relation conformance context",
				"A relation-only context with no required ordering, geometry, units, source, or observer.",
			),
			StructuralModules: modules("relations"),
		},
		CapabilityReport: CapabilityReport{
			Schema:           ReportSchema,
			EvidenceRegistry: softwareEvidenceRegistry(),
			LawContext: LawContextRef{
				ID:      "conformance.relation.atemporal",
				Version: "v0alpha1",
			},
			Capabilities: []Capability{availableMetadataCapability()},
		},
	},
	{
		Schema: InspectionSchema,
		LawContext: Context{
			ID:       "conformance.opaque.minimal",
			Version:  "v0alpha1",
			Maturity: "schema-conformance",
		},
		CapabilityReport: CapabilityReport{
			Schema:           ReportSchema,
			EvidenceRegistry: softwareEvidenceRegistry(),
			LawContext: LawContextRef{
				ID:      "conformance.opaque.minimal",
				Version: "v0alpha1",
			},
			Capabilities: []Capability{presentationFreeMetadataCapability()},
		},
	},
}

func presentations(messageKey, name, description string) []LocalizedPresentation {
	return []LocalizedPresentation{{
		Locale:      "en",
		MessageKey:  messageKey,
		Name:        name,
		Description: description,
	}}
}

func modules(ids ...string) []StructuralModule {
	result := make([]StructuralModule, len(ids))
	for index, id := range ids {
		result[index] = StructuralModule{ID: id}
	}
	return result
}

func softwareEvidenceRegistry() []EvidenceRecord {
	return []EvidenceRecord{
		{
			ID:        "test:law-catalog-inspection",
			Scope:     "software",
			Kind:      "go-test",
			GoPackage: "./internal/lawcatalog",
			GoTest:    "TestBuiltInCatalog",
		},
		{
			ID:        "test:law-cli-fixtures",
			Scope:     "software",
			Kind:      "go-test",
			GoPackage: ".",
			GoTest:    "TestLawCLITextAndJSONFixtures",
		},
	}
}

func availableMetadataCapability() Capability {
	return Capability{
		ID: "catalog.inspect",
		Presentations: presentations(
			"capability.catalog-inspect",
			"",
			"Read the built-in candidate law description and capability report.",
		),
		LawDefinition:  Assessment{Status: "not-applicable", ReasonCode: "application_capability"},
		Implementation: Assessment{Status: "available"},
		Closure:        Assessment{Status: "not-required"},
		Applicability:  Assessment{Status: "applicable"},
		Evidence:       Assessment{Status: "verified", ReasonCode: "software_fixture"},
		EvidenceReferences: []string{
			"test:law-catalog-inspection",
			"test:law-cli-fixtures",
		},
		Trust:               Assessment{Status: "built-in-candidate"},
		BackendFeasibility:  Assessment{Status: "not-required", ReasonCode: "application_capability"},
		ResourceFeasibility: Assessment{Status: "within-default-budget"},
	}
}

func presentationFreeMetadataCapability() Capability {
	capability := availableMetadataCapability()
	capability.Presentations = nil
	return capability
}

func plannedPhysicsCapability(id, description string) Capability {
	return Capability{
		ID: id,
		Presentations: presentations(
			"capability."+id,
			"",
			description,
		),
		LawDefinition:  Assessment{Status: "candidate"},
		Implementation: Assessment{Status: "unavailable", ReasonCode: "not_implemented"},
		Closure: Assessment{
			Status:     "undetermined",
			ReasonCode: "scenario_not_evaluated",
		},
		Applicability: Assessment{
			Status:     "undetermined",
			ReasonCode: "scenario_not_evaluated",
		},
		Evidence: Assessment{
			Status:     "design-only",
			ReasonCode: "implementation_evidence_unavailable",
		},
		Trust: Assessment{Status: "undetermined", ReasonCode: "operation_not_evaluated"},
		BackendFeasibility: Assessment{
			Status:     "not-applicable",
			ReasonCode: "implementation_unavailable",
		},
		ResourceFeasibility: Assessment{
			Status:     "not-applicable",
			ReasonCode: "implementation_unavailable",
		},
	}
}

// List returns a stable, defensive copy in catalog order.
func List() ListDocument {
	summaries := make([]Summary, len(builtInInspections))
	for index, inspection := range builtInInspections {
		context := inspection.LawContext
		summaries[index] = Summary{
			ID:            context.ID,
			Version:       context.Version,
			Maturity:      context.Maturity,
			Presentations: slices.Clone(context.Presentations),
		}
	}
	return ListDocument{Schema: ListSchema, LawContexts: summaries}
}

// Inspect accepts either an ID with one catalog version or an exact id@version.
// An unversioned ID becomes unresolved if multiple versions enter the catalog.
func Inspect(reference string) (Inspection, bool) {
	inspection, resolution := Resolve(reference)
	return inspection, resolution == ResolutionFound
}

// Resolution distinguishes a missing reference from an ambiguous unversioned
// reference without exposing catalog internals to the CLI.
type Resolution uint8

const (
	ResolutionNotFound Resolution = iota
	ResolutionFound
	ResolutionAmbiguous
)

// Resolve returns one defensively copied catalog inspection and its resolution.
func Resolve(reference string) (Inspection, Resolution) {
	var match *Inspection
	for index := range builtInInspections {
		inspection := &builtInInspections[index]
		exactReference := inspection.LawContext.ID + "@" + inspection.LawContext.Version
		if reference == exactReference {
			return cloneInspection(*inspection), ResolutionFound
		}
		if reference == inspection.LawContext.ID {
			if match != nil {
				return Inspection{}, ResolutionAmbiguous
			}
			match = inspection
		}
	}
	if match == nil {
		return Inspection{}, ResolutionNotFound
	}
	return cloneInspection(*match), ResolutionFound
}

func cloneInspection(source Inspection) Inspection {
	result := source
	result.LawContext.Presentations = slices.Clone(source.LawContext.Presentations)
	result.LawContext.StructuralModules = slices.Clone(source.LawContext.StructuralModules)
	result.LawContext.ExtensionRoles = slices.Clone(source.LawContext.ExtensionRoles)
	result.CapabilityReport.Capabilities = make([]Capability, len(source.CapabilityReport.Capabilities))
	result.CapabilityReport.EvidenceRegistry = slices.Clone(
		source.CapabilityReport.EvidenceRegistry,
	)
	for index, capability := range source.CapabilityReport.Capabilities {
		result.CapabilityReport.Capabilities[index] = capability
		result.CapabilityReport.Capabilities[index].Presentations = slices.Clone(capability.Presentations)
		result.CapabilityReport.Capabilities[index].EvidenceReferences = slices.Clone(
			capability.EvidenceReferences,
		)
	}
	return result
}

// ValidateInspection checks the candidate schema without assuming an Earth
// source, language, spatial geometry, dimensional values, ordering, or observer.
func ValidateInspection(inspection Inspection) error {
	if inspection.Schema != InspectionSchema {
		return fmt.Errorf("inspection schema %q is not supported", inspection.Schema)
	}
	if err := validateToken("law context id", inspection.LawContext.ID); err != nil {
		return err
	}
	if err := validateToken("law context version", inspection.LawContext.Version); err != nil {
		return err
	}
	if err := validateToken("law context maturity", inspection.LawContext.Maturity); err != nil {
		return err
	}
	if err := validatePresentations(inspection.LawContext.Presentations); err != nil {
		return fmt.Errorf("law context presentation: %w", err)
	}
	if inspection.CapabilityReport.Schema != ReportSchema {
		return fmt.Errorf(
			"capability report schema %q is not supported",
			inspection.CapabilityReport.Schema,
		)
	}
	if err := validateLawContextRef(inspection.CapabilityReport.LawContext); err != nil {
		return fmt.Errorf("capability report: %w", err)
	}
	if inspection.CapabilityReport.LawContext.ID != inspection.LawContext.ID ||
		inspection.CapabilityReport.LawContext.Version != inspection.LawContext.Version {
		return fmt.Errorf("capability report law context does not match inspected law context")
	}
	if err := validateEvidenceRegistry(inspection.CapabilityReport.EvidenceRegistry); err != nil {
		return err
	}
	moduleIDs := make([]string, len(inspection.LawContext.StructuralModules))
	for index, module := range inspection.LawContext.StructuralModules {
		moduleIDs[index] = module.ID
	}
	if err := validateUniqueTokens("structural module id", moduleIDs); err != nil {
		return err
	}
	if err := validateUniqueTokens(
		"extension role",
		inspection.LawContext.ExtensionRoles,
	); err != nil {
		return err
	}
	seenCapabilities := make(map[string]struct{}, len(inspection.CapabilityReport.Capabilities))
	for _, capability := range inspection.CapabilityReport.Capabilities {
		if _, exists := seenCapabilities[capability.ID]; exists {
			return fmt.Errorf("capability id %q is duplicated", capability.ID)
		}
		if err := validateCapability(
			capability,
			inspection.CapabilityReport.EvidenceRegistry,
		); err != nil {
			return err
		}
		seenCapabilities[capability.ID] = struct{}{}
	}
	return nil
}

func validateLawContextRef(reference LawContextRef) error {
	if err := validateToken("law context reference id", reference.ID); err != nil {
		return err
	}
	return validateToken("law context reference version", reference.Version)
}

func validateCapability(capability Capability, evidenceRegistry []EvidenceRecord) error {
	if err := validateToken("capability id", capability.ID); err != nil {
		return err
	}
	if err := validatePresentations(capability.Presentations); err != nil {
		return fmt.Errorf("capability %q presentation: %w", capability.ID, err)
	}
	if err := validateUniqueTokens("evidence reference", capability.EvidenceReferences); err != nil {
		return err
	}
	for _, reference := range capability.EvidenceReferences {
		if !evidenceRegistryContains(evidenceRegistry, reference) {
			return fmt.Errorf("evidence reference %q does not resolve", reference)
		}
	}
	if (capability.Evidence.Status == "verified" || capability.Evidence.Status == "validated") &&
		len(capability.EvidenceReferences) == 0 {
		return fmt.Errorf("capability %q evidence status requires a reference", capability.ID)
	}
	axes := []struct {
		name       string
		assessment Assessment
	}{
		{"law definition", capability.LawDefinition},
		{"implementation", capability.Implementation},
		{"closure", capability.Closure},
		{"applicability", capability.Applicability},
		{"evidence", capability.Evidence},
		{"trust", capability.Trust},
		{"backend feasibility", capability.BackendFeasibility},
		{"resource feasibility", capability.ResourceFeasibility},
	}
	for _, axis := range axes {
		if err := validateAssessment(capability.ID, axis.name, axis.assessment); err != nil {
			return err
		}
	}
	return nil
}

func validateEvidenceRegistry(values []EvidenceRecord) error {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if err := validateToken("evidence record id", value.ID); err != nil {
			return err
		}
		if _, exists := seen[value.ID]; exists {
			return fmt.Errorf("evidence record id %q is duplicated", value.ID)
		}
		seen[value.ID] = struct{}{}
		if value.Scope != "software" {
			return fmt.Errorf("evidence record %q has unsupported scope %q", value.ID, value.Scope)
		}
		if value.Kind != "go-test" {
			return fmt.Errorf("evidence record %q has unsupported kind %q", value.ID, value.Kind)
		}
		if err := validateGoPackage(value.GoPackage); err != nil {
			return fmt.Errorf("evidence record %q: %w", value.ID, err)
		}
		if err := validateGoTest(value.GoTest); err != nil {
			return fmt.Errorf("evidence record %q: %w", value.ID, err)
		}
	}
	return nil
}

func evidenceRegistryContains(values []EvidenceRecord, id string) bool {
	for _, value := range values {
		if value.ID == id {
			return true
		}
	}
	return false
}

func validateGoPackage(value string) error {
	if value == "." {
		return nil
	}
	relative := strings.TrimPrefix(value, "./")
	if len(value) > 256 || relative == value || relative == "" ||
		strings.Contains(value, "\\") || path.Clean(relative) != relative ||
		strings.HasPrefix(relative, "../") {
		return fmt.Errorf("go package must be a clean repository-relative package")
	}
	for _, character := range relative {
		if (character >= 'a' && character <= 'z') ||
			(character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') ||
			strings.ContainsRune("._-/", character) {
			continue
		}
		return fmt.Errorf("go package must contain only portable ASCII path characters")
	}
	return nil
}

func validateGoTest(value string) error {
	if len(value) < 5 || len(value) > 128 || !strings.HasPrefix(value, "Test") {
		return fmt.Errorf("go test must be a bounded Test-prefixed identifier")
	}
	for _, character := range value {
		if (character >= 'a' && character <= 'z') ||
			(character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || character == '_' {
			continue
		}
		return fmt.Errorf("go test %q is not an ASCII identifier", value)
	}
	return nil
}

func validateAssessment(capabilityID, axis string, assessment Assessment) error {
	if err := validateToken(
		fmt.Sprintf("capability %q %s status", capabilityID, axis),
		assessment.Status,
	); err != nil {
		return err
	}
	if assessment.ReasonCode != "" {
		if err := validateToken(
			fmt.Sprintf("capability %q %s reason code", capabilityID, axis),
			assessment.ReasonCode,
		); err != nil {
			return err
		}
	}
	if !validAssessmentPair(axis, assessment.Status, assessment.ReasonCode) {
		return fmt.Errorf(
			"capability %q %s status %q does not permit reason %q",
			capabilityID,
			axis,
			assessment.Status,
			assessment.ReasonCode,
		)
	}
	return nil
}

func validAssessmentPair(axis, status, reason string) bool {
	switch axis {
	case "law definition":
		return pairMatches(status, reason,
			"candidate", "",
			"declared", "",
			"not-applicable", "application_capability",
			"not-declared", "law_does_not_define",
			"unknown", "not_evaluated",
		)
	case "implementation":
		return pairMatches(status, reason,
			"available", "",
			"unavailable", "not_implemented",
			"unknown", "not_evaluated",
		)
	case "closure":
		return pairMatches(status, reason,
			"available", "",
			"unavailable", "closure_unavailable",
			"not-required", "",
			"undetermined", "scenario_not_evaluated",
		)
	case "applicability":
		return pairMatches(status, reason,
			"applicable", "",
			"not-applicable", "law_does_not_define",
			"not-applicable", "outside_validity",
			"undetermined", "scenario_not_evaluated",
		)
	case "evidence":
		return pairMatches(status, reason,
			"verified", "software_fixture",
			"validated", "empirical_reference",
			"design-only", "implementation_evidence_unavailable",
			"none", "no_evidence",
			"unknown", "not_evaluated",
		)
	case "trust":
		return pairMatches(status, reason,
			"built-in-candidate", "",
			"permitted", "",
			"refused", "forbidden_by_policy",
			"untrusted", "untrusted_pack",
			"undetermined", "operation_not_evaluated",
		)
	case "backend feasibility":
		return pairMatches(status, reason,
			"available", "",
			"unavailable", "backend_unavailable",
			"not-required", "application_capability",
			"not-applicable", "implementation_unavailable",
			"undetermined", "not_evaluated",
		)
	case "resource feasibility":
		return pairMatches(status, reason,
			"within-default-budget", "",
			"insufficient", "resource_budget_exceeded",
			"refused", "resource_policy_refusal",
			"not-applicable", "implementation_unavailable",
			"undetermined", "scenario_not_evaluated",
		)
	default:
		return false
	}
}

func pairMatches(status, reason string, pairs ...string) bool {
	for index := 0; index < len(pairs); index += 2 {
		if pairs[index] == status && pairs[index+1] == reason {
			return true
		}
	}
	return false
}

func validatePresentations(values []LocalizedPresentation) error {
	seenLocales := make(map[string]struct{}, len(values))
	for _, value := range values {
		if err := validateLocale(value.Locale); err != nil {
			return err
		}
		localeKey := strings.ToLower(value.Locale)
		if _, exists := seenLocales[localeKey]; exists {
			return fmt.Errorf("locale %q is duplicated", value.Locale)
		}
		seenLocales[localeKey] = struct{}{}
		if err := validateToken("message key", value.MessageKey); err != nil {
			return err
		}
		if value.Name == "" && value.Description == "" {
			return fmt.Errorf("localized presentation requires a name or description")
		}
		if err := validateDisplayText("localized name", value.Name); err != nil {
			return err
		}
		if err := validateDisplayText("localized description", value.Description); err != nil {
			return err
		}
	}
	return nil
}

func validateLocale(value string) error {
	if value == "" || len(value) > 63 || strings.HasPrefix(value, "-") ||
		strings.HasSuffix(value, "-") || strings.Contains(value, "--") {
		return fmt.Errorf("locale must be a bounded provisional ASCII language tag")
	}
	for _, character := range value {
		if (character >= 'a' && character <= 'z') ||
			(character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || character == '-' {
			continue
		}
		return fmt.Errorf("locale %q is not an ASCII language tag", value)
	}
	return nil
}

func validateDisplayText(name, value string) error {
	if len(value) > 4096 || !utf8.ValidString(value) {
		return fmt.Errorf("%s must be valid UTF-8 within 4096 bytes", name)
	}
	for _, character := range value {
		if character < 0x20 || (character >= 0x7f && character <= 0x9f) {
			return fmt.Errorf("%s contains a control character", name)
		}
	}
	return nil
}

func validateUniqueTokens(name string, values []string) error {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if err := validateToken(name, value); err != nil {
			return err
		}
		if _, exists := seen[value]; exists {
			return fmt.Errorf("%s %q is duplicated", name, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func validateToken(name, value string) error {
	if value == "" || len(value) > 128 {
		return fmt.Errorf("%s must contain 1 to 128 ASCII token bytes", name)
	}
	for _, character := range value {
		if (character >= 'a' && character <= 'z') ||
			(character >= '0' && character <= '9') ||
			strings.ContainsRune("._:-", character) {
			continue
		}
		return fmt.Errorf("%s %q is not a lowercase ASCII token", name, value)
	}
	return nil
}
