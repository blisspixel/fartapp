// Package walkevidence retains exact authored inputs and versioned software
// reports. Integrity and internal binding are not scientific certification.
package walkevidence

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/blisspixel/fartapp/internal/strictjson"
	"github.com/blisspixel/fartapp/internal/walkcase"
)

const (
	Schema           = "fart.walk-evidence/v0alpha1"
	SummarySchema    = "fart.evidence-inspection/v0alpha1"
	MaxArtifactBytes = 24 << 20
)

var (
	ErrInvalidArchive    = errors.New("invalid evidence archive")
	ErrInvalidRequest    = errors.New("invalid capture request")
	ErrTooLarge          = errors.New("evidence byte limit exceeded")
	ErrDestinationExists = errors.New("evidence destination already exists")
)

type member struct {
	ByteLength int    `json:"byte_length"`
	SHA256     string `json:"sha256"`
	Base64     string `json:"base64"`
}

type envelope struct {
	Schema  string `json:"schema"`
	Request member `json:"request"`
	Report  member `json:"report"`
}

// Record is valid only after Capture, Decode, or Read. Accessors return copies.
type Record struct {
	encoded, request, report []byte
	retained                 walkcase.RetainedWitness
}

type Summary struct {
	Schema               string `json:"schema"`
	Status               string `json:"status"`
	ArtifactSchema       string `json:"artifact_schema"`
	ArtifactSHA256       string `json:"artifact_sha256"`
	RequestSHA256        string `json:"request_sha256"`
	ReportSHA256         string `json:"report_sha256"`
	RequestBytes         int    `json:"request_bytes"`
	ReportBytes          int    `json:"report_bytes"`
	InputDigest          string `json:"input_digest"`
	Witness              string `json:"witness"`
	MemberIntegrity      string `json:"member_integrity"`
	InputBinding         string `json:"normalized_input_binding"`
	AccountConsistency   string `json:"account_witness_consistency"`
	SolverReexecution    string `json:"solver_reexecution"`
	Authenticity         string `json:"authenticity"`
	ScientificValidation string `json:"scientific_validation"`
}

func Capture(request []byte) (Record, error) {
	if len(request) > walkcase.MaxInputBytes {
		return Record{}, ErrTooLarge
	}
	report := walkcase.Run(request, "witness")
	if !report.Predicted() {
		return Record{}, fmt.Errorf("%w: %s", ErrInvalidRequest, report.Diagnostics[0].ReasonCode)
	}
	reportBytes, err := json.Marshal(report)
	if err != nil {
		return Record{}, fmt.Errorf("%w: report encoding", ErrInvalidRequest)
	}
	if len(reportBytes) > walkcase.MaxRetainedReportBytes {
		return Record{}, ErrTooLarge
	}
	encoded, err := json.Marshal(envelope{Schema: Schema, Request: encodeMember(request), Report: encodeMember(reportBytes)})
	if err != nil {
		return Record{}, err
	}
	return Decode(append(encoded, '\n'))
}

func encodeMember(data []byte) member {
	return member{ByteLength: len(data), SHA256: digest(data), Base64: base64.StdEncoding.EncodeToString(data)}
}

func decodeMember(value member, limit int) ([]byte, error) {
	if value.ByteLength <= 0 || value.ByteLength > limit ||
		len(value.Base64) > base64.StdEncoding.EncodedLen(limit) || len(value.SHA256) != 64 {
		return nil, fmt.Errorf("%w: member bounds", ErrInvalidArchive)
	}
	data, err := base64.StdEncoding.Strict().DecodeString(value.Base64)
	if err != nil || len(data) != value.ByteLength || base64.StdEncoding.EncodeToString(data) != value.Base64 {
		return nil, fmt.Errorf("%w: noncanonical base64 or byte count", ErrInvalidArchive)
	}
	if digest(data) != value.SHA256 {
		return nil, fmt.Errorf("%w: member digest mismatch", ErrInvalidArchive)
	}
	return data, nil
}

func Decode(encoded []byte) (Record, error) {
	if len(encoded) > MaxArtifactBytes {
		return Record{}, ErrTooLarge
	}
	if issue := strictjson.Inspect(encoded, strictjson.Limits{MaximumDepth: 8, MaximumMemberNameBytes: 128}); issue != nil {
		return Record{}, fmt.Errorf("%w: envelope syntax %s", ErrInvalidArchive, issue.Kind)
	}
	// The envelope has no collections. Reject arrays and unknown fields using
	// its closed typed decoder before InspectShape materializes JSON values.
	// Otherwise a small array of empty objects can expand substantially in memory
	// despite the encoded-byte bound. The shape pass still enforces exact names
	// and non-null members after this allocation gate.
	var document envelope
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&document); err != nil || document.Schema != Schema {
		return Record{}, fmt.Errorf("%w: unsupported envelope", ErrInvalidArchive)
	}
	if issue := strictjson.InspectShape[envelope](encoded); issue != nil {
		return Record{}, fmt.Errorf("%w: envelope shape %s", ErrInvalidArchive, issue.Kind)
	}
	request, err := decodeMember(document.Request, walkcase.MaxInputBytes)
	if err != nil {
		return Record{}, err
	}
	report, err := decodeMember(document.Report, walkcase.MaxRetainedReportBytes)
	if err != nil {
		return Record{}, err
	}
	retained, err := walkcase.VerifyRetainedWitnessReport(report)
	if err != nil {
		return Record{}, fmt.Errorf("%w: retained report: %v", ErrInvalidArchive, err)
	}
	inputDigest, err := walkcase.NormalizeRequestFingerprint(request)
	if err != nil || inputDigest != retained.InputDigest() {
		return Record{}, fmt.Errorf("%w: authored request binding", ErrInvalidArchive)
	}
	return Record{encoded: bytes.Clone(encoded), request: request, report: report, retained: retained}, nil
}

func Read(reader io.Reader) (Record, error) {
	if reader == nil {
		return Record{}, fmt.Errorf("%w: nil reader", ErrInvalidArchive)
	}
	data, err := io.ReadAll(io.LimitReader(reader, MaxArtifactBytes+1))
	if err != nil {
		return Record{}, err
	}
	return Decode(data)
}

func (record Record) Bytes() []byte   { return bytes.Clone(record.encoded) }
func (record Record) Request() []byte { return bytes.Clone(record.request) }
func (record Record) Replay() []byte  { return bytes.Clone(record.report) }
func (record Record) Reconstruct() walkcase.Report {
	return record.retained.Reconstruct(record.request)
}

func (record Record) Summary() Summary {
	if len(record.encoded) == 0 {
		return Summary{Schema: SummarySchema, Status: "invalid"}
	}
	return Summary{
		Schema: SummarySchema, Status: "verified", ArtifactSchema: Schema,
		ArtifactSHA256: digest(record.encoded), RequestSHA256: digest(record.request), ReportSHA256: digest(record.report),
		RequestBytes: len(record.request), ReportBytes: len(record.report),
		InputDigest: record.retained.InputDigest(), Witness: record.retained.Witness(),
		MemberIntegrity: "verified", InputBinding: "verified", AccountConsistency: "verified",
		SolverReexecution: "not-performed", Authenticity: "not-established", ScientificValidation: "not-assessed",
	}
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
