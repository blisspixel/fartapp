package cli

import (
	"fmt"
	"runtime/debug"
)

const releaseVersion = "v0.8.0-alpha.3"

const versionHelp = `Print the release tag and the build commit.

Usage:
  fartapp version
  fartapp --version

The commit is the first 12 hexadecimal digits when the build recorded a VCS
revision, otherwise unknown. This does not change the permanent toy output.
`

func versionLine() string {
	revision := "unknown"
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" && setting.Value != "" {
				revision = setting.Value
				if len(revision) > 12 {
					revision = revision[:12]
				}
				break
			}
		}
	}
	return fmt.Sprintf("fartapp %s commit %s\n", releaseVersion, revision)
}
