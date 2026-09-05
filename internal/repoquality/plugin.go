package repoquality

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/blisspixel/fartapp/internal/strictjson"
)

const (
	pluginSchema      = "https://agent-plugins.org/schemas/1.0.0/plugin.schema.json"
	recipeSchema      = "fart.agent-cli-recipes/v0alpha1"
	pluginSkillPath   = "skills/fartapp-lab/SKILL.md"
	pluginRecipesPath = "skills/fartapp-lab/recipes.json"
	maxPluginBytes    = 65_536
)

// The project intentionally produces a smaller subset of Agent Plugins 1.0.0:
// string metadata, one skill, and no MCP configuration or client extensions.
// These offline producer checks are not a general plugin loader or YAML parser.
type pluginManifest struct {
	Schema      string `json:"$schema"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Repository  string `json:"repository"`
	License     string `json:"license"`
}

// CLIRecipe contains literal arguments for the local executable. InputArgument
// identifies an existing package fixture; it is never shell interpolation.
type CLIRecipe struct {
	ID            string               `json:"id"`
	Args          []string             `json:"args"`
	InputArgument *int                 `json:"input_argument,omitempty"`
	Expect        CLIRecipeExpectation `json:"expect"`
}

type CLIRecipeExpectation struct {
	ExitCode *int                       `json:"exit_code"`
	Equals   map[string]json.RawMessage `json:"equals"`
}

func CheckPlugin(root string) (CheckResult, error) {
	var manifest pluginManifest
	if err := decodePluginJSON(root, "plugin.json", &manifest); err != nil {
		return CheckResult{}, err
	}
	if manifest.Schema != pluginSchema || !validPluginName(manifest.Name, true) {
		return CheckResult{}, fmt.Errorf("plugin.json: invalid Agent Plugins 1.0.0 schema or name")
	}
	if manifest.Name != "fartapp" || strings.TrimSpace(manifest.Version) == "" ||
		strings.TrimSpace(manifest.Description) == "" ||
		manifest.Repository != "https://github.com/blisspixel/fartapp" || manifest.License != "Apache-2.0" {
		return CheckResult{}, fmt.Errorf("plugin.json: incomplete or incorrect project metadata")
	}
	if _, err := os.Lstat(filepath.Join(root, "mcp.json")); !os.IsNotExist(err) {
		return CheckResult{}, fmt.Errorf("skill-only package must not contain mcp.json")
	}
	if _, err := containedPath(root, filepath.Join(root, "skills")); err != nil {
		return CheckResult{}, err
	}
	entries, err := os.ReadDir(filepath.Join(root, "skills"))
	if err != nil {
		return CheckResult{}, err
	}
	if len(entries) != 1 || entries[0].Name() != "fartapp-lab" {
		return CheckResult{}, fmt.Errorf("skill-only package requires exactly skills/fartapp-lab")
	}
	skill, err := readPluginFile(root, pluginSkillPath)
	if err != nil {
		return CheckResult{}, err
	}
	if err := validatePluginSkill(skill); err != nil {
		return CheckResult{}, fmt.Errorf("%s: %w", pluginSkillPath, err)
	}
	recipes, err := ReadPluginRecipes(root)
	if err != nil {
		return CheckResult{}, err
	}
	return CheckResult{Report: fmt.Sprintf("portable agent plugin verified: Agent Plugins 1.0.0, 1 skill, %d CLI recipes\n", len(recipes))}, nil
}

func validPluginName(name string, periods bool) bool {
	if len(name) == 0 || len(name) > 64 || strings.Contains(name, "--") || strings.Contains(name, "..") {
		return false
	}
	for index, character := range name {
		if character >= 'a' && character <= 'z' || character >= '0' && character <= '9' {
			continue
		}
		if index == 0 || index == len(name)-1 || character != '-' && (!periods || character != '.') {
			return false
		}
	}
	return true
}

func readPluginFile(root, relative string) ([]byte, error) {
	path, err := repositoryPath(root, relative)
	if err != nil {
		return nil, err
	}
	if _, err := containedPath(root, path); err != nil {
		return nil, fmt.Errorf("%s: %w", relative, err)
	}
	data, err := readFileLimited(path, maxPluginBytes)
	if err != nil {
		return nil, err
	}
	if !utf8.Valid(data) {
		return nil, fmt.Errorf("%s: invalid UTF-8", relative)
	}
	return data, nil
}

func decodePluginJSON(root, relative string, destination any) error {
	data, err := readPluginFile(root, relative)
	if err != nil {
		return err
	}
	if issue := strictjson.Inspect(data, strictjson.Limits{MaximumDepth: 16, MaximumMemberNameBytes: 256}); issue != nil {
		return fmt.Errorf("%s: invalid JSON: %s at %s", relative, issue.Kind, issue.Path)
	}
	if _, manifest := destination.(*pluginManifest); manifest {
		if issue := strictjson.InspectShape[pluginManifest](data); issue != nil {
			return fmt.Errorf("%s: invalid manifest: %s at %s", relative, issue.Kind, issue.Path)
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("%s: %w", relative, err)
	}
	return nil
}

func validatePluginSkill(data []byte) error {
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	if !strings.HasPrefix(text, "---\n") {
		return fmt.Errorf("missing YAML frontmatter")
	}
	metadata, body, ok := strings.Cut(text[4:], "\n---\n")
	if !ok || strings.TrimSpace(body) == "" {
		return fmt.Errorf("missing frontmatter boundary or skill instructions")
	}
	limits := map[string]int{"name": 64, "description": 1024}
	fields := make(map[string]string)
	for _, line := range strings.Split(metadata, "\n") {
		key, raw, ok := strings.Cut(line, ": ")
		limit, known := limits[key]
		var value string
		if !ok || !known || json.Unmarshal([]byte(raw), &value) != nil ||
			strings.TrimSpace(value) == "" || utf8.RuneCountInString(value) > limit {
			return fmt.Errorf("frontmatter requires bounded JSON-quoted name and description strings")
		}
		if _, exists := fields[key]; exists {
			return fmt.Errorf("duplicate frontmatter field %s", key)
		}
		fields[key] = value
	}
	if fields["name"] != "fartapp-lab" || fields["description"] == "" {
		return fmt.Errorf("missing required project skill metadata or name does not match directory")
	}
	return nil
}

func ReadPluginRecipes(root string) ([]CLIRecipe, error) {
	var corpus struct {
		Schema           string      `json:"schema"`
		WorkingDirectory string      `json:"working_directory"`
		Recipes          []CLIRecipe `json:"recipes"`
	}
	if err := decodePluginJSON(root, pluginRecipesPath, &corpus); err != nil {
		return nil, err
	}
	if corpus.Schema != recipeSchema || corpus.WorkingDirectory != "plugin-root" || len(corpus.Recipes) == 0 || len(corpus.Recipes) > 32 {
		return nil, fmt.Errorf("invalid CLI recipe schema, working directory, or recipe count")
	}
	seen := make(map[string]bool)
	for _, recipe := range corpus.Recipes {
		if !validPluginName(recipe.ID, false) || seen[recipe.ID] {
			return nil, fmt.Errorf("invalid or duplicate CLI recipe id: %s", quote(recipe.ID))
		}
		seen[recipe.ID] = true
		if err := validateRecipe(root, recipe); err != nil {
			return nil, fmt.Errorf("CLI recipe %s: %w", recipe.ID, err)
		}
	}
	return corpus.Recipes, nil
}

func validateRecipe(root string, recipe CLIRecipe) error {
	args := recipe.Args
	if len(args) < 4 || len(args) > 9 || args[len(args)-2] != "--format" || args[len(args)-1] != "json" {
		return fmt.Errorf("expected literal CLI arguments ending in --format json")
	}
	operation := args[0] + " " + args[1]
	input := true
	switch operation {
	case "law list", "law inspect", "assurance inspect":
		input = false
	case "scenario validate", "reservoir predict", "restriction predict", "restriction history",
		"walk predict", "walk simulate", "walk inspect", "walk explain", "walk branch", "walk certify", "walk witness", "walk reconstruct", "walk refine":
	default:
		return fmt.Errorf("command is outside the implemented CLI recipe profile")
	}
	expectedArguments := 5
	if operation == "law list" {
		expectedArguments = 4
	} else if operation == "walk refine" {
		expectedArguments = 9
	}
	if len(args) != expectedArguments {
		return fmt.Errorf("incorrect command argument count")
	}
	if operation == "walk refine" {
		relative, relativeErr := strconv.ParseFloat(args[4], 64)
		budget, budgetErr := strconv.Atoi(args[6])
		if args[3] != "--relative-tolerance" || args[5] != "--max-evaluations" ||
			relativeErr != nil || math.IsNaN(relative) || relative < 1e-12 || relative > 0.1 ||
			budgetErr != nil || budget < 15 || budget > 1000000 {
			return fmt.Errorf("refinement recipe requires bounded literal tolerance and evaluation budget")
		}
	}
	if input {
		if recipe.InputArgument == nil || *recipe.InputArgument != 2 || !strings.HasPrefix(args[2], "testdata/") || !strings.HasSuffix(args[2], ".json") {
			return fmt.Errorf("input_argument must select a package testdata JSON fixture")
		}
		if _, err := readPluginFile(root, args[2]); err != nil {
			return err
		}
	} else if recipe.InputArgument != nil {
		return fmt.Errorf("catalog commands do not consume input files")
	}
	if recipe.Expect.ExitCode == nil || *recipe.Expect.ExitCode < 0 || *recipe.Expect.ExitCode > 1 || len(recipe.Expect.Equals) == 0 || len(recipe.Expect.Equals) > 32 {
		return fmt.Errorf("missing or invalid CLI result expectations")
	}
	var schema string
	if json.Unmarshal(recipe.Expect.Equals["/schema"], &schema) != nil || !strings.HasPrefix(schema, "fart.") {
		return fmt.Errorf("result expectations must identify a fart. report schema")
	}
	for pointer := range recipe.Expect.Equals {
		if len(pointer) > 256 || !strings.HasPrefix(pointer, "/") || strings.ContainsAny(pointer, "~\x00\r\n") {
			return fmt.Errorf("expectations require simple absolute JSON pointers")
		}
	}
	return nil
}
