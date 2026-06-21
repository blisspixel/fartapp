package main

import (
	"bytes"
	"testing"
)

func TestPick(t *testing.T) {
	cases := []struct {
		intensity int
		want      string
	}{
		{1, "pfft"},
		{2, "toot"},
		{3, "braaap"},
		{4, "blorp"},
		{5, "KABLAM"},
	}
	for _, tc := range cases {
		if got := Pick(tc.intensity); got != tc.want {
			t.Errorf("Pick(%d) = %q, want %q", tc.intensity, got, tc.want)
		}
	}
}

func TestRate(t *testing.T) {
	cases := []struct {
		intensity int
		want      string
	}{
		{1, "gentle"},
		{2, "respectable"},
		{3, "respectable"},
		{4, "respectable"},
		{5, "mighty"},
	}
	for _, tc := range cases {
		if got := Rate(tc.intensity); got != tc.want {
			t.Errorf("Rate(%d) = %q, want %q", tc.intensity, got, tc.want)
		}
	}
}

func TestCLIInvalidInput(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"missing intensity", []string{"fartapp"}},
		{"non-integer", []string{"fartapp", "nope"}},
		{"below range", []string{"fartapp", "0"}},
		{"above range", []string{"fartapp", "6"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer
			if code := run(tt.args, &bytes.Buffer{}, &stderr); code != 1 {
				t.Fatalf("exit code = %d, want 1", code)
			}
			if stderr.Len() == 0 {
				t.Fatal("expected error message on stderr")
			}
		})
	}
}

func TestCLISuccess(t *testing.T) {
	var stdout bytes.Buffer
	if code := run([]string{"fartapp", "3"}, &stdout, &bytes.Buffer{}); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	want := "braaap (respectable)\n"
	if got := stdout.String(); got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}
