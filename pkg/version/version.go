package version

import (
	"fmt"
	"runtime"
	"time"
)

// Version information - using semantic versioning
const (
	Major     = 2
	Minor     = 5
	Patch     = 0
	PreRelease = "" // e.g., "alpha", "beta", "rc1"
	BuildMetadata = "" // e.g., "20231201.1"
	GitCommit = ""
	BuildDate = ""
)

// Version returns the semantic version string
func Version() string {
	version := fmt.Sprintf("%d.%d.%d", Major, Minor, Patch)
	
	if PreRelease != "" {
		version += "-" + PreRelease
	}
	
	if BuildMetadata != "" {
		version += "+" + BuildMetadata
	}
	
	return version
}

// BuildInfo contains comprehensive build information
type BuildInfo struct {
	Version       string `json:"version"`
	Major         int    `json:"major"`
	Minor         int    `json:"minor"`
	Patch         int    `json:"patch"`
	PreRelease    string `json:"pre_release,omitempty"`
	BuildMetadata string `json:"build_metadata,omitempty"`
	GitCommit     string `json:"git_commit,omitempty"`
	BuildDate     string `json:"build_date,omitempty"`
	GoVersion     string `json:"go_version"`
	Platform      string `json:"platform"`
	SDKName       string `json:"sdk_name"`
}

// GetVersion returns the current semantic version
func GetVersion() string {
	return Version()
}

// GetBuildInfo returns complete build information
func GetBuildInfo() *BuildInfo {
	buildDate := BuildDate
	if buildDate == "" {
		buildDate = time.Now().Format("2006-01-02T15:04:05Z07:00")
	}
	
	return &BuildInfo{
		Version:       Version(),
		Major:         Major,
		Minor:         Minor,
		Patch:         Patch,
		PreRelease:    PreRelease,
		BuildMetadata: BuildMetadata,
		GitCommit:     GitCommit,
		BuildDate:     buildDate,
		GoVersion:     runtime.Version(),
		Platform:      fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		SDKName:       "Teneo Agent SDK",
	}
}

// GetVersionString returns a formatted version string
func GetVersionString() string {
	buildInfo := GetBuildInfo()
	if buildInfo.GitCommit != "" && len(buildInfo.GitCommit) >= 7 {
		return fmt.Sprintf("%s (%s)", buildInfo.Version, buildInfo.GitCommit[:7])
	}
	return buildInfo.Version
}

// GetFullVersionString returns a complete version string with build info
func GetFullVersionString() string {
	buildInfo := GetBuildInfo()
	result := fmt.Sprintf("%s v%s", buildInfo.SDKName, buildInfo.Version)
	
	if buildInfo.GitCommit != "" && len(buildInfo.GitCommit) >= 7 {
		result += fmt.Sprintf(" (commit: %s)", buildInfo.GitCommit[:7])
	}
	
	if buildInfo.BuildDate != "" {
		result += fmt.Sprintf(" (built: %s)", buildInfo.BuildDate)
	}
	
	result += fmt.Sprintf(" (go: %s, platform: %s)", buildInfo.GoVersion, buildInfo.Platform)
	
	return result
}

// GetBanner returns a formatted banner for application startup
func GetBanner() string {
	buildInfo := GetBuildInfo()
	banner := fmt.Sprintf(`
┌─────────────────────────────────────────────────────────────────┐
│                         %s                        │
│                           v%-8s                          │
├─────────────────────────────────────────────────────────────────┤
│ Go Version: %-25s Platform: %-15s │
│ Build Date: %-51s │`, 
		buildInfo.SDKName,
		buildInfo.Version,
		buildInfo.GoVersion,
		buildInfo.Platform,
		buildInfo.BuildDate,
	)
	
	if buildInfo.GitCommit != "" && len(buildInfo.GitCommit) >= 7 {
		banner += fmt.Sprintf(`
│ Git Commit: %-51s │`, buildInfo.GitCommit[:7])
	}
	
	banner += `
└─────────────────────────────────────────────────────────────────┘`
	
	return banner
}

// IsPreRelease returns true if this is a pre-release version
func IsPreRelease() bool {
	return PreRelease != ""
}

// IsCompatible checks if the given version is compatible with the current version
func IsCompatible(otherMajor, otherMinor int) bool {
	// Same major version is required for compatibility
	if Major != otherMajor {
		return false
	}
	
	// Current version must be >= other version for backward compatibility
	return Minor >= otherMinor
}

// CompareVersions compares two semantic versions
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func CompareVersions(v1Major, v1Minor, v1Patch, v2Major, v2Minor, v2Patch int) int {
	if v1Major != v2Major {
		if v1Major < v2Major {
			return -1
		}
		return 1
	}
	
	if v1Minor != v2Minor {
		if v1Minor < v2Minor {
			return -1
		}
		return 1
	}
	
	if v1Patch != v2Patch {
		if v1Patch < v2Patch {
			return -1
		}
		return 1
	}
	
	return 0
}
