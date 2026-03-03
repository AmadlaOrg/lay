package install

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/AmadlaOrg/lay/binary/forge"
	"github.com/stretchr/testify/assert"
)

type mockForgeSvc struct {
	assets         []forge.Asset
	tag            string
	assetsErr      error
	tagAssetsErr   error
	calledByTag    bool
	calledByLatest bool
}

func (m *mockForgeSvc) GetLatestReleaseAssets(owner, repo string) ([]forge.Asset, string, error) {
	m.calledByLatest = true
	return m.assets, m.tag, m.assetsErr
}

func (m *mockForgeSvc) GetReleaseAssetsByTag(owner, repo, tag string) ([]forge.Asset, string, error) {
	m.calledByTag = true
	return m.assets, m.tag, m.tagAssetsErr
}

func TestService_Install_InvalidShorthand(t *testing.T) {
	origNewForge := newForgeService
	defer func() { newForgeService = origNewForge }()

	newForgeService = func(name string) (forge.Forge, error) {
		return &mockForgeSvc{
			assetsErr: assert.AnError,
		}, nil
	}

	s := &Service{}
	err := s.Install("sharkdp/fd", "/tmp/test-bin")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get release assets")
}

func TestService_Install_NoAssets(t *testing.T) {
	origNewForge := newForgeService
	defer func() { newForgeService = origNewForge }()

	newForgeService = func(name string) (forge.Forge, error) {
		return &mockForgeSvc{
			assets: nil,
		}, nil
	}

	s := &Service{}
	err := s.Install("sharkdp/fd", "/tmp/test-bin")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no release assets found")
}

func TestService_Install_VersionSpecific(t *testing.T) {
	origNewForge := newForgeService
	defer func() { newForgeService = origNewForge }()

	mock := &mockForgeSvc{
		tagAssetsErr: assert.AnError,
	}

	newForgeService = func(name string) (forge.Forge, error) {
		return mock, nil
	}

	s := &Service{}
	err := s.Install("sharkdp/fd@v9.0.0", "/tmp/test-bin")

	assert.Error(t, err)
	assert.True(t, mock.calledByTag, "should call GetReleaseAssetsByTag for versioned shorthand")
	assert.False(t, mock.calledByLatest, "should not call GetLatestReleaseAssets for versioned shorthand")
}

func TestService_Install_LatestRelease(t *testing.T) {
	origNewForge := newForgeService
	defer func() { newForgeService = origNewForge }()

	mock := &mockForgeSvc{
		assetsErr: assert.AnError,
	}

	newForgeService = func(name string) (forge.Forge, error) {
		return mock, nil
	}

	s := &Service{}
	err := s.Install("sharkdp/fd", "/tmp/test-bin")

	assert.Error(t, err)
	assert.True(t, mock.calledByLatest, "should call GetLatestReleaseAssets for unversioned shorthand")
	assert.False(t, mock.calledByTag, "should not call GetReleaseAssetsByTag for unversioned shorthand")
}

func TestService_Install_GitLabPrefix(t *testing.T) {
	origNewForge := newForgeService
	defer func() { newForgeService = origNewForge }()

	var calledWithForge string
	mock := &mockForgeSvc{
		assetsErr: assert.AnError,
	}

	newForgeService = func(name string) (forge.Forge, error) {
		calledWithForge = name
		return mock, nil
	}

	s := &Service{}
	_ = s.Install("gitlab:user/project", "/tmp/test-bin")

	assert.Equal(t, "gitlab", calledWithForge, "should request gitlab forge")
	assert.True(t, mock.calledByLatest)
}

func TestService_Install_CodebergPrefix(t *testing.T) {
	origNewForge := newForgeService
	defer func() { newForgeService = origNewForge }()

	var calledWithForge string
	mock := &mockForgeSvc{
		assetsErr: assert.AnError,
	}

	newForgeService = func(name string) (forge.Forge, error) {
		calledWithForge = name
		return mock, nil
	}

	s := &Service{}
	_ = s.Install("codeberg:user/repo", "/tmp/test-bin")

	assert.Equal(t, "codeberg", calledWithForge, "should request codeberg forge")
	assert.True(t, mock.calledByLatest)
}

