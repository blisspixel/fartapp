package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestPublishWAVRefusesExistingUnlessForced(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "out.wav")
	if err := publishWAV(dest, []byte("one"), false); err != nil {
		t.Fatal(err)
	}
	if err := publishWAV(dest, []byte("two"), false); !errors.Is(err, errOutputExists) {
		t.Fatalf("existing file err = %v", err)
	}
	got, err := os.ReadFile(dest)
	if err != nil || string(got) != "one" {
		t.Fatalf("existing bytes = %q, %v", got, err)
	}
	if err := publishWAV(dest, []byte("two"), true); err != nil {
		t.Fatal(err)
	}
	got, err = os.ReadFile(dest)
	if err != nil || string(got) != "two" {
		t.Fatalf("replaced bytes = %q, %v", got, err)
	}
	assertNoWAVStagingFiles(t, dir)

	if err := publishWAV("", []byte("x"), true); err == nil {
		t.Fatal("empty path published")
	}
	if err := publishWAV(dir, []byte("x"), true); err == nil {
		t.Fatal("directory published")
	}
	missing := filepath.Join(dir, "missing", "out.wav")
	if err := publishWAV(missing, []byte("x"), false); err == nil {
		t.Fatal("missing parent published")
	}
	if _, err := os.Lstat(missing); !os.IsNotExist(err) {
		t.Fatalf("missing parent created: %v", err)
	}
}

func TestPublishWAVPreservesSymlinkTarget(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.wav")
	if err := os.WriteFile(target, []byte("kept"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link.wav")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if err := publishWAV(link, []byte("RIFF"), false); !errors.Is(err, errOutputExists) {
		t.Fatalf("symlink err = %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil || string(got) != "kept" {
		t.Fatalf("target = %q, %v", got, err)
	}
	info, err := os.Lstat(link)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("symlink replaced: %v", err)
	}
	if err := publishWAV(link, []byte("RIFF"), true); err != nil {
		t.Fatal(err)
	}
	got, err = os.ReadFile(target)
	if err != nil || string(got) != "kept" {
		t.Fatalf("forced replace changed target: %q, %v", got, err)
	}
	info, err = os.Lstat(link)
	if err != nil || info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("force left a symlink: %v", err)
	}
	got, err = os.ReadFile(link)
	if err != nil || string(got) != "RIFF" {
		t.Fatalf("link bytes = %q, %v", got, err)
	}
	assertNoWAVStagingFiles(t, dir)
}

func TestWAVCommandsRefuseExistingFileUnlessForced(t *testing.T) {
	dir := t.TempDir()
	original := []byte("keep\n")
	intensityPath := filepath.Join(dir, "y.wav")
	if err := os.WriteFile(intensityPath, original, 0o644); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := runCLI(t, "fartapp", "3", "--wav", intensityPath)
	assertExistingWAVRefused(t, code, stdout, stderr, intensityPath, original)

	code, stdout, stderr = runCLI(t, "fartapp", "3", "--wav="+intensityPath, "--force")
	if code != 0 || stdout != "braaap (respectable)\n" || stderr != "" {
		t.Fatalf("forced intensity=(%d,%q,%q)", code, stdout, stderr)
	}
	if got, err := os.ReadFile(intensityPath); err != nil || !bytes.HasPrefix(got, []byte("RIFF")) {
		t.Fatalf("forced intensity wav = %q, %v", got, err)
	}

	sonifyPath := filepath.Join(dir, "x.wav")
	if err := os.WriteFile(sonifyPath, original, 0o644); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join("testdata", "walk", "ordinary-low-pressure.json")
	code, stdout, stderr = runCLI(t, "fartapp", "walk", "sonify", source, "--output", sonifyPath)
	assertExistingWAVRefused(t, code, stdout, stderr, sonifyPath, original)

	code, stdout, stderr = runCLI(t, "fartapp", "walk", "sonify", source, "--output", sonifyPath, "--force")
	if code != 0 || stderr != "" || !strings.Contains(stdout, "Output: "+sonifyPath) {
		t.Fatalf("forced sonify=(%d,%q,%q)", code, stdout, stderr)
	}
	if got, err := os.ReadFile(sonifyPath); err != nil || !bytes.HasPrefix(got, []byte("RIFF")) || bytes.Equal(got, original) {
		t.Fatalf("forced sonify wav = %q, %v", got, err)
	}
	assertNoWAVStagingFiles(t, dir)
}

