package version

import (
	"fmt"
	"runtime"
)

// These variables are set at build time using ldflags
var (
	// Version is the semantic version of the application
	Version = "dev"

	// Commit is the git commit hash
	Commit = "unknown"

	// BuildTime is the build timestamp
	BuildTime = "unknown"

	// GoVersion is the Go version used to build the application
	GoVersion = runtime.Version()
)

// Info contains version information
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
}

// Get returns the current version information
func Get() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		BuildTime: BuildTime,
		GoVersion: GoVersion,
	}
}

// String returns a formatted version string
func String() string {
	return fmt.Sprintf("Bazaruto %s (commit: %s, built: %s, go: %s)",
		Version, Commit, BuildTime, GoVersion)
}

// Short returns a short version string
func Short() string {
	return fmt.Sprintf("v%s", Version)
}
