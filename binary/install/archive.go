package install

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ArchiveType represents the type of archive
type ArchiveType string

const (
	ArchiveTarGz ArchiveType = "tar.gz"
	ArchiveTarXz ArchiveType = "tar.xz"
	ArchiveZip   ArchiveType = "zip"
	ArchiveRaw   ArchiveType = "raw"
)

// DetectArchiveType determines the archive type from the filename
func DetectArchiveType(filename string) ArchiveType {
	lower := strings.ToLower(filename)

	if strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz") {
		return ArchiveTarGz
	}
	if strings.HasSuffix(lower, ".tar.xz") || strings.HasSuffix(lower, ".txz") {
		return ArchiveTarXz
	}
	if strings.HasSuffix(lower, ".zip") {
		return ArchiveZip
	}

	return ArchiveRaw
}

// Extract extracts an archive to a temporary directory and returns the path
func Extract(archivePath string, archiveType ArchiveType) (string, error) {
	destDir, err := os.MkdirTemp("", "lay-extract-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	switch archiveType {
	case ArchiveTarGz:
		if err := extractTarGz(archivePath, destDir); err != nil {
			os.RemoveAll(destDir)
			return "", err
		}
	case ArchiveZip:
		if err := extractZip(archivePath, destDir); err != nil {
			os.RemoveAll(destDir)
			return "", err
		}
	case ArchiveRaw:
		// Copy raw file as-is
		base := filepath.Base(archivePath)
		dest := filepath.Join(destDir, base)
		if err := copyFile(archivePath, dest); err != nil {
			os.RemoveAll(destDir)
			return "", err
		}
	default:
		os.RemoveAll(destDir)
		return "", fmt.Errorf("unsupported archive type: %s", archiveType)
	}

	return destDir, nil
}

// FindBinary searches the extracted directory for the most likely binary file.
// hint is used to match the expected binary name (e.g., repo name).
func FindBinary(extractedDir string, hint string) (string, error) {
	var candidates []string

	err := filepath.Walk(extractedDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		name := info.Name()

		// Skip non-binary files
		lower := strings.ToLower(name)
		for _, skip := range []string{".md", ".txt", ".1", ".fish", ".zsh", ".bash", ".ps1", ".elv"} {
			if strings.HasSuffix(lower, skip) {
				return nil
			}
		}
		// Skip common non-binary directories/files
		if strings.Contains(path, "/completions/") || strings.Contains(path, "/man/") ||
			strings.Contains(path, "/doc/") || strings.Contains(path, "/autocomplete/") {
			return nil
		}

		// Check if executable
		if info.Mode()&0111 != 0 {
			candidates = append(candidates, path)
		}

		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to walk extracted dir: %w", err)
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("no executable binary found in %s", extractedDir)
	}

	// Prefer candidate matching hint
	if hint != "" {
		for _, c := range candidates {
			if filepath.Base(c) == hint {
				return c, nil
			}
		}
	}

	// Return first candidate
	return candidates[0], nil
}

func extractTarGz(archivePath, destDir string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar entry: %w", err)
		}

		target := filepath.Join(destDir, filepath.Clean(header.Name))

		// Prevent path traversal
		if !strings.HasPrefix(target, filepath.Clean(destDir)+string(os.PathSeparator)) && target != filepath.Clean(destDir) {
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}

func extractZip(archivePath, destDir string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		target := filepath.Join(destDir, filepath.Clean(f.Name))

		// Prevent path traversal
		if !strings.HasPrefix(target, filepath.Clean(destDir)+string(os.PathSeparator)) && target != filepath.Clean(destDir) {
			continue
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
