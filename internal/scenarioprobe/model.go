// Package scenarioprobe validates the deliberately small v0.7 scenario probe.
// It does not create an occurrence, assign defaults, compute identity, inspect
// the host environment, or invoke a solver.
package scenarioprobe

import (
	"errors"
	"io"
	"slices"

	"github.com/blisspixel/fartapp/internal/lawcatalog"
)

const (
	DocumentSchema = "fart.scenario-probe/v0alpha1"
	ReportSchema   = "fart.scenario-validation/v0alpha1"
	MaxInputBytes  = 64 * 1024
)

var ErrInputTooLarge = errors.New("scenario probe input exceeds 65536 bytes")

type Document struct {
	Schema             string
	LawContextSet      LawContextSet
	Scope              Scope
	CapabilityRequests []CapabilityRequest
}

type LawContextSet struct {
	Contexts []ScopedLawContext
}

type ScopedLawContext struct {
	ID      string
	Version string
	ScopeID string
}

type Scope struct {
	ID string `json:"id"`
}

type CapabilityRequest struct {
	ID string
}

type StageAssessment struct {
	Status     string `json:"status"`
	ReasonCode string `json:"reason_code,omitempty"`
}

type ValidationStages struct {
	Syntax               StageAssessment `json:"syntax"`
	Schema               StageAssessment `json:"schema"`
	LawResolution        StageAssessment `json:"law_resolution"`
	CapabilityResolution StageAssessment `json:"capability_resolution"`
}

type Diagnostic struct {
	Code       string `json:"code"`
	Stage      string `json:"stage"`
	Path       string `json:"path"`
	ReasonCode string `json:"reason_code"`
	ByteOffset int64  `json:"byte_offset,omitempty"`
}

type CapabilityResult struct {
	ID                  string                `json:"id"`
	Resolution          string                `json:"resolution"`
	LawDefinition       lawcatalog.Assessment `json:"law_definition"`
	Implementation      lawcatalog.Assessment `json:"implementation"`
	Closure             lawcatalog.Assessment `json:"closure"`
	Applicability       lawcatalog.Assessment `json:"applicability"`
	Evidence            lawcatalog.Assessment `json:"evidence"`
	EvidenceReferences  []string              `json:"evidence_references,omitempty"`
	Trust               lawcatalog.Assessment `json:"trust"`
	BackendFeasibility  lawcatalog.Assessment `json:"backend_feasibility"`
	ResourceFeasibility lawcatalog.Assessment `json:"resource_feasibility"`
}

type ValidationEnvironment struct {
	ConsultedInputs []string `json:"consulted_inputs"`
	AmbientInputs   []string `json:"ambient_inputs"`
}

type Report struct {
	Schema           string                      `json:"schema"`
	DocumentStatus   string                      `json:"document_status"`
	ValidationStages ValidationStages            `json:"validation_stages"`
	DocumentSchema   string                      `json:"document_schema,omitempty"`
	LawContext       *lawcatalog.LawContextRef   `json:"law_context,omitempty"`
	Scope            *Scope                      `json:"scope,omitempty"`
	Capabilities     []CapabilityResult          `json:"capabilities,omitempty"`
	EvidenceRegistry []lawcatalog.EvidenceRecord `json:"evidence_registry,omitempty"`
	Environment      ValidationEnvironment       `json:"validation_environment"`
	Admission        StageAssessment             `json:"realization_admission"`
	Realization      StageAssessment             `json:"realization"`
	Diagnostics      []Diagnostic                `json:"diagnostics,omitempty"`
}

func baseReport(consultedInputs ...string) Report {
	return Report{
		Schema:         ReportSchema,
		DocumentStatus: "invalid",
		Environment: ValidationEnvironment{
			ConsultedInputs: slices.Clone(consultedInputs),
			AmbientInputs:   []string{},
		},
		Admission: StageAssessment{
			Status:     "not-evaluated",
			ReasonCode: "admission_policy_unratified",
		},
		Realization: StageAssessment{
			Status:     "not-performed",
			ReasonCode: "validation_only",
		},
	}
}

func Failure(diagnostic Diagnostic) Report {
	report := baseReport("document_bytes")
	if diagnostic.ReasonCode == "unsupported_schema" {
		report.DocumentStatus = "unsupported-schema"
	}
	report.ValidationStages = failedValidationStages(diagnostic)
	report.Diagnostics = []Diagnostic{diagnostic}
	return report
}

