package repoquality

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testPluginManifest = `{
  "$schema": "https://agent-plugins.org/schemas/1.0.0/plugin.schema.json",
  "name": "fartapp",
  "version": "0.7.0-alpha.1",
  "description": "Inspect scientific CLI reports.",
  "repository": "https://github.com/blisspixel/fartapp",
  "license": "Apache-2.0"
}`

const testPluginSkill = "---\nname: \"fartapp-lab\"\ndescription: \"Inspect scientific CLI reports.\"\n---\nUse the CLI.\n"

const testPluginRecipes = `{
  "schema": "fart.agent-cli-recipes/v0alpha1",
  "working_directory": "plugin-root",
  "recipes": [{
    "id": "scenario-validation",
    "args": ["scenario", "validate", "testdata/case.json", "--format", "json"],
    "input_argument": 2,
    "expect": {"exit_code": 0, "equals": {"/schema": "fart.scenario-validation/v0alpha2"}}
  }]
}`

func writePluginFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "plugin.json"), testPluginManifest)
	writeFile(t, filepath.Join(root, filepath.FromSlash(pluginSkillPath)), testPluginSkill)
	writeFile(t, filepath.Join(root, filepath.FromSlash(pluginRecipesPath)), testPluginRecipes)
	writeFile(t, filepath.Join(root, "testdata", "case.json"), "{}\n")
	return root
}

func TestPluginOfflineProducerContract(t *testing.T) {
	root := writePluginFixture(t)
	result, err := CheckPlugin(root)
	if err != nil || len(result.Failures) != 0 || !strings.Contains(result.Report, "1 skill, 1 CLI recipes") {
		t.Fatalf("valid package: %+v, %v", result, err)
	}
	root, err = FindRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := CheckPlugin(root); err != nil {
		t.Fatalf("project package: %v", err)
	}
}

func TestPluginRejectsInvalidManifest(t *testing.T) {
	cases := map[string]string{
		"draft-schema":     strings.Replace(testPluginManifest, "/1.0.0/", "/1.1.0/", 1),
		"unknown-field":    strings.Replace(testPluginManifest, `"name":`, `"skills": [], "name":`, 1),
		"author-metadata":  strings.Replace(testPluginManifest, `"name":`, `"author": {}, "name":`, 1),
		"case-alias":       strings.Replace(testPluginManifest, `"name":`, `"Name":`, 1),
		"duplicate-member": strings.Replace(testPluginManifest, `"name":`, `"name": "other", "name":`, 1),
		"null-name":        strings.Replace(testPluginManifest, `"name": "fartapp"`, `"name": null`, 1),
		"null-description": strings.Replace(testPluginManifest, `"description": "Inspect scientific CLI reports."`, `"description": null`, 1),
		"wrong-name":       strings.Replace(testPluginManifest, `"name": "fartapp"`, `"name": "not-fartapp"`, 1),
		"invalid-name":     strings.Replace(testPluginManifest, `"name": "fartapp"`, `"name": "Fartapp"`, 1),
		"missing-version":  strings.Replace(testPluginManifest, `"version": "0.7.0-alpha.1",`, "", 1),
		"wrong-source":     strings.Replace(testPluginManifest, "https://github.com/blisspixel/fartapp", "https://example.invalid", 1),
		"number-version":   strings.Replace(testPluginManifest, `"version": "0.7.0-alpha.1"`, `"version": 7`, 1),
		"invalid-json":     "{",
		"trailing-json":    testPluginManifest + " {}",
		"null-root":        "null",
		"invalid-utf8":     "\xff" + testPluginManifest,
		"oversized":        strings.Repeat(" ", maxPluginBytes) + testPluginManifest,
	}
	for name, manifest := range cases {
		t.Run(name, func(t *testing.T) {
			root := writePluginFixture(t)
			writeFile(t, filepath.Join(root, "plugin.json"), manifest)
			if _, err := CheckPlugin(root); err == nil {
				t.Fatal("invalid manifest was accepted")
			}
		})
	}
}

func TestPluginNameConstraints(t *testing.T) {
	for _, name := range []string{"a", "fartapp", "acme.tools", "a-9.b", strings.Repeat("a", 64)} {
		if !validPluginName(name, true) {
			t.Errorf("valid plugin name refused: %q", name)
		}
	}
	for _, name := range []string{"", "A", "_a", "a_", "-a", "a-", ".a", "a.", "a--b", "a..b", "a/b", "a b", "f\u00e5rt", strings.Repeat("a", 65)} {
		if validPluginName(name, true) {
			t.Errorf("invalid plugin name accepted: %q", name)
		}
	}
	if validPluginName("a.b", false) || !validPluginName("a-b", false) {
		t.Fatal("skill/recipe names must exclude periods")
	}
}

