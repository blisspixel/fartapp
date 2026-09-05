package repoquality

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestCurrentRepositoryPassesPolicy(t *testing.T) {
	root, err := FindRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	result, err := CheckRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Failures) != 0 {
		t.Fatalf("current repository failed: %s\n%s", joinFailures(result.Failures), result.Report)
	}
	if !strings.Contains(result.Report, "dependency policy verified:") ||
		!strings.Contains(result.Report, "local Markdown links verified:") ||
		!strings.Contains(result.Report, "media manifest verified:") {
		t.Fatalf("report = %q", result.Report)
	}
}

func TestCheckDependenciesRejectsPolicyViolations(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/lab\n")
	writeFile(t, filepath.Join(root, "package-lock.json"), `{
		"lockfileVersion": 2,
		"packages": {
			"": {},
			"node_modules/evil": {
				"license": "UNLICENSED",
				"integrity": "sha1-nope",
				"resolved": "https://example.invalid/evil"
			}
		}
	}`)
	result, err := CheckDependencies(root)
	if err != nil {
		t.Fatal(err)
	}
	joined := joinFailures(result.Failures)
	for _, want := range []string{
		"unsupported npm lockfile version",
		"unreviewed npm license",
		"not development-only",
		"invalid npm integrity",
		"unexpected npm source",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %q in %s", want, joined)
		}
	}
}

func TestCheckLinksFindsMissingAndEscapingTargets(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/lab\n")
	writeFile(t, filepath.Join(root, "README.md"), "[ok](docs/exists.md) [gone](docs/missing.md) [web](https://example.com) [escape](../outside.md) ![img](docs/exists.md)\n")
	writeFile(t, filepath.Join(root, "docs", "exists.md"), "ok\n")
	writeFile(t, filepath.Join(root, "node_modules", "ignored.md"), "[broken](nope.md)\n")
	writeFile(t, filepath.Join(filepath.Dir(root), "outside.md"), "secret\n")
	result, err := CheckLinks(root)
	if err != nil {
		t.Fatal(err)
	}
	joined := joinFailures(result.Failures)
	if !strings.Contains(joined, "missing local link") || !strings.Contains(joined, "escapes repository") {
		t.Fatalf("failures = %s", joined)
	}
	if strings.Contains(joined, "node_modules") {
		t.Fatal("node_modules was scanned")
	}
}

func TestCheckMediaDetectsOrphansDuplicatesAndDrift(t *testing.T) {
	root := t.TempDir()
	pngPath := filepath.Join(root, "docs", "media", "ok.png")
	writePNG(t, pngPath, 2, 2)
	info, err := os.Stat(pngPath)
	if err != nil {
		t.Fatal(err)
	}
	digest, err := fileSHA256(pngPath)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "README.md"), "docs/media/ok.png\n")
	writeFile(t, filepath.Join(root, "docs", "media", "manifest.json"), `{
		"schemaVersion": 2,
		"assets": [
			{
				"path": "docs/media/ok.png",
				"status": "current",
				"width": 2,
				"height": 2,
				"bytes": `+strconv.FormatInt(info.Size(), 10)+`,
				"maxBytes": 10000,
				"sha256": "`+digest+`",
				"origin": "project-authored-vector",
				"alt": "ok",
				"license": "Apache-2.0",
				"rightsStatus": "approved-for-public-repository",
				"rightsBasis": "test",
				"rightsReviewedAt": "2026-09-02",
				"reviewerRole": "test",
				"sourceRevision": "abc",
				"replacementHistory": [],
				"referencedBy": ["README.md"]
			},
			{
				"path": "docs/media/ok.png",
				"status": "current",
				"width": 2,
				"height": 2,
				"bytes": 1,
				"maxBytes": 1,
				"sha256": "deadbeef",
				"origin": "project-authored-vector",
				"alt": "dup",
				"license": "Apache-2.0",
				"rightsStatus": "approved-for-public-repository",
				"rightsBasis": "test",
				"rightsReviewedAt": "2026-09-02",
				"reviewerRole": "test",
				"sourceRevision": "abc",
				"replacementHistory": [],
				"referencedBy": ["README.md"]
			}
		]
	}`)
	writeFile(t, filepath.Join(root, "docs", "media", "orphan.bin"), "x")
	writeFile(t, filepath.Join(root, "brand", "source", "keep.txt"), "y")
	result, err := CheckMedia(root)
	if err != nil {
		t.Fatal(err)
	}
	joined := joinFailures(result.Failures)
	for _, want := range []string{"duplicate media manifest path", "orphan media file: docs/media/orphan.bin", "orphan media file: brand/source/keep.txt"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %q in %s", want, joined)
		}
	}
}

