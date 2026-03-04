package manifest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// For mocking
var osUserHomeDir = os.UserHomeDir

// Entry represents an installed binary
type Entry struct {
	Name        string `json:"name"`
	Source      string `json:"source"`
	Version     string `json:"version"`
	Path        string `json:"path"`
	Checksum    string `json:"checksum,omitempty"`
	HashAlgo    string `json:"hash_algo,omitempty"`
	InstalledAt string `json:"installed_at"`
}

// Manifest holds a list of installed binary entries
type Manifest struct {
	Entries []Entry `json:"entries"`
}

// Add appends an entry to the manifest
func (m *Manifest) Add(e Entry) {
	m.Entries = append(m.Entries, e)
}

// Remove deletes an entry by name. Returns true if found and removed.
func (m *Manifest) Remove(name string) bool {
	for i, e := range m.Entries {
		if e.Name == name {
			m.Entries = append(m.Entries[:i], m.Entries[i+1:]...)
			return true
		}
	}
	return false
}

// Find returns a pointer to the entry with the given name, or nil
func (m *Manifest) Find(name string) *Entry {
	for i := range m.Entries {
		if m.Entries[i].Name == name {
			return &m.Entries[i]
		}
	}
	return nil
}

// Update modifies the version and checksum of an entry. Returns true if found.
func (m *Manifest) Update(name, version, checksum string) bool {
	e := m.Find(name)
	if e == nil {
		return false
	}
	e.Version = version
	e.Checksum = checksum
	e.InstalledAt = time.Now().UTC().Format(time.RFC3339)
	return true
}

// Store defines persistence for the manifest
type Store interface {
	Load() (*Manifest, error)
	Save(m *Manifest) error
	Path() string
}

// FileStore implements Store using a JSON file
type FileStore struct {
	path string
}

// NewFileStore creates a FileStore at ~/.local/share/lay/installed.json
func NewFileStore() (*FileStore, error) {
	home, err := osUserHomeDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(home, ".local", "share", "lay")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &FileStore{path: filepath.Join(dir, "installed.json")}, nil
}

// NewFileStoreAt creates a FileStore at a specific path (for testing)
func NewFileStoreAt(path string) *FileStore {
	return &FileStore{path: path}
}

// Path returns the file path
func (s *FileStore) Path() string {
	return s.path
}

// Load reads the manifest from disk. Returns an empty manifest if the file doesn't exist.
func (s *FileStore) Load() (*Manifest, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Manifest{}, nil
		}
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// Save writes the manifest to disk
func (s *FileStore) Save(m *Manifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}