func TestService_Install_GitLabWithVersion(t *testing.T) {
	origNewForge := newForgeService
	defer func() { newForgeService = origNewForge }()

	var calledWithForge string
	mock := &mockForgeSvc{
		tagAssetsErr: assert.AnError,
	}

	newForgeService = func(name string) (forge.Forge, error) {
		calledWithForge = name
		return mock, nil
	}

	s := &Service{}
	_ = s.Install("gitlab:user/project@v2.0", "/tmp/test-bin")

	assert.Equal(t, "gitlab", calledWithForge)
	assert.True(t, mock.calledByTag, "should call GetReleaseAssetsByTag for versioned shorthand")
}

func TestService_Install_UnsupportedForgePrefix(t *testing.T) {
	s := &Service{}
	err := s.Install("bitbucket:user/repo", "/tmp/test-bin")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported forge: bitbucket")
}

func TestService_Install_ForgeServiceCreationError(t *testing.T) {
	origNewForge := newForgeService
	defer func() { newForgeService = origNewForge }()

	newForgeService = func(name string) (forge.Forge, error) {
		return nil, fmt.Errorf("forge unavailable: %s", name)
	}

	s := &Service{}
	err := s.Install("sharkdp/fd", "/tmp/test-bin")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forge unavailable: github")
}

func TestService_Install_NoCompatibleAsset(t *testing.T) {
	origNewForge := newForgeService
	defer func() { newForgeService = origNewForge }()

	newForgeService = func(name string) (forge.Forge, error) {
		return &mockForgeSvc{
			assets: []forge.Asset{
				{Name: "tool-freebsd-riscv64.tar.gz", DownloadURL: "https://example.com/tool-freebsd-riscv64.tar.gz"},
				{Name: "tool-plan9-mips.tar.gz", DownloadURL: "https://example.com/tool-plan9-mips.tar.gz"},
			},
			tag: "v1.0.0",
		}, nil
	}

	s := &Service{}
	err := s.Install("user/tool", "/tmp/test-bin")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no compatible asset found")
}

func TestService_Install_HappyPath(t *testing.T) {
	origNewForge := newForgeService
	origHTTPGet := httpGet
	defer func() {
		newForgeService = origNewForge
		httpGet = origHTTPGet
	}()

	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Build asset name that matches current platform
	assetName := fmt.Sprintf("mytool-%s-%s.tar.gz", goos, goarch)
	binaryName := "mytool"
	binaryContent := "#!/bin/sh\necho hello"

	tarBuf := createTestTarGz(t, binaryName, binaryContent)

	newForgeService = func(name string) (forge.Forge, error) {
		return &mockForgeSvc{
			assets: []forge.Asset{
				{Name: assetName, DownloadURL: "https://example.com/" + assetName},
			},
			tag: "v1.0.0",
		}, nil
	}

	httpGet = func(url string) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader(tarBuf)),
		}, nil
	}

	targetDir := t.TempDir()
	s := &Service{}
	err := s.Install("user/mytool", targetDir)

	assert.NoError(t, err)

	installedPath := filepath.Join(targetDir, binaryName)
	info, err := os.Stat(installedPath)
	assert.NoError(t, err)
	assert.True(t, info.Mode()&0111 != 0, "binary should be executable")

	content, err := os.ReadFile(installedPath)
	assert.NoError(t, err)
	assert.Equal(t, binaryContent, string(content))
}

// createTestTarGz builds an in-memory tar.gz containing a single executable file.
func createTestTarGz(t *testing.T, name, content string) []byte {
	t.Helper()

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	hdr := &tar.Header{
		Name: name,
		Mode: 0755,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(tw, strings.NewReader(content)); err != nil {
		t.Fatal(err)
	}

	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}

	return buf.Bytes()
}

func TestNewInstallService(t *testing.T) {
	svc := NewInstallService()
	assert.NotNil(t, svc)
}
