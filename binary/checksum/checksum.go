package checksum

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"
)

// Algorithm represents a checksum algorithm
type Algorithm string

const (
	SHA256 Algorithm = "sha256"
	SHA512 Algorithm = "sha512"
)

// Detect determines the checksum algorithm from a file extension
func Detect(filename string) (Algorithm, bool) {
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".sha256") || strings.HasSuffix(lower, ".sha256sum"):
		return SHA256, true
	case strings.HasSuffix(lower, ".sha512") || strings.HasSuffix(lower, ".sha512sum"):
		return SHA512, true
	default:
		return "", false
	}
}

// FindChecksumAsset looks for a checksum file among release assets.
// It searches for patterns like: <binary>.sha256, <binary>.sha256sum, SHA256SUMS, checksums.txt
func FindChecksumAsset(binaryName string, allAssets []string) (assetName string, algo Algorithm, found bool) {
	baseName := strings.TrimSuffix(binaryName, ".tar.gz")
	baseName = strings.TrimSuffix(baseName, ".tgz")
	baseName = strings.TrimSuffix(baseName, ".tar.xz")
	baseName = strings.TrimSuffix(baseName, ".zip")

	// Per-binary checksum files (highest priority)
	perBinaryPatterns := []struct {
		suffix string
		algo   Algorithm
	}{
		{".sha256", SHA256},
		{".sha256sum", SHA256},
		{".sha512", SHA512},
		{".sha512sum", SHA512},
	}

	for _, asset := range allAssets {
		for _, p := range perBinaryPatterns {
			if asset == binaryName+p.suffix || asset == baseName+p.suffix {
				return asset, p.algo, true
			}
		}
	}

	// Project-wide checksum files
	projectWide := []struct {
		name string
		algo Algorithm
	}{
		{"SHA256SUMS", SHA256},
		{"SHA256SUMS.txt", SHA256},
		{"sha256sums.txt", SHA256},
		{"checksums.txt", SHA256},
		{"SHA512SUMS", SHA512},
		{"SHA512SUMS.txt", SHA512},
	}

	for _, asset := range allAssets {
		for _, p := range projectWide {
			if asset == p.name {
				return asset, p.algo, true
			}
		}
	}

	return "", "", false
}

// ParseChecksumFile extracts the hash for targetFilename from a checksum file.
// Supports format: "<hash>  <filename>" or "<hash> <filename>"
func ParseChecksumFile(content, targetFilename string) (string, error) {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Format: "hash  filename" or "hash *filename" (binary mode)
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		filename := parts[len(parts)-1]
		// Strip binary mode indicator
		filename = strings.TrimPrefix(filename, "*")

		if filename == targetFilename {
			return parts[0], nil
		}
	}
	return "", fmt.Errorf("checksum not found for %s", targetFilename)
}

// VerifyFile computes the hash of filePath and compares it to the expected value
func VerifyFile(filePath, expected string, algo Algorithm) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	var h hash.Hash
	switch algo {
	case SHA256:
		h = sha256.New()
	case SHA512:
		h = sha512.New()
	default:
		return fmt.Errorf("unsupported algorithm: %s", algo)
	}

	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	computed := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(computed, expected) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, computed)
	}

	return nil
}
