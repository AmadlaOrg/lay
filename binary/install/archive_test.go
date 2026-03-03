package install

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectArchiveType(t *testing.T) {
	tests := []struct {
		filename string
		expected ArchiveType
	}{
		{"file.tar.gz", ArchiveTarGz},
		{"file.tgz", ArchiveTarGz},
		{"FILE.TAR.GZ", ArchiveTarGz},
		{"file.tar.xz", ArchiveTarXz},
		{"file.txz", ArchiveTarXz},
		{"file.zip", ArchiveZip},
		{"file.ZIP", ArchiveZip},
		{"binary", ArchiveRaw},
		{"file.deb", ArchiveRaw},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			assert.Equal(t, tt.expected, DetectArchiveType(tt.filename))
		})
	}
}

func TestFindBinary(t *testing.T) {
	t.Run("finds executable with hint", func(t *testing.T) {
		dir := t.TempDir()

		// Create a binary matching the hint
		binPath := filepath.Join(dir, "fd")
		err := os.WriteFile(binPath, []byte("binary"), 0755)
		assert.NoError(t, err)

		// Create a non-matching file
		readmePath := filepath.Join(dir, "README.md")
		err = os.WriteFile(readmePath, []byte("readme"), 0644)
		assert.NoError(t, err)

		result, err := FindBinary(dir, "fd")
		assert.NoError(t, err)
		assert.Equal(t, binPath, result)
	})

	t.Run("finds executable without hint", func(t *testing.T) {
		dir := t.TempDir()

		binPath := filepath.Join(dir, "mybinary")
		err := os.WriteFile(binPath, []byte("binary"), 0755)
		assert.NoError(t, err)

		result, err := FindBinary(dir, "")
		assert.NoError(t, err)
		assert.Equal(t, binPath, result)
	})

	t.Run("skips non-executable files", func(t *testing.T) {
		dir := t.TempDir()

		// Only non-executable files
		err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("readme"), 0644)
		assert.NoError(t, err)
		err = os.WriteFile(filepath.Join(dir, "LICENSE.txt"), []byte("license"), 0644)
		assert.NoError(t, err)

		_, err = FindBinary(dir, "")
		assert.Error(t, err)
	})

	t.Run("skips completions directory", func(t *testing.T) {
		dir := t.TempDir()

		compDir := filepath.Join(dir, "completions")
		assert.NoError(t, os.MkdirAll(compDir, 0755))
		assert.NoError(t, os.WriteFile(filepath.Join(compDir, "fd.bash"), []byte("completion"), 0755))

		binPath := filepath.Join(dir, "fd")
		assert.NoError(t, os.WriteFile(binPath, []byte("binary"), 0755))

		result, err := FindBinary(dir, "fd")
		assert.NoError(t, err)
		assert.Equal(t, binPath, result)
	})

	t.Run("empty directory", func(t *testing.T) {
		dir := t.TempDir()

		_, err := FindBinary(dir, "fd")
		assert.Error(t, err)
	})
}
