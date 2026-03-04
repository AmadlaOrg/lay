package scoop

import (
	"bufio"
	"os"
	"os/exec"
	"strings"

	"github.com/AmadlaOrg/lay/package_manager/types"
)

// For mocking
var execCommand = exec.Command

// Client implements the package manager interface for scoop
type Client struct{}

// Name returns the package manager name
func (s *Client) Name() string {
	return "scoop"
}

// Install installs packages using scoop
func (s *Client) Install(packages []string) error {
	args := append([]string{"install"}, packages...)
	cmd := execCommand("scoop", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Remove removes packages using scoop uninstall
func (s *Client) Remove(packages []string) error {
	args := append([]string{"uninstall"}, packages...)
	cmd := execCommand("scoop", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Search searches for packages using scoop search
// Output format: header + separator + "name version source" rows
func (s *Client) Search(query string) ([]types.SearchResult, error) {
	cmd := execCommand("scoop", "search", query)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseScoopSearch(string(out)), nil
}

// Update runs scoop update to refresh package lists
func (s *Client) Update() error {
	cmd := execCommand("scoop", "update")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Upgrade upgrades packages. If packages is empty, upgrades all packages.
// If packages are specified, only upgrades those packages.
func (s *Client) Upgrade(packages []string) error {
	var args []string
	if len(packages) == 0 {
		args = []string{"update", "*"}
	} else {
		args = append([]string{"update"}, packages...)
	}

	cmd := execCommand("scoop", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// List returns all installed packages
func (s *Client) List() ([]types.PackageInfo, error) {
	cmd := execCommand("scoop", "list")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseScoopList(string(out)), nil
}

func parseScoopList(output string) []types.PackageInfo {
	var results []types.PackageInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	headerPassed := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Skip lines until we see the "----" separator
		if !headerPassed {
			if strings.HasPrefix(line, "----") || strings.HasPrefix(line, "---") {
				headerPassed = true
			}
			continue
		}
		// Format: "name version source updated info"
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

// IsInstalled checks whether a package is installed using scoop info
func (s *Client) IsInstalled(pkg string) (bool, error) {
	cmd := execCommand("scoop", "info", pkg)
	err := cmd.Run()
	if err != nil {
		return false, nil
	}
	return true, nil
}

func parseScoopSearch(output string) []types.SearchResult {
	var results []types.SearchResult
	scanner := bufio.NewScanner(strings.NewReader(output))
	headerPassed := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Skip header line (e.g., "Name  Version  Source")
		if !headerPassed {
			if strings.HasPrefix(line, "Name") || strings.HasPrefix(line, "----") || strings.HasPrefix(line, "---") {
				if strings.HasPrefix(line, "----") || strings.HasPrefix(line, "---") {
					headerPassed = true
				}
				continue
			}
			// If first non-empty line isn't a header, treat as data
			headerPassed = true
		}
		// Skip separator lines
		if strings.HasPrefix(line, "----") || strings.HasPrefix(line, "---") {
			continue
		}
		// Format: "name version source"
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
