// Copyright (c) 2024-2025 Neomantra Corp

// Package version reports the version and build provenance of dbn-go binaries.
package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

// Version is the current version of the application.
// This can be set at build time using ldflags:
// -ldflags="-X github.com/NimbleMarkets/dbn-go/internal/version.Version=v1.0.0"
var Version = ""

// Info holds all version-related metadata.
type Info struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Revision  string `json:"revision,omitempty"`    // full VCS revision (vcs.revision)
	Modified  bool   `json:"modified,omitempty"`    // working tree had local changes (vcs.modified)
	Time      string `json:"commit_time,omitempty"` // VCS commit time, RFC3339 (vcs.time)
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"` // GOOS/GOARCH of the running binary
}

// GetInfo returns a structured Info object.
func GetInfo(name string) Info {
	bi, ok := debug.ReadBuildInfo()
	return infoFrom(name, Version, bi, ok)
}

// Get returns the version string, including build info if available.
// It is a single token, suitable for embedding in client identifiers.
func Get() string {
	return GetInfo("").Version
}

// String returns a fully formatted version and build summary.
// The first line is "<name> version <version>"; further indented
// lines describe the revision, commit time, and Go toolchain.
func String(name string) string {
	return GetInfo(name).String()
}

// String formats the Info as a multi-line summary.
func (i Info) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s version %s", i.Name, i.Version)
	if i.Revision != "" {
		rev := shorten(i.Revision, 12)
		if i.Modified {
			rev += "-dirty"
		}
		fmt.Fprintf(&sb, "\n  revision:    %s", rev)
	}
	if i.Time != "" {
		fmt.Fprintf(&sb, "\n  commit time: %s", i.Time)
	}
	fmt.Fprintf(&sb, "\n  go version:  %s %s", i.GoVersion, i.Platform)
	return sb.String()
}

// infoFrom builds an Info from an ldflags version and (optional) build info.
// The ldflags version takes precedence over the main module version, which
// takes precedence over a "dev-<revision>" fallback, then "dev".
func infoFrom(name, ldflagsVersion string, bi *debug.BuildInfo, ok bool) Info {
	info := Info{
		Name:      name,
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}
	mainVersion := ""
	if ok && bi != nil {
		if bi.GoVersion != "" {
			info.GoVersion = bi.GoVersion
		}
		mainVersion = bi.Main.Version
		for _, setting := range bi.Settings {
			switch setting.Key {
			case "vcs.revision":
				info.Revision = setting.Value
			case "vcs.time":
				info.Time = setting.Value
			case "vcs.modified":
				info.Modified = setting.Value == "true"
			}
		}
	}

	switch {
	case ldflagsVersion != "":
		info.Version = ldflagsVersion
	case mainVersion != "" && mainVersion != "(devel)":
		info.Version = mainVersion
	case info.Revision != "":
		info.Version = "dev-" + shorten(info.Revision, 7)
	default:
		info.Version = "dev"
	}
	return info
}

func shorten(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
