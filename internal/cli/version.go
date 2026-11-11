package cli

import (
	"fmt"
	"runtime/debug"
	"strings"
)

const releaseVersion = "v0.8.0-alpha.7"

// versionOverride and commitOverride are optional link-time stamps:
//
//	go build -ldflags "-X github.com/blisspixel/fartapp/internal/cli.versionOverride=v0.8.0-alpha.7 -X github.com/blisspixel/fartapp/internal/cli.commitOverride=0123456789ab"
var (
	versionOverride string
	commitOverride  string
)

const versionHelp = `Print the version and the commit when the build recorded one.

Usage:
  fartapp version
  fartapp --version

The version is the module version when the build recorded one, otherwise the
release tag in this source. The commit is the first 12 characters of the VCS
revision when one was recorded. When no commit was recorded, the line prints
the version only. Link-time -X settings for the version and commit override
those values. This does not change the permanent toy output.
`

type versionFacts struct {
	moduleVersion string
	revision      string
}

func versionLine() string {
	return formatVersion(versionOverride, commitOverride, currentVersionFacts())
}

func currentVersionFacts() versionFacts {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return versionFacts{}
	}
	facts := versionFacts{moduleVersion: info.Main.Version}
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" && setting.Value != "" {
			facts.revision = setting.Value
			break
		}
	}
	return facts
}

func formatVersion(overrideVersion, overrideCommit string, facts versionFacts) string {
	version := strings.TrimSpace(overrideVersion)
	if version == "" {
		moduleVersion := strings.TrimSpace(facts.moduleVersion)
		// "(devel)" is the local-build placeholder, not a released module version.
		if moduleVersion != "" && moduleVersion != "(devel)" {
			version = moduleVersion
		}
	}
	if version == "" {
		version = releaseVersion
	}
	commit := shortCommit(overrideCommit)
	if commit == "" {
		commit = shortCommit(facts.revision)
	}
	if commit == "" {
		return fmt.Sprintf("fartapp %s\n", version)
	}
	return fmt.Sprintf("fartapp %s commit %s\n", version, commit)
}

func shortCommit(commit string) string {
	commit = strings.TrimSpace(commit)
	if commit == "" {
		return ""
	}
	count := 0
	for index := range commit {
		if count == 12 {
			return commit[:index]
		}
		count++
	}
	return commit
}
