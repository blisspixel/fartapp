package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/blisspixel/fartapp/internal/walkcase"
)

func TestVersionAndCrossCLIHint(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"fartapp", "--version"}, strings.NewReader(""), &stdout, &stderr); code != 0 || stderr.Len() != 0 || !strings.HasPrefix(stdout.String(), "fartapp v0.8.0-alpha.7 commit ") {
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
		"go install github.com/blisspixel/fartapp/cmd/fartapp@v0.8.0-alpha.7",
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
	// Installed binaries must offer useful next steps outside a source checkout.
	t.Chdir(t.TempDir())
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

func TestQuickPlayExplanationsAndLayout(t *testing.T) {
	for _, tc := range []struct {
		level      string
		preset     string
		wantStop   string
		wantRegime string
		wantBudget bool
		avoidNext  string
		wantWav    string
	}{
		{"1", "gentle", "max-steps", "subsonic", true, "fartapp quickplay 1", "fartapp 1 --wav pfft.wav"},
		{"3", "ordinary", "equalized", "subsonic", false, "fartapp quickplay 3", "fartapp 3 --wav braaap.wav"},
		{"5", "choked", "max-time", "sonic", true, "fartapp quickplay 5", "fartapp 5 --wav kablam.wav"},
	} {
		var stdout, stderr bytes.Buffer
		code := Run([]string{"fartapp", "quickplay", tc.level, "--format", "json"}, strings.NewReader(""), &stdout, &stderr)
		if code != 0 || stderr.Len() != 0 {
			t.Fatalf("quickplay %s failed: code=%d, stderr=%s", tc.level, code, stderr.String())
		}
		var report quickplayReport
		if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
			t.Fatalf("unmarshal quickplay %s: %v", tc.level, err)
		}
		if report.Stop != tc.wantStop {
			t.Fatalf("level %s stop = %q, want %q", tc.level, report.Stop, tc.wantStop)
		}
		if report.Preset != tc.preset {
			t.Fatalf("level %s preset = %q, want %q", tc.level, report.Preset, tc.preset)
		}
		if strings.Contains(report.Explanation, "The walk account did not include an explanation") {
			t.Fatalf("level %s has fallback explanation", tc.level)
		}
		if strings.Contains(report.Explanation, "This is one explicit synthetic SI reservoir") {
			t.Fatalf("level %s has generic disclaimer explanation", tc.level)
		}
		if !strings.Contains(strings.ToLower(report.Explanation), tc.wantRegime) {
			t.Fatalf("level %s explanation %q missing regime %q", tc.level, report.Explanation, tc.wantRegime)
		}
		if tc.wantBudget && !strings.Contains(strings.ToLower(report.Explanation), "budget") {
			t.Fatalf("level %s explanation %q missing budget stop note", tc.level, report.Explanation)
		}
		if !tc.wantBudget && !strings.Contains(strings.ToLower(report.Explanation), "roundoff") {
			t.Fatalf("level %s explanation %q missing roundoff stop note", tc.level, report.Explanation)
		}
		foundWav := false
		for _, cmd := range report.TryNext {
			if cmd == tc.avoidNext {
				t.Fatalf("level %s suggests itself: %s", tc.level, cmd)
			}
			if cmd == tc.wantWav {
				foundWav = true
			}
		}
		if !foundWav {
			t.Fatalf("level %s missing wav command %q in try_next: %v", tc.level, tc.wantWav, report.TryNext)
		}

		// Text layout checks
		text := formatQuickplay(report)
		lines := strings.Split(text, "\n")
		if len(lines) > 25 {
			t.Fatalf("level %s text has %d lines, want <= 25", tc.level, len(lines))
		}
		for lineIdx, line := range lines {
			if len(line) > 80 {
				t.Fatalf("level %s line %d exceeds 80 columns (%d chars): %q", tc.level, lineIdx+1, len(line), line)
			}
		}
		if !strings.Contains(text, "recorded only") {
			t.Fatalf("level %s text missing 'recorded only': %s", tc.level, text)
		}
	}
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

func TestQuickPlayHelpNamesAudioWithoutAcceptingWav(t *testing.T) {
	const usage = "  fartapp quickplay [1-5] [--seed <integer>] [--format text|json]"
	if !strings.Contains(quickplayHelp, usage+"\n") {
		t.Fatal("quickplay usage line changed")
	}
	for _, line := range strings.Split(quickplayHelp, "\n") {
		if len(line) > 80 {
			t.Fatalf("help line is %d columns: %q", len(line), line)
		}
		if strings.Contains(line, "fartapp quickplay [1-5]") && strings.Contains(line, "--wav") {
			t.Fatalf("usage accepts --wav: %q", line)
		}
	}
	for _, command := range []string{
		"fartapp 3 --wav braaap.wav",
		"fartapp walk sonify <case> --output out.wav",
	} {
		if !strings.Contains(quickplayHelp, "\n  "+command+"\n") {
			t.Fatalf("help missing copyable command %s", command)
		}
	}
	for _, args := range [][]string{
		{"fartapp", "quickplay", "--help"},
		{"fartapp", "quickplay", "-h"},
		{"fartapp", "help", "quickplay"},
	} {
		var stdout, stderr bytes.Buffer
		code := Run(args, strings.NewReader(""), &stdout, &stderr)
		if code != 0 || stderr.Len() != 0 || stdout.String() != quickplayHelp {
			t.Fatalf("%v=(%d,%q,%q)", args, code, stdout.String(), stderr.String())
		}
	}
	script, err := completionScript("bash")
	if err != nil {
		t.Fatal(err)
	}
	const quickplayCompletion = `quickplay) COMPREPLY=( $(compgen -W "1 2 3 4 5" -- "${cur}") )`
	if !strings.Contains(script, quickplayCompletion) {
		t.Fatalf("quickplay completion changed: %s", script)
	}
}

func TestQuickPlayWavNamesBothAudioCommands(t *testing.T) {
	cases := []struct {
		args   []string
		level  int
		target string
	}{
		{args: []string{"fartapp", "quickplay", "3", "--wav", "out.wav"}, level: 3, target: `"out.wav"`},
		{args: []string{"fartapp", "quickplay", "5", "--wav=kablam.wav"}, level: 5, target: `"kablam.wav"`},
		{args: []string{"fartapp", "quickplay", "--seed", "4", "--wav", "out.wav"}, level: 3, target: `"out.wav"`},
		{args: []string{"fartapp", "quickplay", "--seed=4", "--format", "json", "1", "--wav", "out.wav"}, level: 1, target: `"out.wav"`},
		{args: []string{"fartapp", "quickplay", "--wav"}, level: 3, target: "braaap.wav"},
		{args: []string{"fartapp", "quickplay", "5", "--wav"}, level: 5, target: "kablam.wav"},
		{args: []string{"fartapp", "quickplay", "1", "--wav"}, level: 1, target: "pfft.wav"},
		{args: []string{"fartapp", "quickplay", "4", "--wav"}, level: 4, target: "blorp.wav"},
		{args: []string{"fartapp", "quickplay", "2", "--wav"}, level: 2, target: "toot.wav"},
		{args: []string{"fartapp", "quickplay", "9", "--wav", "out.wav"}, level: 3, target: `"out.wav"`},
		{args: []string{"fartapp", "quickplay", "--wav", "out.wav", "5"}, level: 5, target: `"out.wav"`},
		{args: []string{"fartapp", "quickplay", "--wav", "out.wav", "03"}, level: 3, target: `"out.wav"`},
		{args: []string{"fartapp", "quickplay", "--seed", "--wav", "out.wav"}, level: 3, target: `"out.wav"`},
		{args: []string{"fartapp", "quickplay", "--format", "--wav", "out.wav"}, level: 3, target: `"out.wav"`},
		{args: []string{"fartapp", "quickplay", "--help", "--wav", "out.wav"}, level: 3, target: `"out.wav"`},
		{args: []string{"fartapp", "quickplay", "--wav", "out.wav", "--help"}, level: 3, target: `"out.wav"`},
		{args: []string{"fartapp", "quickplay", "--wav="}, level: 3, target: `""`},
		{args: []string{"fartapp", "quickplay", "--wav", "out file.wav"}, level: 3, target: `"out file.wav"`},
		{args: []string{"fartapp", "quickplay", "--wav", "out\nwav"}, level: 3, target: `"out\nwav"`},
		{args: []string{"fartapp", "quickplay", "--wav", "--help"}, level: 3, target: "braaap.wav"},
		{args: []string{"fartapp", "quickplay", "5", "1", "--wav"}, level: 5, target: "kablam.wav"},
	}
	for _, tc := range cases {
		requireQuickplayAudioRefusal(t, tc.args, tc.level, tc.target)
	}

	longPath := "artifacts/session-recordings/braaap-ordinary.wav"
	if len(longPath) <= maximumDisplayedInputBytes {
		t.Fatalf("fixture path is %d bytes", len(longPath))
	}
	quoted := strconv.QuoteToASCII(longPath)
	var stdout, stderr bytes.Buffer
	code := Run([]string{"fartapp", "quickplay", "3", "--wav", longPath}, strings.NewReader(""), &stdout, &stderr)
	wantLong := quickplayAudioDiagnostic(3, quoted)
	if code != 1 || stdout.Len() != 0 || stderr.String() != wantLong {
		t.Fatalf("long path=(%d,%q) want %q", code, stderr.String(), wantLong)
	}
	if strings.Contains(stderr.String(), quoteInput(longPath)) {
		t.Fatalf("path was truncated to %d bytes: %q", maximumDisplayedInputBytes, stderr.String())
	}
	longLines := strings.Split(strings.TrimSuffix(stderr.String(), "\n"), "\n")
	if len(longLines) != 3 || len(longLines[0]) > 80 || len(longLines[2]) <= 80 {
		t.Fatalf("long path layout = %#v", longLines)
	}

	hostile := strings.Repeat("a", maximumQuickplayPathBytes+1)
	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"fartapp", "quickplay", "--wav", hostile}, strings.NewReader(""), &stdout, &stderr)
	bounded := strconv.QuoteToASCII(strings.Repeat("a", maximumQuickplayPathBytes) + "...")
	if code != 1 || stderr.String() != quickplayAudioDiagnostic(3, bounded) || strings.Contains(stderr.String(), strings.Repeat("a", maximumQuickplayPathBytes+1)) {
		t.Fatalf("hostile path=(%d,len %d)", code, stderr.Len())
	}
	exact := strings.Repeat("b", maximumQuickplayPathBytes)
	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"fartapp", "quickplay", "--wav", exact}, strings.NewReader(""), &stdout, &stderr)
	if code != 1 || stderr.String() != quickplayAudioDiagnostic(3, strconv.QuoteToASCII(exact)) || strings.Contains(stderr.String(), "...") {
		t.Fatalf("exact bound=(%d,%t)", code, strings.Contains(stderr.String(), "..."))
	}

	for _, args := range [][]string{
		{"fartapp", "quickplay", "--wav", "out.wav", "--wav", "other.wav"},
		{"fartapp", "quickplay", "--wav=out.wav", "--wav=other.wav"},
		{"fartapp", "quickplay", "--wav", "out.wav", "--wav=out.wav"},
		{"fartapp", "quickplay", "--wav", "--wav"},
	} {
		stdout.Reset()
		stderr.Reset()
		code = Run(args, strings.NewReader(""), &stdout, &stderr)
		if code != 1 || stdout.Len() != 0 || stderr.String() != "invalid quickplay: --wav may be specified only once\n" || strings.Contains(stderr.String(), "does not write audio") {
			t.Fatalf("duplicate %v=(%d,%q)", args, code, stderr.String())
		}
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"fartapp", "quickplay", "3", "--", "--wav", "out.wav"}, strings.NewReader(""), &stdout, &stderr)
	if code != 1 || stdout.Len() != 0 || strings.Contains(stderr.String(), "does not write audio") {
		t.Fatalf("end of options = (%d,%q,%q)", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"fartapp", "quickplay", "3", "--nope"}, strings.NewReader(""), &stdout, &stderr)
	if code != 1 || stderr.String() != "invalid quickplay: unknown option \"--nope\"\n" {
		t.Fatalf("unknown option = (%d,%q)", code, stderr.String())
	}
}

