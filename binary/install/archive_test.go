package install

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ulikunitz/xz"
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

func TestExtract_Zip(t *testing.T) {
	// Create a real zip archive with files and a subdirectory
	zipPath := filepath.Join(t.TempDir(), "test.zip")
	f, err := os.Create(zipPath)
	assert.NoError(t, err)

	zw := zip.NewWriter(f)

	// Add a file at root
	w, err := zw.Create("hello.txt")
	assert.NoError(t, err)
	_, err = w.Write([]byte("hello world"))
	assert.NoError(t, err)

	// Add a file in a subdirectory
	w, err = zw.Create("sub/nested.txt")
	assert.NoError(t, err)
	_, err = w.Write([]byte("nested content"))
	assert.NoError(t, err)

	assert.NoError(t, zw.Close())
	assert.NoError(t, f.Close())

	destDir, err := Extract(zipPath, ArchiveZip)
	assert.NoError(t, err)
	defer os.RemoveAll(destDir)

	// Verify root file
	content, err := os.ReadFile(filepath.Join(destDir, "hello.txt"))
	assert.NoError(t, err)
	assert.Equal(t, "hello world", string(content))

	// Verify nested file
	content, err = os.ReadFile(filepath.Join(destDir, "sub", "nested.txt"))
	assert.NoError(t, err)
	assert.Equal(t, "nested content", string(content))
}

func TestExtract_Zip_PathTraversal(t *testing.T) {
	zipPath := filepath.Join(t.TempDir(), "evil.zip")
	f, err := os.Create(zipPath)
	assert.NoError(t, err)

	zw := zip.NewWriter(f)

	// Add a path-traversal entry
	w, err := zw.Create("../../evil.txt")
	assert.NoError(t, err)
	_, err = w.Write([]byte("malicious"))
	assert.NoError(t, err)

	// Add a legitimate file
	w, err = zw.Create("safe.txt")
	assert.NoError(t, err)
	_, err = w.Write([]byte("safe"))
	assert.NoError(t, err)

	assert.NoError(t, zw.Close())
	assert.NoError(t, f.Close())

	destDir, err := Extract(zipPath, ArchiveZip)
	assert.NoError(t, err)
	defer os.RemoveAll(destDir)

	// The traversal entry should be skipped
	_, err = os.Stat(filepath.Join(destDir, "..", "..", "evil.txt"))
	assert.True(t, os.IsNotExist(err))

	// The safe file should exist
	content, err := os.ReadFile(filepath.Join(destDir, "safe.txt"))
	assert.NoError(t, err)
	assert.Equal(t, "safe", string(content))
}

func TestExtract_Zip_InvalidFile(t *testing.T) {
	// Write non-zip data to a file
	badPath := filepath.Join(t.TempDir(), "bad.zip")
	assert.NoError(t, os.WriteFile(badPath, []byte("this is not a zip"), 0644))

	_, err := Extract(badPath, ArchiveZip)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open zip")
}

// createTestTarXz builds a tar.xz archive containing a single executable file.
func createTestTarXz(t *testing.T, name, content string) []byte {
	t.Helper()

	var buf bytes.Buffer
	xw, err := xz.NewWriter(&buf)
	if err != nil {
		t.Fatal(err)
	}
	tw := tar.NewWriter(xw)

	hdr := &tar.Header{
		Name: name,
		Mode: 0755,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(tw, strings.NewReader(content)); err != nil {
		t.Fatal(err)
	}

	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := xw.Close(); err != nil {
		t.Fatal(err)
	}

	return buf.Bytes()
}

func TestExtract_TarXz(t *testing.T) {
	data := createTestTarXz(t, "mytool", "#!/bin/sh\necho xz")
	archivePath := filepath.Join(t.TempDir(), "test.tar.xz")
	assert.NoError(t, os.WriteFile(archivePath, data, 0644))

	destDir, err := Extract(archivePath, ArchiveTarXz)
	assert.NoError(t, err)
	defer os.RemoveAll(destDir)

	content, err := os.ReadFile(filepath.Join(destDir, "mytool"))
	assert.NoError(t, err)
	assert.Equal(t, "#!/bin/sh\necho xz", string(content))
}

func TestExtract_TarXz_InvalidFile(t *testing.T) {
	badPath := filepath.Join(t.TempDir(), "bad.tar.xz")
	assert.NoError(t, os.WriteFile(badPath, []byte("not xz data"), 0644))

	_, err := Extract(badPath, ArchiveTarXz)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create xz reader")
}

func TestExtract_TarXz_PathTraversal(t *testing.T) {
	// Build a tar.xz with a path-traversal entry + a safe entry
	var buf bytes.Buffer
	xw, err := xz.NewWriter(&buf)
	assert.NoError(t, err)
	tw := tar.NewWriter(xw)

	// Traversal entry
	assert.NoError(t, tw.WriteHeader(&tar.Header{
		Name: "../../evil.txt",
		Mode: 0644,
		Size: 4,
	}))
	_, err = tw.Write([]byte("evil"))
	assert.NoError(t, err)

	// Safe entry
	assert.NoError(t, tw.WriteHeader(&tar.Header{
		Name: "safe.txt",
		Mode: 0644,
		Size: 4,
	}))
	_, err = tw.Write([]byte("safe"))
	assert.NoError(t, err)

	assert.NoError(t, tw.Close())
	assert.NoError(t, xw.Close())

	archivePath := filepath.Join(t.TempDir(), "evil.tar.xz")
	assert.NoError(t, os.WriteFile(archivePath, buf.Bytes(), 0644))

	destDir, err := Extract(archivePath, ArchiveTarXz)
	assert.NoError(t, err)
	defer os.RemoveAll(destDir)

	// Traversal entry should be skipped
	_, err = os.Stat(filepath.Join(destDir, "..", "..", "evil.txt"))
	assert.True(t, os.IsNotExist(err))

	// Safe file should exist
	content, err := os.ReadFile(filepath.Join(destDir, "safe.txt"))
	assert.NoError(t, err)
	assert.Equal(t, "safe", string(content))
}

func TestExtract_UnsupportedType(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.7z")
	assert.NoError(t, os.WriteFile(tmpFile, []byte("data"), 0644))

	_, err := Extract(tmpFile, ArchiveType("7z"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported archive type")
}
