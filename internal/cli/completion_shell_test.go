package cli

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestZshCompletionDispatch(t *testing.T) {
	zsh := os.Getenv("FARTAPP_ZSH")
	if zsh == "" {
		var err error
		zsh, err = exec.LookPath("zsh")
		if err != nil {
			t.Skip("zsh is not installed; the Linux quality job requires it")
		}
	}
	script := filepath.Join(t.TempDir(), "_fartapp")
	if err := os.WriteFile(script, []byte(zshCompletion(strings.Join(catalogCommands(), " "))), 0o600); err != nil {
		t.Fatal(err)
	}
	// Execute both loading routes in zsh. Capture what the native dispatch passes
	// to _describe rather than requiring a terminal line editor in the test suite.
	const harness = `
compdef() { [[ $1 == _fartapp && $2 == fartapp ]] && registered=1; }
_describe() { local array_name=$2; print -rl -- "${(@P)array_name}"; }
if [[ $2 == source ]]; then
  source "$1"
  [[ $registered == 1 ]] || exit 1
else
  fpath=("${1:h}" $fpath)
  autoload -Uz _fartapp
fi
CURRENT=$3
words=(fartapp "$4" '')
_fartapp || true
`
	for _, mode := range []string{"source", "autoload"} {
		for _, tc := range []struct {
			name    string
			current string
			command string
			want    []string
		}{
			{"commands", "2", "", catalogCommands()},
			{"walk", "3", "walk", []string{"predict", "simulate", "inspect", "explain", "branch", "certify", "witness", "reconstruct", "refine", "sonify", "new"}},
			{"play", "3", "play", []string{"run", "replay", "reconstruct"}},
			{"shells", "3", "completion", []string{"bash", "zsh", "fish"}},
			{"no_operations", "3", "version", []string{}},
			{"unknown", "3", "unknown", []string{}},
			{"later_argument", "4", "walk", []string{}},
		} {
			t.Run(mode+"/"+tc.name, func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, zsh, "-dfc", harness, "--", script, mode, tc.current, tc.command)
				output, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("zsh completion: %v: %s", err, output)
				}
				if got := strings.Fields(string(output)); !reflect.DeepEqual(got, tc.want) {
					t.Fatalf("completion = %q, want %q", got, tc.want)
				}
			})
		}
	}
}
