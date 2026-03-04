package manifest

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManifest_Add(t *testing.T) {
	m := &Manifest{}
	m.Add(Entry{Name: "fd", Version: "9.0.0"})
	assert.Len(t, m.Entries, 1)
	assert.Equal(t, "fd", m.Entries[0].Name)
}

func TestManifest_Find(t *testing.T) {
	m := &Manifest{
		Entries: []Entry{
			{Name: "fd", Version: "9.0.0"},
			{Name: "rg", Version: "14.0.0"},
		},
	}

	e := m.Find("rg")
	assert.NotNil(t, e)
	assert.Equal(t, "14.0.0", e.Version)

	e = m.Find("nonexistent")
	assert.Nil(t, e)
}

func TestManifest_Remove(t *testing.T) {
	m := &Manifest{
		Entries: []Entry{
			{Name: "fd", Version: "9.0.0"},
			{Name: "rg", Version: "14.0.0"},
		},
	}

	ok := m.Remove("fd")
	assert.True(t, ok)
	assert.Len(t, m.Entries, 1)
	assert.Equal(t, "rg", m.Entries[0].Name)

	ok = m.Remove("nonexistent")
	assert.False(t, ok)
}

func TestManifest_Update(t *testing.T) {
	m := &Manifest{
		Entries: []Entry{
			{Name: "fd", Version: "9.0.0"},
		},
	}

	ok := m.Update("fd", "10.0.0", "abc123")
	assert.True(t, ok)
	assert.Equal(t, "10.0.0", m.Find("fd").Version)
	assert.Equal(t, "abc123", m.Find("fd").Checksum)
	assert.NotEmpty(t, m.Find("fd").InstalledAt)

	ok = m.Update("nonexistent", "1.0", "xxx")
	assert.False(t, ok)
}

func TestFileStore_LoadNonExistent(t *testing.T) {
	store := NewFileStoreAt(filepath.Join(t.TempDir(), "installed.json"))
	m, err := store.Load()
	assert.NoError(t, err)
	assert.NotNil(t, m)
	assert.Empty(t, m.Entries)
}

func TestFileStore_SaveAndLoad(t *testing.T) {
	store := NewFileStoreAt(filepath.Join(t.TempDir(), "installed.json"))

	m := &Manifest{
		Entries: []Entry{
			{Name: "fd", Source: "sharkdp/fd", Version: "9.0.0", Path: "/usr/local/bin/fd"},
			{Name: "rg", Source: "BurntSushi/ripgrep", Version: "14.0.0", Path: "/usr/local/bin/rg", Checksum: "abc", HashAlgo: "sha256"},
		},
	}

	err := store.Save(m)
	assert.NoError(t, err)

	loaded, err := store.Load()
	assert.NoError(t, err)
	assert.Len(t, loaded.Entries, 2)
	assert.Equal(t, "fd", loaded.Entries[0].Name)
	assert.Equal(t, "sharkdp/fd", loaded.Entries[0].Source)
	assert.Equal(t, "abc", loaded.Entries[1].Checksum)
}

func TestFileStore_Path(t *testing.T) {
	store := NewFileStoreAt("/tmp/test.json")
	assert.Equal(t, "/tmp/test.json", store.Path())
}

func TestNewFileStore(t *testing.T) {
	origHome := osUserHomeDir
	defer func() { osUserHomeDir = origHome }()

	tmpDir := t.TempDir()
	osUserHomeDir = func() (string, error) { return tmpDir, nil }

	store, err := NewFileStore()
	assert.NoError(t, err)
	assert.Contains(t, store.Path(), "installed.json")
}
