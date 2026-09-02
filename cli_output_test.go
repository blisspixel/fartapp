package main

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"syscall"
	"testing"
)

func TestParseOutputFormat(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		positional []string
		format     outputFormat
		wantError  string
	}{
		{name: "default", args: []string{"value"}, positional: []string{"value"}, format: outputText},
		{name: "separate json", args: []string{"--format", "json", "value"}, positional: []string{"value"}, format: outputJSON},
		{name: "joined text", args: []string{"value", "--format=text"}, positional: []string{"value"}, format: outputText},
		{name: "stdin", args: []string{"-"}, positional: []string{"-"}, format: outputText},
		{name: "terminator", args: []string{"--", "--format=json"}, positional: []string{"--format=json"}, format: outputText},
		{name: "missing value", args: []string{"--format"}, wantError: "requires text or json"},
		{name: "empty joined value", args: []string{"--format="}, wantError: "unsupported format"},
		{name: "unsupported", args: []string{"--format", "yaml"}, wantError: "unsupported format"},
		{name: "duplicate mixed", args: []string{"--format=json", "--format", "text"}, wantError: "only once"},
		{name: "unknown option", args: []string{"--wat"}, wantError: "unknown option"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			positional, format, err := parseOutputFormat(test.args)
			if test.wantError != "" {
				if err == nil || !bytes.Contains([]byte(err.Error()), []byte(test.wantError)) {
					t.Fatalf("error = %v, want containing %q", err, test.wantError)
				}
				return
			}
			if err != nil || fmt.Sprint(positional) != fmt.Sprint(test.positional) || format != test.format {
				t.Fatalf("result = (%q, %v, %v), want (%q, %v, nil)", positional, format, err, test.positional, test.format)
			}
		})
	}
}

func TestHelpOptionHonorsTerminator(t *testing.T) {
	if !hasHelpOption([]string{"value", "--help"}) || !hasHelpOption([]string{"-h"}) {
		t.Fatal("help option not detected")
	}
	if hasHelpOption([]string{"--", "--help"}) || hasHelpOption(nil) {
		t.Fatal("non-option help was detected")
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
