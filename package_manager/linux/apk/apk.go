package apk

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

// Client implements the package manager interface for apk (Alpine)
type Client struct{}

// Name returns the package manager name
func (s *Client) Name() string {
	return "apk"
}

// Install installs packages using apk add
func (s *Client) Install(packages []string) error {
	args := append([]string{"add"}, packages...)

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"apk"}, args...)...)
	} else {
		cmd = execCommand("apk", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Remove removes packages using apk del
func (s *Client) Remove(packages []string) error {
	args := append([]string{"del"}, packages...)

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"apk"}, args...)...)
	} else {
		cmd = execCommand("apk", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Search searches for packages using apk search -v
// Output format: "package-version - description"
func (s *Client) Search(query string) ([]types.SearchResult, error) {
	cmd := execCommand("apk", "search", "-v", query)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseApkSearch(string(out)), nil
}

func parseApkSearch(output string) []types.SearchResult {
	var results []types.SearchResult
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Format: "package-version - description"
		parts := strings.SplitN(line, " - ", 2)
		if len(parts) == 2 {
			nameVersion := strings.TrimSpace(parts[0])
			// Split name from version: last hyphen before a digit
			name, version := splitApkNameVersion(nameVersion)
			results = append(results, types.SearchResult{
				Name:        name,
				Version:     version,
				Description: strings.TrimSpace(parts[1]),
			})
		}
	}
	return results
}

// Update runs apk update to refresh package lists
func (s *Client) Update() error {
	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", "apk", "update")
	} else {
		cmd = execCommand("apk", "update")
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
		args = []string{"upgrade"}
	} else {
		args = append([]string{"upgrade"}, packages...)
	}

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"apk"}, args...)...)
	} else {
		cmd = execCommand("apk", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// List returns all installed packages
func (s *Client) List() ([]types.PackageInfo, error) {
	cmd := execCommand("apk", "list", "--installed")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseApkList(string(out)), nil
}

func parseApkList(output string) []types.PackageInfo {
	var results []types.PackageInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Format: "name-version arch {repo} (license)"
		fields := strings.Fields(line)
		if len(fields) < 1 {
			continue
		}
		name, version := splitApkNameVersion(fields[0])
		if name != "" {
			results = append(results, types.PackageInfo{
				Name:    name,
				Version: version,
			})
		}
	}
	return results
}

// IsInstalled checks whether a package is installed using apk info -e
func (s *Client) IsInstalled(pkg string) (bool, error) {
	cmd := execCommand("apk", "info", "-e", pkg)
	err := cmd.Run()
	if err != nil {
		return false, nil
	}
	return true, nil
}

// splitApkNameVersion splits "curl-8.5.0-r0" into ("curl", "8.5.0-r0")
func splitApkNameVersion(nameVersion string) (string, string) {
	// Find the last '-' that precedes a digit
	for i := len(nameVersion) - 1; i >= 0; i-- {
		if nameVersion[i] == '-' && i+1 < len(nameVersion) && nameVersion[i+1] >= '0' && nameVersion[i+1] <= '9' {
			return nameVersion[:i], nameVersion[i+1:]
		}
	}
	return nameVersion, ""
}
