package install

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectPlatform(t *testing.T) {
	p := DetectPlatform()
	assert.NotEmpty(t, p.OS)
	assert.NotEmpty(t, p.Arch)
}

func TestMatchAsset(t *testing.T) {
	tests := []struct {
		name     string
		asset    string
		platform PlatformInfo
		expected bool
	}{
		{
			name:     "linux amd64 tar.gz",
			asset:    "fd-v9.0.0-x86_64-unknown-linux-gnu.tar.gz",
			platform: PlatformInfo{OS: "linux", Arch: "amd64"},
			expected: true,
		},
		{
			name:     "linux arm64",
			asset:    "fd-v9.0.0-aarch64-unknown-linux-gnu.tar.gz",
			platform: PlatformInfo{OS: "linux", Arch: "arm64"},
			expected: true,
		},
		{
			name:     "darwin amd64",
			asset:    "fd-v9.0.0-x86_64-apple-darwin.tar.gz",
			platform: PlatformInfo{OS: "darwin", Arch: "amd64"},
			expected: true,
		},
		{
			name:     "macOS arm64",
			asset:    "ripgrep-14.0.0-aarch64-apple-darwin.tar.gz",
			platform: PlatformInfo{OS: "darwin", Arch: "arm64"},
			expected: true,
		},
		{
			name:     "wrong OS",
			asset:    "fd-v9.0.0-x86_64-unknown-linux-gnu.tar.gz",
			platform: PlatformInfo{OS: "darwin", Arch: "amd64"},
			expected: false,
		},
		{
			name:     "wrong arch",
			asset:    "fd-v9.0.0-x86_64-unknown-linux-gnu.tar.gz",
			platform: PlatformInfo{OS: "linux", Arch: "arm64"},
			expected: false,
		},
		{
			name:     "skip sha256",
			asset:    "fd-v9.0.0-x86_64-unknown-linux-gnu.tar.gz.sha256",
			platform: PlatformInfo{OS: "linux", Arch: "amd64"},
			expected: false,
		},
		{
			name:     "skip sig file",
			asset:    "fd-v9.0.0-x86_64-unknown-linux-gnu.tar.gz.sig",
			platform: PlatformInfo{OS: "linux", Arch: "amd64"},
			expected: false,
		},
		{
			name:     "skip asc file",
			asset:    "fd-v9.0.0-x86_64-unknown-linux-gnu.tar.gz.asc",
			platform: PlatformInfo{OS: "linux", Arch: "amd64"},
			expected: false,
		},
		{
			name:     "windows amd64 zip",
			asset:    "fd-v9.0.0-x86_64-pc-windows-msvc.zip",
			platform: PlatformInfo{OS: "windows", Arch: "amd64"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, MatchAsset(tt.asset, tt.platform))
		})
	}
}

func TestRankAssets(t *testing.T) {
	assets := []string{
		"fd-v9.0.0-x86_64-unknown-linux-gnu.tar.gz",
		"fd-v9.0.0-x86_64-unknown-linux-musl.tar.gz",
		"fd-v9.0.0-x86_64-pc-windows-msvc.zip",
		"fd-v9.0.0-x86_64-unknown-linux-gnu.tar.gz.sha256",
		"fd-v9.0.0-aarch64-unknown-linux-gnu.tar.gz",
	}

	platform := PlatformInfo{OS: "linux", Arch: "amd64"}
	ranked := RankAssets(assets, platform)

	assert.Len(t, ranked, 2)
	// musl should be ranked first (tar.gz + musl bonus)
	assert.Equal(t, "fd-v9.0.0-x86_64-unknown-linux-musl.tar.gz", ranked[0])
	assert.Equal(t, "fd-v9.0.0-x86_64-unknown-linux-gnu.tar.gz", ranked[1])
}

func TestRankAssets_NoMatch(t *testing.T) {
	assets := []string{
		"fd-v9.0.0-x86_64-pc-windows-msvc.zip",
	}

	platform := PlatformInfo{OS: "linux", Arch: "amd64"}
	ranked := RankAssets(assets, platform)

	assert.Empty(t, ranked)
}

func TestRankAssets_PrefersTarGzOverZip(t *testing.T) {
	assets := []string{
		"tool-linux-amd64.zip",
		"tool-linux-amd64.tar.gz",
	}

	platform := PlatformInfo{OS: "linux", Arch: "amd64"}
	ranked := RankAssets(assets, platform)

	assert.Len(t, ranked, 2)
	assert.Equal(t, "tool-linux-amd64.tar.gz", ranked[0])
}
