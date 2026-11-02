package repoquality

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClaudeGuidanceImportsAgents(t *testing.T) {
	root, err := FindRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	claude, err := os.ReadFile(filepath.Join(root, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(claude) != "@AGENTS.md\n" {
		t.Fatalf("CLAUDE.md = %q", claude)
	}
	agents, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(agents), "## Start tidy, stop tidy") {
		t.Fatal("AGENTS.md lost the start-tidy guidance")
	}
}
