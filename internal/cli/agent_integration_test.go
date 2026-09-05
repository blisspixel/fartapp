package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/blisspixel/fartapp/internal/repoquality"
)

func TestAgentCLIRecipes(t *testing.T) {
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	recipes, err := repoquality.ReadPluginRecipes(root)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "fartapp")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	buildContext, cancelBuild := context.WithTimeout(t.Context(), time.Minute)
	defer cancelBuild()
	build := exec.CommandContext(buildContext, "go", "build", "-o", binary, "./cmd/fartapp")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build recipe executable: %v\n%s", err, output)
	}
	for _, recipe := range recipes {
		t.Run(recipe.ID, func(t *testing.T) {
			var fixture string
			var input []byte
			if recipe.InputArgument != nil {
				fixture = filepath.FromSlash(recipe.Args[*recipe.InputArgument])
				input, err = os.ReadFile(fixture)
				if err != nil {
					t.Fatal(err)
				}
			}
			stdout, document, code := runAgentRecipe(t, binary, recipe.Args, nil)
			if code != *recipe.Expect.ExitCode {
				t.Fatalf("exit = %d, want %d; %s", code, *recipe.Expect.ExitCode, stdout)
			}
			for pointer, expected := range recipe.Expect.Equals {
				var want any
				decoder := json.NewDecoder(bytes.NewReader(expected))
				decoder.UseNumber()
				if err := decoder.Decode(&want); err != nil {
					t.Fatal(err)
				}
				if got := agentRecipeValue(t, document, pointer); !reflect.DeepEqual(got, want) {
					t.Errorf("%s = %v, want %v", pointer, got, want)
				}
			}
			if recipe.InputArgument == nil {
				return
			}
			inputIndex := *recipe.InputArgument
			args := append([]string(nil), recipe.Args...)
			args[inputIndex] = "-"
			streamed, _, streamedCode := runAgentRecipe(t, binary, args, input)
			if streamedCode != code || !bytes.Equal(streamed, stdout) {
				t.Fatalf("file and stdin results differ: file status %d, stdin status %d\nfile %s\nstdin %s", code, streamedCode, stdout, streamed)
			}
			after, err := os.ReadFile(fixture)
			if err != nil || !bytes.Equal(after, input) {
				t.Fatalf("recipe changed its input fixture: %v", err)
			}
		})
	}
	t.Run("malformed-input-retains-structured-refusal", func(t *testing.T) {
		_, document, code := runAgentRecipe(t, binary, []string{"scenario", "validate", "-", "--format", "json"}, []byte(`{"schema":`))
		if code != 1 || agentRecipeValue(t, document, "/document_status") != "invalid" ||
			agentRecipeValue(t, document, "/validation_stages/syntax/status") != "invalid" {
			t.Fatalf("malformed input was not refused: status %d, %v", code, document)
		}
	})
	t.Run("oversized-input-retains-structured-refusal", func(t *testing.T) {
		_, document, code := runAgentRecipe(t, binary, []string{"walk", "predict", "-", "--format", "json"}, bytes.Repeat([]byte(" "), 65_537))
		if code != 1 || agentRecipeValue(t, document, "/status") != "invalid" ||
			agentRecipeValue(t, document, "/diagnostics/0/reason_code") != "input_too_large" {
			t.Fatalf("oversized input was not refused: status %d, %v", code, document)
		}
	})
}

func runAgentRecipe(t *testing.T, binary string, args []string, input []byte) ([]byte, any, int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary, args...)
	command.Stdin = bytes.NewReader(input)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	code := 0
	if err != nil {
		var failure *exec.ExitError
		if !errors.As(err, &failure) || ctx.Err() != nil {
			t.Fatalf("run recipe %v: %v", args, err)
		}
		code = failure.ExitCode()
	}
	if stderr.Len() != 0 {
		t.Fatalf("structured recipe wrote stderr: %q", stderr.String())
	}
	decoder := json.NewDecoder(bytes.NewReader(stdout.Bytes()))
	decoder.UseNumber()
	var document any
	if err := decoder.Decode(&document); err != nil {
		t.Fatalf("recipe stdout is not JSON: %v; %q", err, stdout.String())
	}
	if err := decoder.Decode(new(any)); err != io.EOF || !bytes.HasSuffix(stdout.Bytes(), []byte("\n")) {
		t.Fatalf("recipe stdout must contain one newline-terminated JSON document: %q", stdout.String())
	}
	return stdout.Bytes(), document, code
}

func agentRecipeValue(t *testing.T, document any, pointer string) any {
	t.Helper()
	value := document
	for _, component := range strings.Split(strings.TrimPrefix(pointer, "/"), "/") {
		switch current := value.(type) {
		case map[string]any:
			var exists bool
			value, exists = current[component]
			if !exists {
				t.Fatalf("missing result member at %s", pointer)
			}
		case []any:
			index, err := strconv.Atoi(component)
			if err != nil || index < 0 || index >= len(current) {
				t.Fatalf("missing result array element at %s", pointer)
			}
			value = current[index]
		default:
			t.Fatalf("result path descends through a scalar at %s", pointer)
		}
	}
	return value
}