func TestAssetDimensions(t *testing.T) {
	dir := t.TempDir()
	pngPath := filepath.Join(dir, "a.png")
	writePNG(t, pngPath, 3, 5)
	width, height, err := assetDimensions(pngPath)
	if err != nil || width != 3 || height != 5 {
		t.Fatalf("png = (%d,%d,%v)", width, height, err)
	}
	svgPath := filepath.Join(dir, "a.svg")
	writeFile(t, svgPath, `<svg width="16" height="9" xmlns="http://www.w3.org/2000/svg"></svg>`)
	width, height, err = assetDimensions(svgPath)
	if err != nil || width != 16 || height != 9 {
		t.Fatalf("svg = (%d,%d,%v)", width, height, err)
	}
	webp := bytes.Repeat([]byte{0}, 30)
	copy(webp[0:4], "RIFF")
	copy(webp[8:12], "WEBP")
	copy(webp[12:16], "VP8X")
	binary.LittleEndian.PutUint32(webp[4:8], 22)
	binary.LittleEndian.PutUint32(webp[16:20], 10)
	webp[24] = 9
	webp[27] = 4
	webpPath := filepath.Join(dir, "a.webp")
	if err := os.WriteFile(webpPath, webp, 0o644); err != nil {
		t.Fatal(err)
	}
	width, height, err = assetDimensions(webpPath)
	if err != nil || width != 10 || height != 5 {
		t.Fatalf("webp = (%d,%d,%v)", width, height, err)
	}

	lossy := bytes.Repeat([]byte{0}, 30)
	copy(lossy[0:4], "RIFF")
	copy(lossy[8:12], "WEBP")
	copy(lossy[12:16], "VP8 ")
	binary.LittleEndian.PutUint32(lossy[4:8], 22)
	binary.LittleEndian.PutUint32(lossy[16:20], 10)
	lossy[23], lossy[24], lossy[25] = 0x9d, 0x01, 0x2a
	binary.LittleEndian.PutUint16(lossy[26:28], 7)
	binary.LittleEndian.PutUint16(lossy[28:30], 11)
	lossyPath := filepath.Join(dir, "lossy.webp")
	if err := os.WriteFile(lossyPath, lossy, 0o644); err != nil {
		t.Fatal(err)
	}
	width, height, err = assetDimensions(lossyPath)
	if err != nil || width != 7 || height != 11 {
		t.Fatalf("lossy webp = (%d,%d,%v)", width, height, err)
	}

	lossless := bytes.Repeat([]byte{0}, 30)
	copy(lossless[0:4], "RIFF")
	copy(lossless[8:12], "WEBP")
	copy(lossless[12:16], "VP8L")
	binary.LittleEndian.PutUint32(lossless[4:8], 22)
	binary.LittleEndian.PutUint32(lossless[16:20], 10)
	lossless[20] = 0x2f
	lossless[21] = 3
	losslessPath := filepath.Join(dir, "lossless.webp")
	if err := os.WriteFile(losslessPath, lossless, 0o644); err != nil {
		t.Fatal(err)
	}
	width, height, err = assetDimensions(losslessPath)
	if err != nil || width != 4 || height != 1 {
		t.Fatalf("lossless webp = (%d,%d,%v)", width, height, err)
	}

	if _, _, err := assetDimensions(filepath.Join(dir, "a.gif")); err == nil {
		t.Fatal("gif was accepted")
	}
	writeFile(t, filepath.Join(dir, "bad.svg"), `<svg></svg>`)
	if _, _, err := assetDimensions(filepath.Join(dir, "bad.svg")); err == nil {
		t.Fatal("svg without dimensions was accepted")
	}
}

func TestFindRootAndPathEscape(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "internal", "pkg")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/lab\n")
	found, err := FindRoot(nested)
	if err != nil {
		t.Fatal(err)
	}
	resolvedFound, _ := filepath.EvalSymlinks(found)
	resolvedRoot, _ := filepath.EvalSymlinks(root)
	if resolvedFound != resolvedRoot {
		t.Fatalf("FindRoot = %s, want %s", found, root)
	}
	if _, err := FindRoot(t.TempDir()); err == nil {
		t.Fatal("FindRoot accepted a non-module tree")
	}
	if _, err := containedPath(root, filepath.Join(filepath.Dir(root), "outside.txt")); err == nil {
		t.Fatal("escape was accepted")
	}
	if skipWalkDir("node_modules/x") != true || skipWalkDir("docs/a.md") != false {
		t.Fatal("skipWalkDir changed")
	}
}

func writePNG(t *testing.T, path string, width, height int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	if err := png.Encode(file, img); err != nil {
		t.Fatal(err)
	}
}
