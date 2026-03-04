package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AmadlaOrg/lay/binary/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupBinaryRemoveTest(t *testing.T) (string, *manifest.FileStore) {
	t.Helper()
	binDir := filepath.Join(t.TempDir(), "bin")
	require.NoError(t, os.MkdirAll(binDir, 0755))

	// Create a fake binary
	binPath := filepath.Join(binDir, "fd")
	require.NoError(t, os.WriteFile(binPath, []byte("binary-content"), 0755))

	// Create a manifest with the entry
	storePath := filepath.Join(t.TempDir(), "installed.json")
	store := manifest.NewFileStoreAt(storePath)
	m := &manifest.Manifest{}
	m.Add(manifest.Entry{
		Name:        "fd",
		Source:      "sharkdp/fd",
		Version:     "9.0.0",
		Path:        binPath,
		Checksum:    "abc123",
		HashAlgo:    "sha256",
		InstalledAt: "2025-01-01T00:00:00Z",
	})
	require.NoError(t, store.Save(m))

	return binPath, store
}

func TestRunBinaryRemove_Success(t *testing.T) {
	binPath, store := setupBinaryRemoveTest(t)

	origStore := binaryRemoveManifestStore
	defer func() { binaryRemoveManifestStore = origStore }()
	binaryRemoveManifestStore = func() (manifest.Store, error) { return store, nil }

	err := runBinaryRemove([]string{"fd"}, newTestOut())
	require.NoError(t, err)

	// Binary should be gone
	_, err = os.Stat(binPath)
	assert.True(t, os.IsNotExist(err))

	// Manifest should no longer have the entry
	m, err := store.Load()
	require.NoError(t, err)
	assert.Nil(t, m.Find("fd"))
}

func TestRunBinaryRemove_NotFound(t *testing.T) {
	_, store := setupBinaryRemoveTest(t)

	origStore := binaryRemoveManifestStore
	defer func() { binaryRemoveManifestStore = origStore }()
	binaryRemoveManifestStore = func() (manifest.Store, error) { return store, nil }

	// Should not error, just print "not found"
	err := runBinaryRemove([]string{"nonexistent"}, newTestOut())
	assert.NoError(t, err)
}

func TestRunBinaryRemove_DryRun(t *testing.T) {
	binPath, store := setupBinaryRemoveTest(t)

	origStore := binaryRemoveManifestStore
	origDryRun := flagDryRun
	defer func() {
		binaryRemoveManifestStore = origStore
		flagDryRun = origDryRun
	}()
	binaryRemoveManifestStore = func() (manifest.Store, error) { return store, nil }
	flagDryRun = true

	err := runBinaryRemove([]string{"fd"}, newTestOut())
	require.NoError(t, err)

	// Binary should still exist (dry run)
	_, err = os.Stat(binPath)
	assert.NoError(t, err)

	// Manifest should still have the entry
	m, err := store.Load()
	require.NoError(t, err)
	assert.NotNil(t, m.Find("fd"))
}

func TestRunBinaryRemove_JarLauncher(t *testing.T) {
	binDir := filepath.Join(t.TempDir(), "bin")
	jarDir := filepath.Join(t.TempDir(), "jars")
	require.NoError(t, os.MkdirAll(binDir, 0755))
	require.NoError(t, os.MkdirAll(jarDir, 0755))

	// Create a JAR file
	jarPath := filepath.Join(jarDir, "tika-app-3.2.3.jar")
	require.NoError(t, os.WriteFile(jarPath, []byte("fake-jar-content"), 0644))

	// Create a launcher script
	launcherPath := filepath.Join(binDir, "tika-app")
	launcherContent := "#!/bin/sh\nexec java -jar \"" + jarPath + "\" \"$@\"\n"
	require.NoError(t, os.WriteFile(launcherPath, []byte(launcherContent), 0755))

	// Create manifest
	storePath := filepath.Join(t.TempDir(), "installed.json")
	store := manifest.NewFileStoreAt(storePath)
	m := &manifest.Manifest{}
	m.Add(manifest.Entry{
		Name:    "tika-app",
		Source:  "apache/tika",
		Version: "3.2.3",
		Path:    launcherPath,
	})
	require.NoError(t, store.Save(m))

	origStore := binaryRemoveManifestStore
	defer func() { binaryRemoveManifestStore = origStore }()
	binaryRemoveManifestStore = func() (manifest.Store, error) { return store, nil }

	err := runBinaryRemove([]string{"tika-app"}, newTestOut())
	require.NoError(t, err)

	// Both launcher and JAR should be gone
	_, err = os.Stat(launcherPath)
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(jarPath)
	assert.True(t, os.IsNotExist(err))

	// Manifest should be empty
	m, err = store.Load()
	require.NoError(t, err)
	assert.Nil(t, m.Find("tika-app"))
}

func TestRunBinaryRemove_Multiple(t *testing.T) {
	binDir := filepath.Join(t.TempDir(), "bin")
	require.NoError(t, os.MkdirAll(binDir, 0755))

	// Create two binaries
	fdPath := filepath.Join(binDir, "fd")
	batPath := filepath.Join(binDir, "bat")
	require.NoError(t, os.WriteFile(fdPath, []byte("fd-binary"), 0755))
	require.NoError(t, os.WriteFile(batPath, []byte("bat-binary"), 0755))

	storePath := filepath.Join(t.TempDir(), "installed.json")
	store := manifest.NewFileStoreAt(storePath)
	m := &manifest.Manifest{}
	m.Add(manifest.Entry{Name: "fd", Path: fdPath})
	m.Add(manifest.Entry{Name: "bat", Path: batPath})
	require.NoError(t, store.Save(m))

	origStore := binaryRemoveManifestStore
	defer func() { binaryRemoveManifestStore = origStore }()
	binaryRemoveManifestStore = func() (manifest.Store, error) { return store, nil }

	err := runBinaryRemove([]string{"fd", "bat"}, newTestOut())
	require.NoError(t, err)

	// Both should be gone
	_, err = os.Stat(fdPath)
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(batPath)
	assert.True(t, os.IsNotExist(err))

	m, err = store.Load()
	require.NoError(t, err)
	assert.Nil(t, m.Find("fd"))
	assert.Nil(t, m.Find("bat"))
}
