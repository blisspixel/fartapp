package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/blisspixel/fartapp/internal/walkcase"
)

func TestVersionAndCrossCLIHint(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"fartapp", "--version"}, strings.NewReader(""), &stdout, &stderr); code != 0 || stderr.Len() != 0 || !strings.HasPrefix(stdout.String(), "fartapp v0.8.0-alpha.3 commit ") {
		t.Fatalf("version=(%d,%q,%q)", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code := Run([]string{"fartapp", "play"}, strings.NewReader(""), &stdout, &stderr)
	want := "'play' is available in the Rust CLI: fart play run testdata/play/reservoir-session.jsonl\n"
	if code != 2 || stdout.Len() != 0 || stderr.String() != want {
		t.Fatalf("play=(%d,%q,%q)", code, stdout.String(), stderr.String())
	}
}

func TestEveryCatalogCommandIsRecognized(t *testing.T) {
	for command, row := range parityIndex {
		var stdout, stderr bytes.Buffer
		code := Run([]string{"fartapp", command}, strings.NewReader(""), &stdout, &stderr)
		if row.home == "rust" {
			if code != 2 || !strings.Contains(stderr.String(), "Rust CLI") {
				t.Fatalf("%s=(%d,%q,%q)", command, code, stdout.String(), stderr.String())
			}
			continue
		}
		combined := stdout.String() + stderr.String()
		if strings.Contains(combined, "unknown command") || strings.Contains(stderr.String(), "usage: fartapp <intensity>") {
			t.Fatalf("%s fell through to intensity usage: %d %q %q", command, code, stdout.String(), stderr.String())
		}
	}
}

func TestCompletionListsCatalogCommandsAndOperations(t *testing.T) {
	required := catalogCommands()
	for _, operations := range completionOperations {
		required = append(required, operations...)
	}
	for _, shell := range []string{"bash", "zsh", "fish"} {
		var stdout, stderr bytes.Buffer
		if code := Run([]string{"fartapp", "completion", shell}, strings.NewReader(""), &stdout, &stderr); code != 0 || stderr.Len() != 0 {
			t.Fatalf("%s=(%d,%q)", shell, code, stderr.String())
		}
		script := stdout.String()
		for _, word := range required {
			if !strings.Contains(script, word) {
				t.Errorf("%s completion missing %q", shell, word)
			}
		}
	}
}

func TestReadmeMatchesParityCatalog(t *testing.T) {
	readme, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(readme)
	start := strings.Index(text, "## Command parity")
	if start < 0 {
		t.Fatal("README has no command parity section")
	}
	section := text[start:]
	if end := strings.Index(section[len("## Command parity"):], "\n## "); end >= 0 {
		section = section[:len("## Command parity")+end]
	}
	for command := range parityIndex {
		if !strings.Contains(section, "| "+command+" |") {
			t.Errorf("parity table missing %s", command)
		}
	}
	for _, phrase := range []string{
		"go install github.com/blisspixel/fartapp/cmd/fartapp@v0.8.0-alpha.3",
		"go install ./cmd/fartapp",
		"cargo install --locked --path crates/fart-cli",
		"fartapp completion bash",
	} {
		if !strings.Contains(text, phrase) {
			t.Errorf("README missing %s", phrase)
		}
	}
}

func TestQuickPlayPresetsBoundsAndTryNext(t *testing.T) {
	started := time.Now()
	var first, second bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"fartapp", "quickplay", "3", "--format", "json", "--seed", "7"}, strings.NewReader(""), &first, &stderr); code != 0 || stderr.Len() != 0 {
		t.Fatalf("quickplay 3=(%d,%q,%s)", code, stderr.String(), first.String())
	}
	if elapsed := time.Since(started); elapsed > 100*time.Millisecond {
		t.Fatalf("quickplay 3 took %s", elapsed)
	}
	if code := Run([]string{"fartapp", "quickplay", "3", "--seed", "7", "--format", "json"}, strings.NewReader(""), &second, &stderr); code != 0 || !bytes.Equal(first.Bytes(), second.Bytes()) {
		t.Fatal("same seed was not byte-identical")
	}
	var report quickplayReport
	if err := json.Unmarshal(first.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Preset != "ordinary" || report.Sound != "braaap" || report.Choked || len(report.TryNext) < 3 {
		t.Fatalf("report=%+v", report)
	}
	if strings.Count(formatQuickplay(report), "\n") > 24 {
		t.Fatalf("quickplay text has %d lines", strings.Count(formatQuickplay(report), "\n"))
	}
	for _, command := range report.TryNext {
		fields := strings.Fields(command)
		if len(fields) == 0 || fields[0] != "fartapp" {
			t.Fatalf("try next %q", command)
		}
		var out, errBuf bytes.Buffer
		if code := Run(fields, strings.NewReader(""), &out, &errBuf); code != 0 {
			t.Fatalf("%s failed (%d): %s", command, code, errBuf.String())
		}
	}
	assertQuickRegime(t, "1", false)
	assertQuickRegime(t, "5", true)
}

