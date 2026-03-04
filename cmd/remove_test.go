package cmd

import (
	"errors"
	"testing"

	pm "github.com/AmadlaOrg/lay/package_manager"
	"github.com/stretchr/testify/assert"
)

func TestRemoveCmd_RequiresArgs(t *testing.T) {
	err := removeCmd.Args(removeCmd, []string{})
	assert.Error(t, err)
}

func TestRemoveCmd_AcceptsArgs(t *testing.T) {
	err := removeCmd.Args(removeCmd, []string{"curl"})
	assert.NoError(t, err)
}

func TestRunRemove_Success(t *testing.T) {
	detector := &mockDetector{name: "apt"}
	mgr := &mockManager{name: "apt", isInstalledMap: map[string]bool{"curl": true}}
	managerFn := func(name string) (pm.Manager, error) { return mgr, nil }

	err := runRemove(detector, managerFn, "", []string{"curl"}, newTestOut())
	assert.NoError(t, err)
	assert.Equal(t, []string{"curl"}, mgr.removedPkgs)
}

func TestRunRemove_DetectError(t *testing.T) {
	detector := &mockDetector{err: errors.New("no pm found")}
	managerFn := func(name string) (pm.Manager, error) { return nil, nil }

	err := runRemove(detector, managerFn, "", []string{"curl"}, newTestOut())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no pm found")
}

func TestRunRemove_ManagerError(t *testing.T) {
	detector := &mockDetector{name: "bad"}
	managerFn := func(name string) (pm.Manager, error) { return nil, errors.New("unsupported") }

	err := runRemove(detector, managerFn, "", []string{"curl"}, newTestOut())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported")
}

func TestRunRemove_RemoveError(t *testing.T) {
	detector := &mockDetector{name: "apt"}
	mgr := &mockManager{
		name:           "apt",
		removeErr:      errors.New("remove failed"),
		isInstalledMap: map[string]bool{"badpkg": true},
	}
	managerFn := func(name string) (pm.Manager, error) { return mgr, nil }

	err := runRemove(detector, managerFn, "", []string{"badpkg"}, newTestOut())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "remove failed")
}

func TestRunRemove_SkipsNotInstalled(t *testing.T) {
	detector := &mockDetector{name: "apt"}
	mgr := &mockManager{
		name:           "apt",
		isInstalledMap: map[string]bool{"curl": true, "wget": false},
	}
	managerFn := func(name string) (pm.Manager, error) { return mgr, nil }

	err := runRemove(detector, managerFn, "", []string{"curl", "wget"}, newTestOut())
	assert.NoError(t, err)
	assert.Equal(t, []string{"curl"}, mgr.removedPkgs)
}

func TestRunRemove_NoneInstalled(t *testing.T) {
	detector := &mockDetector{name: "apt"}
	mgr := &mockManager{
		name:           "apt",
		isInstalledMap: map[string]bool{"curl": false},
	}
	managerFn := func(name string) (pm.Manager, error) { return mgr, nil }

	err := runRemove(detector, managerFn, "", []string{"curl"}, newTestOut())
	assert.NoError(t, err)
	assert.Nil(t, mgr.removedPkgs, "should not call Remove when none installed")
}

func TestRunRemove_IsInstalledErrorFallsThrough(t *testing.T) {
	detector := &mockDetector{name: "apt"}
	mgr := &mockManager{
		name:      "apt",
		isInstErr: errors.New("check failed"),
	}
	managerFn := func(name string) (pm.Manager, error) { return mgr, nil }

	err := runRemove(detector, managerFn, "", []string{"curl"}, newTestOut())
	assert.NoError(t, err)
	assert.Equal(t, []string{"curl"}, mgr.removedPkgs, "should remove when IsInstalled errors")
}

func TestRunRemove_DryRun(t *testing.T) {
	origDryRun := flagDryRun
	defer func() { flagDryRun = origDryRun }()
	flagDryRun = true

	detector := &mockDetector{name: "apt"}
	mgr := &mockManager{name: "apt"}
	managerFn := func(name string) (pm.Manager, error) { return mgr, nil }

	err := runRemove(detector, managerFn, "", []string{"curl"}, newTestOut())
	assert.NoError(t, err)
	assert.Nil(t, mgr.removedPkgs, "should not call Remove in dry-run")
}
