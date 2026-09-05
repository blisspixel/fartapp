package repoquality

import (
	"bytes"
	"encoding/json"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestFuzzTargetsMatchEveryDeclaredTarget(t *testing.T) {
	root, err := FindRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("go", "list", "-json", "./...")
	command.Dir = root
	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	declared := map[fuzzTarget]bool{}
	decoder := json.NewDecoder(bytes.NewReader(output))
	for {
		var pkg struct {
			Dir          string
			TestGoFiles  []string
			XTestGoFiles []string
		}
		if err := decoder.Decode(&pkg); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			t.Fatal(err)
		}
		relative, err := filepath.Rel(root, pkg.Dir)
		if err != nil {
			t.Fatal(err)
		}
		packagePath := "."
		if relative != "." {
			packagePath += "/" + filepath.ToSlash(relative)
		}
		for _, name := range append(pkg.TestGoFiles, pkg.XTestGoFiles...) {
			file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(pkg.Dir, name), nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			for _, declaration := range file.Decls {
				function, ok := declaration.(*ast.FuncDecl)
				if ok && function.Recv == nil && strings.HasPrefix(function.Name.Name, "Fuzz") {
					declared[fuzzTarget{packagePath, function.Name.Name}] = true
				}
			}
		}
	}
	for _, target := range fuzzTargets {
		if !declared[target] {
			t.Errorf("fuzz runner contains an absent or duplicate target: %#v", target)
		}
		delete(declared, target)
	}
	for target := range declared {
		t.Errorf("fuzz runner omits declared target: %#v", target)
	}
}

func TestRunFuzzRejectsNonpositiveDurationBeforeStartingGo(t *testing.T) {
	if err := RunFuzz(t.TempDir(), 0, io.Discard, io.Discard); err == nil {
		t.Fatal("zero duration accepted")
	}
}