func assertQuickRegime(t *testing.T, intensity string, choked bool) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"fartapp", "quickplay", intensity, "--format", "json"}, strings.NewReader(""), &stdout, &stderr); code != 0 {
		t.Fatalf("quickplay %s=(%d,%s)", intensity, code, stderr.String())
	}
	var report quickplayReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Choked != choked {
		t.Fatalf("quickplay %s choked=%t preset=%s", intensity, report.Choked, report.Preset)
	}
}

func TestWalkNewOrdinarySimulatesLikeTheFixture(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"fartapp", "walk", "new", "--preset", "ordinary"}, strings.NewReader(""), &stdout, &stderr); code != 0 || stderr.Len() != 0 {
		t.Fatalf("walk new=(%d,%q)", code, stderr.String())
	}
	fixture, err := os.ReadFile(filepath.FromSlash("testdata/walk/ordinary-low-pressure.json"))
	if err != nil {
		t.Fatal(err)
	}
	fromPreset := walkcase.Run(stdout.Bytes(), "simulate")
	fromFixture := walkcase.Run(fixture, "simulate")
	if !fromPreset.Predicted() || !fromFixture.Predicted() || fromPreset.Signature == nil || fromFixture.Signature == nil {
		t.Fatal("ordinary preset did not simulate")
	}
	if *fromPreset.MassOutKilograms != *fromFixture.MassOutKilograms || fromPreset.Signature.ChokedOccurred != fromFixture.Signature.ChokedOccurred {
		t.Fatal("walk new ordinary diverged from the fixture")
	}
}

func TestIntensityWavKeepsToyStdout(t *testing.T) {
	path := filepath.Join(t.TempDir(), "toy.wav")
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"fartapp", "3", "--wav", path}, strings.NewReader(""), &stdout, &stderr); code != 0 || stderr.Len() != 0 || stdout.String() != "braaap (respectable)\n" {
		t.Fatalf("wav intensity=(%d,%q,%q)", code, stdout.String(), stderr.String())
	}
	first, err := os.ReadFile(path)
	if err != nil || !bytes.HasPrefix(first, []byte("RIFF")) {
		t.Fatalf("wav missing: %v", err)
	}
	again := filepath.Join(t.TempDir(), "toy-again.wav")
	stdout.Reset()
	if code := Run([]string{"fartapp", "3", "--wav=" + again}, strings.NewReader(""), &stdout, &stderr); code != 0 {
		t.Fatal(stderr.String())
	}
	second, err := os.ReadFile(again)
	if err != nil || !bytes.Equal(first, second) {
		t.Fatal("intensity wav was not deterministic")
	}
}

