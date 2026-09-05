package repoquality

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrNotRepository = errors.New("working directory is not inside the Go module")
	ErrPathEscape    = errors.New("path escapes the repository root")
)

func FindRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(dir)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		dir = filepath.Dir(dir)
	}
	for {
		if info, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && info.Mode().IsRegular() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrNotRepository
		}
		dir = parent
	}
}

func containedPath(root, candidate string) (string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	absCandidate, err := filepath.Abs(candidate)
	if err != nil {
		return "", err
	}
	resolvedRoot, err := filepath.EvalSymlinks(absRoot)
	if err != nil {
		return "", err
	}
	lexical, err := filepath.Rel(absRoot, absCandidate)
	if err != nil || lexical == ".." || strings.HasPrefix(lexical, ".."+string(os.PathSeparator)) {
		return "", ErrPathEscape
	}
	resolvedCandidate, err := resolveExistingParent(absCandidate)
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(resolvedRoot, resolvedCandidate)
	if err != nil {
		return "", err
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return "", ErrPathEscape
	}
	return filepath.ToSlash(relative), nil
}

// Missing leaves still inherit the containment of their nearest existing parent.
// A dangling or unreadable symlink is never treated as an ordinary path.
func resolveExistingParent(candidate string) (string, error) {
	var missing []string
	for {
		resolved, err := filepath.EvalSymlinks(candidate)
		if err == nil {
			for index := len(missing) - 1; index >= 0; index-- {
				resolved = filepath.Join(resolved, missing[index])
			}
			return resolved, nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		if _, statErr := os.Lstat(candidate); !os.IsNotExist(statErr) {
			return "", err
		}
		parent := filepath.Dir(candidate)
		if parent == candidate {
			return "", err
		}
		missing = append(missing, filepath.Base(candidate))
		candidate = parent
	}
}

// Manifest and coverage identities use canonical slash-separated relative paths
// on every operating system, including hosts where backslashes are legal names.
func repositoryPath(root, relative string) (string, error) {
	if relative == "." || !fs.ValidPath(relative) || strings.ContainsAny(relative, "\\:\x00") {
		return "", fmt.Errorf("invalid repository-relative path: %s", quote(relative))
	}
	return filepath.Join(root, filepath.FromSlash(relative)), nil
}

func skipWalkDir(relative string) bool {
	relative = filepath.ToSlash(relative)
	for _, prefix := range []string{"node_modules/", "vendor/", "target/", "artifacts/", ".git/", ".steward/DECISIONS/"} {
		if relative == strings.TrimSuffix(prefix, "/") || strings.HasPrefix(relative, prefix) {
			return true
		}
	}
	return false
}
