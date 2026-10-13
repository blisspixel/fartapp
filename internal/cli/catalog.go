package cli

import (
	_ "embed"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

//go:embed parity.tsv
var parityTable string

const maximumDiagnosticBytes = 1024

type parityRow struct {
	home    string
	example string
}

func mustParity() map[string]parityRow {
	rows := make(map[string]parityRow)
	for _, line := range strings.Split(parityTable, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		command, rest, ok := strings.Cut(line, "\t")
		home, example, ok2 := strings.Cut(rest, "\t")
		if !ok || !ok2 || command == "" || example == "" {
			panic("invalid parity.tsv row: " + line)
		}
		switch home {
		case "go", "rust", "both":
		default:
			panic("invalid parity.tsv home: " + line)
		}
		if _, exists := rows[command]; exists {
			panic("duplicate parity.tsv command: " + command)
		}
		rows[command] = parityRow{home: home, example: example}
	}
	if len(rows) == 0 {
		panic("parity.tsv is empty")
	}
	return rows
}

var parityIndex = mustParity()

func rustCLIHint(command string) (string, bool) {
	row, ok := parityIndex[command]
	if !ok || row.home != "rust" {
		return "", false
	}
	return fmt.Sprintf("'%s' is available in the Rust CLI: %s\n", command, row.example), true
}

func looksNumeric(value string) bool {
	if value == "" {
		return false
	}
	digit := false
	for _, r := range value {
		if unicode.IsDigit(r) {
			digit = true
			continue
		}
		switch r {
		case '+', '-', '.', ' ', '\t':
		default:
			return false
		}
	}
	return digit
}

func boundDiagnostic(value string) string {
	if len(value) <= maximumDiagnosticBytes {
		return value
	}
	cut := maximumDiagnosticBytes
	for cut > 0 && !utf8.RuneStart(value[cut]) {
		cut--
	}
	return value[:cut] + "..."
}
