package repoquality

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDependencyAuditMatchesCIAndLeavesRequiredChecks(t *testing.T) {
	root, err := FindRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	ci, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	audit, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "dependency-audit.yml"))
	if err != nil {
		t.Fatal(err)
	}
	ciText := string(ci)
	auditText := string(audit)
	for _, command := range []string{
		"npm audit --audit-level=low",
		"go run golang.org/x/vuln/cmd/govulncheck@v1.7.0 ./...",
		"cargo install --locked --version 0.20.2 cargo-deny",
		"cargo deny --locked check",
		"actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1",
		"actions/setup-node@949feb2413d6458794dcd2491c4babbbce0c15c1",
		"actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e",
	} {
		if !strings.Contains(ciText, command) || !strings.Contains(auditText, command) {
			t.Fatalf("command %q is not shared by CI and the scheduled audit", command)
		}
	}
	for _, piece := range []string{
		"workflow_dispatch:",
		"schedule:",
		`cron: "17 12 * * 1"`,
		"contents: read",
	} {
		if !strings.Contains(auditText, piece) {
			t.Fatalf("audit workflow missing %s", piece)
		}
	}
	if strings.Contains(auditText, "pull_request:") {
		t.Fatal("scheduled audit must stay off the required pull-request checks")
	}
	for _, name := range []string{
		"name: Build, lint, fuzz, race, and coverage",
		"name: Test on ${{ matrix.os }}",
		"name: Rust lint, dependencies, parity, and coverage",
		"name: Rust test on ${{ matrix.os }}",
	} {
		if !strings.Contains(ciText, name) {
			t.Fatalf("CI lost required context %s", name)
		}
		if strings.Contains(auditText, name) {
			t.Fatalf("audit workflow reuses required context %s", name)
		}
	}
}
