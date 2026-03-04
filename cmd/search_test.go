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

func TestSearchCmd_RequiresExactlyOneArg(t *testing.T) {
	err := searchCmd.Args(searchCmd, []string{})
	assert.Error(t, err)
}

func TestSearchCmd_AcceptsOneArg(t *testing.T) {
	err := searchCmd.Args(searchCmd, []string{"curl"})
	assert.NoError(t, err)
}

func TestSearchCmd_RejectsTwoArgs(t *testing.T) {
	err := searchCmd.Args(searchCmd, []string{"curl", "wget"})
	assert.Error(t, err)
}

func TestRunSearch_Success_WithVersions(t *testing.T) {
	detector := &mockDetector{name: "pacman"}
	mgr := &mockManager{
		name: "pacman",
		searchRes: []types.SearchResult{
			{Name: "curl", Version: "8.5.0", Description: "URL transfer tool"},
		},
	}
	managerFn := func(name string) (pm.Manager, error) { return mgr, nil }

	var buf bytes.Buffer
	out := output.NewWriter(&buf, output.ModeNormal)
	err := runSearch(detector, managerFn, "", "curl", &buf, out)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "curl")
	assert.Contains(t, buf.String(), "8.5.0")
}

func TestRunSearch_Success_WithoutVersions(t *testing.T) {
	detector := &mockDetector{name: "apt"}
	mgr := &mockManager{
		name: "apt",
		searchRes: []types.SearchResult{
			{Name: "curl", Description: "URL transfer tool"},
		},
	}
	managerFn := func(name string) (pm.Manager, error) { return mgr, nil }

	var buf bytes.Buffer
	out := output.NewWriter(&buf, output.ModeNormal)
	err := runSearch(detector, managerFn, "", "curl", &buf, out)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "curl")
	assert.Contains(t, buf.String(), "Search results")
}

func TestRunSearch_EmptyResults(t *testing.T) {
	detector := &mockDetector{name: "apt"}
	mgr := &mockManager{name: "apt", searchRes: nil}
	managerFn := func(name string) (pm.Manager, error) { return mgr, nil }

	var buf bytes.Buffer
	out := output.NewWriter(&buf, output.ModeNormal)
	err := runSearch(detector, managerFn, "", "nonexistent", &buf, out)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "No packages found")
}

func TestRunSearch_DetectError(t *testing.T) {
	detector := &mockDetector{err: errors.New("no pm found")}
	managerFn := func(name string) (pm.Manager, error) { return nil, nil }

	var buf bytes.Buffer
	out := output.NewWriter(&buf, output.ModeNormal)
	err := runSearch(detector, managerFn, "", "curl", &buf, out)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no pm found")
}

func TestRunSearch_ManagerError(t *testing.T) {
	detector := &mockDetector{name: "bad"}
	managerFn := func(name string) (pm.Manager, error) { return nil, errors.New("unsupported") }

	var buf bytes.Buffer
	out := output.NewWriter(&buf, output.ModeNormal)
	err := runSearch(detector, managerFn, "", "curl", &buf, out)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported")
}

func TestRunSearch_SearchError(t *testing.T) {
	detector := &mockDetector{name: "apt"}
	mgr := &mockManager{name: "apt", searchErr: errors.New("search failed")}
	managerFn := func(name string) (pm.Manager, error) { return mgr, nil }

	var buf bytes.Buffer
	out := output.NewWriter(&buf, output.ModeNormal)
	err := runSearch(detector, managerFn, "", "curl", &buf, out)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "search failed")
}

func TestRunSearch_JSONOutput(t *testing.T) {
	detector := &mockDetector{name: "apt"}
	mgr := &mockManager{
		name: "apt",
		searchRes: []types.SearchResult{
			{Name: "curl", Version: "8.5.0", Description: "URL transfer tool"},
		},
	}
	managerFn := func(name string) (pm.Manager, error) { return mgr, nil }

	var buf bytes.Buffer
	out := output.NewWriter(&buf, output.ModeJSON)
	err := runSearch(detector, managerFn, "", "curl", &buf, out)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), `"name": "curl"`)
	assert.Contains(t, buf.String(), `"version": "8.5.0"`)
}

func TestSearchCmd_Run_Error(t *testing.T) {
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

	searchCmd.Run(searchCmd, []string{"curl"})
	assert.Equal(t, 1, exitCode)
}
