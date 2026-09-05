package walkevidence

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/blisspixel/fartapp/internal/strictjson"
	"github.com/blisspixel/fartapp/internal/walkcase"
)

func fixture(t testing.TB) []byte {
	t.Helper()
	data, err := os.ReadFile("../../testdata/walk/ordinary-low-pressure.json")
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func capture(t testing.TB) Record {
	t.Helper()
	record, err := Capture(fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	return record
}

func TestCapturePreservesAuthoredBytesAndReconstructs(t *testing.T) {
	request := append([]byte(" \n"), fixture(t)...)
	record, err := Capture(request)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(request, record.Request()) {
		t.Fatal("authored bytes changed")
	}
	report := record.Reconstruct()
	if !report.Predicted() || report.WitnessMatch == nil || !*report.WitnessMatch {
		t.Fatalf("reconstruction failed: %#v", report.Diagnostics)
	}
	for _, getter := range []func() []byte{record.Bytes, record.Request, record.Replay} {
		a := getter()
		a[0] ^= 1
		if bytes.Equal(a, getter()) {
			t.Fatal("mutable storage escaped")
		}
	}
	reloaded, err := Read(bytes.NewReader(record.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	summary := reloaded.Summary()
	if summary.Status != "verified" || summary.RequestSHA256 != digest(request) ||
		summary.ReportSHA256 != digest(record.Replay()) || summary.ArtifactSHA256 != digest(record.Bytes()) ||
		summary.SolverReexecution != "not-performed" || summary.Authenticity != "not-established" {
		t.Fatalf("wrong evidence scope: %#v", summary)
	}
	if !bytes.Equal(record.Replay(), reloaded.Replay()) {
		t.Fatal("replay changed retained report")
	}
}

func TestHostileEnvelopesAndMemberSubstitution(t *testing.T) {
	record := capture(t)
	var original envelope
	if err := json.Unmarshal(record.Bytes(), &original); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name string
		edit func(*envelope)
	}{
		{"schema", func(e *envelope) { e.Schema = "future" }},
		{"missing request", func(e *envelope) { e.Request = member{} }},
		{"negative length", func(e *envelope) { e.Report.ByteLength = -1 }},
		{"length overflow", func(e *envelope) { e.Report.ByteLength = walkcase.MaxRetainedReportBytes + 1 }},
		{"digest case", func(e *envelope) { e.Report.SHA256 = strings.ToUpper(e.Report.SHA256) }},
		{"digest length", func(e *envelope) { e.Report.SHA256 = "abc" }},
		{"wrong digest", func(e *envelope) { e.Report.SHA256 = strings.Repeat("0", 64) }},
		{"base64 spelling", func(e *envelope) { e.Report.Base64 = "\n" + e.Report.Base64 }},
		{"invalid base64", func(e *envelope) { e.Report.Base64 = "!!!" }},
		{"wrong length", func(e *envelope) { e.Request.ByteLength++ }},
		{"invalid authored JSON with valid hash", func(e *envelope) { e.Request = encodeMember([]byte("{}")) }},
		{"different authored request with valid hash", func(e *envelope) {
			request := bytes.Replace(record.Request(), []byte("0.000001"), []byte("0.000003"), 1)
			e.Request = encodeMember(request)
		}},
		{"invalid report with valid hash", func(e *envelope) { e.Report = encodeMember([]byte("{}")) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			e := original
			test.edit(&e)
			data, err := json.Marshal(e)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := Decode(data); !errors.Is(err, ErrInvalidArchive) {
				t.Fatalf("accepted attack: %v", err)
			}
		})
	}
	for _, raw := range [][]byte{
		[]byte("null"), []byte("[]"), []byte("{}{}"),
		[]byte(`{"schema":"x","schema":"y"}`),
		[]byte(`{"schema":"x","request":null}`),
		bytes.Replace(record.Bytes(), []byte(`"request":`), []byte(`"Request":`), 1),
		bytes.Replace(record.Bytes(), []byte(`"request":`), []byte(`"../request.json":`), 1),
		append([]byte("{\"extra\":0,"), record.Bytes()[1:]...),
	} {
		if _, err := Decode(raw); !errors.Is(err, ErrInvalidArchive) {
			t.Fatalf("accepted envelope: %v", err)
		}
	}
}

func TestLimitsAndReadFailures(t *testing.T) {
	if _, err := Capture(bytes.Repeat([]byte(" "), walkcase.MaxInputBytes+1)); !errors.Is(err, ErrTooLarge) {
		t.Fatal(err)
	}
	if _, err := Capture([]byte("{}")); !errors.Is(err, ErrInvalidRequest) {
		t.Fatal(err)
	}
	if _, err := Read(nil); !errors.Is(err, ErrInvalidArchive) {
		t.Fatal(err)
	}
	if _, err := Read(errorReader{}); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatal(err)
	}
	if _, err := Decode(make([]byte, MaxArtifactBytes+1)); !errors.Is(err, ErrTooLarge) {
		t.Fatal(err)
	}
	if (Record{}).Summary().Status != "invalid" {
		t.Fatal("zero record is valid")
	}
	if (Record{}).Reconstruct().Predicted() {
		t.Fatal("zero record reconstructed")
	}
}

func TestIllegalEnvelopeArrayDoesNotAllocateDecodedElements(t *testing.T) {
	// No envelope member admits an array. Beyond bounded syntax inspection,
	// rejection must not allocate an interface value/map for every attacker-
	// supplied element. Measure that additional allocation cost against the
	// same syntax pass rather than a platform-dependent total memory ceiling.
	raw := []byte(`{"schema":"` + Schema + `","request":[` + strings.Repeat(`{},`, 4095) + `{}` + `]}`)
	var syntaxIssue *strictjson.Issue
	syntaxAllocations := testing.AllocsPerRun(1, func() {
		syntaxIssue = strictjson.Inspect(raw, strictjson.Limits{MaximumDepth: 8, MaximumMemberNameBytes: 128})
	})
	if syntaxIssue != nil {
		t.Fatalf("fixture is not valid bounded syntax: %#v", syntaxIssue)
	}
	var decodeError error
	decodeAllocations := testing.AllocsPerRun(1, func() {
		_, decodeError = Decode(raw)
	})
	if !errors.Is(decodeError, ErrInvalidArchive) {
		t.Fatalf("illegal collection accepted: %v", decodeError)
	}
	if decodeAllocations > syntaxAllocations+512 {
		t.Fatalf("collection rejection allocated per-element values: syntax=%g, decode=%g", syntaxAllocations, decodeAllocations)
	}
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func TestAtomicNoClobberAndConcurrentPublication(t *testing.T) {
	record := capture(t)
	directory := t.TempDir()
	destination := filepath.Join(directory, "run.fartevidence")
	const contenders = 8
	var wg sync.WaitGroup
	results := make(chan error, contenders)
	for range contenders {
		wg.Go(func() { results <- WriteNew(context.Background(), destination, record) })
	}
	wg.Wait()
	close(results)
	winners := 0
	for err := range results {
		if err == nil {
			winners++
		} else if !errors.Is(err, ErrDestinationExists) {
			t.Fatal(err)
		}
	}
	if winners != 1 {
		t.Fatalf("publication winners: %d", winners)
	}
	stored, err := os.ReadFile(destination)
	if err != nil || !bytes.Equal(stored, record.Bytes()) {
		t.Fatalf("partial or wrong publication: %v", err)
	}
	entries, err := os.ReadDir(directory)
	if err != nil || len(entries) != 1 {
		t.Fatalf("temporary files remain: %v %v", entries, err)
	}
	if err := WriteNew(context.Background(), directory, record); !errors.Is(err, ErrDestinationExists) {
		t.Fatal(err)
	}
}

func TestPublicationCancellationAndExistingSymlink(t *testing.T) {
	record := capture(t)
	directory := t.TempDir()
	destination := filepath.Join(directory, "record")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := WriteNew(ctx, destination, record); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := os.Lstat(destination); !os.IsNotExist(err) {
		t.Fatalf("cancelled publication exists: %v", err)
	}
	//lint:ignore SA1012 Verify the explicit nil-context refusal without a panic.
	if err := WriteNew(nil, destination, record); !errors.Is(err, ErrInvalidArchive) {
		t.Fatal(err)
	}
	if err := WriteNew(context.Background(), destination, Record{}); !errors.Is(err, ErrInvalidArchive) {
		t.Fatal(err)
	}
	if err := WriteNew(context.Background(), filepath.Join(directory, "missing", "record"), record); err == nil {
		t.Fatal("missing parent created")
	}
	target := filepath.Join(t.TempDir(), "target")
	if err := os.WriteFile(target, []byte("retained"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, destination); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if err := WriteNew(context.Background(), destination, record); !errors.Is(err, ErrDestinationExists) {
		t.Fatal(err)
	}
	data, err := os.ReadFile(target)
	if err != nil || string(data) != "retained" {
		t.Fatalf("symlink target replaced: %v", err)
	}
}

type faultStage struct {
	writes                int
	syncs                 int
	short                 bool
	writeError, syncError error
	cancel                context.CancelFunc
}

func (f *faultStage) Write(data []byte) (int, error) {
	f.writes++
	if f.cancel != nil {
		f.cancel()
	}
	if f.writeError != nil {
		return 0, f.writeError
	}
	if f.short {
		return len(data) - 1, nil
	}
	return len(data), nil
}
func (f *faultStage) Sync() error {
	f.syncs++
	return f.syncError
}

func TestStagingFaultsDoNotReachPublication(t *testing.T) {
	for _, test := range []struct {
		stage faultStage
		want  error
	}{
		{faultStage{short: true}, io.ErrShortWrite},
		{faultStage{writeError: io.ErrClosedPipe}, io.ErrClosedPipe},
		{faultStage{syncError: io.ErrUnexpectedEOF}, io.ErrUnexpectedEOF},
	} {
		if err := stage(context.Background(), &test.stage, []byte("evidence")); !errors.Is(err, test.want) {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	writer := &faultStage{cancel: cancel}
	if err := stage(ctx, writer, make([]byte, 128<<10)); !errors.Is(err, context.Canceled) || writer.writes != 1 {
		t.Fatalf("cancellation: %v", err)
	}
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	writer = &faultStage{cancel: cancel}
	if err := stage(ctx, writer, []byte("last chunk")); !errors.Is(err, context.Canceled) || writer.writes != 1 || writer.syncs != 0 {
		t.Fatalf("cancellation during final write still synced: error=%v, writer=%#v", err, writer)
	}
}

func FuzzDecode(f *testing.F) {
	// Keep the valid seed small so short fuzz campaigns mutate the complete
	// envelope instead of spending their budget minimizing thousands of samples.
	var request map[string]any
	if err := json.Unmarshal(fixture(f), &request); err != nil {
		f.Fatal(err)
	}
	request["restriction"].(map[string]any)["area"].(map[string]any)["prescribed_m2"] = 0
	closed, err := json.Marshal(request)
	if err != nil {
		f.Fatal(err)
	}
	record, err := Capture(closed)
	if err != nil {
		f.Fatal(err)
	}
	f.Add(record.Bytes())
	f.Add([]byte("{}"))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > MaxArtifactBytes {
			return
		}
		record, err := Decode(data)
		if err != nil {
			return
		}
		if record.Summary().Status != "verified" || !bytes.Equal(record.Bytes(), data) {
			t.Fatal("accepted record lost evidence")
		}
		again, err := Decode(record.Bytes())
		if err != nil || again.Summary() != record.Summary() {
			t.Fatal("verification is not deterministic")
		}
	})
}
