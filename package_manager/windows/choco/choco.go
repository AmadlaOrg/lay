package choco

import (
	"bufio"
	"os"
	"os/exec"
	"strings"

	"github.com/AmadlaOrg/lay/package_manager/types"
)

// For mocking
var execCommand = exec.Command

// Client implements the package manager interface for choco
type Client struct{}

// Name returns the package manager name
func (s *Client) Name() string {
	return "choco"
}

// Install installs packages using choco
func (s *Client) Install(packages []string) error {
	args := append([]string{"install", "-y"}, packages...)
	cmd := execCommand("choco", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Remove removes packages using choco uninstall
func (s *Client) Remove(packages []string) error {
	args := append([]string{"uninstall", "-y"}, packages...)
	cmd := execCommand("choco", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Search searches for packages using choco search
// Output format: "name version\n...\nN packages found."
func (s *Client) Search(query string) ([]types.SearchResult, error) {
	cmd := execCommand("choco", "search", query)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseChocoSearch(string(out)), nil
}

func parseChocoSearch(output string) []types.SearchResult {
	var results []types.SearchResult
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Skip summary line like "3 packages found."
		if strings.HasSuffix(line, "packages found.") {
			continue
		}
		// Format: "name version"
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			results = append(results, types.SearchResult{
				Name:    parts[0],
				Version: parts[1],
			})
		}
	}
	return results
}

// Update is a no-op for choco; it auto-checks for updates
func (s *Client) Update() error {
	return nil
}

// Upgrade upgrades packages. If packages is empty, upgrades all packages.
// If packages are specified, only upgrades those packages.
func (s *Client) Upgrade(packages []string) error {
	var args []string
	if len(packages) == 0 {
		args = []string{"upgrade", "all", "-y"}
	} else {
		args = append([]string{"upgrade", "-y"}, packages...)
	}

	cmd := execCommand("choco", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// List returns all installed packages
func (s *Client) List() ([]types.PackageInfo, error) {
	cmd := execCommand("choco", "list")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseChocoList(string(out)), nil
}

func parseChocoList(output string) []types.PackageInfo {
	var results []types.PackageInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Skip summary line like "2 packages installed."
		if strings.HasSuffix(line, "packages installed.") {
			continue
		}
		// Skip Chocolatey version line
		if strings.HasPrefix(line, "Chocolatey v") {
			continue
		}
		// Format: "name version"
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			results = append(results, types.PackageInfo{
				Name:    parts[0],
				Version: parts[1],
			})
		}
	}
	return results
}

// IsInstalled checks whether a package is installed using choco list --exact
func (s *Client) IsInstalled(pkg string) (bool, error) {
	cmd := execCommand("choco", "list", "--exact", pkg)
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return len(parseChocoList(string(out))) > 0, nil
}
