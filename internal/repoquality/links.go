package repoquality

import (
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var markdownLink = regexp.MustCompile(`!?\[[^\]]*\]\((<[^>]+>|[^)\s]+)`)

func CheckLinks(root string) (CheckResult, error) {
	var failures []string
	checkedLinks := 0
	markdownFiles := 0
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := containedPath(root, path)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if skipWalkDir(relative) {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.EqualFold(filepath.Ext(path), ".md") || skipWalkDir(relative) {
			return nil
		}
		markdownFiles++
		data, err := readFileLimited(path, maxPolicyFileBytes)
		if err != nil {
			return err
		}
		for _, match := range markdownLink.FindAllSubmatch(data, -1) {
			target := strings.Trim(string(match[1]), "<>")
			lowerTarget := strings.ToLower(target)
			if strings.HasPrefix(lowerTarget, "http://") || strings.HasPrefix(lowerTarget, "https://") ||
				strings.HasPrefix(lowerTarget, "mailto:") || strings.HasPrefix(target, "#") {
				continue
			}
			pathPart, _, _ := strings.Cut(target, "#")
			pathPart, _, _ = strings.Cut(pathPart, "?")
			if strings.TrimSpace(pathPart) == "" {
				continue
			}
			decoded, err := url.PathUnescape(pathPart)
			if err != nil || strings.HasPrefix(decoded, "/") || strings.ContainsAny(decoded, "\\:\x00") {
				failures = append(failures, fmt.Sprintf("invalid local link in %s: %s", filepath.Base(path), target))
				continue
			}
			candidate := filepath.Join(filepath.Dir(path), filepath.FromSlash(decoded))
			checkedLinks++
			if _, err := containedPath(root, candidate); err != nil {
				failures = append(failures, fmt.Sprintf("local link escapes repository: %s -> %s", filepath.Base(path), target))
				continue
			}
			if _, err := os.Stat(candidate); err != nil {
				failures = append(failures, fmt.Sprintf("missing local link: %s -> %s", relative, target))
			}
		}
		return nil
	})
	if err != nil {
		return CheckResult{}, err
	}
	report := fmt.Sprintf("local Markdown links verified: %d links in %d files\n", checkedLinks, markdownFiles)
	return CheckResult{Report: report, Failures: failures}, nil
}
