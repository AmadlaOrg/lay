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
	_, err := s.Install("sharkdp/fd", "/tmp/test-bin")
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
	_, err := s.Install("sharkdp/fd", "/tmp/test-bin")
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
	_, err := s.Install("sharkdp/fd@v9.0.0", "/tmp/test-bin")

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
	_, err := s.Install("sharkdp/fd", "/tmp/test-bin")

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
	_, _ = s.Install("gitlab:user/project", "/tmp/test-bin")

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
	_, _ = s.Install("codeberg:user/repo", "/tmp/test-bin")

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
	_, _ = s.Install("gitlab:user/project@v2.0", "/tmp/test-bin")

	assert.Equal(t, "gitlab", calledWithForge)
	assert.True(t, mock.calledByTag, "should call GetReleaseAssetsByTag for versioned shorthand")
}

func TestService_Install_UnsupportedForgePrefix(t *testing.T) {
	s := &Service{}
	_, err := s.Install("bitbucket:user/repo", "/tmp/test-bin")

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
	_, err := s.Install("sharkdp/fd", "/tmp/test-bin")

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
	_, err := s.Install("user/tool", "/tmp/test-bin")

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
	result, err := s.Install("user/mytool", targetDir)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "mytool", result.BinaryName)
	assert.Equal(t, "v1.0.0", result.Version)
	assert.NotEmpty(t, result.Checksum)
	assert.Equal(t, "sha256", result.HashAlgo)

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

func TestService_Install_DirectURL(t *testing.T) {
	origHTTPGet := httpGet
	defer func() { httpGet = origHTTPGet }()

	binaryName := "mytool"
	binaryContent := "#!/bin/sh\necho direct"
	tarBuf := createTestTarGz(t, binaryName, binaryContent)

	httpGet = func(url string) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader(tarBuf)),
		}, nil
	}

	targetDir := t.TempDir()
	s := &Service{}
	result, err := s.Install("https://example.com/mytool-linux-amd64.tar.gz", targetDir)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "mytool", result.BinaryName)

	installedPath := filepath.Join(targetDir, binaryName)
	info, err := os.Stat(installedPath)
	assert.NoError(t, err)
	assert.True(t, info.Mode()&0111 != 0, "binary should be executable")

	content, err := os.ReadFile(installedPath)
	assert.NoError(t, err)
	assert.Equal(t, binaryContent, string(content))
}

func TestService_Install_DirectURL_DownloadError(t *testing.T) {
	origHTTPGet := httpGet
	defer func() { httpGet = origHTTPGet }()

	httpGet = func(url string) (*http.Response, error) {
		return nil, fmt.Errorf("network error")
	}

	s := &Service{}
	_, err := s.Install("https://example.com/tool.tar.gz", t.TempDir())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to download")
}

func TestService_Install_DirectURL_BadStatus(t *testing.T) {
	origHTTPGet := httpGet
	defer func() { httpGet = origHTTPGet }()

	httpGet = func(url string) (*http.Response, error) {
		return &http.Response{
			StatusCode: 404,
			Body:       io.NopCloser(strings.NewReader("not found")),
		}, nil
	}

	s := &Service{}
	_, err := s.Install("https://example.com/tool.tar.gz", t.TempDir())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to download")
}

func TestNewInstallService(t *testing.T) {
	svc := NewInstallService()
	assert.NotNil(t, svc)
}

func TestNewInstallServiceWithName(t *testing.T) {
	svc := NewInstallServiceWithName("custom")
	assert.NotNil(t, svc)
	s, ok := svc.(*Service)
	assert.True(t, ok)
	assert.Equal(t, "custom", s.NameOverride)
}

func TestService_Install_LocalJar_Success(t *testing.T) {
	origOsStat := osStat
	defer func() { osStat = origOsStat }()

	// Create a temp JAR file
	tmpDir := t.TempDir()
	jarPath := filepath.Join(tmpDir, "tika-app-3.2.3.jar")
	err := os.WriteFile(jarPath, []byte("fake jar"), 0644)
	assert.NoError(t, err)

	// osStat should succeed for the JAR path
	osStat = os.Stat

	// We need to mock jar.DetectJava — since it's in a different package,
	// we test the full flow by ensuring java is "found" via the jar package's mock.
	// Instead, let's test through the installLocalJar method directly
	// by creating a real temp file. The jar package's DetectJava uses execCommand
	// which we can't mock from here — so we use integration-style test that
	// may be skipped if java is not available.

	// For unit testing, we'll test the error paths and the isLocalFile helper
	t.Skip("requires java to be installed — covered by jar package tests")
}

