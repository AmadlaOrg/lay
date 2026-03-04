package dnf

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

// Client implements the package manager interface for dnf
type Client struct{}

// Name returns the package manager name
func (s *Client) Name() string {
	return "dnf"
}

// Install installs packages using dnf
func (s *Client) Install(packages []string) error {
	args := append([]string{"install", "-y"}, packages...)

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"dnf"}, args...)...)
	} else {
		cmd = execCommand("dnf", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Remove removes packages using dnf
func (s *Client) Remove(packages []string) error {
	args := append([]string{"remove", "-y"}, packages...)

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"dnf"}, args...)...)
	} else {
		cmd = execCommand("dnf", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Search searches for packages using dnf search
// Output format lines like: "name.arch : description"
func (s *Client) Search(query string) ([]types.SearchResult, error) {
	cmd := execCommand("dnf", "search", "-q", query)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseDnfSearch(string(out)), nil
}

func parseDnfSearch(output string) []types.SearchResult {
	var results []types.SearchResult
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "=") {
			continue
		}
		// Format: "name.arch : description"
		parts := strings.SplitN(line, " : ", 2)
		if len(parts) == 2 {
			nameArch := strings.TrimSpace(parts[0])
			// Strip .arch suffix (e.g., "curl.x86_64" → "curl")
			name := nameArch
			if idx := strings.LastIndex(nameArch, "."); idx > 0 {
				name = nameArch[:idx]
			}
			results = append(results, types.SearchResult{
				Name:        name,
				Description: strings.TrimSpace(parts[1]),
			})
		}
	}
	return results
}

// Update refreshes the package metadata cache using dnf makecache
func (s *Client) Update() error {
	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", "dnf", "makecache")
	} else {
		cmd = execCommand("dnf", "makecache")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Upgrade upgrades packages using dnf upgrade.
// If packages is empty, all packages are upgraded.
// If specific packages are provided, only those are upgraded.
func (s *Client) Upgrade(packages []string) error {
	args := []string{"upgrade", "-y"}
	if len(packages) > 0 {
		args = append(args, packages...)
	}

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"dnf"}, args...)...)
	} else {
		cmd = execCommand("dnf", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// List returns all installed packages using dnf list installed
func (s *Client) List() ([]types.PackageInfo, error) {
	cmd := execCommand("dnf", "list", "installed", "-q")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseDnfList(string(out)), nil
}

func parseDnfList(output string) []types.PackageInfo {
	var results []types.PackageInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Format: "name.arch    version    repo"
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		nameArch := fields[0]
		// Strip .arch suffix (e.g., "curl.x86_64" → "curl")
		name := nameArch
		if idx := strings.LastIndex(nameArch, "."); idx > 0 {
			name = nameArch[:idx]
		}
		results = append(results, types.PackageInfo{
			Name:    name,
			Version: fields[1],
		})
	}
	return results
}

// IsInstalled checks if a specific package is installed using rpm -q
func (s *Client) IsInstalled(pkg string) (bool, error) {
	cmd := execCommand("rpm", "-q", pkg)
	err := cmd.Run()
	if err != nil {
		// Exit code non-zero means not installed
		return false, nil
	}
	return true, nil
}
