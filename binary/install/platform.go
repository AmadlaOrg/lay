package install

import (
	"runtime"
	"sort"
	"strings"
)

// PlatformInfo holds the current OS and architecture
type PlatformInfo struct {
	OS   string
	Arch string
}

// DetectPlatform returns the current platform info
func DetectPlatform() PlatformInfo {
	return PlatformInfo{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}

// OS aliases map canonical Go OS names to common variants found in release assets
var osAliases = map[string][]string{
	"linux":   {"linux", "Linux"},
	"darwin":  {"darwin", "Darwin", "macos", "macOS", "mac", "apple"},
	"windows": {"windows", "Windows", "win", "win64", "win32"},
}

// Arch aliases map canonical Go arch names to common variants
var archAliases = map[string][]string{
	"amd64": {"amd64", "x86_64", "x64"},
	"arm64": {"arm64", "aarch64"},
	"386":   {"386", "i686", "i386", "x86"},
}

// Checksum/signature extensions to skip
var skipExtensions = []string{".sha256", ".sha512", ".sig", ".asc", ".sbom", ".sha256sum", ".md5"}

// MatchAsset returns true if the asset name matches the given platform
func MatchAsset(name string, platform PlatformInfo) bool {
	lower := strings.ToLower(name)

	// Skip checksum/signature files
	for _, ext := range skipExtensions {
		if strings.HasSuffix(lower, ext) {
			return false
		}
	}

	osMatch := false
	if aliases, ok := osAliases[platform.OS]; ok {
		for _, alias := range aliases {
			if strings.Contains(name, alias) {
				osMatch = true
				break
			}
		}
	}

	archMatch := false
	if aliases, ok := archAliases[platform.Arch]; ok {
		for _, alias := range aliases {
			if strings.Contains(name, alias) {
				archMatch = true
				break
			}
		}
	}

	return osMatch && archMatch
}

// RankAssets filters and sorts assets for the given platform.
// Prefers .tar.gz over .zip, prefers musl/static variants.
func RankAssets(names []string, platform PlatformInfo) []string {
	var matched []string
	for _, name := range names {
		if MatchAsset(name, platform) {
			matched = append(matched, name)
		}
	}

	sort.Slice(matched, func(i, j int) bool {
		si := scoreAsset(matched[i])
		sj := scoreAsset(matched[j])
		return si > sj
	})

	return matched
}

func scoreAsset(name string) int {
	score := 0
	lower := strings.ToLower(name)

	// Prefer tar.gz
	if strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz") {
		score += 10
	}
	// tar.xz also good
	if strings.HasSuffix(lower, ".tar.xz") {
		score += 8
	}
	// zip is less preferred
	if strings.HasSuffix(lower, ".zip") {
		score += 5
	}

	// Prefer musl/static builds
	if strings.Contains(lower, "musl") || strings.Contains(lower, "static") {
		score += 3
	}

	return score
}
