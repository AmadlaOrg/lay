package cmd

import (
	"errors"
	"testing"

	"github.com/AmadlaOrg/lay/binary/install"
	"github.com/AmadlaOrg/lay/binary/manifest"
	"github.com/stretchr/testify/assert"
)

func TestBinaryUpdateCmd_NoArgs(t *testing.T) {
	err := binaryUpdateCmd.Args(binaryUpdateCmd, []string{})
	assert.NoError(t, err)
}

func TestBinaryUpdateCmd_AcceptsArgs(t *testing.T) {
	err := binaryUpdateCmd.Args(binaryUpdateCmd, []string{"fd", "rg"})
	assert.NoError(t, err)
}

func TestRunBinaryUpdate_Empty(t *testing.T) {
	store := &mockStore{}
	inst := &mockInstaller{}

	err := runBinaryUpdate(store, inst, nil, newTestOut())
	assert.NoError(t, err)
}

func TestRunBinaryUpdate_UpdateAvailable(t *testing.T) {
	store := &mockStore{
		manifest: &manifest.Manifest{
			Entries: []manifest.Entry{
				{Name: "fd", Version: "v8.0.0", Source: "sharkdp/fd", Path: "/usr/local/bin/fd"},
			},
		},
	}
	inst := &mockInstaller{result: &install.InstallResult{
		BinaryName: "fd",
		Version:    "v9.0.0",
		Path:       "/usr/local/bin/fd",
		Checksum:   "newchecksum",
	}}

	err := runBinaryUpdate(store, inst, nil, newTestOut())
	assert.NoError(t, err)
	assert.NotNil(t, store.saved)
	e := store.saved.Find("fd")
	assert.NotNil(t, e)
	assert.Equal(t, "v9.0.0", e.Version)
}

func TestRunBinaryUpdate_AlreadyLatest(t *testing.T) {
	store := &mockStore{
		manifest: &manifest.Manifest{
			Entries: []manifest.Entry{
				{Name: "fd", Version: "v9.0.0", Source: "sharkdp/fd", Path: "/usr/local/bin/fd"},
			},
		},
	}
	inst := &mockInstaller{result: &install.InstallResult{
		BinaryName: "fd",
		Version:    "v9.0.0",
		Path:       "/usr/local/bin/fd",
	}}

	err := runBinaryUpdate(store, inst, nil, newTestOut())
	assert.NoError(t, err)
}

func TestRunBinaryUpdate_EntryNotFound(t *testing.T) {
	store := &mockStore{
		manifest: &manifest.Manifest{
			Entries: []manifest.Entry{
				{Name: "fd", Version: "v9.0.0", Source: "sharkdp/fd", Path: "/usr/local/bin/fd"},
			},
		},
	}
	inst := &mockInstaller{}

	err := runBinaryUpdate(store, inst, []string{"nonexistent"}, newTestOut())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "binary not found in manifest")
}

func TestRunBinaryUpdate_DryRun(t *testing.T) {
	origDryRun := flagDryRun
	defer func() { flagDryRun = origDryRun }()
	flagDryRun = true

	store := &mockStore{
		manifest: &manifest.Manifest{
			Entries: []manifest.Entry{
				{Name: "fd", Version: "v8.0.0", Source: "sharkdp/fd", Path: "/usr/local/bin/fd"},
			},
		},
	}
	inst := &mockInstaller{installErr: errors.New("should not be called")}

	err := runBinaryUpdate(store, inst, nil, newTestOut())
	assert.NoError(t, err)
}

func TestRunBinaryUpdate_LoadError(t *testing.T) {
	store := &mockStore{loadErr: errors.New("load error")}
	inst := &mockInstaller{}

	err := runBinaryUpdate(store, inst, nil, newTestOut())
	assert.Error(t, err)
}
