package zypper

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

// Client implements the package manager interface for zypper
type Client struct{}

// Name returns the package manager name
func (s *Client) Name() string {
	return "zypper"
}

// Install installs packages using zypper
func (s *Client) Install(packages []string) error {
	args := append([]string{"install", "-y"}, packages...)

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"zypper"}, args...)...)
	} else {
		cmd = execCommand("zypper", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Remove removes packages using zypper
func (s *Client) Remove(packages []string) error {
	args := append([]string{"remove", "-y"}, packages...)

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"zypper"}, args...)...)
	} else {
		cmd = execCommand("zypper", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Search searches for packages using zypper search
// Output table format with | delimiters:
// S | Name | Summary | Type
func (s *Client) Search(query string) ([]types.SearchResult, error) {
	cmd := execCommand("zypper", "search", query)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseZypperSearch(string(out)), nil
}

// Update refreshes the package repository metadata using zypper refresh
func (s *Client) Update() error {
	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", "zypper", "refresh")
	} else {
		cmd = execCommand("zypper", "refresh")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Upgrade upgrades packages using zypper update.
// If packages is empty, all packages are upgraded.
// If specific packages are provided, only those are upgraded.
func (s *Client) Upgrade(packages []string) error {
	args := []string{"update", "-y"}
	if len(packages) > 0 {
		args = append(args, packages...)
	}

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"zypper"}, args...)...)
	} else {
		cmd = execCommand("zypper", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// List returns all installed packages using rpm query
func (s *Client) List() ([]types.PackageInfo, error) {
	cmd := execCommand("rpm", "-qa", "--queryformat", "%{NAME}\\t%{VERSION}-%{RELEASE}\\n")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseZypperList(string(out)), nil
}

func parseZypperList(output string) []types.PackageInfo {
	var results []types.PackageInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Format: "name\tversion-release"
		parts := strings.SplitN(line, "\t", 2)
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

func parseZypperSearch(output string) []types.SearchResult {
	var results []types.SearchResult
	scanner := bufio.NewScanner(strings.NewReader(output))
	headerPassed := false
	for scanner.Scan() {
		line := scanner.Text()
		// Skip until we pass the header separator (--+--)
		if strings.Contains(line, "--+--") {
			headerPassed = true
			continue
		}
		if !headerPassed || strings.TrimSpace(line) == "" {
			continue
		}
		// Format: "S | name | summary | type"
		cols := strings.Split(line, "|")
		if len(cols) >= 3 {
			results = append(results, types.SearchResult{
				Name:        strings.TrimSpace(cols[1]),
				Description: strings.TrimSpace(cols[2]),
			})
		}
	}
	return results
}
