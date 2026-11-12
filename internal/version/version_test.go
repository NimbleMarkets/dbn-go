// Copyright (c) 2025 Neomantra Corp

package version

import (
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
)

const testRevision = "f312e2e0123456789abcdef0123456789abcdef0"

func testBuildInfo(mainVersion string, settings ...debug.BuildSetting) *debug.BuildInfo {
	return &debug.BuildInfo{
		GoVersion: "go1.99.1",
		Main:      debug.Module{Path: "github.com/NimbleMarkets/dbn-go", Version: mainVersion},
		Settings:  settings,
	}
}

func vcsSettings(revision, time, modified string) []debug.BuildSetting {
	return []debug.BuildSetting{
		{Key: "vcs", Value: "git"},
		{Key: "vcs.revision", Value: revision},
		{Key: "vcs.time", Value: time},
		{Key: "vcs.modified", Value: modified},
	}
}

func TestInfoFromLdflagsVersionTakesPrecedence(t *testing.T) {
	bi := testBuildInfo("v0.0.0-20261110120000-f312e2e01234+dirty", vcsSettings(testRevision, "2026-11-10T12:00:00Z", "true")...)
	info := infoFrom("dbn-go-hist", "v1.2.3", bi, true)
	if info.Version != "v1.2.3" {
		t.Fatalf("Version = %q, want ldflags value v1.2.3", info.Version)
	}
	// Build-info fields are still reported alongside the ldflags version.
	if info.Revision != testRevision {
		t.Errorf("Revision = %q, want %q", info.Revision, testRevision)
	}
	if !info.Modified {
		t.Errorf("Modified = false, want true")
	}
	if info.Time != "2026-11-10T12:00:00Z" {
		t.Errorf("Time = %q, want vcs.time", info.Time)
	}
}

func TestInfoFromMainModuleVersionWithoutLdflags(t *testing.T) {
	bi := testBuildInfo("v0.9.0", vcsSettings(testRevision, "2026-11-10T12:00:00Z", "false")...)
	info := infoFrom("dbn-go-live", "", bi, true)
	if info.Version != "v0.9.0" {
		t.Fatalf("Version = %q, want module version v0.9.0", info.Version)
	}
	if info.Modified {
		t.Errorf("Modified = true, want false")
	}
}

func TestInfoFromDevelFallsBackToShortRevision(t *testing.T) {
	for _, mainVersion := range []string{"(devel)", ""} {
		bi := testBuildInfo(mainVersion, vcsSettings(testRevision, "", "false")...)
		info := infoFrom("dbn-go-tui", "", bi, true)
		if info.Version != "dev-f312e2e" {
			t.Errorf("main version %q: Version = %q, want dev-f312e2e", mainVersion, info.Version)
		}
	}
}

func TestInfoFromShortRevisionDoesNotPanic(t *testing.T) {
	bi := testBuildInfo("(devel)", debug.BuildSetting{Key: "vcs.revision", Value: "abc"})
	info := infoFrom("dbn-go-tui", "", bi, true)
	if info.Version != "dev-abc" {
		t.Fatalf("Version = %q, want dev-abc", info.Version)
	}
}

func TestInfoFromMissingBuildInfo(t *testing.T) {
	info := infoFrom("dbn-go-file", "", nil, false)
	want := Info{
		Name:      "dbn-go-file",
		Version:   "dev",
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}
	if info != want {
		t.Fatalf("info = %+v, want %+v", info, want)
	}

	// The ldflags version still applies without build info.
	if got := infoFrom("dbn-go-file", "v2.0.0", nil, false).Version; got != "v2.0.0" {
		t.Errorf("Version = %q, want v2.0.0", got)
	}
}

func TestInfoFromGoVersionAndPlatform(t *testing.T) {
	info := infoFrom("x", "", testBuildInfo("(devel)"), true)
	if info.GoVersion != "go1.99.1" {
		t.Errorf("GoVersion = %q, want go1.99.1 from build info", info.GoVersion)
	}
	if info.Platform != runtime.GOOS+"/"+runtime.GOARCH {
		t.Errorf("Platform = %q", info.Platform)
	}
}

func TestInfoFromModifiedFlag(t *testing.T) {
	cases := map[string]bool{"true": true, "false": false, "": false, "garbage": false}
	for value, want := range cases {
		settings := []debug.BuildSetting{{Key: "vcs.revision", Value: testRevision}}
		if value != "" {
			settings = append(settings, debug.BuildSetting{Key: "vcs.modified", Value: value})
		}
		info := infoFrom("x", "v1.0.0", testBuildInfo("(devel)", settings...), true)
		if info.Modified != want {
			t.Errorf("vcs.modified=%q: Modified = %v, want %v", value, info.Modified, want)
		}
	}
}

func TestInfoStringFull(t *testing.T) {
	info := Info{
		Name:      "dbn-go-hist",
		Version:   "v1.2.3",
		Revision:  testRevision,
		Modified:  true,
		Time:      "2026-11-10T12:00:00Z",
		GoVersion: "go1.99.1",
		Platform:  "darwin/arm64",
	}
	want := "dbn-go-hist version v1.2.3\n" +
		"  revision:    f312e2e01234-dirty\n" +
		"  commit time: 2026-11-10T12:00:00Z\n" +
		"  go version:  go1.99.1 darwin/arm64"
	if got := info.String(); got != want {
		t.Fatalf("String() =\n%s\nwant\n%s", got, want)
	}
}

func TestInfoStringCleanRevision(t *testing.T) {
	info := Info{Name: "n", Version: "v1", Revision: testRevision, GoVersion: "go1.99.1", Platform: "linux/amd64"}
	if got := info.String(); !strings.Contains(got, "  revision:    f312e2e01234\n") || strings.Contains(got, "dirty") {
		t.Fatalf("String() = %q, want clean 12-char revision", got)
	}
}

func TestInfoStringOmitsUnknownFields(t *testing.T) {
	info := Info{Name: "dbn-go-file", Version: "dev", GoVersion: "go1.99.1", Platform: "linux/amd64"}
	want := "dbn-go-file version dev\n" +
		"  go version:  go1.99.1 linux/amd64"
	if got := info.String(); got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}

func TestStringFirstLineIsBackwardCompatible(t *testing.T) {
	first, _, _ := strings.Cut(String("dbn-go-live"), "\n")
	if want := "dbn-go-live version " + Get(); first != want {
		t.Fatalf("first line = %q, want %q", first, want)
	}
}

func TestGetMatchesGetInfo(t *testing.T) {
	if Get() != GetInfo("x").Version {
		t.Fatalf("Get() = %q, GetInfo().Version = %q", Get(), GetInfo("x").Version)
	}
}
