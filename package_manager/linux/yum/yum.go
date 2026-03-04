package yum

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

// Client implements the package manager interface for yum
type Client struct{}

// Name returns the package manager name
func (s *Client) Name() string {
	return "yum"
}

// Install installs packages using yum
func (s *Client) Install(packages []string) error {
	args := append([]string{"install", "-y"}, packages...)

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"yum"}, args...)...)
	} else {
		cmd = execCommand("yum", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Remove removes packages using yum
func (s *Client) Remove(packages []string) error {
	args := append([]string{"remove", "-y"}, packages...)

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"yum"}, args...)...)
	} else {
		cmd = execCommand("yum", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Search searches for packages using yum search
// Output format same as dnf: "name.arch : description"
func (s *Client) Search(query string) ([]types.SearchResult, error) {
	cmd := execCommand("yum", "search", "-q", query)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseYumSearch(string(out)), nil
}

// Update refreshes the yum package cache
func (s *Client) Update() error {
	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", "yum", "makecache")
	} else {
		cmd = execCommand("yum", "makecache")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Upgrade upgrades packages using yum update
// If packages is empty, all packages are upgraded; otherwise only the specified packages.
func (s *Client) Upgrade(packages []string) error {
	args := []string{"update", "-y"}
	if len(packages) > 0 {
		args = append(args, packages...)
	}

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"yum"}, args...)...)
	} else {
		cmd = execCommand("yum", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// List returns all installed packages
// Uses "yum list installed -q" which outputs lines like: "name.arch version repo"
func (s *Client) List() ([]types.PackageInfo, error) {
	cmd := execCommand("yum", "list", "installed", "-q")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseYumList(string(out)), nil
}

// IsInstalled checks whether a specific package is installed using rpm -q
func (s *Client) IsInstalled(pkg string) (bool, error) {
	cmd := execCommand("rpm", "-q", pkg)
	err := cmd.Run()
	if err != nil {
		return false, nil
	}
	return true, nil
}

func parseYumList(output string) []types.PackageInfo {
	var packages []types.PackageInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Format: "name.arch    version    repo"
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			nameArch := fields[0]
			version := fields[1]
			name := nameArch
			if idx := strings.LastIndex(nameArch, "."); idx > 0 {
				name = nameArch[:idx]
			}
			packages = append(packages, types.PackageInfo{
				Name:    name,
				Version: version,
			})
		}
	}
	return packages
}

func parseYumSearch(output string) []types.SearchResult {
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
