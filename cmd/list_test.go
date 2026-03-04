package cmd

import (
	"bytes"
	"errors"
	"testing"

	"github.com/AmadlaOrg/lay/output"
	pm "github.com/AmadlaOrg/lay/package_manager"
	"github.com/AmadlaOrg/lay/package_manager/types"
	"github.com/stretchr/testify/assert"
)

func TestListCmd_NoArgs(t *testing.T) {
	err := listCmd.Args(listCmd, []string{})
	assert.NoError(t, err)
}

func TestRunList_Success(t *testing.T) {
	detector := &mockDetector{name: "apt"}
	mgr := &mockManager{
		name: "apt",
		listRes: []types.PackageInfo{
			{Name: "curl", Version: "7.81.0"},
			{Name: "wget", Version: "1.21.2"},
		},
	}
	managerFn := func(name string) (pm.Manager, error) { return mgr, nil }

	var buf bytes.Buffer
	out := output.NewWriter(&buf, output.ModeNormal)
	err := runList(detector, managerFn, "", &buf, out)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "curl")
	assert.Contains(t, buf.String(), "wget")
}

func TestRunList_EmptyResults(t *testing.T) {
	detector := &mockDetector{name: "apt"}
	mgr := &mockManager{name: "apt", listRes: nil}
	managerFn := func(name string) (pm.Manager, error) { return mgr, nil }

	var buf bytes.Buffer
	out := output.NewWriter(&buf, output.ModeNormal)
	err := runList(detector, managerFn, "", &buf, out)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "No installed packages found")
}

func TestRunList_DetectError(t *testing.T) {
	detector := &mockDetector{err: errors.New("no pm found")}
	managerFn := func(name string) (pm.Manager, error) { return nil, nil }

	var buf bytes.Buffer
	out := output.NewWriter(&buf, output.ModeNormal)
	err := runList(detector, managerFn, "", &buf, out)

	assert.Error(t, err)
}

func TestRunList_ListError(t *testing.T) {
	detector := &mockDetector{name: "apt"}
	mgr := &mockManager{name: "apt", listErr: errors.New("list failed")}
	managerFn := func(name string) (pm.Manager, error) { return mgr, nil }

	var buf bytes.Buffer
	out := output.NewWriter(&buf, output.ModeNormal)
	err := runList(detector, managerFn, "", &buf, out)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "list failed")
}

func TestRunList_JSONOutput(t *testing.T) {
	detector := &mockDetector{name: "apt"}
	mgr := &mockManager{
		name: "apt",
		listRes: []types.PackageInfo{
			{Name: "curl", Version: "7.81.0"},
		},
	}
	managerFn := func(name string) (pm.Manager, error) { return mgr, nil }

	var buf bytes.Buffer
	out := output.NewWriter(&buf, output.ModeJSON)
	err := runList(detector, managerFn, "", &buf, out)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), `"name": "curl"`)
	assert.Contains(t, buf.String(), `"version": "7.81.0"`)
}

func TestListCmd_Run_Error(t *testing.T) {
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

	listCmd.Run(listCmd, []string{})
	assert.Equal(t, 1, exitCode)
}