func TestSonifyIsDeterministicBoundedAndDistinct(t *testing.T) {
	ordinary := sonifyFile(t, "testdata/walk/ordinary-low-pressure.json")
	repeat := sonifyFile(t, "testdata/walk/ordinary-low-pressure.json")
	if !bytes.Equal(ordinary.wav, repeat.wav) {
		t.Fatal("sonify was not byte-identical")
	}
	sum := sha256.Sum256(ordinary.wav)
	got := hex.EncodeToString(sum[:])
	const ordinaryGolden = "9ba90624acb3ce686ccf80b970e2bda9af3c7d6d3f2370a5146586c91c2fae3b"
	if got != ordinaryGolden {
		t.Fatalf("ordinary wav sha256 = %s", got)
	}
	choked := sonifyFile(t, "testdata/walk/isothermal-choked.json")
	if bytes.Equal(ordinary.wav, choked.wav) {
		t.Fatal("ordinary and choked sonifications were identical")
	}
	if !strings.Contains(string(ordinary.wav), sonifyMappingID+"/"+sonifyMappingVersion) || !strings.Contains(string(ordinary.wav), sonifyNonclaim) {
		t.Fatal("wav metadata missing mapping or nonclaim")
	}
	assertPeakAtMost(t, ordinary.wav, defaultGainDBFS)
	assertPeakAtMost(t, choked.wav, defaultGainDBFS)
	if ordinary.meta.ChokedStartSample != nil {
		t.Fatal("ordinary sonification reported choking")
	}
	if choked.meta.ChokedStartS == nil || choked.meta.ChokedEndS == nil {
		t.Fatal("choked sonification omitted the choking interval")
	}
	account := walkcase.Run(mustRead(t, "testdata/walk/isothermal-choked.json"), "simulate")
	minT, maxT, seen := 0.0, 0.0, false
	for _, sample := range account.History {
		if sample.Regime != "choked" {
			continue
		}
		if !seen || sample.TimeSeconds < minT {
			minT = sample.TimeSeconds
		}
		if !seen || sample.TimeSeconds > maxT {
			maxT = sample.TimeSeconds
		}
		seen = true
	}
	if !seen {
		t.Fatal("fixture history never choked")
	}
	if *choked.meta.ChokedStartS+1e-9 < minT || *choked.meta.ChokedEndS > maxT+1e-6 {
		t.Fatalf("choke audio [%g,%g] outside history [%g,%g]", *choked.meta.ChokedStartS, *choked.meta.ChokedEndS, minT, maxT)
	}
}

type sonifyArtifact struct {
	wav  []byte
	meta sonifyReport
}

func sonifyFile(t *testing.T, source string) sonifyArtifact {
	t.Helper()
	path := filepath.Join(t.TempDir(), "out.wav")
	var stdout, stderr bytes.Buffer
	code := Run([]string{"fartapp", "walk", "sonify", source, "--output", path, "--format", "json"}, strings.NewReader(""), &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("sonify %s=(%d,%q,%s)", source, code, stderr.String(), stdout.String())
	}
	var meta sonifyReport
	if err := json.Unmarshal(stdout.Bytes(), &meta); err != nil {
		t.Fatal(err)
	}
	if meta.MappingID != sonifyMappingID || meta.Nonclaims[0] != sonifyNonclaim || meta.PeakDBFS > defaultGainDBFS+1e-9 {
		t.Fatalf("meta=%+v", meta)
	}
	wav, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return sonifyArtifact{wav: wav, meta: meta}
}

func assertPeakAtMost(t *testing.T, wav []byte, gain float64) {
	t.Helper()
	samples := pcmSamples(t, wav)
	target := int(math.Floor(32767 * math.Pow(10, gain/20)))
	for _, sample := range samples {
		magnitude := int(sample)
		if magnitude < 0 {
			magnitude = -magnitude
		}
		if magnitude > target {
			t.Fatalf("sample %d exceeds dBFS target %d", magnitude, target)
		}
	}
}

func pcmSamples(t *testing.T, data []byte) []int16 {
	t.Helper()
	if len(data) < 12 || string(data[:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		t.Fatal("not a wav")
	}
	offset := 12
	for offset+8 <= len(data) {
		id := string(data[offset : offset+4])
		size := int(binary.LittleEndian.Uint32(data[offset+4 : offset+8]))
		offset += 8
		if size < 0 || offset+size > len(data) {
			t.Fatal("truncated wav")
		}
		if id == "data" {
			payload := data[offset : offset+size]
			samples := make([]int16, len(payload)/2)
			for index := range samples {
				samples[index] = int16(binary.LittleEndian.Uint16(payload[index*2:]))
			}
			return samples
		}
		offset += size
		if size%2 == 1 {
			offset++
		}
	}
	t.Fatal("wav has no data chunk")
	return nil
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.FromSlash(path))
	if err != nil {
		t.Fatal(err)
	}
	return data
}
