//go:build integration

package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func podmanAvailable() bool {
	_, err := exec.LookPath("podman")
	return err == nil
}

func buildLayBinary(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "lay")

	cmd := exec.Command("go", "build", "-o", binPath, "../")
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS=linux", "GOARCH=amd64")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to build lay binary: %s", string(out))

	return binPath
}

func buildContainer(t *testing.T, containerfile, binPath, tag string) {
	t.Helper()
	ctx := filepath.Dir(binPath)

	cmd := exec.Command("podman", "build",
		"-f", containerfile,
		"-t", tag,
		ctx)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to build container %s: %s", tag, string(out))
}

func runInContainer(t *testing.T, tag string, args ...string) (string, error) {
	t.Helper()
	cmdArgs := append([]string{"run", "--rm", tag}, args...)
	cmd := exec.Command("podman", cmdArgs...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestInstall_Ubuntu_Apt(t *testing.T) {
	if !podmanAvailable() {
		t.Skip("podman not available, skipping integration test")
	}

	binPath := buildLayBinary(t)

	containerfile, err := filepath.Abs("testdata/Containerfile.ubuntu")
	require.NoError(t, err)

	tag := "lay-test-ubuntu"
	buildContainer(t, containerfile, binPath, tag)
	defer exec.Command("podman", "rmi", tag).Run()

	out, err := runInContainer(t, tag, "lay", "install", "curl")
	assert.NoError(t, err, "lay install curl failed: %s", out)
	assert.Contains(t, out, "Using package manager: apt")
}

func TestInstall_Fedora_Dnf(t *testing.T) {
	if !podmanAvailable() {
		t.Skip("podman not available, skipping integration test")
	}

	binPath := buildLayBinary(t)

	containerfile, err := filepath.Abs("testdata/Containerfile.fedora")
	require.NoError(t, err)

	tag := "lay-test-fedora"
	buildContainer(t, containerfile, binPath, tag)
	defer exec.Command("podman", "rmi", tag).Run()

	out, err := runInContainer(t, tag, "lay", "install", "curl")
	assert.NoError(t, err, "lay install curl failed: %s", out)
	assert.Contains(t, out, "Using package manager: dnf")
}
