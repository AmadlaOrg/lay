package snap

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

// Client implements the package manager interface for snap
type Client struct{}

// Name returns the package manager name
func (s *Client) Name() string {
	return "snap"
}

// Install installs packages using snap install
func (s *Client) Install(packages []string) error {
	args := append([]string{"install"}, packages...)

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"snap"}, args...)...)
	} else {
		cmd = execCommand("snap", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Remove removes packages using snap remove
func (s *Client) Remove(packages []string) error {
	args := append([]string{"remove"}, packages...)

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"snap"}, args...)...)
	} else {
		cmd = execCommand("snap", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Search searches for packages using snap find
// Output table format:
// Name  Version  Publisher  Notes  Summary
func (s *Client) Search(query string) ([]types.SearchResult, error) {
	cmd := execCommand("snap", "find", query)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseSnapSearch(string(out)), nil
}

// Update is a no-op for snap since snap auto-updates
func (s *Client) Update() error {
	return nil
}

// Upgrade refreshes snap packages. If packages is empty, refreshes all packages.
// If packages are specified, only refreshes those packages.
func (s *Client) Upgrade(packages []string) error {
	var args []string
	if len(packages) == 0 {
		args = []string{"refresh"}
	} else {
		args = append([]string{"refresh"}, packages...)
	}

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"snap"}, args...)...)
	} else {
		cmd = execCommand("snap", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// List returns all installed snap packages
func (s *Client) List() ([]types.PackageInfo, error) {
	cmd := execCommand("snap", "list")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseSnapList(string(out)), nil
}

func parseSnapList(output string) []types.PackageInfo {
	var results []types.PackageInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	firstLine := true
	for scanner.Scan() {
		line := scanner.Text()
		// Skip header line
		if firstLine {
			firstLine = false
			continue
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		// Fields are whitespace-delimited: Name Version Rev Tracking Publisher Notes
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		results = append(results, types.PackageInfo{
			Name:    fields[0],
			Version: fields[1],
		})
	}
	return results
}

// IsInstalled checks whether a snap package is installed
func (s *Client) IsInstalled(pkg string) (bool, error) {
	cmd := execCommand("snap", "list", pkg)
	err := cmd.Run()
	if err != nil {
		return false, nil
	}
	return true, nil
}

func parseSnapSearch(output string) []types.SearchResult {
	var results []types.SearchResult
	scanner := bufio.NewScanner(strings.NewReader(output))
	firstLine := true
	for scanner.Scan() {
		line := scanner.Text()
		// Skip header line
		if firstLine {
			firstLine = false
			continue
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		// Fields are whitespace-delimited: Name Version Publisher Notes Summary...
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		name := fields[0]
		version := fields[1]
		// Summary is everything from field 4 onward
		desc := strings.Join(fields[4:], " ")

		results = append(results, types.SearchResult{
			Name:        name,
			Version:     version,
			Description: desc,
		})
	}
	return results
}
