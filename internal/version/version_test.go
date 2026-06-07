// AI-generated test fixture

package version

import (
	"strings"
	"testing"
)

func TestSemver(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{"strips v prefix", "v1.0.0", "1.0.0"},
		{"no v prefix", "1.0.0", "1.0.0"},
		{"dev version", "v0.9.0-dev", "0.9.0-dev"},
		{"empty string", "", ""},
		{"just v", "v", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := Version
			Version = tt.version
			t.Cleanup(func() { Version = orig })
			if got := Semver(); got != tt.want {
				t.Errorf("Semver() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsPrerelease(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    bool
	}{
		{"dev suffix", "v0.9.0-dev", true},
		{"alpha suffix", "v1.0.0-alpha", true},
		{"beta suffix", "v1.0.0-beta.1", true},
		{"rc suffix", "v1.0.0-rc1", true},
		{"full release", "v1.0.0", false},
		{"no suffix", "0.1.0", false},
		{"empty string", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := Version
			Version = tt.version
			t.Cleanup(func() { Version = orig })
			if got := IsPrerelease(); got != tt.want {
				t.Errorf("IsPrerelease() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAPIVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{"v1 major", "v1.0.0", "v1"},
		{"v2 major", "v2.3.4", "v2"},
		{"dev v0", "v0.9.0-dev", "v0"},
		{"no prefix", "1.2.3", "v1"},
		{"empty version", "", "v0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := Version
			Version = tt.version
			t.Cleanup(func() { Version = orig })
			if got := APIVersion(); got != tt.want {
				t.Errorf("APIVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestShort(t *testing.T) {
	orig := Version
	Version = "v1.0.0"
	t.Cleanup(func() { Version = orig })

	want := "ailinter version v1.0.0"
	if got := Short(); got != want {
		t.Errorf("Short() = %q, want %q", got, want)
	}
}

func TestString(t *testing.T) {
	origVersion := Version
	origCommit := Commit
	origBuildDate := BuildDate
	Version = "v1.0.0"
	Commit = "abc1234"
	BuildDate = "2026-06-07T12:00:00Z"
	t.Cleanup(func() {
		Version = origVersion
		Commit = origCommit
		BuildDate = origBuildDate
	})

	result := String()
	if !strings.HasPrefix(result, "ailinter ") {
		t.Errorf("String() should start with 'ailinter ', got: %s", result)
	}
	if !strings.Contains(result, "v1.0.0") {
		t.Errorf("String() should contain version, got: %s", result)
	}
	if !strings.Contains(result, "abc1234") {
		t.Errorf("String() should contain commit, got: %s", result)
	}
	if !strings.Contains(result, "2026-06-07T12:00:00Z") {
		t.Errorf("String() should contain build date, got: %s", result)
	}
	if !strings.Contains(result, "go") {
		t.Errorf("String() should contain go version, got: %s", result)
	}
}

func TestDefaults(t *testing.T) {
	// Test default values (when no ldflags are set)
	if Version != "0.0.0-dev" {
		t.Errorf("default Version = %q, want %q", Version, "0.0.0-dev")
	}
	if Commit != "unknown" {
		t.Errorf("default Commit = %q, want %q", Commit, "unknown")
	}
	if BuildDate != "unknown" {
		t.Errorf("default BuildDate = %q, want %q", BuildDate, "unknown")
	}
}
