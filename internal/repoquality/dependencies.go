package repoquality

import (
	"encoding/base64"
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

var allowedLicenses = map[string]struct{}{
	"Apache-2.0":   {},
	"BSD-2-Clause": {},
	"BSD-3-Clause": {},
	"ISC":          {},
	"MIT":          {},
	"Python-2.0":   {},
}

type npmLockfile struct {
	LockfileVersion int                   `json:"lockfileVersion"`
	Packages        map[string]npmPackage `json:"packages"`
}

type npmPackage struct {
	Dev              bool   `json:"dev"`
	License          string `json:"license"`
	Integrity        string `json:"integrity"`
	Resolved         string `json:"resolved"`
	HasInstallScript bool   `json:"hasInstallScript"`
}

func CheckDependencies(root string) (CheckResult, error) {
	var lock npmLockfile
	if err := readPolicyJSON(filepath.Join(root, "package-lock.json"), &lock); err != nil {
		return CheckResult{}, fmt.Errorf("invalid package-lock.json: %w", err)
	}
	var failures []string
	if lock.LockfileVersion != 3 {
		failures = append(failures, fmt.Sprintf("unsupported npm lockfile version: %d", lock.LockfileVersion))
	}
	if _, exists := lock.Packages[""]; !exists {
		failures = append(failures, "npm lockfile is missing the root package")
	}
	packageCount := 0
	licenseCounts := map[string]int{}
	packageNames := make([]string, 0, len(lock.Packages))
	for name := range lock.Packages {
		packageNames = append(packageNames, name)
	}
	sort.Strings(packageNames)
	for _, name := range packageNames {
		pkg := lock.Packages[name]
		if name == "" {
			continue
		}
		packageCount++
		if _, allowed := allowedLicenses[pkg.License]; !allowed {
			failures = append(failures, fmt.Sprintf("unreviewed npm license %s: %s", pkg.License, name))
		} else {
			licenseCounts[pkg.License]++
		}
		if !pkg.Dev {
			failures = append(failures, "npm package is not development-only: "+name)
		}
		if pkg.HasInstallScript {
			failures = append(failures, "npm install script is not permitted: "+name)
		}
		if !validNPMIntegrity(pkg.Integrity) {
			failures = append(failures, "missing or invalid npm integrity: "+name)
		}
		if !strings.HasPrefix(pkg.Resolved, "https://registry.npmjs.org/") {
			failures = append(failures, "unexpected npm source: "+name)
		}
	}

	modulePath, err := modulePathFromGoMod(root)
	if err != nil {
		return CheckResult{}, err
	}
	command := exec.Command("go", "list", "-m", "all")
	command.Dir = root
	output, err := command.Output()
	if err != nil {
		failures = append(failures, "go list -m all failed")
	} else {
		modules := nonEmptyLines(string(output))
		if len(modules) != 1 || modules[0] != modulePath {
			failures = append(failures, "external Go modules require a reviewed dependency and license manifest")
		}
	}

	names := make([]string, 0, len(licenseCounts))
	for name := range licenseCounts {
		names = append(names, name)
	}
	sort.Strings(names)
	summary := make([]string, 0, len(names))
	for _, name := range names {
		summary = append(summary, fmt.Sprintf("%s=%d", name, licenseCounts[name]))
	}
	report := fmt.Sprintf(
		"dependency policy verified: %d npm development packages (%s); no external Go modules\n",
		packageCount,
		strings.Join(summary, ", "),
	)
	return CheckResult{Report: report, Failures: failures}, nil
}

func validNPMIntegrity(value string) bool {
	encoded, ok := strings.CutPrefix(value, "sha512-")
	if !ok {
		return false
	}
	digest, err := base64.StdEncoding.Strict().DecodeString(encoded)
	return err == nil && len(digest) == 64 && base64.StdEncoding.EncodeToString(digest) == encoded
}

func nonEmptyLines(value string) []string {
	var lines []string
	for _, line := range strings.Split(value, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}
