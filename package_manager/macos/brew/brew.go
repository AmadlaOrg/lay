package brew

import (
	"bufio"
	"os"
	"os/exec"
	"strings"

	"github.com/AmadlaOrg/lay/package_manager/types"
)

// For mocking
var execCommand = exec.Command

// Client implements the package manager interface for brew
type Client struct{}

// Name returns the package manager name
func (s *Client) Name() string {
	return "brew"
}

// Install installs packages using brew (no sudo — brew refuses root)
func (s *Client) Install(packages []string) error {
	args := append([]string{"install"}, packages...)
	cmd := execCommand("brew", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Remove removes packages using brew uninstall (no sudo — brew refuses root)
func (s *Client) Remove(packages []string) error {
	args := append([]string{"uninstall"}, packages...)
	cmd := execCommand("brew", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Search searches for packages using brew search
// Output format:
//
//	==> Formulae
//	name
//	...
//	==> Casks
//	name
//	...
func (s *Client) Search(query string) ([]types.SearchResult, error) {
	cmd := execCommand("brew", "search", query)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseBrewSearch(string(out)), nil
}

// Update runs brew update to refresh the package index (no sudo)
func (s *Client) Update() error {
	cmd := execCommand("brew", "update")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Upgrade upgrades packages. If packages is empty, upgrades all packages.
// If packages are specified, only upgrades those packages. (no sudo)
func (s *Client) Upgrade(packages []string) error {
	var args []string
	if len(packages) == 0 {
		args = []string{"upgrade"}
	} else {
		args = append([]string{"upgrade"}, packages...)
	}

	cmd := execCommand("brew", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// List returns all installed brew packages with versions
func (s *Client) List() ([]types.PackageInfo, error) {
	cmd := execCommand("brew", "list", "--versions")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseBrewList(string(out)), nil
}

func parseBrewList(output string) []types.PackageInfo {
	var results []types.PackageInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Format: "name version"
		fields := strings.SplitN(line, " ", 2)
		if len(fields) >= 2 {
			results = append(results, types.PackageInfo{
				Name:    fields[0],
				Version: strings.TrimSpace(fields[1]),
			})
		}
	}
	return results
}

// IsInstalled checks whether a brew package is installed using brew list
func (s *Client) IsInstalled(pkg string) (bool, error) {
	cmd := execCommand("brew", "list", pkg)
	err := cmd.Run()
	if err != nil {
		return false, nil
	}
	return true, nil
}

func parseBrewSearch(output string) []types.SearchResult {
	var results []types.SearchResult
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Skip section headers like "==> Formulae" or "==> Casks"
		if strings.HasPrefix(line, "==>") {
			continue
		}
		results = append(results, types.SearchResult{
			Name: line,
		})
	}
	return results
}
