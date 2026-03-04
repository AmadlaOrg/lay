package cmd

import (
	"errors"
	"testing"

	"github.com/AmadlaOrg/lay/binary/install"
	"github.com/AmadlaOrg/lay/binary/manifest"
	"github.com/AmadlaOrg/lay/binary/target"
	"github.com/stretchr/testify/assert"
)

func TestBinaryInstallCmd_RequiresExactlyOneArg(t *testing.T) {
	err := binaryInstallCmd.Args(binaryInstallCmd, []string{})
	assert.Error(t, err)
}

func TestBinaryInstallCmd_AcceptsOneArg(t *testing.T) {
	err := binaryInstallCmd.Args(binaryInstallCmd, []string{"sharkdp/fd"})
	assert.NoError(t, err)
}

func TestBinaryInstallCmd_RejectsTwoArgs(t *testing.T) {
	err := binaryInstallCmd.Args(binaryInstallCmd, []string{"sharkdp/fd", "extra"})
	assert.Error(t, err)
}

type mockTarget struct {
	resolveDir string
	resolveErr error
	ensureErr  error
}

func (m *mockTarget) Resolve(flagOverride string) (string, error) {
	return m.resolveDir, m.resolveErr
}

func (m *mockTarget) Ensure(path string) error {
	return m.ensureErr
}

type mockInstaller struct {
	result     *install.InstallResult
	installErr error
}

func (m *mockInstaller) Install(source string, targetDir string) (*install.InstallResult, error) {
	return m.result, m.installErr
}

type mockStore struct {
	manifest *manifest.Manifest
	loadErr  error
	saveErr  error
	saved    *manifest.Manifest
}

func (m *mockStore) Load() (*manifest.Manifest, error) {
	if m.manifest != nil {
		return m.manifest, m.loadErr
	}
	return &manifest.Manifest{}, m.loadErr
}

func (m *mockStore) Save(man *manifest.Manifest) error {
	m.saved = man
	return m.saveErr
}

func (m *mockStore) Path() string { return "/tmp/test-manifest.json" }

func TestRunBinaryInstall_Success(t *testing.T) {
	origStore := newManifestStore
	defer func() { newManifestStore = origStore }()

	store := &mockStore{}
	newManifestStore = func() (manifest.Store, error) { return store, nil }

	tgt := &mockTarget{resolveDir: t.TempDir()}
	inst := &mockInstaller{result: &install.InstallResult{
		BinaryName: "fd",
		Version:    "v9.0.0",
		Path:       "/usr/local/bin/fd",
		Checksum:   "abc123",
		HashAlgo:   "sha256",
	}}

	err := runBinaryInstall(tgt, inst, "sharkdp/fd", "", newTestOut())
	assert.NoError(t, err)
	assert.NotNil(t, store.saved)
	assert.Len(t, store.saved.Entries, 1)
	assert.Equal(t, "fd", store.saved.Entries[0].Name)
}

func TestRunBinaryInstall_ResolveError(t *testing.T) {
	tgt := &mockTarget{resolveErr: errors.New("home dir error")}
	inst := &mockInstaller{}

	err := runBinaryInstall(tgt, inst, "sharkdp/fd", "", newTestOut())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "resolving target directory")
}

func TestRunBinaryInstall_EnsureError(t *testing.T) {
	tgt := &mockTarget{resolveDir: "/bad/path", ensureErr: errors.New("permission denied")}
	inst := &mockInstaller{}

	err := runBinaryInstall(tgt, inst, "sharkdp/fd", "", newTestOut())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "creating target directory")
}

func TestRunBinaryInstall_InstallError(t *testing.T) {
	tgt := &mockTarget{resolveDir: "/usr/local/bin"}
	inst := &mockInstaller{installErr: errors.New("download failed")}

	origStore := newManifestStore
	defer func() { newManifestStore = origStore }()
	newManifestStore = func() (manifest.Store, error) { return &mockStore{}, nil }

	err := runBinaryInstall(tgt, inst, "sharkdp/fd", "", newTestOut())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "download failed")
}

func TestRunBinaryInstall_DryRun(t *testing.T) {
	origDryRun := flagDryRun
	defer func() { flagDryRun = origDryRun }()
	flagDryRun = true

	tgt := &mockTarget{resolveDir: t.TempDir()}
	inst := &mockInstaller{installErr: errors.New("should not be called")}

	err := runBinaryInstall(tgt, inst, "sharkdp/fd", "", newTestOut())
	assert.NoError(t, err)
}

