package install

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/AmadlaOrg/lay/binary/forge"
	"github.com/AmadlaOrg/lay/binary/forge/codeberg"
	"github.com/AmadlaOrg/lay/binary/forge/github"
	"github.com/AmadlaOrg/lay/binary/forge/gitlab"
)

// For mocking
var (
	httpGet         = http.Get
	newForgeService = func(name string) (forge.Forge, error) {
		switch name {
		case "github":
			return github.NewService(), nil
		case "gitlab":
			return gitlab.NewService(), nil
		case "codeberg":
			return codeberg.NewService(), nil
		default:
			return nil, fmt.Errorf("unsupported forge: %s", name)
		}
	}
)

// Installer defines the binary installation interface
type Installer interface {
	Install(source string, targetDir string) error
}

// Service implements Installer
type Service struct{}

// Install downloads, extracts, and places a binary into targetDir
func (s *Service) Install(source string, targetDir string) error {
	var downloadURL string
	var hint string

	info, err := forge.ParseSource(source)
	if err != nil {
		return err
	}
	if info.IsForge {
		f, err := newForgeService(info.Forge)
		if err != nil {
			return err
		}
		hint = info.Repo

		var assets []forge.Asset
		if info.Version != "" {
			assets, _, err = f.GetReleaseAssetsByTag(info.Owner, info.Repo, info.Version)
		} else {
			assets, _, err = f.GetLatestReleaseAssets(info.Owner, info.Repo)
		}
		if err != nil {
			return fmt.Errorf("failed to get release assets: %w", err)
		}

		if len(assets) == 0 {
			return fmt.Errorf("no release assets found for %s", source)
		}

		platform := DetectPlatform()
		assetNames := make([]string, len(assets))
		for i, a := range assets {
			assetNames[i] = a.Name
		}

		ranked := RankAssets(assetNames, platform)
		if len(ranked) == 0 {
			return fmt.Errorf("no compatible asset found for %s/%s (os=%s, arch=%s)", platform.OS, platform.Arch, platform.OS, platform.Arch)
		}

		// Find the matching asset URL
		for _, a := range assets {
			if a.Name == ranked[0] {
				downloadURL = a.DownloadURL
				break
			}
		}
	} else {
		downloadURL = source
		// Extract hint from URL path
		hint = strings.TrimSuffix(filepath.Base(downloadURL), filepath.Ext(downloadURL))
	}

	// Download to temp file
	tmpDir, err := os.MkdirTemp("", "lay-download-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	filename := filepath.Base(downloadURL)
	tmpFile := filepath.Join(tmpDir, filename)

	if err := downloadFile(downloadURL, tmpFile); err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}

	// Detect and extract
	archiveType := DetectArchiveType(filename)
	extractDir, err := Extract(tmpFile, archiveType)
	if err != nil {
		return fmt.Errorf("failed to extract: %w", err)
	}
	defer os.RemoveAll(extractDir)

	// Find binary
	binaryPath, err := FindBinary(extractDir, hint)
	if err != nil {
		return fmt.Errorf("failed to find binary: %w", err)
	}

	// Copy to target
	destPath := filepath.Join(targetDir, filepath.Base(binaryPath))
	if err := copyFile(binaryPath, destPath); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	if err := os.Chmod(destPath, 0755); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	fmt.Printf("Installed %s to %s\n", filepath.Base(binaryPath), destPath)
	return nil
}

func downloadFile(url string, dest string) error {
	resp, err := httpGet(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
