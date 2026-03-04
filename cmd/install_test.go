package cmd

import (
	"errors"
	"testing"

	"github.com/AmadlaOrg/lay/output"
	pm "github.com/AmadlaOrg/lay/package_manager"
	"github.com/AmadlaOrg/lay/package_manager/types"
	"github.com/stretchr/testify/assert"
)

func TestInstallCmd_RequiresArgs(t *testing.T) {
	err := installCmd.Args(installCmd, []string{})
	assert.Error(t, err)
}

func TestInstallCmd_AcceptsArgs(t *testing.T) {
	err := installCmd.Args(installCmd, []string{"curl"})
	assert.NoError(t, err)
}

func TestInstallCmd_AcceptsMultipleArgs(t *testing.T) {
	err := installCmd.Args(installCmd, []string{"curl", "wget", "vim"})
	assert.NoError(t, err)
}

type mockDetector struct {
	name string
	err  error
}

func (m *mockDetector) Detect(override string) (string, error) {
	return m.name, m.err
}

type mockManager struct {
	name           string
	installErr     error
	installedPkgs  []string // packages that were passed to Install
	removeErr      error
	removedPkgs    []string // packages that were passed to Remove
	searchRes      []types.SearchResult
	searchErr      error
	updateErr      error
	upgradeErr     error
	upgradedPkgs   []string // packages that were passed to Upgrade
	listRes        []types.PackageInfo
	listErr        error
	isInstalledMap map[string]bool // pkg -> installed
	isInstErr      error
}

func (m *mockManager) Name() string { return m.name }
func (m *mockManager) Install(packages []string) error {
	m.installedPkgs = packages
	return m.installErr
}
func (m *mockManager) Remove(packages []string) error {
	m.removedPkgs = packages
	return m.removeErr
}
func (m *mockManager) Search(query string) ([]types.SearchResult, error) {
	return m.searchRes, m.searchErr
}
func (m *mockManager) Update() error { return m.updateErr }
func (m *mockManager) Upgrade(packages []string) error {
	m.upgradedPkgs = packages
	return m.upgradeErr
}
func (m *mockManager) List() ([]types.PackageInfo, error) { return m.listRes, m.listErr }
func (m *mockManager) IsInstalled(pkg string) (bool, error) {
	if m.isInstalledMap != nil {
		return m.isInstalledMap[pkg], m.isInstErr
	}
	return false, m.isInstErr
}

func newTestOut() *output.Writer {
	return output.NewWriter(&discard{}, output.ModeNormal)
}

type discard struct{}

func (d *discard) Write(p []byte) (int, error) { return len(p), nil }

func TestRunInstall_Success(t *testing.T) {
	detector := &mockDetector{name: "apt"}
	mgr := &mockManager{name: "apt"}
	managerFn := func(name string) (pm.Manager, error) { return mgr, nil }

	err := runInstall(detector, managerFn, "", []string{"curl"}, newTestOut())
	assert.NoError(t, err)
}

func TestRunInstall_DetectError(t *testing.T) {
	detector := &mockDetector{err: errors.New("no pm found")}
	managerFn := func(name string) (pm.Manager, error) { return nil, nil }

	err := runInstall(detector, managerFn, "", []string{"curl"}, newTestOut())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no pm found")
}

func TestRunInstall_ManagerError(t *testing.T) {
	detector := &mockDetector{name: "bad"}
	managerFn := func(name string) (pm.Manager, error) { return nil, errors.New("unsupported") }

	err := runInstall(detector, managerFn, "", []string{"curl"}, newTestOut())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported")
}

func TestRunInstall_InstallError(t *testing.T) {
	detector := &mockDetector{name: "apt"}
	mgr := &mockManager{name: "apt", installErr: errors.New("install failed")}
	managerFn := func(name string) (pm.Manager, error) { return mgr, nil }

	err := runInstall(detector, managerFn, "", []string{"badpkg"}, newTestOut())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "install failed")
}

func TestRunInstall_IdempotencySkipsInstalled(t *testing.T) {
	detector := &mockDetector{name: "apt"}
	mgr := &mockManager{
		name:           "apt",
		isInstalledMap: map[string]bool{"curl": true, "wget": false},
	}
	managerFn := func(name string) (pm.Manager, error) { return mgr, nil }

	err := runInstall(detector, managerFn, "", []string{"curl", "wget"}, newTestOut())
	assert.NoError(t, err)
	assert.Equal(t, []string{"wget"}, mgr.installedPkgs)
}

func TestRunInstall_AllAlreadyInstalled(t *testing.T) {
	detector := &mockDetector{name: "apt"}
	mgr := &mockManager{
		name:           "apt",
		isInstalledMap: map[string]bool{"curl": true, "wget": true},
	}
	managerFn := func(name string) (pm.Manager, error) { return mgr, nil }

	err := runInstall(detector, managerFn, "", []string{"curl", "wget"}, newTestOut())
	assert.NoError(t, err)
	assert.Nil(t, mgr.installedPkgs, "should not call Install when all installed")
}

func TestRunInstall_IsInstalledErrorFallsThrough(t *testing.T) {
	detector := &mockDetector{name: "apt"}
	mgr := &mockManager{
		name:      "apt",
		isInstErr: errors.New("check failed"),
	}
	managerFn := func(name string) (pm.Manager, error) { return mgr, nil }

	err := runInstall(detector, managerFn, "", []string{"curl"}, newTestOut())
	assert.NoError(t, err)
	assert.Equal(t, []string{"curl"}, mgr.installedPkgs, "should install when IsInstalled errors")
}

func TestRunInstall_DryRun(t *testing.T) {
	origDryRun := flagDryRun
	defer func() { flagDryRun = origDryRun }()
	flagDryRun = true

	detector := &mockDetector{name: "apt"}
	mgr := &mockManager{name: "apt"}
	managerFn := func(name string) (pm.Manager, error) { return mgr, nil }

	err := runInstall(detector, managerFn, "", []string{"curl"}, newTestOut())
	assert.NoError(t, err)
	assert.Nil(t, mgr.installedPkgs, "should not call Install in dry-run")
}

// Integration test: exercises the cobra Run closure with mocked dependencies
func TestInstallCmd_Run_Error(t *testing.T) {
	origDetector := newDetector
	origExit := osExit
	defer func() {
		newDetector = origDetector
		osExit = origExit
	}()

	newDetector = func() pm.Detector {
		return &mockDetector{err: errors.New("no pm")}
	}

	var exitCode int
	osExit = func(code int) { exitCode = code }

	installCmd.Run(installCmd, []string{"curl"})
	assert.Equal(t, 1, exitCode)
}

func TestInstallCmd_Run_Success(t *testing.T) {
	origDetector := newDetector
	origManager := newManagerByName
	origExit := osExit
	defer func() {
		newDetector = origDetector
		newManagerByName = origManager
		osExit = origExit
	}()

	newDetector = func() pm.Detector {
		return &mockDetector{name: "apt"}
	}
	newManagerByName = func(name string) (pm.Manager, error) {
		return &mockManager{name: "apt"}, nil
	}

	var exitCode int = -1
	osExit = func(code int) { exitCode = code }

	installCmd.Run(installCmd, []string{"curl"})
	assert.Equal(t, -1, exitCode, "should not call osExit on success")
}
