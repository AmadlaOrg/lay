package install

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/AmadlaOrg/lay/binary/checksum"
	"github.com/AmadlaOrg/lay/binary/forge"
	"github.com/AmadlaOrg/lay/binary/forge/codeberg"
	"github.com/AmadlaOrg/lay/binary/forge/github"
	"github.com/AmadlaOrg/lay/binary/forge/gitlab"
	"github.com/AmadlaOrg/lay/binary/jar"
)

// For mocking
var (
	osStat          = os.Stat
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

// InstallResult contains metadata about a successfully installed binary
type InstallResult struct {
	BinaryName string
	Version    string
	Path       string
	Checksum   string
	HashAlgo   string
}

// Installer defines the binary installation interface
type Installer interface {
	Install(source string, targetDir string) (*InstallResult, error)
}

// Service implements Installer
type Service struct {
	NameOverride string
}

// Install downloads, extracts, and places a binary into targetDir
func (s *Service) Install(source string, targetDir string) (*InstallResult, error) {
	// Check if source is a local file
	if isLocalFile(source) {
		return s.installLocalJar(source, targetDir)
	}

	var downloadURL string
	var hint string
	var tagName string
	var allAssets []forge.Asset

	info, err := forge.ParseSource(source)
	if err != nil {
		return nil, err
	}
	if info.IsForge {
		f, err := newForgeService(info.Forge)
		if err != nil {
			return nil, err
		}
		hint = info.Repo

		if info.Version != "" {
			allAssets, tagName, err = f.GetReleaseAssetsByTag(info.Owner, info.Repo, info.Version)
		} else {
			allAssets, tagName, err = f.GetLatestReleaseAssets(info.Owner, info.Repo)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get release assets: %w", err)
		}

		if len(allAssets) == 0 {
			return nil, fmt.Errorf("no release assets found for %s", source)
		}

		platform := DetectPlatform()
		assetNames := make([]string, len(allAssets))
		for i, a := range allAssets {
			assetNames[i] = a.Name
		}

		ranked := RankAssets(assetNames, platform)
		if len(ranked) == 0 {
			// Try JAR assets as fallback
			jarAssets := jar.FindJarAssets(assetNames)
			if len(jarAssets) > 0 {
				for _, a := range allAssets {
					if a.Name == jarAssets[0] {
						return s.installJarFromURL(a.DownloadURL, jarAssets[0], tagName, targetDir)
					}
				}
			}
			return nil, fmt.Errorf("no compatible asset found for %s/%s (os=%s, arch=%s)", platform.OS, platform.Arch, platform.OS, platform.Arch)
		}

		// Find the matching asset URL
		for _, a := range allAssets {
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
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	filename := filepath.Base(downloadURL)
	tmpFile := filepath.Join(tmpDir, filename)

	if err := downloadFile(downloadURL, tmpFile); err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}

	// Checksum verification (forge sources only)
	if len(allAssets) > 0 {
		allAssetNames := make([]string, len(allAssets))
		for i, a := range allAssets {
			allAssetNames[i] = a.Name
		}
		checksumAsset, algo, found := checksum.FindChecksumAsset(filename, allAssetNames)
		if found {
			// Find checksum asset URL
			var checksumURL string
			for _, a := range allAssets {
				if a.Name == checksumAsset {
					checksumURL = a.DownloadURL
					break
				}
			}
			if checksumURL != "" {
				checksumContent, err := downloadToString(checksumURL)
				if err == nil {
					expectedHash, err := checksum.ParseChecksumFile(checksumContent, filename)
					if err == nil {
						if verifyErr := checksum.VerifyFile(tmpFile, expectedHash, algo); verifyErr != nil {
							return nil, fmt.Errorf("checksum verification failed: %w", verifyErr)
						}
						fmt.Println("Checksum verified")
					}
				}
			}
		}
	}

	// Detect and extract
	archiveType := DetectArchiveType(filename)
	extractDir, err := Extract(tmpFile, archiveType)
	if err != nil {
		return nil, fmt.Errorf("failed to extract: %w", err)
	}
	defer os.RemoveAll(extractDir)

	// Find binary
	binaryPath, err := FindBinary(extractDir, hint)
	if err != nil {
		return nil, fmt.Errorf("failed to find binary: %w", err)
	}

	// Copy to target
	destPath := filepath.Join(targetDir, filepath.Base(binaryPath))
	if err := copyFile(binaryPath, destPath); err != nil {
		return nil, fmt.Errorf("failed to copy binary: %w", err)
	}

	if err := os.Chmod(destPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to set permissions: %w", err)
	}

	// Compute SHA256 of installed binary
	checksum, err := computeSHA256(destPath)
	if err != nil {
		// Non-fatal: proceed without checksum
		checksum = ""
	}

	fmt.Printf("Installed %s to %s\n", filepath.Base(binaryPath), destPath)
	return &InstallResult{
		BinaryName: filepath.Base(binaryPath),
		Version:    tagName,
		Path:       destPath,
		Checksum:   checksum,
		HashAlgo:   "sha256",
	}, nil
}

func computeSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func downloadToString(url string) (string, error) {
	resp, err := httpGet(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
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

// isLocalFile returns true if source points to an existing file on disk.
func isLocalFile(source string) bool {
	_, err := osStat(source)
	return err == nil
}

// installLocalJar installs a JAR file from the local filesystem.
func (s *Service) installLocalJar(source string, targetDir string) (*InstallResult, error) {
	filename := filepath.Base(source)
	if !jar.IsJarFile(filename) {
		return nil, fmt.Errorf("local file install only supports .jar files: %s", filename)
	}

	javaInfo, err := jar.DetectJava()
	if err != nil {
		return nil, fmt.Errorf("java is required to run JAR files: %w", err)
	}

	fmt.Printf("Found Java %s\n", javaInfo.Version)

	commandName := s.NameOverride
	if commandName == "" {
		commandName = jar.DeriveCommandName(filename)
	}

	jarDest, err := jar.CopyJar(source)
	if err != nil {
		return nil, fmt.Errorf("failed to copy JAR: %w", err)
	}

	scriptPath, err := jar.CreateLauncherScript(jarDest, commandName, targetDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create launcher script: %w", err)
	}

	chk, _ := computeSHA256(jarDest)

	fmt.Printf("Installed %s (launcher: %s)\n", filename, scriptPath)
	return &InstallResult{
		BinaryName: commandName,
		Version:    "",
		Path:       scriptPath,
		Checksum:   chk,
		HashAlgo:   "sha256",
	}, nil
}

// installJarFromURL downloads a JAR from a URL and installs it.
func (s *Service) installJarFromURL(downloadURL, filename, tagName, targetDir string) (*InstallResult, error) {
	javaInfo, err := jar.DetectJava()
	if err != nil {
		return nil, fmt.Errorf("java is required to run JAR files: %w", err)
	}

	fmt.Printf("Found Java %s\n", javaInfo.Version)

	// Download to temp file
	tmpDir, err := os.MkdirTemp("", "lay-jar-download-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, filename)
	if err := downloadFile(downloadURL, tmpFile); err != nil {
		return nil, fmt.Errorf("failed to download JAR: %w", err)
	}

	commandName := s.NameOverride
	if commandName == "" {
		commandName = jar.DeriveCommandName(filename)
	}

	jarDest, err := jar.CopyJar(tmpFile)
	if err != nil {
		return nil, fmt.Errorf("failed to copy JAR: %w", err)
	}

	scriptPath, err := jar.CreateLauncherScript(jarDest, commandName, targetDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create launcher script: %w", err)
	}

	chk, _ := computeSHA256(jarDest)

	fmt.Printf("Installed %s (launcher: %s)\n", filename, scriptPath)
	return &InstallResult{
		BinaryName: commandName,
		Version:    tagName,
		Path:       scriptPath,
		Checksum:   chk,
		HashAlgo:   "sha256",
	}, nil
}
