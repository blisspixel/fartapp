package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestHelpRoutes(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "root short", args: []string{"fartapp", "-h"}, want: rootHelp},
		{name: "root long", args: []string{"fartapp", "--help"}, want: rootHelp},
		{name: "root command", args: []string{"fartapp", "help"}, want: rootHelp},
		{name: "law topic", args: []string{"fartapp", "help", "law"}, want: lawHelp},
		{name: "law family", args: []string{"fartapp", "law", "help"}, want: lawHelp},
		{name: "law list topic", args: []string{"fartapp", "help", "law", "list"}, want: lawListHelp},
		{name: "law list leaf", args: []string{"fartapp", "law", "list", "--format", "json", "--help"}, want: lawListHelp},
		{name: "law inspect topic", args: []string{"fartapp", "help", "law", "inspect"}, want: lawInspectHelp},
		{name: "law inspect leaf", args: []string{"fartapp", "law", "inspect", "unresolved.context", "-h"}, want: lawInspectHelp},
		{name: "scenario topic", args: []string{"fartapp", "help", "scenario"}, want: scenarioHelp},
		{name: "scenario family", args: []string{"fartapp", "scenario", "help"}, want: scenarioHelp},
		{name: "scenario validate topic", args: []string{"fartapp", "help", "scenario", "validate"}, want: scenarioValidateHelp},
		{name: "scenario validate leaf", args: []string{"fartapp", "scenario", "validate", "unopened.json", "--format", "json", "--help"}, want: scenarioValidateHelp},
		{name: "reservoir topic", args: []string{"fartapp", "help", "reservoir"}, want: reservoirHelp},
		{name: "reservoir family", args: []string{"fartapp", "reservoir", "help"}, want: reservoirHelp},
		{name: "reservoir predict topic", args: []string{"fartapp", "help", "reservoir", "predict"}, want: reservoirPredictHelp},
		{name: "reservoir predict leaf", args: []string{"fartapp", "reservoir", "predict", "unopened.json", "--format", "json", "--help"}, want: reservoirPredictHelp},
		{name: "restriction topic", args: []string{"fartapp", "help", "restriction"}, want: restrictionHelp},
		{name: "restriction family", args: []string{"fartapp", "restriction", "help"}, want: restrictionHelp},
		{name: "restriction predict topic", args: []string{"fartapp", "help", "restriction", "predict"}, want: restrictionPredictHelp},
		{name: "restriction predict leaf", args: []string{"fartapp", "restriction", "predict", "unopened.json", "--format", "json", "--help"}, want: restrictionPredictHelp},
		{name: "restriction history topic", args: []string{"fartapp", "help", "restriction", "history"}, want: restrictionHistoryHelp},
		{name: "restriction history leaf", args: []string{"fartapp", "restriction", "history", "unopened.json", "--help"}, want: restrictionHistoryHelp},
		{name: "walk topic", args: []string{"fartapp", "help", "walk"}, want: walkHelp},
		{name: "walk family", args: []string{"fartapp", "walk", "help"}, want: walkHelp},
		{name: "walk simulate topic", args: []string{"fartapp", "help", "walk", "simulate"}, want: walkHelpForOperation("simulate")},
		{name: "walk simulate leaf", args: []string{"fartapp", "walk", "simulate", "unopened.json", "--help"}, want: walkHelpForOperation("simulate")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			code := Run(test.args, scenarioErrorReader{}, &stdout, &stderr)
			if code != 0 || stdout.String() != test.want || stderr.Len() != 0 {
				t.Fatalf("result = (%d, %q, %q), want (0, %q, empty)", code, stdout.String(), stderr.String(), test.want)
			}
		})
	}
}

