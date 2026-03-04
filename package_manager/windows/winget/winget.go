package winget

import (
	"bufio"
	"os"
	"os/exec"
	"strings"

	"github.com/AmadlaOrg/lay/package_manager/types"
)

// For mocking
var execCommand = exec.Command

// Client implements the package manager interface for winget
type Client struct{}

// Name returns the package manager name
func (s *Client) Name() string {
	return "winget"
}

// Install installs packages using winget (one at a time)
func (s *Client) Install(packages []string) error {
	for _, pkg := range packages {
		cmd := execCommand("winget", "install", pkg)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

// Remove removes packages using winget uninstall (one at a time)
func (s *Client) Remove(packages []string) error {
	for _, pkg := range packages {
		cmd := execCommand("winget", "uninstall", pkg)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

// Search searches for packages using winget search
// Output format: header + dash-separator + "Name Id Version Source" rows
func (s *Client) Search(query string) ([]types.SearchResult, error) {
	cmd := execCommand("winget", "search", query)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseWingetSearch(string(out)), nil
}

// Update runs winget source update to refresh package sources
func (s *Client) Update() error {
	cmd := execCommand("winget", "source", "update")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Upgrade upgrades packages. If packages is empty, upgrades all packages.
// If packages are specified, only upgrades those packages.
func (s *Client) Upgrade(packages []string) error {
	if len(packages) == 0 {
		cmd := execCommand("winget", "upgrade", "--all")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		return cmd.Run()
	}
	for _, pkg := range packages {
		cmd := execCommand("winget", "upgrade", pkg)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

// List returns all installed winget packages
func (s *Client) List() ([]types.PackageInfo, error) {
	cmd := execCommand("winget", "list")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseWingetList(string(out)), nil
}

func parseWingetList(output string) []types.PackageInfo {
	var results []types.PackageInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	separatorSeen := false
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// Look for the dash separator line
		if !separatorSeen {
			if strings.HasPrefix(trimmed, "---") {
				separatorSeen = true
			}
			continue
		}
		// Format: "Name  Id  Version" (space-separated columns)
		parts := strings.Fields(trimmed)
		if len(parts) >= 3 {
			results = append(results, types.PackageInfo{
				Name:    parts[0],
				Version: parts[2],
			})
		}
	}
	return results
}

// IsInstalled checks whether a winget package is installed using winget list --exact
func (s *Client) IsInstalled(pkg string) (bool, error) {
	cmd := execCommand("winget", "list", "--exact", pkg)
	err := cmd.Run()
	if err != nil {
		return false, nil
	}
	return true, nil
}

func parseWingetSearch(output string) []types.SearchResult {
	var results []types.SearchResult
	scanner := bufio.NewScanner(strings.NewReader(output))
	separatorSeen := false
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// Look for the dash separator line
		if !separatorSeen {
			if strings.HasPrefix(trimmed, "---") {
				separatorSeen = true
			}
			continue
		}
		// Format: "Name  Id  Version  Source" (space-separated columns)
		parts := strings.Fields(trimmed)
		if len(parts) >= 3 {
			results = append(results, types.SearchResult{
				Name:    parts[0],
				Version: parts[2],
			})
		}
	}
	return results
}
