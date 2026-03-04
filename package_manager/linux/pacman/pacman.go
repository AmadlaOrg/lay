package pacman

import (
	"os"
	"os/exec"
	"strings"

	"github.com/AmadlaOrg/lay/package_manager/types"
)

// For mocking
var (
	execCommand = exec.Command
	osGetuid    = os.Getuid
)

// Client implements the package manager interface for pacman
type Client struct{}

// Name returns the package manager name
func (s *Client) Name() string {
	return "pacman"
}

// Install installs packages using pacman
func (s *Client) Install(packages []string) error {
	args := append([]string{"-S", "--noconfirm"}, packages...)

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"pacman"}, args...)...)
	} else {
		cmd = execCommand("pacman", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Remove removes packages using pacman
func (s *Client) Remove(packages []string) error {
	args := append([]string{"-R", "--noconfirm"}, packages...)

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"pacman"}, args...)...)
	} else {
		cmd = execCommand("pacman", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Search searches for packages using pacman -Ss
// Output format:
//
//	repo/package version
//	    description
func (s *Client) Search(query string) ([]types.SearchResult, error) {
	cmd := execCommand("pacman", "-Ss", query)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parsePacmanSearch(string(out)), nil
}

// Update refreshes the package database using pacman -Sy
func (s *Client) Update() error {
	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", "pacman", "-Sy")
	} else {
		cmd = execCommand("pacman", "-Sy")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Upgrade upgrades packages. If packages is empty, upgrades all with pacman -Syu.
// If packages are specified, installs/upgrades those specific packages with pacman -S.
func (s *Client) Upgrade(packages []string) error {
	var args []string
	if len(packages) == 0 {
		args = []string{"-Syu", "--noconfirm"}
	} else {
		args = append([]string{"-S", "--noconfirm"}, packages...)
	}

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"pacman"}, args...)...)
	} else {
		cmd = execCommand("pacman", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// List returns all installed packages using pacman -Q
func (s *Client) List() ([]types.PackageInfo, error) {
	cmd := execCommand("pacman", "-Q")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parsePacmanList(string(out)), nil
}

// IsInstalled checks whether a specific package is installed using pacman -Q
func (s *Client) IsInstalled(pkg string) (bool, error) {
	cmd := execCommand("pacman", "-Q", pkg)
	err := cmd.Run()
	if err != nil {
		return false, nil
	}
	return true, nil
}

func parsePacmanList(output string) []types.PackageInfo {
	var results []types.PackageInfo
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		results = append(results, types.PackageInfo{
			Name:    parts[0],
			Version: parts[1],
		})
	}
	return results
}

func parsePacmanSearch(output string) []types.SearchResult {
	var results []types.SearchResult
	lines := strings.Split(output, "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if line == "" || strings.HasPrefix(line, " ") {
			continue
		}
		// Header line: "repo/name version [installed]"
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		repoName := parts[0]
		version := parts[1]

		name := repoName
		if idx := strings.Index(repoName, "/"); idx >= 0 {
			name = repoName[idx+1:]
		}

		var desc string
		if i+1 < len(lines) {
			desc = strings.TrimSpace(lines[i+1])
			i++ // skip description line
		}

		results = append(results, types.SearchResult{
			Name:        name,
			Version:     version,
			Description: desc,
		})
	}
	return results
}
