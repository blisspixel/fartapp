package repoquality

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type mediaManifest struct {
	SchemaVersion int          `json:"schemaVersion"`
	Assets        []mediaAsset `json:"assets"`
}

type mediaAsset struct {
	Path               string           `json:"path"`
	Status             string           `json:"status"`
	Width              int              `json:"width"`
	Height             int              `json:"height"`
	Bytes              int64            `json:"bytes"`
	MaxBytes           int64            `json:"maxBytes"`
	SHA256             string           `json:"sha256"`
	Origin             string           `json:"origin"`
	Alt                string           `json:"alt"`
	License            string           `json:"license"`
	RightsStatus       string           `json:"rightsStatus"`
	RightsBasis        string           `json:"rightsBasis"`
	RightsReviewedAt   string           `json:"rightsReviewedAt"`
	ReviewerRole       string           `json:"reviewerRole"`
	SourceRevision     string           `json:"sourceRevision"`
	ReplacementHistory json.RawMessage  `json:"replacementHistory"`
	ReferencedBy       []string         `json:"referencedBy"`
	Provenance         *mediaProvenance `json:"provenance"`
}

type mediaProvenance struct {
	Generator         string            `json:"generator"`
	Model             string            `json:"model"`
	GeneratedAt       string            `json:"generatedAt"`
	SanitizedPrompt   string            `json:"sanitizedPrompt"`
	SourceAssetSHA256 string            `json:"sourceAssetSha256"`
	PostProcessing    string            `json:"postProcessing"`
	InputAssets       []mediaInputAsset `json:"inputAssets"`
}

type mediaInputAsset struct {
	SHA256                     string `json:"sha256"`
	Role                       string `json:"role"`
	RepositoryPathAtGeneration string `json:"repositoryPathAtGeneration"`
}

var sha256Pattern = regexp.MustCompile(`^[0-9a-fA-F]{64}$`)

