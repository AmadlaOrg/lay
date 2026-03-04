package flatpak

import (
	"bufio"
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

// Client implements the package manager interface for flatpak
type Client struct{}

// Name returns the package manager name
func (s *Client) Name() string {
	return "flatpak"
}

// Install installs packages using flatpak install
func (s *Client) Install(packages []string) error {
	args := append([]string{"install", "-y"}, packages...)

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"flatpak"}, args...)...)
	} else {
		cmd = execCommand("flatpak", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Remove removes packages using flatpak uninstall
func (s *Client) Remove(packages []string) error {
	args := append([]string{"uninstall", "-y"}, packages...)
	cmd := execCommand("flatpak", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Search searches for packages using flatpak search
// Output columns: Name\tDescription\tApplication ID\tVersion\tBranch\tRemotes
func (s *Client) Search(query string) ([]types.SearchResult, error) {
	cmd := execCommand("flatpak", "search", query)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseFlatpakSearch(string(out)), nil
}

// Update runs flatpak update --appstream to refresh appstream data
func (s *Client) Update() error {
	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", "flatpak", "update", "--appstream", "-y")
	} else {
		cmd = execCommand("flatpak", "update", "--appstream", "-y")
	}

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
		args = []string{"update", "-y"}
	} else {
		args = append([]string{"update", "-y"}, packages...)
	}

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"flatpak"}, args...)...)
	} else {
		cmd = execCommand("flatpak", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// List returns all installed flatpak packages
func (s *Client) List() ([]types.PackageInfo, error) {
	cmd := execCommand("flatpak", "list", "--columns=name,version")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseFlatpakList(string(out)), nil
}

func parseFlatpakList(output string) []types.PackageInfo {
	var results []types.PackageInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Format: "Name\tVersion"
		cols := strings.Split(line, "\t")
		if len(cols) >= 2 {
			results = append(results, types.PackageInfo{
				Name:    strings.TrimSpace(cols[0]),
				Version: strings.TrimSpace(cols[1]),
			})
		}
	}
	return results
}

// IsInstalled checks whether a flatpak package is installed using flatpak info
func (s *Client) IsInstalled(pkg string) (bool, error) {
	cmd := execCommand("flatpak", "info", pkg)
	err := cmd.Run()
	if err != nil {
		return false, nil
	}
	return true, nil
}

func parseFlatpakSearch(output string) []types.SearchResult {
	var results []types.SearchResult
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		// Flatpak uses tab-separated columns:
		// Name\tDescription\tApplication ID\tVersion\tBranch\tRemotes
		cols := strings.Split(line, "\t")
		if len(cols) >= 4 {
			results = append(results, types.SearchResult{
				Name:        strings.TrimSpace(cols[0]),
				Version:     strings.TrimSpace(cols[3]),
				Description: strings.TrimSpace(cols[1]),
			})
		}
	}
	return results
}