func requireQuickplayAudioRefusal(t *testing.T, args []string, level int, target string) {
	t.Helper()
	want := quickplayAudioDiagnostic(level, target)
	var stdout, stderr bytes.Buffer
	code := Run(args, strings.NewReader(""), &stdout, &stderr)
	if code != 1 || stdout.Len() != 0 || stderr.String() != want {
		t.Fatalf("%v=(%d,%q) want %q", args, code, stderr.String(), want)
	}
	if strings.Count(stderr.String(), "\n") != 3 {
		t.Fatalf("%v broke the diagnostic into %d lines: %q", args, strings.Count(stderr.String(), "\n"), stderr.String())
	}
	for _, line := range strings.Split(strings.TrimSuffix(stderr.String(), "\n"), "\n") {
		if len(line) > 80 {
			t.Fatalf("%v line is %d columns: %q", args, len(line), line)
		}
	}
}

func quickplayAudioDiagnostic(level int, target string) string {
	return fmt.Sprintf(
		"invalid quickplay: this command does not write audio.\n"+
			"  Toy sound: fartapp %d --wav %s\n"+
			"  Walk sound: fartapp walk sonify <case> --output %s\n",
		level,
		target,
		target,
	)
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

type mockTerminalBuffer struct {
	bytes.Buffer
}

func (m *mockTerminalBuffer) IsTerminal() bool {
	return true
}

func TestIntensityWavReceiptOnTerminal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "braaap.wav")
	var stdout bytes.Buffer
	var stderr mockTerminalBuffer
	code := Run([]string{"fartapp", "3", "--wav", path}, strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code %d, stderr: %s", code, stderr.String())
	}
	if stdout.String() != "braaap (respectable)\n" {
		t.Fatalf("stdout = %q, want %q", stdout.String(), "braaap (respectable)\n")
	}
	wav, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(wav)
	fullDigest := hex.EncodeToString(sum[:])
	shortDigest := fullDigest[:12]
	wantReceipt := fmt.Sprintf("wrote %s  1.11 s  sha256 %s  regenerate: fartapp 3 --wav %s  Conservation of Ass: satisfied within roundoff\n",
		path, shortDigest, path)
	if stderr.String() != wantReceipt {
		t.Fatalf("stderr = %q, want %q", stderr.String(), wantReceipt)
	}

	// Two runs print the same digest
	againPath := filepath.Join(t.TempDir(), "again.wav")
	var stdoutAgain bytes.Buffer
	var stderrAgain mockTerminalBuffer
	code = Run([]string{"fartapp", "3", "--wav", againPath}, strings.NewReader(""), &stdoutAgain, &stderrAgain)
	if code != 0 {
		t.Fatalf("exit code %d, stderr: %s", code, stderrAgain.String())
	}
	wantAgainReceipt := fmt.Sprintf("wrote %s  1.11 s  sha256 %s  regenerate: fartapp 3 --wav %s  Conservation of Ass: satisfied within roundoff\n",
		againPath, shortDigest, againPath)
	if stderrAgain.String() != wantAgainReceipt {
		t.Fatalf("stderr = %q, want %q", stderrAgain.String(), wantAgainReceipt)
	}

	// Redirected stderr (non-terminal) prints nothing extra to stderr
	var stderrRedirected bytes.Buffer
	var stdoutRedirected bytes.Buffer
	code = Run([]string{"fartapp", "3", "--wav", filepath.Join(t.TempDir(), "redir.wav")}, strings.NewReader(""), &stdoutRedirected, &stderrRedirected)
	if code != 0 || stderrRedirected.Len() != 0 {
		t.Fatalf("redirected stderr was not empty: code=%d, stderr=%q", code, stderrRedirected.String())
	}

	// Receipt without mass-ledger claim does not show Conservation of Ass
	withoutClaim := formatIntensityReceipt("test.wav", 1.5, "1234567890abcdef", 1, nil)
	if strings.Contains(withoutClaim, "Conservation of Ass") {
		t.Fatalf("unexpected deadpan label: %s", withoutClaim)
	}
	wantWithout := "wrote test.wav  1.50 s  sha256 1234567890ab  regenerate: fartapp 1 --wav test.wav"
	if withoutClaim != wantWithout {
		t.Fatalf("without claim = %q, want %q", withoutClaim, wantWithout)
	}

	failedClaim := formatIntensityReceipt("test.wav", 1.5, "1234567890abcdef", 1, []walkcase.Claim{
		{ID: "walk.mass-ledger", Status: "failed"},
	})
	if strings.Contains(failedClaim, "Conservation of Ass") {
		t.Fatalf("unexpected deadpan label on failed claim: %s", failedClaim)
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
	if ordinary.meta.SHA256 != ordinaryGolden {
		t.Fatalf("ordinary sonifyReport.SHA256 = %s, want %s", ordinary.meta.SHA256, ordinaryGolden)
	}
	chokedSum := sha256.Sum256(choked.wav)
	if choked.meta.SHA256 != hex.EncodeToString(chokedSum[:]) {
		t.Fatalf("choked sonifyReport.SHA256 = %s, want %s", choked.meta.SHA256, hex.EncodeToString(chokedSum[:]))
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

func TestReportsFit80ColumnsAndSentenceCase(t *testing.T) {
	commands := [][]string{
		{"walk", "explain", "testdata/walk/ordinary-low-pressure.json"},
		{"walk", "simulate", "testdata/walk/ordinary-low-pressure.json"},
		{"reservoir", "predict", "testdata/reservoir/synthetic-mixture-adiabatic.json"},
		{"restriction", "predict", "testdata/restriction/gamma15-choked.json"},
		{"restriction", "history", "testdata/restriction/gamma15-choked-history.json"},
	}

	for _, cmd := range commands {
		t.Run(strings.Join(cmd, "_"), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			args := append([]string{"fartapp"}, cmd...)
			args[len(args)-1] = filepath.FromSlash(args[len(args)-1])
			code := Run(args, strings.NewReader(""), &stdout, &stderr)
			if code != 0 || stderr.Len() != 0 {
				t.Fatalf("command failed: code=%d stderr=%s", code, stderr.String())
			}

			out := stdout.String()
			lines := strings.Split(out, "\n")

			// Check first line (report title) is sentence case, not uppercase.
			if len(lines) > 0 && len(lines[0]) > 0 {
				first := lines[0]
				if strings.ToUpper(first) == first && strings.ToLower(first) != first {
					t.Errorf("report title %q is uppercase; expected sentence case", first)
				}
			}

			// If Explanation section is present, check every explanation line is <= 80 cols.
			if idx := strings.Index(out, "Explanation\n"); idx != -1 {
				rest := out[idx+len("Explanation\n"):]
				end := strings.Index(rest, "\n\n")
				if end != -1 {
					rest = rest[:end]
				}
				for _, line := range strings.Split(rest, "\n") {
					if len(line) > 80 {
						t.Errorf("explanation line exceeds 80 columns (%d cols): %q", len(line), line)
					}
				}
			}

			// Check identifier lists (Assumptions and Nonclaims) format: one per line, indented, no commas.
			for _, label := range []string{"Assumptions:\n", "Model nonclaims:\n", "Operation nonclaims:\n", "Evidence nonclaims:\n"} {
				if idx := strings.Index(out, label); idx != -1 {
					rest := out[idx+len(label):]
					end := strings.Index(rest, "\n\n")
					if end != -1 {
						rest = rest[:end]
					}
					for _, item := range strings.Split(rest, "\n") {
						if strings.TrimSpace(item) == "" {
							continue
						}
						if strings.Contains(item, ",") {
							t.Errorf("identifier list contained comma: %q", item)
						}
						if !strings.HasPrefix(item, "  ") {
							t.Errorf("identifier not indented by 2 spaces: %q", item)
						}
					}
				}
			}
		})
	}
}