func TestWAVForceUsageAndUnwritableDestinations(t *testing.T) {
	dir := t.TempDir()
	fresh := filepath.Join(dir, "fresh.wav")
	code, stdout, stderr := runCLI(t, "fartapp", "3", "--wav", fresh, "--force", "--force")
	if code == 0 || stdout != "" || !strings.Contains(stderr, "--force") {
		t.Fatalf("repeated force=(%d,%q,%q)", code, stdout, stderr)
	}
	if _, err := os.Lstat(fresh); !os.IsNotExist(err) {
		t.Fatalf("repeated force created %s: %v", fresh, err)
	}

	code, stdout, stderr = runCLI(t, "fartapp", "3", "--force")
	if code == 0 || stdout != "" || !strings.Contains(stderr, "[--force]") {
		t.Fatalf("force without wav=(%d,%q,%q)", code, stdout, stderr)
	}

	created := filepath.Join(dir, "new.wav")
	code, stdout, stderr = runCLI(t, "fartapp", "3", "--force", "--wav", created)
	if code != 0 || stdout != "braaap (respectable)\n" {
		t.Fatalf("new file with force=(%d,%q,%q)", code, stdout, stderr)
	}
	if got, err := os.ReadFile(created); err != nil || !bytes.HasPrefix(got, []byte("RIFF")) {
		t.Fatalf("new wav = %q, %v", got, err)
	}

	missing := filepath.Join(dir, "missing", "out.wav")
	code, stdout, stderr = runCLI(t, "fartapp", "3", "--wav", missing)
	if code == 0 || stdout != "" {
		t.Fatalf("missing parent=(%d,%q,%q)", code, stdout, stderr)
	}
	if _, err := os.Lstat(missing); !os.IsNotExist(err) {
		t.Fatalf("missing parent created: %v", err)
	}

	code, _, stderr = runCLI(t, "fartapp", "3", "--wav", dir, "--force")
	if code == 0 || !strings.Contains(stderr, dir) {
		t.Fatalf("directory wav=(%d,%q)", code, stderr)
	}
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		t.Fatalf("destination directory changed: %v", err)
	}

	source := filepath.Join("testdata", "walk", "ordinary-low-pressure.json")
	code, stdout, stderr = runCLI(t, "fartapp", "walk", "sonify", source, "--output", fresh, "--force", "--force")
	if code == 0 || stdout != "" || !strings.Contains(stderr, "--force may be specified only once") {
		t.Fatalf("repeated sonify force=(%d,%q,%q)", code, stdout, stderr)
	}
	if _, err := os.Lstat(fresh); !os.IsNotExist(err) {
		t.Fatalf("repeated sonify force created %s: %v", fresh, err)
	}
	assertNoWAVStagingFiles(t, dir)
}

func TestHelpAndReadmeDocumentWAVForce(t *testing.T) {
	if !strings.Contains(rootHelp, "[--wav <out.wav>] [--force]") {
		t.Fatal("root help omits --force")
	}
	for _, help := range []string{sonifyHelp, walkHelpForOperation("sonify")} {
		if !strings.Contains(help, "--force") || !strings.Contains(help, "Refused if that file exists.") {
			t.Fatalf("sonify help omits the refusal: %s", help)
		}
	}
	readme, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(readme), "unless `--force` is also given") {
		t.Fatal("README omits --force")
	}
	code, stdout, stderr := runCLI(t, "fartapp", "walk", "sonify", "--help")
	if code != 0 || stderr != "" || !strings.Contains(stdout, "--force") {
		t.Fatalf("sonify help=(%d,%q,%q)", code, stdout, stderr)
	}
}

func runCLI(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := Run(args, strings.NewReader(""), &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func assertExistingWAVRefused(t *testing.T, code int, stdout, stderr, path string, original []byte) {
	t.Helper()
	if code == 0 || stdout != "" {
		t.Fatalf("refused=(%d,%q,%q)", code, stdout, stderr)
	}
	if !strings.Contains(stderr, strconv.QuoteToASCII(path)) || !strings.Contains(stderr, "--force") {
		t.Fatalf("stderr = %q", stderr)
	}
	got, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(got, original) {
		t.Fatalf("file changed: %v %q", err, got)
	}
}

func assertNoWAVStagingFiles(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".fartwav-") {
			t.Fatalf("left staging file %s", entry.Name())
		}
	}
}
