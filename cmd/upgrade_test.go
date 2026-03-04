package cmd

import (
	"errors"
	"testing"

	pm "github.com/AmadlaOrg/lay/package_manager"
	"github.com/stretchr/testify/assert"
)

func TestUpgradeCmd_NoArgs(t *testing.T) {
	err := upgradeCmd.Args(upgradeCmd, []string{})
	assert.NoError(t, err)
}

func TestUpgradeCmd_AcceptsArgs(t *testing.T) {
	err := upgradeCmd.Args(upgradeCmd, []string{"curl", "wget"})
	assert.NoError(t, err)
}

func TestRunUpgrade_SuccessAll(t *testing.T) {
	detector := &mockDetector{name: "apt"}
	mgr := &mockManager{name: "apt"}
	managerFn := func(name string) (pm.Manager, error) { return mgr, nil }

	err := runUpgrade(detector, managerFn, "", nil, newTestOut())
	assert.NoError(t, err)
	assert.Nil(t, mgr.upgradedPkgs)
}

func TestRunUpgrade_SuccessSpecific(t *testing.T) {
	detector := &mockDetector{name: "apt"}
	mgr := &mockManager{name: "apt"}
	managerFn := func(name string) (pm.Manager, error) { return mgr, nil }

	err := runUpgrade(detector, managerFn, "", []string{"curl", "wget"}, newTestOut())
	assert.NoError(t, err)
	assert.Equal(t, []string{"curl", "wget"}, mgr.upgradedPkgs)
}

func TestRunUpgrade_DetectError(t *testing.T) {
	detector := &mockDetector{err: errors.New("no pm found")}
	managerFn := func(name string) (pm.Manager, error) { return nil, nil }

	err := runUpgrade(detector, managerFn, "", nil, newTestOut())
	assert.Error(t, err)
}

func TestRunUpgrade_UpgradeError(t *testing.T) {
	detector := &mockDetector{name: "apt"}
	mgr := &mockManager{name: "apt", upgradeErr: errors.New("upgrade failed")}
	managerFn := func(name string) (pm.Manager, error) { return mgr, nil }

	err := runUpgrade(detector, managerFn, "", nil, newTestOut())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "upgrade failed")
}

func TestRunUpgrade_DryRun(t *testing.T) {
	origDryRun := flagDryRun
	defer func() { flagDryRun = origDryRun }()
	flagDryRun = true

	detector := &mockDetector{name: "apt"}
	mgr := &mockManager{name: "apt", upgradeErr: errors.New("should not be called")}
	managerFn := func(name string) (pm.Manager, error) { return mgr, nil }

	err := runUpgrade(detector, managerFn, "", []string{"curl"}, newTestOut())
	assert.NoError(t, err)
}

func TestUpgradeCmd_Run_Error(t *testing.T) {
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

	upgradeCmd.Run(upgradeCmd, []string{})
	assert.Equal(t, 1, exitCode)
}