func CheckMedia(root string) (CheckResult, error) {
	mediaRoot := filepath.Join(root, "docs", "media")
	brandRoot := filepath.Join(root, "brand", "source")
	manifestPath := filepath.Join(mediaRoot, "manifest.json")
	if _, err := containedPath(root, manifestPath); err != nil {
		return CheckResult{}, err
	}
	var manifest mediaManifest
	if err := readPolicyJSON(manifestPath, &manifest); err != nil {
		return CheckResult{}, fmt.Errorf("invalid media manifest: %w", err)
	}
	var failures []string
	if manifest.SchemaVersion != 2 {
		failures = append(failures, fmt.Sprintf("unsupported media manifest schema: %d", manifest.SchemaVersion))
	}
	if manifest.Assets == nil {
		failures = append(failures, "media manifest is missing an assets array")
	}
	registered := map[string]struct{}{}
	for _, asset := range manifest.Assets {
		normalized := asset.Path
		absolute, err := repositoryPath(root, normalized)
		if err != nil {
			failures = append(failures, "invalid asset path: "+normalized)
			continue
		}
		if _, exists := registered[normalized]; exists {
			failures = append(failures, "duplicate media manifest path: "+normalized)
			continue
		}
		registered[normalized] = struct{}{}
		if _, err := containedPath(root, absolute); err != nil {
			failures = append(failures, "asset path escapes repository: "+normalized)
			continue
		}
		if _, err := containedPath(mediaRoot, absolute); err != nil {
			if _, brandErr := containedPath(brandRoot, absolute); brandErr != nil {
				failures = append(failures, "asset path escapes approved roots: "+normalized)
				continue
			}
		}
		info, err := os.Stat(absolute)
		if err != nil || !info.Mode().IsRegular() {
			failures = append(failures, "missing media file: "+normalized)
			continue
		}
		if info.Size() != asset.Bytes {
			failures = append(failures, fmt.Sprintf("byte count drift for %s: expected %d, got %d", normalized, asset.Bytes, info.Size()))
		}
		if info.Size() > asset.MaxBytes {
			failures = append(failures, fmt.Sprintf("byte budget exceeded for %s: %d > %d", normalized, info.Size(), asset.MaxBytes))
		}
		if info.Size() > maxMediaBytes {
			failures = append(failures, fmt.Sprintf("media exceeds checker limit of %d bytes: %s", maxMediaBytes, normalized))
			continue
		}
		if asset.Width <= 0 || asset.Height <= 0 || asset.Bytes <= 0 || asset.MaxBytes <= 0 {
			failures = append(failures, "media dimensions and byte budgets must be positive: "+normalized)
		}
		width, height, dimErr := assetDimensions(absolute)
		if dimErr != nil {
			failures = append(failures, dimErr.Error())
		} else if width != asset.Width || height != asset.Height {
			failures = append(failures, fmt.Sprintf("dimension drift for %s: expected %dx%d, got %dx%d", normalized, asset.Width, asset.Height, width, height))
		}
		digest, err := fileSHA256(absolute)
		if err != nil {
			return CheckResult{}, err
		}
		if !strings.EqualFold(digest, asset.SHA256) {
			failures = append(failures, "SHA-256 drift for "+normalized)
		}
		if !sha256Pattern.MatchString(asset.SHA256) {
			failures = append(failures, "invalid SHA-256 format for "+normalized)
		}
		if strings.TrimSpace(asset.Alt) == "" {
			failures = append(failures, "missing alt text for "+normalized)
		}
		if asset.Status != "current" && asset.Status != "planned" {
			failures = append(failures, "invalid status for "+normalized+": "+asset.Status)
		}
		for _, field := range []struct {
			name  string
			value string
		}{
			{"origin", asset.Origin},
			{"license", asset.License},
			{"rightsStatus", asset.RightsStatus},
			{"rightsBasis", asset.RightsBasis},
			{"rightsReviewedAt", asset.RightsReviewedAt},
			{"reviewerRole", asset.ReviewerRole},
		} {
			if strings.TrimSpace(field.value) == "" {
				failures = append(failures, "missing "+field.name+" for "+normalized)
			}
		}
		if asset.RightsStatus != "approved-for-public-repository" {
			failures = append(failures, "asset is not approved for public repository: "+normalized)
		}
		var replacementHistory []json.RawMessage
		if json.Unmarshal(asset.ReplacementHistory, &replacementHistory) != nil || replacementHistory == nil {
			failures = append(failures, "missing replacementHistory for "+normalized)
		}
		if asset.Origin == "project-generated-concept" {
			if asset.Provenance == nil {
				failures = append(failures, "missing provenance for "+normalized)
			} else {
				failures = append(failures, provenanceFailures(normalized, *asset.Provenance)...)
			}
		} else if strings.TrimSpace(asset.SourceRevision) == "" {
			failures = append(failures, "missing sourceRevision for "+normalized)
		}
		for _, reference := range asset.ReferencedBy {
			referencePath, err := repositoryPath(root, reference)
			if err != nil {
				failures = append(failures, fmt.Sprintf("invalid reference path %s for %s", reference, normalized))
				continue
			}
			if _, err := containedPath(root, referencePath); err != nil {
				failures = append(failures, fmt.Sprintf("reference escapes repository: %s for %s", reference, normalized))
				continue
			}
			if info, err := os.Stat(referencePath); err != nil || !info.Mode().IsRegular() {
				failures = append(failures, fmt.Sprintf("missing reference file %s for %s", reference, normalized))
				continue
			}
			text, err := readFileLimited(referencePath, maxPolicyFileBytes)
			if err != nil {
				return CheckResult{}, err
			}
			if !strings.Contains(string(text), normalized) {
				failures = append(failures, fmt.Sprintf("%s does not reference %s", reference, normalized))
			}
		}
	}

	for _, assetRoot := range []string{mediaRoot, brandRoot} {
		if _, err := os.Stat(assetRoot); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return CheckResult{}, err
		}
		err := filepath.WalkDir(assetRoot, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				return nil
			}
			if entry.Name() == "README.md" || entry.Name() == "manifest.json" {
				return nil
			}
			relative, err := containedPath(root, path)
			if err != nil {
				return err
			}
			if _, ok := registered[relative]; !ok {
				failures = append(failures, "orphan media file: "+relative)
			}
			return nil
		})
		if err != nil {
			return CheckResult{}, err
		}
	}
	report := fmt.Sprintf("media manifest verified: %d assets\n", len(registered))
	return CheckResult{Report: report, Failures: failures}, nil
}

func provenanceFailures(path string, provenance mediaProvenance) []string {
	var failures []string
	for _, field := range []struct {
		name  string
		value string
	}{
		{"provenance.generator", provenance.Generator},
		{"provenance.model", provenance.Model},
		{"provenance.generatedAt", provenance.GeneratedAt},
		{"provenance.sanitizedPrompt", provenance.SanitizedPrompt},
		{"provenance.sourceAssetSha256", provenance.SourceAssetSHA256},
		{"provenance.postProcessing", provenance.PostProcessing},
	} {
		if strings.TrimSpace(field.value) == "" {
			failures = append(failures, "missing "+field.name+" for "+path)
		}
	}
	if !sha256Pattern.MatchString(provenance.SourceAssetSHA256) {
		failures = append(failures, "invalid provenance source hash for "+path)
	}
	if provenance.InputAssets == nil {
		failures = append(failures, "missing provenance.inputAssets for "+path)
		return failures
	}
	for _, input := range provenance.InputAssets {
		if !sha256Pattern.MatchString(input.SHA256) {
			failures = append(failures, "invalid provenance input hash for "+path)
		}
		if strings.TrimSpace(input.Role) == "" {
			failures = append(failures, "missing provenance input role for "+path)
		}
		if strings.TrimSpace(input.RepositoryPathAtGeneration) == "" {
			failures = append(failures, "missing provenance input identity for "+path)
		}
	}
	return failures
}

func fileSHA256(path string) (string, error) {
	file, err := openRegularFile(path, maxMediaBytes)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	count, err := io.Copy(hash, io.LimitReader(file, maxMediaBytes+1))
	if err != nil {
		return "", err
	}
	if count > maxMediaBytes {
		return "", fmt.Errorf("media exceeds checker limit of %d bytes: %s", maxMediaBytes, path)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
