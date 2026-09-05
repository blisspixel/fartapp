package cli

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"syscall"
	"testing"
)

func TestParseOutputOptions(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		positional []string
		format     outputFormat
		help       bool
		wantError  string
	}{
		{name: "default", args: []string{"value"}, positional: []string{"value"}, format: outputText},
		{name: "separate json", args: []string{"--format", "json", "value"}, positional: []string{"value"}, format: outputJSON},
		{name: "joined text", args: []string{"value", "--format=text"}, positional: []string{"value"}, format: outputText},
		{name: "long help", args: []string{"--help"}, format: outputText, help: true},
		{name: "short help after value", args: []string{"value", "-h"}, positional: []string{"value"}, format: outputText, help: true},
		{name: "help with format", args: []string{"--format=json", "--help"}, format: outputJSON, help: true},
		{name: "stdin", args: []string{"-"}, positional: []string{"-"}, format: outputText},
		{name: "terminator", args: []string{"--", "--format=json"}, positional: []string{"--format=json"}, format: outputText},
		{name: "help after terminator", args: []string{"--", "--help"}, positional: []string{"--help"}, format: outputText},
		{name: "missing value", args: []string{"--format"}, wantError: "requires text or json"},
		{name: "help consumed as value", args: []string{"--format", "--help"}, wantError: "unsupported format"},
		{name: "empty joined value", args: []string{"--format="}, wantError: "unsupported format"},
		{name: "unsupported", args: []string{"--format", "yaml"}, wantError: "unsupported format"},
		{name: "duplicate mixed", args: []string{"--format=json", "--format", "text"}, wantError: "only once"},
		{name: "duplicate help", args: []string{"--help", "-h"}, wantError: "only once"},
		{name: "unknown option", args: []string{"--wat"}, wantError: "unknown option"},
		{name: "help with value", args: []string{"--help=1"}, wantError: "unknown option"},
		{name: "long single dash", args: []string{"-help"}, wantError: "unknown option"},
		{name: "question help", args: []string{"-?"}, wantError: "unknown option"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			options, err := parseOutputOptions(test.args)
			if test.wantError != "" {
				if err == nil || !bytes.Contains([]byte(err.Error()), []byte(test.wantError)) {
					t.Fatalf("error = %v, want containing %q", err, test.wantError)
				}
				return
			}
			if err != nil || fmt.Sprint(options.positional) != fmt.Sprint(test.positional) ||
				options.format != test.format || options.help != test.help {
				t.Fatalf("result = (%#v, %v), want positional %q format %v help %t", options, err, test.positional, test.format, test.help)
			}
		})
	}
}

type brokenPipeWriter struct {
	err error
}

func (writer brokenPipeWriter) Write([]byte) (int, error) {
	return 0, writer.err
}

func TestBrokenPipeIsQuiet(t *testing.T) {
	errorsToTest := []error{
		syscall.EPIPE,
		fmt.Errorf("wrapped: %w", syscall.EPIPE),
	}
	if runtime.GOOS == "windows" {
		errorsToTest = append(
			errorsToTest,
			syscall.Errno(109),
			fmt.Errorf("wrapped: %w", syscall.Errno(109)),
			syscall.Errno(232),
			fmt.Errorf("wrapped: %w", syscall.Errno(232)),
		)
	}
	for _, err := range errorsToTest {
		var stderr bytes.Buffer
		if code := writeText(brokenPipeWriter{err: err}, &stderr, "output\n"); code != 1 {
			t.Fatalf("exit code = %d", code)
		}
		if stderr.Len() != 0 {
			t.Fatalf("stderr = %q", stderr.String())
		}
	}
	if isBrokenPipe(fmt.Errorf("different")) {
		t.Fatal("ordinary error classified as broken pipe")
	}
}

func TestRealBrokenPipeIsQuiet(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	if err = reader.Close(); err != nil {
		_ = writer.Close()
		t.Fatalf("close reader: %v", err)
	}
	var stderr bytes.Buffer
	if code := writeText(writer, &stderr, "output\n"); code != 1 {
		_ = writer.Close()
		t.Fatalf("exit code = %d", code)
	}
	if err = writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
}