func InputFailure(diagnostic Diagnostic, consultedInput string) Report {
	report := baseReport(consultedInput)
	report.ValidationStages = failedValidationStages(diagnostic)
	report.Diagnostics = []Diagnostic{diagnostic}
	return report
}

func catalogFailure(diagnostic Diagnostic) Report {
	report := baseReport("document_bytes", "built_in_law_catalog")
	report.ValidationStages = failedValidationStages(diagnostic)
	report.Diagnostics = []Diagnostic{diagnostic}
	return report
}

func successfulValidationStages() ValidationStages {
	return ValidationStages{
		Syntax:               StageAssessment{Status: "valid"},
		Schema:               StageAssessment{Status: "valid"},
		LawResolution:        StageAssessment{Status: "resolved"},
		CapabilityResolution: StageAssessment{Status: "resolved"},
	}
}

func failedValidationStages(diagnostic Diagnostic) ValidationStages {
	notEvaluated := StageAssessment{Status: "not-evaluated", ReasonCode: "prior_stage_failed"}
	stages := ValidationStages{
		Syntax:               notEvaluated,
		Schema:               notEvaluated,
		LawResolution:        notEvaluated,
		CapabilityResolution: notEvaluated,
	}
	switch diagnostic.Stage {
	case "input":
		reason := diagnostic.ReasonCode
		stages.Syntax = StageAssessment{Status: "not-evaluated", ReasonCode: reason}
		stages.Schema = StageAssessment{Status: "not-evaluated", ReasonCode: reason}
		stages.LawResolution = StageAssessment{Status: "not-evaluated", ReasonCode: reason}
		stages.CapabilityResolution = StageAssessment{Status: "not-evaluated", ReasonCode: reason}
	case "syntax":
		stages.Syntax = StageAssessment{Status: "invalid", ReasonCode: diagnostic.ReasonCode}
	case "schema":
		stages.Syntax = StageAssessment{Status: "valid"}
		status := "invalid"
		if diagnostic.ReasonCode == "unsupported_schema" {
			status = "unsupported"
		}
		stages.Schema = StageAssessment{Status: status, ReasonCode: diagnostic.ReasonCode}
	case "law-resolution":
		stages.Syntax = StageAssessment{Status: "valid"}
		stages.Schema = StageAssessment{Status: "valid"}
		stages.LawResolution = StageAssessment{Status: "unresolved", ReasonCode: diagnostic.ReasonCode}
	case "capability-resolution":
		stages.Syntax = StageAssessment{Status: "valid"}
		stages.Schema = StageAssessment{Status: "valid"}
		stages.LawResolution = StageAssessment{Status: "resolved"}
		stages.CapabilityResolution = StageAssessment{Status: "unresolved", ReasonCode: diagnostic.ReasonCode}
	}
	return stages
}

func (report Report) Valid() bool {
	return report.DocumentStatus == "valid" && len(report.Diagnostics) == 0 &&
		report.ValidationStages.Syntax.Status == "valid" &&
		report.ValidationStages.Schema.Status == "valid" &&
		report.ValidationStages.LawResolution.Status == "resolved" &&
		report.ValidationStages.CapabilityResolution.Status == "resolved"
}

func ReadBounded(reader io.Reader) ([]byte, error) {
	if reader == nil {
		return nil, errors.New("scenario probe reader is nil")
	}
	data, err := io.ReadAll(io.LimitReader(reader, MaxInputBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > MaxInputBytes {
		return nil, ErrInputTooLarge
	}
	return data, nil
}

func capabilityResult(capability lawcatalog.Capability) CapabilityResult {
	return CapabilityResult{
		ID:                  capability.ID,
		Resolution:          "resolved",
		LawDefinition:       capability.LawDefinition,
		Implementation:      capability.Implementation,
		Closure:             capability.Closure,
		Applicability:       capability.Applicability,
		Evidence:            capability.Evidence,
		EvidenceReferences:  slices.Clone(capability.EvidenceReferences),
		Trust:               capability.Trust,
		BackendFeasibility:  capability.BackendFeasibility,
		ResourceFeasibility: capability.ResourceFeasibility,
	}
}
