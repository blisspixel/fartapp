package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/blisspixel/fartapp/internal/walkcase"
	"github.com/blisspixel/fartapp/internal/walkevidence"
)

func TestEvidenceCLICaptureFileAndStdinAreEquivalent(t *testing.T) {
	request := append([]byte(" \n"), evidenceCLIRequest(t)...)
	directory := t.TempDir()
	source := filepath.Join(directory, "authored.json")
	if err := os.WriteFile(source, request, 0o600); err != nil {
		t.Fatal(err)
	}
	var previousArtifact, previousOutput []byte
	for index, input := range []string{source, "-"} {
		destination := filepath.Join(directory, []string{"file.fartevidence", "stdin.fartevidence"}[index])
		var stdin io.Reader = tripwireReader{}
		if input == "-" {
			stdin = bytes.NewReader(request)
		}
		var stdout, stderr bytes.Buffer
		code := Run([]string{"fartapp", "evidence", "capture", input, "--output=" + destination, "--format=json"}, stdin, &stdout, &stderr)
		if code != 0 || stderr.Len() != 0 {
			t.Fatalf("capture %s=(%d,%q)", input, code, stderr.String())
		}
		artifact, err := os.ReadFile(destination)
		if err != nil {
			t.Fatal(err)
		}
		record, err := walkevidence.Decode(artifact)
		if err != nil || !bytes.Equal(record.Request(), request) {
			t.Fatalf("authored bytes were not retained: %v", err)
		}
		var summary walkevidence.Summary
		if err := json.Unmarshal(stdout.Bytes(), &summary); err != nil || summary != record.Summary() {
			t.Fatalf("capture summary does not describe the committed bytes: %v", err)
		}
		if index > 0 && (!bytes.Equal(artifact, previousArtifact) || !bytes.Equal(stdout.Bytes(), previousOutput)) {
			t.Fatal("input source path changed retained bytes or summary")
		}
		previousArtifact, previousOutput = artifact, bytes.Clone(stdout.Bytes())
	}
}

func TestEvidenceCLIRetainedReplayAndReconstructionHaveDifferentMeanings(t *testing.T) {
	original, err := walkevidence.Capture(evidenceCLIRequest(t))
	if err != nil {
		t.Fatal(err)
	}
	altered := evidenceCLIDifferentRuntime(t, original)
	for _, test := range []struct {
		name    string
		encoded []byte
		matches bool
	}{
		{"current account", original.Bytes(), true},
		{"different retained account", altered, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			record, err := walkevidence.Decode(test.encoded)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), "retained.fartevidence")
			if err := os.WriteFile(path, test.encoded, 0o600); err != nil {
				t.Fatal(err)
			}
			for _, operation := range []string{"inspect", "verify", "replay", "reconstruct"} {
				formats := []string{"text", "json"}
				if operation == "replay" {
					formats = []string{"json"}
				}
				for _, format := range formats {
					var priorOutput []byte
					for _, source := range []string{path, "-"} {
						var stdin io.Reader = tripwireReader{}
						if source == "-" {
							stdin = bytes.NewReader(test.encoded)
						}
						var stdout, stderr bytes.Buffer
						code := Run([]string{"fartapp", "evidence", operation, source, "--format", format}, stdin, &stdout, &stderr)
						want := 0
						if operation == "reconstruct" && !test.matches {
							want = 1
						}
						if code != want || stderr.Len() != 0 {
							t.Fatalf("%s/%s=(%d,%q), want %d", operation, format, code, stderr.String(), want)
						}
						if source == "-" && !bytes.Equal(priorOutput, stdout.Bytes()) {
							t.Fatalf("%s output depends on source path", operation)
						}
						priorOutput = bytes.Clone(stdout.Bytes())
						switch operation {
						case "replay":
							if !bytes.Equal(stdout.Bytes(), append(record.Replay(), '\n')) {
								t.Fatal("replay recalculated, reformatted, or discarded retained report bytes")
							}
						case "reconstruct":
							if format == "json" {
								var report walkcase.Report
								if err := json.Unmarshal(stdout.Bytes(), &report); err != nil || report.WitnessMatch == nil || *report.WitnessMatch != test.matches ||
									report.ExpectedWitness != record.Summary().Witness || report.ReconstructedWitness != original.Summary().Witness || len(report.History) == 0 {
									t.Fatalf("reconstruction lost its comparison evidence: %v", err)
								}
							} else if !strings.Contains(stdout.String(), "WALK RECONSTRUCT") || !strings.Contains(stdout.String(), "Witness match: ") {
								t.Fatal("text reconstruction omitted comparison outcome")
							}
						default:
							if format == "json" {
								var summary walkevidence.Summary
								if err := json.Unmarshal(stdout.Bytes(), &summary); err != nil || summary != record.Summary() ||
									summary.SolverReexecution != "not-performed" || summary.Authenticity != "not-established" || summary.ScientificValidation != "not-assessed" {
									t.Fatalf("inspection overstated retained evidence: %v", err)
								}
							} else if !strings.Contains(stdout.String(), "Evidence: verified") || !strings.Contains(stdout.String(), "Solver reexecution: not-performed") ||
								!strings.Contains(stdout.String(), "Authenticity: not-established") || !strings.Contains(stdout.String(), "Scientific validation: not-assessed") {
								t.Fatal("text inspection omitted evidence limits")
							}
						}
					}
				}
			}
			// The replay default is JSON and preserves the same retained account.
			var stdout, stderr bytes.Buffer
			if code := Run([]string{"fartapp", "evidence", "replay", "-"}, bytes.NewReader(test.encoded), &stdout, &stderr); code != 0 ||
				!bytes.Equal(stdout.Bytes(), append(record.Replay(), '\n')) || stderr.Len() != 0 {
				t.Fatal("default replay was not the exact retained JSON report")
			}
		})
	}
}

