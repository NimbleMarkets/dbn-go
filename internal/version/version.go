package version

import (
	"fmt"
	"runtime/debug"
)

// Version is the current version of the application.
// This can be set at build time using ldflags:
// -ldflags="-X github.com/NimbleMarkets/dbn-go/internal/version.Version=v1.0.0"
var Version = ""

// Info holds all version-related metadata.
type Info struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Revision string `json:"revision,omitempty"`
}

// GetInfo returns a structured Info object.
func GetInfo(name string) Info {
	info := Info{
		Name:    name,
		Version: Get(),
	}

	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range buildInfo.Settings {
			if setting.Key == "vcs.revision" {
				info.Revision = setting.Value
			}
		}
	}

	return info
}

// Get returns the version string, including build info if available.
func Get() string {
	if Version != "" {
		return Version
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "(devel)" {
			return info.Main.Version
		}
		// Try to find a vcs.revision
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return fmt.Sprintf("dev-%s", setting.Value[:7])
			}
		}
	}

	return "dev"
}

// String returns a fully formatted version and build summary.
func String(name string) string {
	return fmt.Sprintf("%s version %s", name, Get())
}
