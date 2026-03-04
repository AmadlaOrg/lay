package apt

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

// Client implements the package manager interface for apt
type Client struct{}

// Name returns the package manager name
func (s *Client) Name() string {
	return "apt"
}

// Install installs packages using apt
func (s *Client) Install(packages []string) error {
	args := append([]string{"install", "-y"}, packages...)

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"apt"}, args...)...)
	} else {
		cmd = execCommand("apt", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Remove removes packages using apt
func (s *Client) Remove(packages []string) error {
	args := append([]string{"remove", "-y"}, packages...)

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"apt"}, args...)...)
	} else {
		cmd = execCommand("apt", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Search searches for packages using apt-cache search
// Output format: "package - description"
func (s *Client) Search(query string) ([]types.SearchResult, error) {
	cmd := execCommand("apt-cache", "search", query)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseAptSearch(string(out)), nil
}

func parseAptSearch(output string) []types.SearchResult {
	var results []types.SearchResult
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Format: "package - description"
		parts := strings.SplitN(line, " - ", 2)
		if len(parts) == 2 {
			results = append(results, types.SearchResult{
				Name:        strings.TrimSpace(parts[0]),
				Description: strings.TrimSpace(parts[1]),
			})
		}
	}
	return results
}

// Update runs apt update to refresh package lists
func (s *Client) Update() error {
	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", "apt", "update")
	} else {
		cmd = execCommand("apt", "update")
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
		args = []string{"upgrade", "-y"}
	} else {
		args = append([]string{"install", "--only-upgrade", "-y"}, packages...)
	}

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"apt"}, args...)...)
	} else {
		cmd = execCommand("apt", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// List returns all installed packages
func (s *Client) List() ([]types.PackageInfo, error) {
	cmd := execCommand("apt", "list", "--installed")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseAptList(string(out)), nil
}

func parseAptList(output string) []types.PackageInfo {
	var results []types.PackageInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "Listing") {
			continue
		}
		// Format: "name/repo version arch [status]"
		slashParts := strings.SplitN(line, "/", 2)
		if len(slashParts) != 2 {
			continue
		}
		name := slashParts[0]
		fields := strings.Fields(slashParts[1])
		if len(fields) < 2 {
			continue
		}
		version := fields[1]
		results = append(results, types.PackageInfo{
			Name:    name,
			Version: version,
		})
	}
	return results
}

// IsInstalled checks whether a package is installed using dpkg -s
func (s *Client) IsInstalled(pkg string) (bool, error) {
	cmd := execCommand("dpkg", "-s", pkg)
	err := cmd.Run()
	if err != nil {
		return false, nil
	}
	return true, nil
}