func TestEvidenceCLIHelpNeverReadsOrCreatesOutput(t *testing.T) {
	destination := filepath.Join(t.TempDir(), "must-not-exist.fartevidence")
	args := [][]string{
		{"evidence", "--help"}, {"evidence", "-h"}, {"evidence", "help"}, {"help", "evidence"},
		{"evidence", "capture", "--help", "--output", destination},
		{"evidence", "capture", "missing.json", "--help", "--output", destination},
	}
	for _, operation := range []string{"capture", "inspect", "verify", "replay", "reconstruct"} {
		args = append(args, []string{"evidence", operation, "-", "--help"}, []string{"help", "evidence", operation})
	}
	for _, args := range args {
		var stdout, stderr bytes.Buffer
		if code := Run(append([]string{"fartapp"}, args...), tripwireReader{}, &stdout, &stderr); code != 0 || stderr.Len() != 0 || stdout.String() != evidenceHelp {
			t.Fatalf("help %v=(%d,%q,%q)", args, code, stdout.String(), stderr.String())
		}
	}
	if _, err := os.Lstat(destination); !os.IsNotExist(err) {
		t.Fatalf("help created an output: %v", err)
	}
}

func TestEvidenceCLIDoesNotReplayOrReconstructCorruptMemberEvidence(t *testing.T) {
	record, err := walkevidence.Capture(evidenceCLIRequest(t))
	if err != nil {
		t.Fatal(err)
	}
	var envelope map[string]any
	if err := json.Unmarshal(record.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	// The retained report is otherwise completely valid. None of the operations
	// may bypass a failed member hash in order to display it or calculate again.
	envelope["report"].(map[string]any)["sha256"] = strings.Repeat("0", 64)
	corrupted, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	for _, operation := range []string{"inspect", "verify", "replay", "reconstruct"} {
		var stdout, stderr bytes.Buffer
		if code := Run([]string{"fartapp", "evidence", operation, "-", "--format=json"}, bytes.NewReader(corrupted), &stdout, &stderr); code != 1 {
			t.Fatalf("%s bypassed corrupt evidence", operation)
		}
		assertEvidenceCLIFailure(t, operation, "invalid_archive", "json", &stdout, &stderr)
	}
}

func TestEvidenceCLIRejectsMalformedArgumentsBeforeInputOrPublication(t *testing.T) {
	directory := t.TempDir()
	destination := filepath.Join(directory, "must-not-exist.fartevidence")
	for _, args := range [][]string{
		{}, {"unknown"},
		{"capture", "--output", destination}, {"capture", "-"},
		{"capture", "-", "--output"}, {"capture", "-", "--output="},
		{"capture", "-", "--output", "--help"}, {"capture", "-", "--output", "-h"},
		{"capture", "-", "--output", destination, "--output", destination},
		{"capture", "-", "--output", destination, "--unknown"},
		{"capture", "-", "--output", destination, "--format", "yaml"},
		{"capture", "-", "--output", destination, "--format="},
		{"capture", "-", "--output", destination, "--format=json", "--format", "text"},
		{"capture", "--output", destination, "--", "-", "--format=json"},
		{"inspect", "-", "extra"}, {"verify", "-", "--output", destination},
		{"replay", "-", "--format=text"}, {"verify", "-", "--help", "--help"},
	} {
		var stdout, stderr bytes.Buffer
		if code := Run(append([]string{"fartapp", "evidence"}, args...), tripwireReader{}, &stdout, &stderr); code != 1 || stdout.Len() != 0 || stderr.Len() == 0 {
			t.Fatalf("arguments %v=(%d,%q,%q)", args, code, stdout.String(), stderr.String())
		}
	}
	// Previously, stripping --output first could reinterpret its following
	// value and turn this malformed command into an actual capture.
	var stdout, stderr bytes.Buffer
	args := []string{"fartapp", "evidence", "capture", "-", "--format", "--output", destination, "json"}
	if code := Run(args, bytes.NewReader(evidenceCLIRequest(t)), &stdout, &stderr); code != 1 || stdout.Len() != 0 || stderr.Len() == 0 {
		t.Fatalf("format value was reinterpreted as an option: %d %q %q", code, stdout.String(), stderr.String())
	}
	entries, err := os.ReadDir(directory)
	if err != nil || len(entries) != 0 {
		t.Fatalf("invalid arguments performed a filesystem mutation: %v %v", entries, err)
	}
}

func TestEvidenceCLIInputIntegrityAndStreamBoundaries(t *testing.T) {
	for _, test := range []struct {
		name, operation, reason string
		input                   func() io.Reader
	}{
		{"nil capture", "capture", "input_or_filesystem_failure", func() io.Reader { return nil }},
		{"capture read failure", "capture", "input_or_filesystem_failure", func() io.Reader { return walkCLIErrorReader{} }},
		{"invalid request", "capture", "invalid_capture_request", func() io.Reader { return strings.NewReader("{}") }},
		{"nil carrier", "verify", "invalid_archive", func() io.Reader { return nil }},
		{"carrier read failure", "verify", "input_or_filesystem_failure", func() io.Reader { return walkCLIErrorReader{} }},
		{"invalid carrier", "verify", "invalid_archive", func() io.Reader { return strings.NewReader("{}") }},
		{"invalid replay", "replay", "invalid_archive", func() io.Reader { return strings.NewReader(`{"report":"must not be replayed"}`) }},
		{"invalid reconstruction", "reconstruct", "invalid_archive", func() io.Reader { return strings.NewReader("{}") }},
	} {
		t.Run(test.name, func(t *testing.T) {
			formats := []string{"text", "json"}
			if test.operation == "replay" {
				formats = []string{"json"}
			}
			for _, format := range formats {
				destination := filepath.Join(t.TempDir(), "rejected.fartevidence")
				args := []string{"fartapp", "evidence", test.operation, "-", "--format", format}
				if test.operation == "capture" {
					args = append(args, "--output", destination)
				}
				var stdout, stderr bytes.Buffer
				if code := Run(args, test.input(), &stdout, &stderr); code != 1 {
					t.Fatalf("invalid input completed with %d", code)
				}
				assertEvidenceCLIFailure(t, test.operation, test.reason, format, &stdout, &stderr)
				if _, err := os.Lstat(destination); !os.IsNotExist(err) {
					t.Fatalf("invalid input was published: %v", err)
				}
			}
		})
	}
	for _, operation := range []string{"capture", "verify"} {
		limit := walkevidence.MaxArtifactBytes
		args := []string{"fartapp", "evidence", operation, "-", "--format=json"}
		if operation == "capture" {
			limit = walkcase.MaxInputBytes
			args = append(args, "--output", filepath.Join(t.TempDir(), "too-large.fartevidence"))
		}
		reader := &evidenceCountingReader{}
		var stdout, stderr bytes.Buffer
		if code := Run(args, reader, &stdout, &stderr); code != 1 || reader.read != limit+1 {
			t.Fatalf("%s overread its input budget: code=%d, read=%d, limit=%d", operation, code, reader.read, limit)
		}
		assertEvidenceCLIFailure(t, operation, "input_too_large", "json", &stdout, &stderr)
	}
}

func TestEvidenceCLINoClobberAndFilesystemFailures(t *testing.T) {
	directory := t.TempDir()
	destination := filepath.Join(directory, "existing.fartevidence")
	const existing = "previous independently owned bytes"
	if err := os.WriteFile(destination, []byte(existing), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, format := range []string{"text", "json"} {
		var stdout, stderr bytes.Buffer
		code := Run([]string{"fartapp", "evidence", "capture", "-", "--output", destination, "--format", format}, bytes.NewReader(evidenceCLIRequest(t)), &stdout, &stderr)
		if code != 1 {
			t.Fatal("capture replaced an existing output")
		}
		assertEvidenceCLIFailure(t, "capture", "destination_exists", format, &stdout, &stderr)
		stored, err := os.ReadFile(destination)
		if err != nil || string(stored) != existing {
			t.Fatalf("existing destination changed: %v", err)
		}
		stdout.Reset()
		stderr.Reset()
		if code := Run([]string{"fartapp", "evidence", "verify", filepath.Join(directory, "missing"), "--format", format}, tripwireReader{}, &stdout, &stderr); code != 1 {
			t.Fatal("missing source was accepted")
		}
		assertEvidenceCLIFailure(t, "verify", "input_or_filesystem_failure", format, &stdout, &stderr)
		stdout.Reset()
		stderr.Reset()
		if code := Run([]string{"fartapp", "evidence", "capture", "-", "--output", filepath.Join(directory, "missing", "result"), "--format", format}, bytes.NewReader(evidenceCLIRequest(t)), &stdout, &stderr); code != 1 {
			t.Fatal("missing parent was created implicitly")
		}
		assertEvidenceCLIFailure(t, "capture", "input_or_filesystem_failure", format, &stdout, &stderr)
	}
	var stdout, stderr bytes.Buffer
	missing := filepath.Join(directory, "missing-\x1b[2J.json")
	if code := Run([]string{"fartapp", "evidence", "verify", missing}, tripwireReader{}, &stdout, &stderr); code != 1 || bytes.ContainsRune(stderr.Bytes(), '\x1b') {
		t.Fatalf("text diagnostic emitted a terminal control: code=%d, stderr=%q", code, stderr.String())
	}
}

func TestEvidenceCLIOutputFailurePreservesCommittedArchive(t *testing.T) {
	for _, writer := range []io.Writer{failingWriter{}, shortWriter{}, brokenPipeWriter{err: syscall.EPIPE}} {
		destination := filepath.Join(t.TempDir(), "committed.fartevidence")
		var stderr bytes.Buffer
		if code := Run([]string{"fartapp", "evidence", "capture", "-", "--output", destination, "--format=json"}, bytes.NewReader(evidenceCLIRequest(t)), writer, &stderr); code != 1 {
			t.Fatal("failed output was reported as success")
		}
		stored, err := os.ReadFile(destination)
		if err != nil {
			t.Fatalf("committed archive was removed after stdout failure: %v", err)
		}
		record, err := walkevidence.Decode(stored)
		if err != nil || record.Summary().Status != "verified" {
			t.Fatalf("stdout failure left a partial archive: %v", err)
		}
		for _, operation := range []string{"inspect", "verify", "replay", "reconstruct"} {
			formats := []string{"text", "json"}
			if operation == "replay" {
				formats = []string{"json"}
			}
			for _, format := range formats {
				stderr.Reset()
				code := Run([]string{"fartapp", "evidence", operation, "-", "--format", format}, bytes.NewReader(stored), writer, &stderr)
				if code != 1 {
					t.Fatalf("%s hid output failure", operation)
				}
				_, broken := writer.(brokenPipeWriter)
				if broken && stderr.Len() != 0 || !broken && stderr.Len() == 0 {
					t.Fatalf("%s output diagnostic=%q, broken pipe=%t", operation, stderr.String(), broken)
				}
			}
		}
		stderr.Reset()
		if code := Run([]string{"fartapp", "evidence", "--help"}, tripwireReader{}, writer, &stderr); code != 1 {
			t.Fatal("help hid output failure")
		}
		stderr.Reset()
		if code := Run([]string{"fartapp", "evidence", "verify", "-", "--format=json"}, strings.NewReader("{}"), writer, &stderr); code != 1 {
			t.Fatal("failed integrity diagnostic output was reported as success")
		}
	}
}

func evidenceCLIRequest(t testing.TB) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.FromSlash("testdata/walk/ordinary-low-pressure.json"))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func assertEvidenceCLIFailure(t testing.TB, operation, reason, format string, stdout, stderr *bytes.Buffer) {
	t.Helper()
	if format == "json" {
		var report struct {
			Schema, Status, Operation string
			Reason                    string `json:"reason_code"`
		}
		if err := json.Unmarshal(stdout.Bytes(), &report); err != nil || report.Schema != "fart.evidence-operation/v0alpha1" || report.Status != "invalid" ||
			report.Operation != operation || report.Reason != reason || stderr.Len() != 0 {
			t.Fatalf("failure=(%s,%q,%v), want operation=%s reason=%s", stdout.String(), stderr.String(), err, operation, reason)
		}
	} else if stdout.Len() != 0 || !strings.Contains(stderr.String(), reason) {
		t.Fatalf("text failure=(%q,%q), want %s", stdout.String(), stderr.String(), reason)
	}
}

type evidenceCountingReader struct{ read int }

func (reader *evidenceCountingReader) Read(data []byte) (int, error) {
	for index := range data {
		data[index] = 'x'
	}
	reader.read += len(data)
	return len(data), nil
}

func evidenceCLIDifferentRuntime(t testing.TB, original walkevidence.Record) []byte {
	t.Helper()
	var report walkcase.Report
	if err := json.Unmarshal(original.Replay(), &report); err != nil {
		t.Fatal(err)
	}
	report.NumericalPolicy.GoVersion = "go1.0-retained-test"
	account := report
	account.Operation = "simulate"
	account.Explanation = nil
	account.Witness = ""
	account.WitnessSchema = ""
	account.WitnessAlgorithm = ""
	account.InputDigest = ""
	account.InputDigestSchema = ""
	accountBytes, err := json.Marshal(account)
	if err != nil {
		t.Fatal(err)
	}
	accountEnvelope := append([]byte(`{"schema":"fart.walk-witness/v0alpha1","account":`), accountBytes...)
	report.Witness = evidenceCLIHash(append(accountEnvelope, '}'))
	reportBytes, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	type member struct {
		ByteLength int    `json:"byte_length"`
		SHA256     string `json:"sha256"`
		Base64     string `json:"base64"`
	}
	encode := func(raw []byte) member {
		return member{len(raw), evidenceCLIHash(raw), base64.StdEncoding.EncodeToString(raw)}
	}
	encoded, err := json.Marshal(struct {
		Schema  string `json:"schema"`
		Request member `json:"request"`
		Report  member `json:"report"`
	}{walkevidence.Schema, encode(original.Request()), encode(reportBytes)})
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func evidenceCLIHash(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
