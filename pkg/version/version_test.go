package version

import (
	"strings"
	"testing"
)

func TestGetVersion(t *testing.T) {
	version := GetVersion()
	if version == "" {
		t.Error("Version should not be empty")
	}

	if !strings.Contains(version, "2.5.0") {
		t.Errorf("Expected version to contain '2.5.0', got: %s", version)
	}
}

func TestSemanticVersionComponents(t *testing.T) {
	if Major != 2 {
		t.Errorf("Expected Major version to be 2, got: %d", Major)
	}
	
	if Minor != 5 {
		t.Errorf("Expected Minor version to be 4, got: %d", Minor)
	}
	
	if Patch != 0 {
		t.Errorf("Expected Patch version to be 0, got: %d", Patch)
	}
}

func TestVersionFormat(t *testing.T) {
	version := Version()
	expected := "2.5.0"
	
	if version != expected {
		t.Errorf("Expected version '%s', got: '%s'", expected, version)
	}
}

func TestGetBuildInfo(t *testing.T) {
	buildInfo := GetBuildInfo()

	if buildInfo.Version == "" {
		t.Error("BuildInfo.Version should not be empty")
	}

	if buildInfo.GoVersion == "" {
		t.Error("BuildInfo.GoVersion should not be empty")
	}

	if buildInfo.Platform == "" {
		t.Error("BuildInfo.Platform should not be empty")
	}
	
	if buildInfo.SDKName != "Teneo Agent SDK" {
		t.Errorf("Expected SDK name 'Teneo Agent SDK', got: %s", buildInfo.SDKName)
	}
	
	if buildInfo.Major != 2 {
		t.Errorf("Expected Major version 2, got: %d", buildInfo.Major)
	}
}

func TestGetVersionString(t *testing.T) {
	versionString := GetVersionString()
	if versionString == "" {
		t.Error("Version string should not be empty")
	}
	
	if !strings.Contains(versionString, "2.5.0") {
		t.Errorf("Version string should contain '2.5.0', got: %s", versionString)
	}
}

func TestGetFullVersionString(t *testing.T) {
	fullVersionString := GetFullVersionString()
	if fullVersionString == "" {
		t.Error("Full version string should not be empty")
	}

	if !strings.Contains(fullVersionString, "Teneo Agent SDK") {
		t.Errorf("Expected full version string to contain 'Teneo Agent SDK', got: %s", fullVersionString)
	}
	
	if !strings.Contains(fullVersionString, "v2.5.0") {
		t.Errorf("Expected full version string to contain 'v2.5.0', got: %s", fullVersionString)
	}
}

func TestGetBanner(t *testing.T) {
	banner := GetBanner()
	if banner == "" {
		t.Error("Banner should not be empty")
	}
	
	if !strings.Contains(banner, "Teneo Agent SDK") {
		t.Error("Banner should contain 'Teneo Agent SDK'")
	}
	
	if !strings.Contains(banner, "v2.5.0") {
		t.Error("Banner should contain 'v2.5.0'")
	}
}

func TestIsPreRelease(t *testing.T) {
	// Since PreRelease is empty by default
	if IsPreRelease() {
		t.Error("Expected IsPreRelease to be false for stable version")
	}
}

func TestIsCompatible(t *testing.T) {
	tests := []struct {
		name         string
		otherMajor   int
		otherMinor   int
		expectedCompatible bool
	}{
		{"same version", 2, 5, true},
		{"newer minor", 2, 6, false},
		{"older minor", 2, -1, true}, // we're backward compatible
		{"different major", 1, 0, false},
		{"different major", 3, 0, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCompatible(tt.otherMajor, tt.otherMinor)
			if result != tt.expectedCompatible {
				t.Errorf("IsCompatible(%d, %d) = %v, expected %v", 
					tt.otherMajor, tt.otherMinor, result, tt.expectedCompatible)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		v1       []int // [major, minor, patch]
		v2       []int
		expected int
	}{
		{"equal versions", []int{2, 1, 0}, []int{2, 1, 0}, 0},
		{"v1 newer major", []int{3, 0, 0}, []int{2, 1, 0}, 1},
		{"v1 older major", []int{1, 0, 0}, []int{2, 1, 0}, -1},
		{"v1 newer minor", []int{2, 2, 0}, []int{2, 1, 0}, 1},
		{"v1 older minor", []int{2, 0, 0}, []int{2, 1, 0}, -1},
		{"v1 newer patch", []int{2, 1, 1}, []int{2, 1, 0}, 1},
		{"v1 older patch", []int{2, 1, 0}, []int{2, 1, 1}, -1},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareVersions(
				tt.v1[0], tt.v1[1], tt.v1[2],
				tt.v2[0], tt.v2[1], tt.v2[2],
			)
			if result != tt.expected {
				t.Errorf("CompareVersions(%v, %v) = %d, expected %d", 
					tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}
