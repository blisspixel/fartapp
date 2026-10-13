package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

const completionHelp = `Print a shell completion script for fartapp.

Usage:
  fartapp completion bash
  fartapp completion zsh
  fartapp completion fish

The script completes top-level commands and their operations, including the
Rust CLI's play operations so a shared catalog stays visible. Source the
script from the shell's completion directory. It does not modify PATH.
`

var completionOperations = map[string][]string{
	"law":         {"list", "inspect"},
	"scenario":    {"validate"},
	"reservoir":   {"predict"},
	"restriction": {"predict", "history"},
	"walk":        {"predict", "simulate", "inspect", "explain", "branch", "certify", "witness", "reconstruct", "refine", "sonify", "new"},
	"evidence":    {"capture", "inspect", "verify", "replay", "reconstruct"},
	"assurance":   {"list", "inspect"},
	"play":        {"run", "replay", "reconstruct"},
	"completion":  {"bash", "zsh", "fish"},
	"quickplay":   {"1", "2", "3", "4", "5"},
}

func catalogCommands() []string {
	commands := make([]string, 0, len(parityIndex))
	for command := range parityIndex {
		commands = append(commands, command)
	}
	sort.Strings(commands)
	return commands
}

func runCompletion(args []string, stdout, stderr io.Writer) int {
	if isHelpRequest(args) {
		return writeText(stdout, stderr, completionHelp)
	}
	if len(args) != 1 {
		writeDiagnostic(stderr, "usage: fartapp completion bash|zsh|fish\n")
		return 1
	}
	script, err := completionScript(args[0])
	if err != nil {
		writeDiagnostic(stderr, "invalid completion: %v\n", err)
		return 1
	}
	return writeText(stdout, stderr, script)
}

func completionScript(shell string) (string, error) {
	commands := strings.Join(catalogCommands(), " ")
	switch shell {
	case "bash":
		return bashCompletion(commands), nil
	case "zsh":
		return zshCompletion(commands), nil
	case "fish":
		return fishCompletion(commands), nil
	default:
		return "", fmt.Errorf("shell must be bash, zsh, or fish")
	}
}

func bashCompletion(commands string) string {
	var output strings.Builder
	output.WriteString("# fartapp bash completion\n")
	output.WriteString("_fartapp_complete() {\n")
	output.WriteString("  local cur prev\n")
	output.WriteString("  cur=\"${COMP_WORDS[COMP_CWORD]}\"\n")
	output.WriteString("  prev=\"${COMP_WORDS[COMP_CWORD-1]}\"\n")
	output.WriteString("  case \"${prev}\" in\n")
	for _, command := range catalogCommands() {
		ops := completionOperations[command]
		if len(ops) == 0 {
			continue
		}
		fmt.Fprintf(&output, "    %s) COMPREPLY=( $(compgen -W \"%s\" -- \"${cur}\") ) ;;\n", command, strings.Join(ops, " "))
	}
	fmt.Fprintf(&output, "    *) COMPREPLY=( $(compgen -W \"%s\" -- \"${cur}\") ) ;;\n", commands)
	output.WriteString("  esac\n}\ncomplete -F _fartapp_complete fartapp\n")
	return output.String()
}

func zshCompletion(commands string) string {
	var output strings.Builder
	output.WriteString("#compdef fartapp\n")
	fmt.Fprintf(&output, "_fartapp_commands=(%s)\n", commands)
	for _, command := range catalogCommands() {
		ops := completionOperations[command]
		if len(ops) == 0 {
			continue
		}
		fmt.Fprintf(&output, "_fartapp_%s=(%s)\n", command, strings.Join(ops, " "))
	}
	output.WriteString("_fartapp() {\n  if (( CURRENT == 2 )); then\n    _describe command _fartapp_commands\n  elif (( CURRENT == 3 )); then\n    case ${words[2]} in\n")
	for _, command := range catalogCommands() {
		if len(completionOperations[command]) != 0 {
			fmt.Fprintf(&output, "      %s) _describe operation _fartapp_%s ;;\n", command, command)
		}
	}
	output.WriteString("      *) return 1 ;;\n    esac\n  else\n    return 1\n  fi\n}\n")
	output.WriteString("if [[ ${funcstack[1]} == _fartapp ]]; then\n  _fartapp \"$@\"\nelif (( $+functions[compdef] )); then\n  compdef _fartapp fartapp\nfi\n")
	return output.String()
}

func fishCompletion(commands string) string {
	var output strings.Builder
	fmt.Fprintf(&output, "complete -c fartapp -n __fish_use_subcommand -a %q\n", commands)
	for _, command := range catalogCommands() {
		ops := completionOperations[command]
		if len(ops) == 0 {
			continue
		}
		fmt.Fprintf(&output, "complete -c fartapp -n \"__fish_seen_subcommand_from %s\" -a %q\n", command, strings.Join(ops, " "))
	}
	return output.String()
}
