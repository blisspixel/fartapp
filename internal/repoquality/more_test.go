package repoquality

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestRunCoverageFlagErrorsAndJoinedValues(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cases := [][]string{
		{"coverage", "--aggregate"},
		{"coverage", "--aggregate", "x"},
		{"coverage", "--aggregate=x"},
		{"coverage", "--package"},
		{"coverage", "--package", "x"},
		{"coverage", "--package=x"},
		{"coverage", "--profile="},
		{"coverage", "--wat"},
		{"fuzz", "--time"},
		{"fuzz", "--time="},
		{"fuzz", "--nope"},
		{"repository", "extra"},
	}
	for _, args := range cases {
		stdout.Reset()
		stderr.Reset()
		if code := Run(args, &stdout, &stderr); code != 1 || stderr.Len() == 0 {
			t.Fatalf("%v = (%d, %q, %q)", args, code, stdout.String(), stderr.String())
		}
	}
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"repository", "--help"}, &stdout, &stderr); code != 0 {
		t.Fatalf("repository help = (%d, %q)", code, stderr.String())
	}
}

func TestRunFuzzTinyTargetAndFailure(t *testing.T) {
	root, err := FindRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	original := fuzzTargets
	t.Cleanup(func() { fuzzTargets = original })

	fuzzTargets = []fuzzTarget{{"./internal/strictjson", "FuzzInspect"}}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := RunFuzz(root, time.Millisecond, &stdout, &stderr); err != nil {
		t.Fatalf("tiny fuzz: %v\n%s%s", err, stdout.String(), stderr.String())
	}

	fuzzTargets = []fuzzTarget{{"./does-not-exist", "FuzzNope"}}
	if err := RunFuzz(root, time.Millisecond, &stdout, &stderr); err == nil {
		t.Fatal("missing package fuzz succeeded")
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(cwd) }()
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"fuzz", "--time=1ms"}, &stdout, &stderr); code != 1 {
		t.Fatalf("fuzz CLI = (%d, %q, %q)", code, stdout.String(), stderr.String())
	}
}

func TestCheckMediaProvenanceAndFieldFailures(t *testing.T) {
	root := t.TempDir()
	pngPath := filepath.Join(root, "docs", "media", "gen.png")
	writePNG(t, pngPath, 1, 1)
	info, err := os.Stat(pngPath)
	if err != nil {
		t.Fatal(err)
	}
	digest, err := fileSHA256(pngPath)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "README.md"), "docs/media/gen.png\n")
	writeFile(t, filepath.Join(root, "docs", "media", "manifest.json"), `{
		"schemaVersion": 1,
		"assets": [{
			"path": "docs/media/gen.png",
			"status": "draft",
			"width": 9,
			"height": 9,
			"bytes": 1,
			"maxBytes": 1,
			"sha256": "not-a-hash",
			"origin": "project-generated-concept",
			"alt": "",
			"license": "",
			"rightsStatus": "private",
			"rightsBasis": "",
			"rightsReviewedAt": "",
			"reviewerRole": "",
			"referencedBy": ["missing.md"]
		}]
	}`)
	result, err := CheckMedia(root)
	if err != nil {
		t.Fatal(err)
	}
	joined := joinFailures(result.Failures)
	for _, want := range []string{
		"unsupported media manifest schema",
		"dimension drift",
		"byte count drift",
		"byte budget exceeded",
		"SHA-256 drift",
		"invalid SHA-256 format",
		"missing alt text",
		"invalid status",
		"missing provenance",
		"not approved for public repository",
		"missing replacementHistory",
		"missing reference file",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %q in %s", want, joined)
		}
	}

	writeFile(t, filepath.Join(root, "docs", "media", "manifest.json"), `{
		"schemaVersion": 2,
		"assets": [{
			"path": "docs/media/gen.png",
			"status": "planned",
			"width": 1,
			"height": 1,
			"bytes": `+strconv.FormatInt(info.Size(), 10)+`,
			"maxBytes": 100000,
			"sha256": "`+digest+`",
			"origin": "project-generated-concept",
			"alt": "gen",
			"license": "Apache-2.0",
			"rightsStatus": "approved-for-public-repository",
			"rightsBasis": "test",
			"rightsReviewedAt": "2026-09-02",
			"reviewerRole": "test",
			"replacementHistory": [],
			"referencedBy": ["README.md"],
			"provenance": {
				"generator": "test",
				"model": "test",
				"generatedAt": "2026-09-02",
				"sanitizedPrompt": "none",
				"sourceAssetSha256": "zzzz",
				"postProcessing": "none",
				"inputAssets": [{
					"sha256": "nope",
					"role": "",
					"repositoryPathAtGeneration": ""
				}]
			}
		}]
	}`)
	result, err = CheckMedia(root)
	if err != nil {
		t.Fatal(err)
	}
	joined = joinFailures(result.Failures)
	for _, want := range []string{
		"invalid provenance source hash",
		"invalid provenance input hash",
		"missing provenance input role",
		"missing provenance input identity",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing provenance %q in %s", want, joined)
		}
	}
}

func TestWebPAndPNGHeaderFailures(t *testing.T) {
	dir := t.TempDir()
	pngPath := filepath.Join(dir, "bad.png")
	if err := os.WriteFile(pngPath, []byte("not png"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := assetDimensions(pngPath); err == nil {
		t.Fatal("invalid png accepted")
	}
	webpPath := filepath.Join(dir, "bad.webp")
	if err := os.WriteFile(webpPath, []byte("RIFF....NOPE"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := assetDimensions(webpPath); err == nil {
		t.Fatal("invalid webp accepted")
	}
	chunk := bytes.Repeat([]byte{0}, 30)
	copy(chunk[0:4], "RIFF")
	copy(chunk[8:12], "WEBP")
	copy(chunk[12:16], "NOPE")
	binary.LittleEndian.PutUint32(chunk[4:8], 22)
	binary.LittleEndian.PutUint32(chunk[16:20], 10)
	nopePath := filepath.Join(dir, "nope.webp")
	if err := os.WriteFile(nopePath, chunk, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := assetDimensions(nopePath); err == nil {
		t.Fatal("unknown webp chunk accepted")
	}
	lossy := bytes.Repeat([]byte{0}, 30)
	copy(lossy[0:4], "RIFF")
	copy(lossy[8:12], "WEBP")
	copy(lossy[12:16], "VP8 ")
	binary.LittleEndian.PutUint32(lossy[4:8], 22)
	binary.LittleEndian.PutUint32(lossy[16:20], 10)
	lossyPath := filepath.Join(dir, "badframe.webp")
	if err := os.WriteFile(lossyPath, lossy, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := assetDimensions(lossyPath); err == nil {
		t.Fatal("invalid VP8 frame accepted")
	}
	lossless := bytes.Repeat([]byte{0}, 30)
	copy(lossless[0:4], "RIFF")
	copy(lossless[8:12], "WEBP")
	copy(lossless[12:16], "VP8L")
	binary.LittleEndian.PutUint32(lossless[4:8], 22)
	binary.LittleEndian.PutUint32(lossless[16:20], 10)
	losslessPath := filepath.Join(dir, "badlossless.webp")
	if err := os.WriteFile(losslessPath, lossless, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := assetDimensions(losslessPath); err == nil {
		t.Fatal("invalid VP8L frame accepted")
	}
}
