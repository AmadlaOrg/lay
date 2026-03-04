package checksum

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		filename string
		algo     Algorithm
		found    bool
	}{
		{"tool.sha256", SHA256, true},
		{"tool.sha256sum", SHA256, true},
		{"tool.SHA256", SHA256, true},
		{"tool.sha512", SHA512, true},
		{"tool.sha512sum", SHA512, true},
		{"tool.tar.gz", "", false},
		{"tool.sig", "", false},
		{"tool", "", false},
	}

	for _, tt := range tests {
		algo, found := Detect(tt.filename)
		assert.Equal(t, tt.found, found, "filename: %s", tt.filename)
		if found {
			assert.Equal(t, tt.algo, algo, "filename: %s", tt.filename)
		}
	}
}

func TestFindChecksumAsset_PerBinary(t *testing.T) {
	assets := []string{
		"tool-linux-amd64.tar.gz",
		"tool-linux-amd64.tar.gz.sha256",
		"tool-darwin-arm64.tar.gz",
	}

	name, algo, found := FindChecksumAsset("tool-linux-amd64.tar.gz", assets)
	assert.True(t, found)
	assert.Equal(t, "tool-linux-amd64.tar.gz.sha256", name)
	assert.Equal(t, SHA256, algo)
}

func TestFindChecksumAsset_ProjectWide(t *testing.T) {
	assets := []string{
		"tool-linux-amd64.tar.gz",
		"tool-darwin-arm64.tar.gz",
		"SHA256SUMS",
	}

	name, algo, found := FindChecksumAsset("tool-linux-amd64.tar.gz", assets)
	assert.True(t, found)
	assert.Equal(t, "SHA256SUMS", name)
	assert.Equal(t, SHA256, algo)
}

func TestFindChecksumAsset_ChecksumsTxt(t *testing.T) {
	assets := []string{
		"tool-linux-amd64.tar.gz",
		"checksums.txt",
	}

	name, algo, found := FindChecksumAsset("tool-linux-amd64.tar.gz", assets)
	assert.True(t, found)
	assert.Equal(t, "checksums.txt", name)
	assert.Equal(t, SHA256, algo)
}

func TestFindChecksumAsset_None(t *testing.T) {
	assets := []string{
		"tool-linux-amd64.tar.gz",
		"tool-darwin-arm64.tar.gz",
	}

	_, _, found := FindChecksumAsset("tool-linux-amd64.tar.gz", assets)
	assert.False(t, found)
}

func TestParseChecksumFile_Standard(t *testing.T) {
	content := "abc123def456  tool-linux-amd64.tar.gz\n789xyz  tool-darwin-arm64.tar.gz\n"
	hash, err := ParseChecksumFile(content, "tool-linux-amd64.tar.gz")
	assert.NoError(t, err)
	assert.Equal(t, "abc123def456", hash)
}

func TestParseChecksumFile_BinaryMode(t *testing.T) {
	content := "abc123def456 *tool-linux-amd64.tar.gz\n"
	hash, err := ParseChecksumFile(content, "tool-linux-amd64.tar.gz")
	assert.NoError(t, err)
	assert.Equal(t, "abc123def456", hash)
}

func TestParseChecksumFile_NotFound(t *testing.T) {
	content := "abc123def456  other-file.tar.gz\n"
	_, err := ParseChecksumFile(content, "tool-linux-amd64.tar.gz")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checksum not found")
}

func TestVerifyFile_Match(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testfile")
	data := []byte("hello world")
	err := os.WriteFile(path, data, 0644)
	assert.NoError(t, err)

	h := sha256.Sum256(data)
	expected := hex.EncodeToString(h[:])

	err = VerifyFile(path, expected, SHA256)
	assert.NoError(t, err)
}

func TestVerifyFile_Mismatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testfile")
	err := os.WriteFile(path, []byte("hello world"), 0644)
	assert.NoError(t, err)

	err = VerifyFile(path, "0000000000000000000000000000000000000000000000000000000000000000", SHA256)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checksum mismatch")
}

func TestVerifyFile_FileNotFound(t *testing.T) {
	err := VerifyFile("/nonexistent/file", "abc", SHA256)
	assert.Error(t, err)
}

func TestVerifyFile_UnsupportedAlgo(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testfile")
	err := os.WriteFile(path, []byte("hello"), 0644)
	assert.NoError(t, err)

	err = VerifyFile(path, "abc", Algorithm("md5"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported algorithm")
}