func TestRunBinaryInstall_ManifestStoreError(t *testing.T) {
	origStore := newManifestStore
	defer func() { newManifestStore = origStore }()
	newManifestStore = func() (manifest.Store, error) { return nil, errors.New("store error") }

	tgt := &mockTarget{resolveDir: t.TempDir()}
	inst := &mockInstaller{result: &install.InstallResult{
		BinaryName: "fd",
		Version:    "v9.0.0",
		Path:       "/usr/local/bin/fd",
	}}

	err := runBinaryInstall(tgt, inst, "sharkdp/fd", "", newTestOut())
	assert.NoError(t, err, "manifest store error should be non-fatal")
}

func TestBinaryInstallCmd_Run_Error(t *testing.T) {
	origTarget := newTargetService
	origExit := osExit
	defer func() {
		newTargetService = origTarget
		osExit = origExit
	}()

	newTargetService = func() target.Target {
		return &mockTarget{resolveErr: errors.New("home error")}
	}

	var exitCode int
	osExit = func(code int) { exitCode = code }

	binaryInstallCmd.Run(binaryInstallCmd, []string{"sharkdp/fd"})
	assert.Equal(t, 1, exitCode)
}

func TestBinaryInstallCmd_Run_Success(t *testing.T) {
	origTarget := newTargetService
	origInstall := newInstallService
	origExit := osExit
	origStore := newManifestStore
	defer func() {
		newTargetService = origTarget
		newInstallService = origInstall
		osExit = origExit
		newManifestStore = origStore
	}()

	newTargetService = func() target.Target {
		return &mockTarget{resolveDir: t.TempDir()}
	}
	newInstallService = func() install.Installer {
		return &mockInstaller{result: &install.InstallResult{
			BinaryName: "fd",
			Version:    "v9.0.0",
			Path:       "/tmp/fd",
		}}
	}
	newManifestStore = func() (manifest.Store, error) { return &mockStore{}, nil }

	var exitCode int = -1
	osExit = func(code int) { exitCode = code }

	binaryInstallCmd.Run(binaryInstallCmd, []string{"sharkdp/fd"})
	assert.Equal(t, -1, exitCode, "should not call osExit on success")
}

func TestBinaryInstallCmd_Run_WithNameFlag(t *testing.T) {
	origTarget := newTargetService
	origInstallWithName := newInstallServiceWithName
	origExit := osExit
	origStore := newManifestStore
	origNameFlag := binaryNameFlag
	defer func() {
		newTargetService = origTarget
		newInstallServiceWithName = origInstallWithName
		osExit = origExit
		newManifestStore = origStore
		binaryNameFlag = origNameFlag
	}()

	binaryNameFlag = "tika"

	var calledWithName string
	newInstallServiceWithName = func(name string) install.Installer {
		calledWithName = name
		return &mockInstaller{result: &install.InstallResult{
			BinaryName: "tika",
			Version:    "",
			Path:       "/tmp/tika",
		}}
	}
	newTargetService = func() target.Target {
		return &mockTarget{resolveDir: t.TempDir()}
	}
	newManifestStore = func() (manifest.Store, error) { return &mockStore{}, nil }

	var exitCode int = -1
	osExit = func(code int) { exitCode = code }

	binaryInstallCmd.Run(binaryInstallCmd, []string{"/tmp/tika-app-3.2.3.jar"})
	assert.Equal(t, -1, exitCode, "should not call osExit on success")
	assert.Equal(t, "tika", calledWithName)
}

func TestRunBinaryInstall_JarLocal(t *testing.T) {
	origStore := newManifestStore
	defer func() { newManifestStore = origStore }()

	store := &mockStore{}
	newManifestStore = func() (manifest.Store, error) { return store, nil }

	tgt := &mockTarget{resolveDir: t.TempDir()}
	inst := &mockInstaller{result: &install.InstallResult{
		BinaryName: "tika-app",
		Version:    "",
		Path:       "/home/user/.local/bin/tika-app",
		Checksum:   "def456",
		HashAlgo:   "sha256",
	}}

	err := runBinaryInstall(tgt, inst, "/tmp/tika-app-3.2.3.jar", "", newTestOut())
	assert.NoError(t, err)
	assert.NotNil(t, store.saved)
	assert.Len(t, store.saved.Entries, 1)
	assert.Equal(t, "tika-app", store.saved.Entries[0].Name)
}
