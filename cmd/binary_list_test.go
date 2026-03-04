package cmd

import (
	"bytes"
	"errors"
	"testing"

	"github.com/AmadlaOrg/lay/binary/manifest"
	"github.com/AmadlaOrg/lay/output"
	"github.com/stretchr/testify/assert"
)

func TestBinaryListCmd_NoArgs(t *testing.T) {
	err := binaryListCmd.Args(binaryListCmd, []string{})
	assert.NoError(t, err)
}

func TestRunBinaryList_Empty(t *testing.T) {
	store := &mockStore{}

	var buf bytes.Buffer
	out := output.NewWriter(&buf, output.ModeNormal)
	err := runBinaryList(store, &buf, out)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "No binaries tracked")
}

func TestRunBinaryList_WithEntries(t *testing.T) {
	store := &mockStore{
		manifest: &manifest.Manifest{
			Entries: []manifest.Entry{
				{Name: "fd", Version: "v9.0.0", Source: "sharkdp/fd", Path: "/usr/local/bin/fd"},
				{Name: "rg", Version: "v14.0.0", Source: "BurntSushi/ripgrep", Path: "/usr/local/bin/rg"},
			},
		},
	}

	var buf bytes.Buffer
	out := output.NewWriter(&buf, output.ModeNormal)
	err := runBinaryList(store, &buf, out)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "fd")
	assert.Contains(t, buf.String(), "rg")
	assert.Contains(t, buf.String(), "v9.0.0")
}

func TestRunBinaryList_LoadError(t *testing.T) {
	store := &mockStore{loadErr: errors.New("load failed")}

	var buf bytes.Buffer
	out := output.NewWriter(&buf, output.ModeNormal)
	err := runBinaryList(store, &buf, out)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "load failed")
}

func TestRunBinaryList_JSONOutput(t *testing.T) {
	store := &mockStore{
		manifest: &manifest.Manifest{
			Entries: []manifest.Entry{
				{Name: "fd", Version: "v9.0.0", Source: "sharkdp/fd", Path: "/usr/local/bin/fd"},
			},
		},
	}

	var buf bytes.Buffer
	out := output.NewWriter(&buf, output.ModeJSON)
	err := runBinaryList(store, &buf, out)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), `"name": "fd"`)
	assert.Contains(t, buf.String(), `"version": "v9.0.0"`)
}
