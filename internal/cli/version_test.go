package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestFormatVersion(t *testing.T) {
	const sha = "0123456789abcdef0123456789abcdef01234567"
	tests := []struct {
		name, version, commit, module, revision, want string
	}{
		{
			name:     "link stamps win over build info",
			version:  "v1.2.3",
			commit:   sha,
			module:   "v9.9.9",
			revision: "ffffffffffffffffffffffffffffffffffffffff",
			want:     "fartapp v1.2.3 commit 0123456789ab\n",
		},
		{
			name:     "module version and revision",
			module:   "v0.8.0-alpha.7",
			revision: sha,
			want:     "fartapp v0.8.0-alpha.7 commit 0123456789ab\n",
		},
		{
			name:   "module version only",
			module: "v0.8.0-alpha.7",
			want:   "fartapp v0.8.0-alpha.7\n",
		},
		{
			name:     "local development build keeps the release tag and revision",
			module:   "(devel)",
			revision: sha,
			want:     "fartapp " + releaseVersion + " commit 0123456789ab\n",
		},
		{
			name:   "local development build without a revision",
			module: "(devel)",
			want:   "fartapp " + releaseVersion + "\n",
		},
		{
			name: "nothing recorded",
			want: "fartapp " + releaseVersion + "\n",
		},
		{
			name:     "blank revision is omitted",
			module:   "v0.8.0-alpha.7",
			revision: "   ",
			want:     "fartapp v0.8.0-alpha.7\n",
		},
		{
			name:     "short revision is kept",
			module:   "v0.1.0",
			revision: "abc",
			want:     "fartapp v0.1.0 commit abc\n",
		},
		{
			name:     "version stamp with the recorded revision",
			version:  "v9",
			revision: sha,
			want:     "fartapp v9 commit 0123456789ab\n",
		},
		{
			name:     "commit stamp with the module version",
			commit:   "deadbeef",
			module:   "v0.8.0-alpha.7",
			revision: sha,
			want:     "fartapp v0.8.0-alpha.7 commit deadbeef\n",
		},
		{
			name:    "whitespace stamps are unset",
			version: "  ",
			commit:  "\t",
			module:  "v0.4.0",
			want:    "fartapp v0.4.0\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := formatVersion(test.version, test.commit, versionFacts{
				moduleVersion: test.module,
				revision:      test.revision,
			})
			if got != test.want {
				t.Fatalf("got %q want %q", got, test.want)
			}
			if strings.Contains(got, "unknown") {
				t.Fatalf("output looks like a missing value: %q", got)
			}
		})
	}
}

func TestVersionCommandOmitsMissingCommit(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"fartapp", "version"}, strings.NewReader(""), &stdout, &stderr)
	want := formatVersion(versionOverride, commitOverride, currentVersionFacts())
	if code != 0 || stderr.Len() != 0 || stdout.String() != want {
		t.Fatalf("version=(%d,%q,%q) want %q", code, stdout.String(), stderr.String(), want)
	}
	if strings.Contains(stdout.String(), "unknown") {
		t.Fatalf("version looks like a missing value: %q", stdout.String())
	}
	if !strings.HasPrefix(stdout.String(), "fartapp ") || strings.Count(stdout.String(), "\n") != 1 {
		t.Fatalf("version = %q", stdout.String())
	}
	if strings.Contains(versionHelp, "unknown") {
		t.Fatal("version help still describes the commit as unknown")
	}
	for _, line := range strings.Split(versionHelp, "\n") {
		if len(line) > 80 {
			t.Fatalf("version help line is %d columns: %q", len(line), line)
		}
	}
}

func TestVersionLinkStampsOverrideBuildInfo(t *testing.T) {
	previousVersion := versionOverride
	previousCommit := commitOverride
	t.Cleanup(func() {
		versionOverride = previousVersion
		commitOverride = previousCommit
	})
	versionOverride = "v9.9.9"
	commitOverride = "abcdef"
	var stdout, stderr bytes.Buffer
	code := Run([]string{"fartapp", "--version"}, strings.NewReader(""), &stdout, &stderr)
	const want = "fartapp v9.9.9 commit abcdef\n"
	if code != 0 || stderr.Len() != 0 || stdout.String() != want {
		t.Fatalf("version=(%d,%q,%q) want %q", code, stdout.String(), stderr.String(), want)
	}
}
