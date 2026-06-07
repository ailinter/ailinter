// Package version provides semantic versioning for the ailinter binary.
//
// Build information is injected via ldflags at build time:
//
//	go build -ldflags="\
//	  -X github.com/ailinter/ailinter/internal/version.Version=v1.0.0 \
//	  -X github.com/ailinter/ailinter/internal/version.Commit=abc1234 \
//	  -X github.com/ailinter/ailinter/internal/version.BuildDate=2026-01-15T10:00:00Z \
//	" ./cmd/ailinter
package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

// Build information — set via ldflags at build time.
var (
	// Version is the semantic version (e.g., "v1.0.0", "v0.9.0-dev").
	// Defaults to "0.0.0-dev" when not set via ldflags.
	Version = "0.0.0-dev"

	// Commit is the short git commit hash (e.g., "abc1234").
	Commit = "unknown"

	// BuildDate is the ISO 8601 build timestamp (e.g., "2026-01-15T10:00:00Z").
	BuildDate = "unknown"
)

// Semver returns the semantic version with the 'v' prefix stripped.
func Semver() string {
	if len(Version) > 0 && Version[0] == 'v' {
		return Version[1:]
	}
	return Version
}

// IsPrerelease returns true if the version contains a pre-release suffix
// (e.g., "-dev", "-alpha", "-beta", "-rc.1").
func IsPrerelease() bool {
	for _, c := range Version {
		if c == '-' {
			return true
		}
	}
	return false
}

// APIVersion returns the API compatibility major version (e.g., "v1").
// This is derived from the major version number in the semantic version string.
func APIVersion() string {
	s := Semver()
	if s == "" {
		return "v0"
	}
	major := string(s[0])
	return "v" + major
}

// String returns a full version string suitable for --version output:
//
//	ailinter v1.0.0 (commit: abc1234, built: 2026-01-15T10:00:00Z, go: go1.25.5, darwin/arm64)
func String() string {
	info, ok := debug.ReadBuildInfo()
	goVersion := runtime.Version()
	if ok && info.GoVersion != "" {
		goVersion = info.GoVersion
	}
	return fmt.Sprintf("ailinter %s (commit: %s, built: %s, go: %s, %s/%s)",
		Version, Commit, BuildDate, goVersion, runtime.GOOS, runtime.GOARCH)
}

// Short returns a short version string.
func Short() string {
	return "ailinter version " + Version
}
