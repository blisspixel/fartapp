// Package assurance exposes declared repository assurance metadata. Reading a
// registry neither runs its checks nor determines applicability to a case.
package assurance

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/blisspixel/fartapp/internal/strictjson"
)

const (
	Schema              = "fart.assurance-registry/v0alpha1"
	ListSchema          = "fart.assurance-list/v0alpha1"
	InspectionSchema    = "fart.assurance-inspection/v0alpha1"
	MaxRegistryBytes    = 1 << 20
	MaxInvariants       = 256
	MaxReferences       = 32
	RegistryPath        = "internal/assurance/registry.json"
	GeneratedPath       = "docs/INVARIANTS.md"
	DesignCandidate     = "design-candidate"
	ExecutableCandidate = "executable-candidate"
)

var (
	ErrInvalidRegistry  = errors.New("invalid assurance registry")
	ErrUnknownInvariant = errors.New("unknown assurance invariant")
	idPattern           = regexp.MustCompile(`^[A-Z]{2,8}-[0-9]{3}$`)
	tokenPattern        = regexp.MustCompile(`^[a-z][a-z0-9-]{0,95}$`)
	checkNamePattern    = regexp.MustCompile(`^(Test|Fuzz)[A-Z0-9_][A-Za-z0-9_]{0,119}$`)
	milestonePattern    = regexp.MustCompile(`^v[0-9]{1,2}\.[0-9]{1,2}$`)
)

//go:embed registry.json
var builtInJSON []byte

// Owner names identify repository responsibilities, not approving people or
// declaration authorities. Lifecycle is metadata, not a check execution status.
type Invariant struct {
	ID                string               `json:"id"`
	Statement         string               `json:"statement"`
	Owner             string               `json:"owner"`
	Applicability     string               `json:"applicability"`
	Tolerance         ToleranceProfile     `json:"tolerance"`
	Lifecycle         string               `json:"lifecycle"`
	Checks            []Check              `json:"checks"`
	Evidence          []EvidenceReference  `json:"evidence"`
	Counterexamples   []EvidenceReference  `json:"counterexamples"`
	Milestone         string               `json:"milestone"`
	Direction         string               `json:"direction"`
	RelatedBenchmarks []BenchmarkReference `json:"related_benchmarks"`
}

type ToleranceProfile struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

// Check names a Go declaration, never a shell command or a test result. Package
// and File must agree; repository validation additionally parses the test file.
type Check struct {
	ID      string `json:"id"`
	Package string `json:"package"`
	File    string `json:"file"`
	Name    string `json:"name"`
}

type EvidenceReference struct {
	Path        string `json:"path"`
	Description string `json:"description"`
}

// BenchmarkReference belongs to the separate VERIFICATION benchmark namespace.
// A relation never promotes the benchmark or asserts full conformance.
type BenchmarkReference struct {
	Namespace    string `json:"namespace"`
	ID           string `json:"id"`
	Relationship string `json:"relationship"`
	Scope        string `json:"scope"`
}

type document struct {
	Schema     string      `json:"schema"`
	Invariants []Invariant `json:"invariants"`
}

// Registry has no public mutable backing data. Its accessors copy all slices.
type Registry struct{ invariants []Invariant }

func BuiltIn() (Registry, error) { return Parse(builtInJSON) }

func Parse(data []byte) (Registry, error) {
	if len(data) > MaxRegistryBytes {
		return Registry{}, fmt.Errorf("%w: exceeds %d bytes", ErrInvalidRegistry, MaxRegistryBytes)
	}
	if issue := strictjson.Inspect(data, strictjson.Limits{MaximumDepth: 12, MaximumMemberNameBytes: 64}); issue != nil {
		return Registry{}, fmt.Errorf("%w: %s at %s", ErrInvalidRegistry, issue.Kind, issue.Path)
	}
	if issue := strictjson.InspectShape[document](data); issue != nil {
		return Registry{}, fmt.Errorf("%w: %s at %s", ErrInvalidRegistry, issue.Kind, issue.Path)
	}
	var decoded document
	if err := json.Unmarshal(data, &decoded); err != nil {
		return Registry{}, fmt.Errorf("%w: %v", ErrInvalidRegistry, err)
	}
	if decoded.Schema != Schema || len(decoded.Invariants) == 0 || len(decoded.Invariants) > MaxInvariants {
		return Registry{}, fmt.Errorf("%w: schema or invariant count", ErrInvalidRegistry)
	}
	seen := make(map[string]bool)
	checks := make(map[string]Check)
	tolerances := make(map[string]ToleranceProfile)
	for _, invariant := range decoded.Invariants {
		if seen[invariant.ID] {
			return Registry{}, fmt.Errorf("%w: duplicate invariant %s", ErrInvalidRegistry, invariant.ID)
		}
		seen[invariant.ID] = true
		if err := validateInvariant(invariant, checks, tolerances); err != nil {
			id := invariant.ID
			if len(id) > 32 {
				id = id[:32] + "..."
			}
			return Registry{}, fmt.Errorf("%w: %s: %v", ErrInvalidRegistry, strconv.QuoteToASCII(id), err)
		}
	}
	slices.SortFunc(decoded.Invariants, func(a, b Invariant) int { return strings.Compare(a.ID, b.ID) })
	return Registry{invariants: decoded.Invariants}, nil
}