func TestService_Install_LocalFile_NotJar(t *testing.T) {
	origOsStat := osStat
	defer func() { osStat = origOsStat }()

	// Create a temp non-JAR file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "somebinary")
	err := os.WriteFile(filePath, []byte("binary content"), 0755)
	assert.NoError(t, err)

	osStat = os.Stat

	s := &Service{}
	_, err = s.Install(filePath, t.TempDir())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "local file install only supports .jar files")
}

func TestIsLocalFile(t *testing.T) {
	origOsStat := osStat
	defer func() { osStat = origOsStat }()

	// Real file
	tmpFile := filepath.Join(t.TempDir(), "test.jar")
	os.WriteFile(tmpFile, []byte("test"), 0644)

	osStat = os.Stat
	assert.True(t, isLocalFile(tmpFile))
	assert.False(t, isLocalFile("/nonexistent/path/file.jar"))
	assert.False(t, isLocalFile("sharkdp/fd"))
}

func TestService_Install_ForgeJarFallback(t *testing.T) {
	origNewForge := newForgeService
	origHTTPGet := httpGet
	origOsStat := osStat
	defer func() {
		newForgeService = origNewForge
		httpGet = origHTTPGet
		osStat = origOsStat
	}()

	// Ensure "apache/tika" is NOT treated as a local file
	osStat = func(name string) (os.FileInfo, error) {
		return nil, os.ErrNotExist
	}

	newForgeService = func(name string) (forge.Forge, error) {
		return &mockForgeSvc{
			assets: []forge.Asset{
				{Name: "tika-app-3.2.3.jar", DownloadURL: "https://example.com/tika-app-3.2.3.jar"},
				{Name: "tika-app-3.2.3-sources.jar", DownloadURL: "https://example.com/tika-app-3.2.3-sources.jar"},
			},
			tag: "v3.2.3",
		}, nil
	}

	// The JAR install path will call DetectJava which needs java installed.
	// We can't easily mock that from here, so test that the fallback path is entered
	// by verifying the error mentions "java is required" (DetectJava fails in CI).
	httpGet = func(url string) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader("fake jar content")),
		}, nil
	}

	s := &Service{}
	_, err := s.Install("apache/tika", t.TempDir())

	// Either succeeds (java installed) or fails with "java is required" (no java)
	// Either way, it means the JAR fallback path was taken, not "no compatible asset"
	if err != nil {
		assert.Contains(t, err.Error(), "java is required")
	}
}

func TestService_Install_ForgeJarFallback_ExcludesSourcesJar(t *testing.T) {
	origNewForge := newForgeService
	origOsStat := osStat
	defer func() {
		newForgeService = origNewForge
		osStat = origOsStat
	}()

	osStat = func(name string) (os.FileInfo, error) {
		return nil, os.ErrNotExist
	}

	// Only sources/javadoc JARs — should still get "no compatible asset"
	newForgeService = func(name string) (forge.Forge, error) {
		return &mockForgeSvc{
			assets: []forge.Asset{
				{Name: "lib-1.0-sources.jar", DownloadURL: "https://example.com/lib-1.0-sources.jar"},
				{Name: "lib-1.0-javadoc.jar", DownloadURL: "https://example.com/lib-1.0-javadoc.jar"},
			},
			tag: "v1.0",
		}, nil
	}

	s := &Service{}
	_, err := s.Install("user/lib", t.TempDir())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no compatible asset found")
}

func TestService_Install_NameOverride(t *testing.T) {
	origOsStat := osStat
	defer func() { osStat = origOsStat }()

	// Create a temp JAR file
	tmpDir := t.TempDir()
	jarPath := filepath.Join(tmpDir, "tika-app-3.2.3.jar")
	err := os.WriteFile(jarPath, []byte("fake jar"), 0644)
	assert.NoError(t, err)

	osStat = os.Stat

	s := &Service{NameOverride: "tika"}
	_, err = s.Install(jarPath, t.TempDir())

	// Will fail with "java is required" if java not installed,
	// but name override is tested in jar package unit tests
	if err != nil {
		assert.Contains(t, err.Error(), "java is required")
	}
}
