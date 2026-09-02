package main

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
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			code := runWithInput(test.args, scenarioErrorReader{}, &stdout, &stderr)
			if code != 0 || stdout.String() != test.want || stderr.Len() != 0 {
				t.Fatalf("result = (%d, %q, %q), want (0, %q, empty)", code, stdout.String(), stderr.String(), test.want)
			}
		})
	}
}

func TestRootHelpAdvertisesOnlyCurrentSurface(t *testing.T) {
	normalized := strings.Join(strings.Fields(rootHelp), " ")
	for _, required := range []string{
		"permanent v0.6 legacy string oracle",
		"law list",
		"law inspect",
		"scenario validate",
		"reservoir predict",
		"no implicit Earth or other world",
		"current probe resolves declared law contexts and capability references only",
		"currently English presentations",
		"does not assert shared language or meaning",
	} {
		if !strings.Contains(normalized, required) {
			t.Errorf("root help omits %q", required)
		}
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
		{args: []string{"fartapp", "--help", "-h"}, want: "invalid help: --help may be specified only once\n"},
		{args: []string{"fartapp", "law", "--help", "-h"}, want: "invalid law help: --help may be specified only once\n"},
		{args: []string{"fartapp", "scenario", "-h", "--help"}, want: "invalid scenario help: --help may be specified only once\n"},
		{args: []string{"fartapp", "reservoir", "-h", "--help"}, want: "invalid reservoir help: --help may be specified only once\n"},
	}
	for _, test := range tests {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		if code := runWithInput(test.args, scenarioErrorReader{}, &stdout, &stderr); code != 1 ||
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