func TestRootHelpAdvertisesOnlyCurrentSurface(t *testing.T) {
	normalized := strings.Join(strings.Fields(rootHelp), " ")
	for _, required := range []string{
		"permanent v0.6 toy output, levels 1 to 5",
		"law List contexts", "scenario Validate", "reservoir Predict",
		"restriction Predict", "walk Explore", "evidence Capture",
		"assurance Inspect declared", "help Show contextual",
		"fartapp help [command [operation]]",
		"Predictors require explicit SI inputs and a named model",
		"executes no mapping or realization",
		"Help and reports use English",
	} {
		if !strings.Contains(normalized, required) {
			t.Errorf("root help omits %q", required)
		}
	}
	if strings.Count(rootHelp, "\n") > 40 || strings.Contains(rootHelp, "<predict|") {
		t.Fatal("root help expanded beyond its compact command index")
	}
	for _, planned := range []string{
		"fartapp quick",
		"fartapp broadcast",
		"fartapp ask",
		"fartapp lab",
		"fartapp mcp",
		"fartapp update",
	} {
		if strings.Contains(normalized, planned) {
			t.Errorf("root help advertises planned command %q", planned)
		}
	}
}

func TestHelpRouteErrors(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{args: []string{"fartapp", "help", "unknown"}, want: "unknown help topic \"unknown\"; run 'fartapp help' for available topics\n"},
		{args: []string{"fartapp", "help", "law", "list", "extra"}, want: "unknown help topic \"law list extra\"; run 'fartapp help' for available topics\n"},
		{args: []string{"fartapp", "law", "help", "unknown"}, want: "usage: fartapp law help\n"},
		{args: []string{"fartapp", "scenario", "help", "validate", "extra"}, want: "usage: fartapp scenario help\n"},
		{args: []string{"fartapp", "reservoir", "help", "predict", "extra"}, want: "usage: fartapp reservoir help\n"},
		{args: []string{"fartapp", "restriction", "help", "predict", "extra"}, want: "usage: fartapp restriction help\n"},
		{args: []string{"fartapp", "walk", "help", "simulate"}, want: "usage: fartapp walk help\n"},
		{args: []string{"fartapp", "walk", "-h", "--help"}, want: "invalid walk help: --help may be specified only once\n"},
		{args: []string{"fartapp", "--help", "-h"}, want: "invalid help: --help may be specified only once\n"},
		{args: []string{"fartapp", "law", "--help", "-h"}, want: "invalid law help: --help may be specified only once\n"},
		{args: []string{"fartapp", "scenario", "-h", "--help"}, want: "invalid scenario help: --help may be specified only once\n"},
		{args: []string{"fartapp", "reservoir", "-h", "--help"}, want: "invalid reservoir help: --help may be specified only once\n"},
		{args: []string{"fartapp", "restriction", "-h", "--help"}, want: "invalid restriction help: --help may be specified only once\n"},
	}
	for _, test := range tests {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		if code := Run(test.args, scenarioErrorReader{}, &stdout, &stderr); code != 1 ||
			stdout.Len() != 0 || stderr.String() != test.want {
			t.Errorf("run(%q) = (%d, %q, %q), want (1, empty, %q)", test.args, code, stdout.String(), stderr.String(), test.want)
		}
	}

	longTopic := strings.Repeat("x", 200)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := run([]string{"fartapp", "help", longTopic, "ignored"}, &stdout, &stderr); code != 1 ||
		!strings.Contains(stderr.String(), `"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx..."`) ||
		strings.Contains(stderr.String(), "ignored") {
		t.Fatalf("bounded topic result = (%d, %q, %q)", code, stdout.String(), stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	hostile := "line\nbreak\x1b[31m"
	if code := run([]string{"fartapp", "help", hostile}, &stdout, &stderr); code != 1 ||
		bytes.ContainsRune(stderr.Bytes(), '\x1b') || strings.Count(stderr.String(), "\n") != 1 {
		t.Fatalf("hostile topic result = (%d, %q, %q)", code, stdout.String(), stderr.String())
	}
}

func TestHelpHelpersAndOutputFailure(t *testing.T) {
	if !isHelpRequest([]string{"-h"}) || !isHelpRequest([]string{"--help"}) ||
		isHelpRequest(nil) || isHelpRequest([]string{"value", "--help"}) {
		t.Fatal("isHelpRequest accepted the wrong grammar")
	}
	if !repeatedHelpRequest([]string{"-h", "--help"}) ||
		repeatedHelpRequest([]string{"--help", "value"}) {
		t.Fatal("repeatedHelpRequest accepted the wrong grammar")
	}
	if _, found := helpOutput([]string{"law\x00list"}); found {
		t.Fatal("embedded separator aliased a help route")
	}
	if got := joinBounded([]string{"law", "list"}); got != "law list" {
		t.Fatalf("joinBounded = %q", got)
	}

	var stderr bytes.Buffer
	if code := run([]string{"fartapp", "--help"}, failingWriter{}, &stderr); code != 1 ||
		!strings.HasPrefix(stderr.String(), "write output: ") {
		t.Fatalf("output failure = (%d, %q)", code, stderr.String())
	}
}

func TestWalkLeafHelpIsSpecificAndDoesNotReadInput(t *testing.T) {
	for _, operation := range []string{"predict", "simulate", "inspect", "explain", "branch", "certify", "witness", "reconstruct"} {
		t.Run(operation, func(t *testing.T) {
			var topic, leaf, stderr bytes.Buffer
			if code := Run([]string{"fartapp", "help", "walk", operation}, tripwireReader{}, &topic, &stderr); code != 0 {
				t.Fatalf("topic help status = %d", code)
			}
			if code := Run([]string{"fartapp", "walk", operation, "unopened.json", "--help"}, tripwireReader{}, &leaf, &stderr); code != 0 || stderr.Len() != 0 {
				t.Fatalf("leaf help = %d, %q", code, stderr.String())
			}
			if !bytes.Equal(topic.Bytes(), leaf.Bytes()) {
				t.Fatal("topic and leaf help differ")
			}
			for _, required := range []string{
				"fartapp walk " + operation + " <case.json|->", "65,536 bytes",
				"standard input", "--format json", "without reading input",
				"testdata/walk/ordinary-low-pressure.json", "Exit status:",
			} {
				if !strings.Contains(leaf.String(), required) {
					t.Errorf("help omits %q", required)
				}
			}
			if strings.Contains(leaf.String(), "<predict|") || strings.Count(leaf.String(), "\n") > 45 {
				t.Fatal("leaf help repeated the command index or unrelated operation guidance")
			}
		})
	}
	if !strings.Contains(walkHelpForOperation("branch"), "1e-6 with 2e-6 m^2") ||
		!strings.Contains(walkHelpForOperation("reconstruct"), "set expected_witness") ||
		!strings.Contains(walkHelpForOperation("simulate"), "budgets may truncate") ||
		strings.Contains(walkHelpForOperation("inspect"), "Set branch.prescribed_area_m2") {
		t.Fatal("operation guidance lost an evidence boundary or leaked unrelated setup")
	}
	if walkHelpForOperation("untrusted\noperation") != "" {
		t.Fatal("unknown operation produced help")
	}
}

func TestScientificLeafHelpProvidesCompleteStartingContract(t *testing.T) {
	for _, args := range [][]string{
		{"reservoir", "predict"}, {"restriction", "predict"}, {"restriction", "history"}, {"walk", "refine"},
	} {
		var stdout, stderr bytes.Buffer
		if code := Run(append([]string{"fartapp", "help"}, args...), tripwireReader{}, &stdout, &stderr); code != 0 || stderr.Len() != 0 {
			t.Fatalf("help %q = %d, %q", args, code, stderr.String())
		}
		for _, required := range []string{"Usage:", "Arguments:", "standard input", "65,536 bytes", "--format", "Example:", "testdata/", "Exit status:"} {
			if !strings.Contains(stdout.String(), required) {
				t.Errorf("help %q omits %q", args, required)
			}
		}
	}
}
