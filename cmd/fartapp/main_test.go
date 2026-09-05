package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestProcessStreamsAndExitStatus(t *testing.T) {
	savedArgs, savedOut, savedErr, savedExit := os.Args, os.Stdout, os.Stderr, exitProcess
	t.Cleanup(func() { os.Args, os.Stdout, os.Stderr, exitProcess = savedArgs, savedOut, savedErr, savedExit })
	for _, test := range []struct {
		arg, stdout, diagnostic string
		code                    int
	}{
		{"3", "braaap (respectable)\n", "", 0},
		{"6", "", "invalid intensity", 1},
	} {
		t.Run(test.arg, func(t *testing.T) {
			out, err := os.CreateTemp(t.TempDir(), "stdout")
			if err != nil {
				t.Fatal(err)
			}
			defer out.Close()
			diagnostics, err := os.CreateTemp(t.TempDir(), "stderr")
			if err != nil {
				t.Fatal(err)
			}
			defer diagnostics.Close()
			os.Args, os.Stdout, os.Stderr = []string{"fartapp", test.arg}, out, diagnostics
			code := -1
			exitProcess = func(value int) { code = value }
			main()
			if _, err := out.Seek(0, io.SeekStart); err != nil {
				t.Fatal(err)
			}
			if _, err := diagnostics.Seek(0, io.SeekStart); err != nil {
				t.Fatal(err)
			}
			text, err := io.ReadAll(out)
			if err != nil {
				t.Fatal(err)
			}
			message, err := io.ReadAll(diagnostics)
			if err != nil {
				t.Fatal(err)
			}
			if code != test.code || string(text) != test.stdout ||
				(test.diagnostic == "" && len(message) != 0) ||
				(test.diagnostic != "" && !strings.Contains(string(message), test.diagnostic)) {
				t.Fatalf("process = (%d, %q, %q)", code, text, message)
			}
		})
	}
}