func validateInvariant(value Invariant, checks map[string]Check, tolerances map[string]ToleranceProfile) error {
	if !idPattern.MatchString(value.ID) || !tokenPattern.MatchString(value.Owner) || !milestonePattern.MatchString(value.Milestone) {
		return errors.New("invalid ID, owner responsibility, or milestone")
	}
	if !validText(value.Statement) || !validText(value.Applicability) || !validText(value.Direction) ||
		!tokenPattern.MatchString(value.Tolerance.ID) || !validText(value.Tolerance.Description) {
		return errors.New("missing, overlong, or unsafe descriptive field")
	}
	if previous, exists := tolerances[value.Tolerance.ID]; exists && previous != value.Tolerance {
		return errors.New("one tolerance profile ID has conflicting definitions")
	}
	tolerances[value.Tolerance.ID] = value.Tolerance
	if value.Checks == nil || value.Evidence == nil || value.Counterexamples == nil || value.RelatedBenchmarks == nil ||
		len(value.Checks) > MaxReferences || len(value.RelatedBenchmarks) > MaxReferences {
		return errors.New("reference arrays must be explicit and bounded")
	}
	switch value.Lifecycle {
	case DesignCandidate:
		if len(value.Checks) != 0 || len(value.Counterexamples) != 0 || value.Tolerance.ID != "planned-v0alpha1" {
			return errors.New("design candidate must keep executable checks and counterexamples open with planned tolerance")
		}
	case ExecutableCandidate:
		if len(value.Checks) == 0 || len(value.Counterexamples) == 0 || value.Tolerance.ID == "planned-v0alpha1" {
			return errors.New("executable candidate requires checks, counterexamples, and a declared tolerance")
		}
	default:
		return errors.New("unsupported lifecycle; this metadata schema cannot ratify or promote a contract")
	}
	if err := validateReferences(value.Evidence, true); err != nil {
		return err
	}
	if err := validateReferences(value.Counterexamples, false); err != nil {
		return err
	}
	localChecks := make(map[string]bool)
	localDeclarations := make(map[string]bool)
	for _, check := range value.Checks {
		declaration := check.File + ":" + check.Name
		if !tokenPattern.MatchString(check.ID) || !ValidRepositoryPath(check.File) || !strings.HasSuffix(check.File, "_test.go") ||
			check.Package != "./"+path.Dir(check.File) || !checkNamePattern.MatchString(check.Name) ||
			localChecks[check.ID] || localDeclarations[declaration] {
			return errors.New("invalid or duplicate Go check reference")
		}
		if previous, exists := checks[check.ID]; exists && previous != check {
			return errors.New("one check ID has conflicting declarations")
		}
		checks[check.ID] = check
		localChecks[check.ID], localDeclarations[declaration] = true, true
	}
	seenBenchmarks := make(map[string]bool)
	for _, reference := range value.RelatedBenchmarks {
		if reference.Namespace != "verification-benchmark" || !idPattern.MatchString(reference.ID) || !validText(reference.Scope) ||
			(reference.Relationship != "partial-support" && reference.Relationship != "planned-conformance") || seenBenchmarks[reference.ID] {
			return errors.New("invalid or duplicate separate benchmark reference")
		}
		if value.Lifecycle == DesignCandidate && reference.Relationship != "planned-conformance" {
			return errors.New("design candidate cannot claim benchmark support")
		}
		seenBenchmarks[reference.ID] = true
	}
	return nil
}

func validateReferences(references []EvidenceReference, required bool) error {
	if len(references) > MaxReferences || (required && len(references) == 0) {
		return errors.New("missing or excessive evidence references")
	}
	seen := make(map[string]bool)
	for _, reference := range references {
		if !ValidRepositoryPath(reference.Path) || !validText(reference.Description) || seen[reference.Path] {
			return errors.New("invalid or duplicate evidence path or description")
		}
		seen[reference.Path] = true
	}
	return nil
}

func validText(value string) bool {
	if strings.TrimSpace(value) != value || value == "" || len(value) > 4096 {
		return false
	}
	for _, char := range value {
		if unicode.IsControl(char) || unicode.Is(unicode.Cf, char) {
			return false
		}
	}
	return true
}

// ValidRepositoryPath is a portable, canonical, bounded source identity. The
// filesystem checker must still establish symlink containment and file presence.
func ValidRepositoryPath(value string) bool {
	if len(value) > 240 || value == "." || !fs.ValidPath(value) {
		return false
	}
	for _, char := range value {
		if !(char >= 'a' && char <= 'z') && !(char >= 'A' && char <= 'Z') && !(char >= '0' && char <= '9') && !strings.ContainsRune("-._/", char) {
			return false
		}
	}
	return true
}

func (registry Registry) List() []Invariant {
	result := make([]Invariant, len(registry.invariants))
	for index, invariant := range registry.invariants {
		result[index] = cloneInvariant(invariant)
	}
	return result
}

func (registry Registry) Inspect(id string) (Invariant, error) {
	if len(registry.invariants) == 0 {
		return Invariant{}, ErrInvalidRegistry
	}
	for _, invariant := range registry.invariants {
		if invariant.ID == id {
			return cloneInvariant(invariant), nil
		}
	}
	return Invariant{}, ErrUnknownInvariant
}

func cloneInvariant(invariant Invariant) Invariant {
	invariant.Checks = slices.Clone(invariant.Checks)
	invariant.Evidence = slices.Clone(invariant.Evidence)
	invariant.Counterexamples = slices.Clone(invariant.Counterexamples)
	invariant.RelatedBenchmarks = slices.Clone(invariant.RelatedBenchmarks)
	return invariant
}
