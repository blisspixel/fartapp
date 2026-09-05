package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunDelegatesToRepoquality(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := run([]string{"help"}, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "repoquality") {
		t.Fatalf("help = (%d, %q, %q)", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := run([]string{"unknown"}, &stdout, &stderr); code != 1 || stdout.Len() != 0 || stderr.Len() == 0 {
		t.Fatalf("unknown = (%d, %q, %q)", code, stdout.String(), stderr.String())
	}
	if code := run([]string{"help"}, nil, nil); code != 0 {
		t.Fatalf("nil writers = %d", code)
	}
}