func TestPluginSkillFrontmatterProfile(t *testing.T) {
	if err := validatePluginSkill([]byte(strings.ReplaceAll(testPluginSkill, "\n", "\r\n"))); err != nil {
		t.Fatalf("CRLF skill: %v", err)
	}
	cases := []string{
		"No frontmatter.",
		"---\nname: \"fartapp-lab\"\n",
		strings.Replace(testPluginSkill, "Use the CLI.", "", 1),
		strings.Replace(testPluginSkill, "name: \"fartapp-lab\"", "name: fartapp-lab", 1),
		strings.Replace(testPluginSkill, "name: \"fartapp-lab\"", "name: null", 1),
		strings.Replace(testPluginSkill, "name: \"fartapp-lab\"", "name: \"another-skill\"", 1),
		strings.Replace(testPluginSkill, "name: \"fartapp-lab\"", "Name: \"fartapp-lab\"", 1),
		strings.Replace(testPluginSkill, "name: \"fartapp-lab\"", "name: \"fartapp-lab\"\nname: \"fartapp-lab\"", 1),
		strings.Replace(testPluginSkill, "description: \"Inspect scientific CLI reports.\"\n", "", 1),
		strings.Replace(testPluginSkill, "Inspect scientific CLI reports.", strings.Repeat("x", 1025), 1),
	}
	for index, data := range cases {
		if err := validatePluginSkill([]byte(data)); err == nil {
			t.Errorf("invalid skill %d was accepted", index)
		}
	}
}

func TestPluginRejectsUnshippedComponentsAndMissingFiles(t *testing.T) {
	for _, path := range []string{"plugin.json", pluginSkillPath, pluginRecipesPath, "testdata/case.json"} {
		t.Run("missing-"+path, func(t *testing.T) {
			root := writePluginFixture(t)
			if err := os.Remove(filepath.Join(root, filepath.FromSlash(path))); err != nil {
				t.Fatal(err)
			}
			if _, err := CheckPlugin(root); err == nil {
				t.Fatal("missing package file was accepted")
			}
		})
	}
	for _, path := range []string{"mcp.json", "skills/another/SKILL.md"} {
		t.Run("unshipped-"+path, func(t *testing.T) {
			root := writePluginFixture(t)
			writeFile(t, filepath.Join(root, filepath.FromSlash(path)), "{}")
			if _, err := CheckPlugin(root); err == nil {
				t.Fatal("component outside this package profile was accepted")
			}
		})
	}
}

func TestPluginRejectsEscapingSymlinks(t *testing.T) {
	for _, relative := range []string{"plugin.json", pluginSkillPath, pluginRecipesPath, "testdata/case.json"} {
		t.Run(relative, func(t *testing.T) {
			root := writePluginFixture(t)
			path := filepath.Join(root, filepath.FromSlash(relative))
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			outside := filepath.Join(t.TempDir(), "outside")
			writeFile(t, outside, string(data))
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(outside, path); err != nil {
				t.Skipf("host does not permit symlinks: %v", err)
			}
			if _, err := CheckPlugin(root); err == nil {
				t.Fatal("package followed a symlink outside its root")
			}
		})
	}
}

func TestPluginRecipeValidation(t *testing.T) {
	cases := map[string]func(map[string]any, map[string]any){
		"wrong-schema":            func(c, r map[string]any) { c["schema"] = "unknown" },
		"wrong-working-directory": func(c, r map[string]any) { c["working_directory"] = ".." },
		"no-recipes":              func(c, r map[string]any) { c["recipes"] = []any{} },
		"duplicate-id":            func(c, r map[string]any) { c["recipes"] = []any{r, r} },
		"invalid-id":              func(c, r map[string]any) { r["id"] = "a..b" },
		"unknown-member":          func(c, r map[string]any) { r["command"] = "sh" },
		"missing-arguments":       func(c, r map[string]any) { r["args"] = nil },
		"not-json":                func(c, r map[string]any) { r["args"].([]any)[4] = "text" },
		"unimplemented-operation": func(c, r map[string]any) { r["args"].([]any)[0] = "play" },
		"wrong-arity":             func(c, r map[string]any) { r["args"] = []any{"scenario", "validate", "--format", "json"} },
		"missing-input-argument":  func(c, r map[string]any) { delete(r, "input_argument") },
		"wrong-input-argument":    func(c, r map[string]any) { r["input_argument"] = 0 },
		"escaped-fixture":         func(c, r map[string]any) { r["args"].([]any)[2] = "testdata/../outside.json" },
		"backslash-fixture":       func(c, r map[string]any) { r["args"].([]any)[2] = "testdata/x\\y.json" },
		"nonfixture-input":        func(c, r map[string]any) { r["args"].([]any)[2] = "package.json" },
		"catalog-with-input":      func(c, r map[string]any) { r["args"] = []any{"law", "list", "--format", "json"} },
		"missing-expectation":     func(c, r map[string]any) { delete(r, "expect") },
		"unsupported-exit-code":   func(c, r map[string]any) { r["expect"].(map[string]any)["exit_code"] = 2 },
		"missing-result-schema":   func(c, r map[string]any) { r["expect"].(map[string]any)["equals"] = map[string]any{"/status": "valid"} },
		"invalid-pointer":         func(c, r map[string]any) { r["expect"].(map[string]any)["equals"].(map[string]any)["status"] = "valid" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			root := writePluginFixture(t)
			var corpus map[string]any
			if err := json.Unmarshal([]byte(testPluginRecipes), &corpus); err != nil {
				t.Fatal(err)
			}
			recipe := corpus["recipes"].([]any)[0].(map[string]any)
			mutate(corpus, recipe)
			data, err := json.Marshal(corpus)
			if err != nil {
				t.Fatal(err)
			}
			writeFile(t, filepath.Join(root, filepath.FromSlash(pluginRecipesPath)), string(data))
			if _, err := ReadPluginRecipes(root); err == nil {
				t.Fatal("invalid recipe corpus was accepted")
			}
		})
	}
}
