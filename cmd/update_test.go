package cmd

import (
	"errors"
	"testing"

	pm "github.com/AmadlaOrg/lay/package_manager"
	"github.com/stretchr/testify/assert"
)

func TestUpdateCmd_NoArgs(t *testing.T) {
	err := updateCmd.Args(updateCmd, []string{})
	assert.NoError(t, err)
}

func TestRunUpdate_Success(t *testing.T) {
	detector := &mockDetector{name: "apt"}
	mgr := &mockManager{name: "apt"}
	managerFn := func(name string) (pm.Manager, error) { return mgr, nil }

	err := runUpdate(detector, managerFn, "", newTestOut())
	assert.NoError(t, err)
}

func TestRunUpdate_DetectError(t *testing.T) {
	detector := &mockDetector{err: errors.New("no pm found")}
	managerFn := func(name string) (pm.Manager, error) { return nil, nil }

	err := runUpdate(detector, managerFn, "", newTestOut())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no pm found")
}

func TestRunUpdate_ManagerError(t *testing.T) {
	detector := &mockDetector{name: "bad"}
	managerFn := func(name string) (pm.Manager, error) { return nil, errors.New("unsupported") }

	err := runUpdate(detector, managerFn, "", newTestOut())
	assert.Error(t, err)
}

func TestRunUpdate_UpdateError(t *testing.T) {
	detector := &mockDetector{name: "apt"}
	mgr := &mockManager{name: "apt", updateErr: errors.New("update failed")}
	managerFn := func(name string) (pm.Manager, error) { return mgr, nil }

	err := runUpdate(detector, managerFn, "", newTestOut())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update failed")
}

func TestRunUpdate_DryRun(t *testing.T) {
	origDryRun := flagDryRun
	defer func() { flagDryRun = origDryRun }()
	flagDryRun = true

	detector := &mockDetector{name: "apt"}
	mgr := &mockManager{name: "apt", updateErr: errors.New("should not be called")}
	managerFn := func(name string) (pm.Manager, error) { return mgr, nil }

	err := runUpdate(detector, managerFn, "", newTestOut())
	assert.NoError(t, err)
}

func TestUpdateCmd_Run_Error(t *testing.T) {
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

	updateCmd.Run(updateCmd, []string{})
	assert.Equal(t, 1, exitCode)
}
